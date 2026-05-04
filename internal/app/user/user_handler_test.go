// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package user_test

import (
	"context"
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	u "github.com/cloudoperators/heureka/internal/app/user"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/cloudoperators/heureka/internal/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	mock "github.com/stretchr/testify/mock"
)

func TestUserHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "User Service Test Suite")
}

var (
	er             event.EventRegistry
	authz          openfga.Authorization
	handlerContext common.HandlerContext
	cfg            *util.Config
	enableLogs     bool
)

var _ = BeforeSuite(func() {
	authEnabled := false
	cfg = common.GetTestConfig(authEnabled)
	enableLogs := false
	db := mocks.NewMockDatabase(GinkgoT())
	authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
	er = event.NewEventRegistry(db, authz)
	handlerContext = common.HandlerContext{
		DB:       db,
		EventReg: er,
		Cache:    nil,
		Authz:    authz,
	}
	handlerContext.Authz.RemoveAllRelations()
})

func getUserFilter() *entity.UserFilter {
	userName := "SomeNotExistingUserName"

	return &entity.UserFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		Name:           []*string{&userName},
		SupportGroupId: nil,
	}
}

var _ = Describe("When listing Users", Label("app", "ListUsers"), func() {
	var (
		db                     *mocks.MockDatabase
		userHandler            u.UserHandler
		ctx                    context.Context
		filter                 *entity.UserFilter
		options                *entity.ListOptions
		handlerContext         common.HandlerContext
		systemUserUniqueUserId string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		ctx = common.NewAdminContext()
		options = entity.NewListOptions()
		filter = getUserFilter()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
		systemUserUniqueUserId = "S0000000"
	})

	When("the list option does include the totalCount", func() {
		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On("GetUsers", mock.Anything, filter).Return([]entity.UserResult{}, nil)
			db.On("CountUsers", mock.Anything, filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUsers(ctx, filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.TotalCount).Should(BeEquivalentTo(int64(1337)), "return correct Totalcount")
		})
	})

	When("the list option does include the PageInfo", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
		})

		DescribeTable(
			"pagination information is correct",
			func(pageSize int, dbElements int, resElements int, hasNextPage bool) {
				authFilter := &entity.UserFilter{UniqueUserID: []*string{&systemUserUniqueUserId}}

				filter.First = &pageSize
				users := []entity.UserResult{}

				for _, user := range test.NNewFakeUserEntities(resElements) {
					cursor, _ := mariadb.EncodeCursor(mariadb.WithUser([]entity.Order{}, user))
					users = append(
						users,
						entity.UserResult{
							WithCursor: entity.WithCursor{Value: cursor},
							User:       new(user),
						},
					)
				}

				cursors := lo.Map(users, func(m entity.UserResult, _ int) string {
					cursor, _ := mariadb.EncodeCursor(mariadb.WithUser([]entity.Order{}, *m.User))
					return cursor
				})

				for i := 0; len(cursors) < dbElements; i++ {
					user := test.NewFakeUserEntity()
					c, _ := mariadb.EncodeCursor(mariadb.WithUser([]entity.Order{}, user))
					cursors = append(cursors, c)
				}

				db.On("GetUsers", mock.Anything, filter).Return(users, nil)
				db.On("GetAllUserCursors", mock.Anything, filter, []entity.Order{}).Return(cursors, nil)
				db.On("GetAllUserIds", mock.Anything, authFilter).Return([]int64{}, nil)
				// db.On("GetAllUserIds", filter).Return(lo.Map(users, func(m entity.UserResult, _
				// int) int64 { return m.User.Id }), nil)

				userHandler = u.NewUserHandler(handlerContext)
				res, err := userHandler.ListUsers(ctx, filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(
					*res.PageInfo.HasNextPage,
				).To(BeEquivalentTo(hasNextPage), "correct hasNextPage indicator")
				Expect(len(res.Elements)).To(BeEquivalentTo(resElements))
				Expect(
					len(res.PageInfo.Pages),
				).To(BeEquivalentTo(int(math.Ceil(float64(dbElements)/float64(pageSize)))), "correct  number of pages")
			},
			Entry("When  pageSize is 1 and the database was returning 2 elements", 1, 2, 1, true),
			Entry(
				"When  pageSize is 10 and the database was returning 9 elements",
				10,
				9,
				9,
				false,
			),
			Entry(
				"When  pageSize is 10 and the database was returning 11 elements",
				10,
				11,
				10,
				true,
			),
		)
	})

	Context("when authz is enabled", func() {
		BeforeEach(func() {
			authEnabled := true
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		AfterEach(func() {
			authEnabled := false
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		Context("and the user has no access to any users", func() {
			BeforeEach(func() {
				sgIds := int64(-1)
				filter.SupportGroupId = []*int64{&sgIds}
				db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
				db.On("GetUsers", mock.Anything, filter).Return([]entity.UserResult{}, nil)
			})

			It("should return no users", func() {
				userHandler = u.NewUserHandler(handlerContext)
				res, err := userHandler.ListUsers(ctx, filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return 0 results")
			})
		})

		Context("and the filter includes a support group Id that has users related to it", func() {
			var user entity.User

			BeforeEach(func() {
				sgId := int64(111)
				systemUserId := int64(1)
				filter.SupportGroupId = []*int64{&sgId}
				user = test.NewFakeUserEntity()
				db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
				db.On("GetUsers", mock.Anything, filter).Return([]entity.UserResult{{User: &user}}, nil)

				relations := []openfga.RelationInput{
					{ // create support group
						UserType:   openfga.TypeRole,
						UserId:     openfga.UserIdFromInt(systemUserId),
						Relation:   openfga.RelRole,
						ObjectType: openfga.TypeSupportGroup,
						ObjectId:   openfga.ObjectIdFromInt(sgId),
					},
					{ // link user to support group
						UserType:   openfga.TypeUser,
						UserId:     openfga.UserIdFromInt(user.Id),
						Relation:   openfga.RelMember,
						ObjectType: openfga.TypeSupportGroup,
						ObjectId:   openfga.ObjectIdFromInt(sgId),
					},
				}

				err := handlerContext.Authz.AddRelationBulk(relations)
				Expect(err).To(BeNil(), "no error should be thrown when adding relations")
			})

			It("should return the expected users in the result", func() {
				userHandler = u.NewUserHandler(handlerContext)
				res, err := userHandler.ListUsers(ctx, filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(1), "return 1 result")
				Expect(
					res.Elements[0].UniqueUserID,
				).To(BeEquivalentTo(user.UniqueUserID))
				// check that the returned user is the expected one
			})
		})
	})
})

var _ = Describe("When creating User", Label("app", "CreateUser"), func() {
	var (
		db             *mocks.MockDatabase
		userHandler    u.UserHandler
		user           entity.User
		filter         *entity.UserFilter
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		user = test.NewFakeUserEntity()
		first := 10
		var after string
		filter = &entity.UserFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	It("creates user", func() {
		filter.UniqueUserID = []*string{&user.UniqueUserID}
		db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
		db.On("CreateUser", &user).Return(&user, nil)
		db.On("GetUsers", mock.Anything, filter).Return([]entity.UserResult{}, nil)
		userHandler = u.NewUserHandler(handlerContext)
		newUser, err := userHandler.CreateUser(common.NewAdminContext(), &user)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newUser.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newUser.Name).To(BeEquivalentTo(user.Name))
			Expect(newUser.UniqueUserID).To(BeEquivalentTo(user.UniqueUserID))
		})
	})
})

var _ = Describe("When updating User", Label("app", "UpdateUser"), func() {
	var (
		db             *mocks.MockDatabase
		userHandler    u.UserHandler
		user           entity.User
		filter         *entity.UserFilter
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		user = test.NewFakeUserEntity()
		first := 10
		var after string
		filter = &entity.UserFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	It("updates user", func() {
		db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
		db.On("UpdateUser", &user).Return(nil)
		userHandler = u.NewUserHandler(handlerContext)
		user.Name = "Sauron"
		filter.Id = []*int64{&user.Id}
		db.On("GetUsers", mock.Anything, filter).Return([]entity.UserResult{
			{
				User: &user,
			},
		}, nil)
		updatedUser, err := userHandler.UpdateUser(common.NewAdminContext(), &user)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedUser.Name).To(BeEquivalentTo(user.Name))
			Expect(updatedUser.UniqueUserID).To(BeEquivalentTo(user.UniqueUserID))
		})
	})
})

var _ = Describe("When deleting User", Label("app", "DeleteUser"), func() {
	var (
		db             *mocks.MockDatabase
		userHandler    u.UserHandler
		id             int64
		filter         *entity.UserFilter
		handlerContext common.HandlerContext
		ctx            context.Context
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after string
		filter = &entity.UserFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
		handlerContext.Authz.RemoveAllRelations()
		ctx = common.NewAdminContext()
	})

	It("deletes user", func() {
		db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
		db.On("DeleteUser", id, mock.Anything).Return(nil)
		userHandler = u.NewUserHandler(handlerContext)
		db.On("GetUsers", mock.Anything, filter).Return([]entity.UserResult{}, nil)
		err := userHandler.DeleteUser(common.NewAdminContext(), id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		users, err := userHandler.ListUsers(ctx, filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(users.Elements).To(BeEmpty(), "no error should be thrown")
	})

	Context("when authz is enabled", func() {
		BeforeEach(func() {
			authEnabled := true
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		AfterEach(func() {
			// Reset authz to disabled after finishing tests
			authEnabled := false
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		Context("when handling a DeleteUserEvent", func() {
			Context("when new user is deleted", func() {
				It("should delete tuples related to that user in openfga", func() {
					// Test OnUserDeleteAuthz against all possible relations
					authz := openfga.NewAuthorizationHandler(cfg, enableLogs)
					userFake := test.NewFakeUserEntity()
					deleteEvent := &u.DeleteUserEvent{
						UserID: userFake.Id,
					}
					userId := openfga.UserIdFromInt(deleteEvent.UserID)

					relations := []openfga.RelationInput{
						{ // user - role
							UserType:   openfga.TypeUser,
							UserId:     userId,
							ObjectId:   openfga.IDRole,
							ObjectType: openfga.TypeRole,
							Relation:   openfga.RelAdmin,
						},
						{ // user - service
							UserType:   openfga.TypeUser,
							UserId:     userId,
							ObjectId:   openfga.IDService,
							ObjectType: openfga.TypeService,
							Relation:   openfga.RelMember,
						},
						{ // user - component_instance
							UserType:   openfga.TypeUser,
							UserId:     userId,
							ObjectId:   openfga.IDComponentInstance,
							ObjectType: openfga.TypeComponentInstance,
							Relation:   openfga.RelCanView,
						},
						{ // user - support_group
							UserType:   openfga.TypeUser,
							UserId:     userId,
							ObjectId:   openfga.IDSupportGroup,
							ObjectType: openfga.TypeSupportGroup,
							Relation:   openfga.RelMember,
						},
						{ // user - issue_match
							UserType:   openfga.TypeUser,
							UserId:     userId,
							ObjectId:   openfga.IDIssueMatch,
							ObjectType: openfga.TypeIssueMatch,
							Relation:   openfga.RelCanView,
						},
						{ // user - component_version
							UserType:   openfga.TypeUser,
							UserId:     userId,
							ObjectId:   openfga.IDComponentVersion,
							ObjectType: openfga.TypeComponentVersion,
							Relation:   openfga.RelCanView,
						},
						{ // user - component
							UserType:   openfga.TypeUser,
							UserId:     userId,
							ObjectId:   openfga.IDComponent,
							ObjectType: openfga.TypeComponent,
							Relation:   openfga.RelCanView,
						},
					}

					handlerContext.Authz.AddRelationBulk(relations)

					// get the number of relations before deletion
					relCountBefore := 0
					for _, r := range relations {
						relations, err := handlerContext.Authz.ListRelations(r)
						Expect(err).To(BeNil(), "no error should be thrown")
						relCountBefore += len(relations)
					}
					Expect(
						relCountBefore,
					).To(Equal(len(relations)), "all relations should exist before deletion")

					// check that relations were created
					for _, r := range relations {
						ok, err := handlerContext.Authz.CheckPermission(r)
						Expect(err).To(BeNil(), "no error should be thrown")
						Expect(ok).To(BeTrue(), "permission should be granted")
					}

					var event event.Event = deleteEvent
					// Simulate event
					u.OnUserDeleteAuthz(db, event, authz)

					// get the number of relations after deletion
					relCountAfter := 0
					for _, r := range relations {
						relations, err := handlerContext.Authz.ListRelations(r)
						Expect(err).To(BeNil(), "no error should be thrown")
						relCountAfter += len(relations)
					}
					Expect(
						relCountAfter < relCountBefore,
					).To(BeTrue(), "less relations after deletion")
					Expect(
						relCountAfter,
					).To(BeEquivalentTo(0), "no relations should exist after deletion")

					// verify that relations were deleted
					for _, r := range relations {
						ok, err := handlerContext.Authz.CheckPermission(r)
						Expect(err).To(BeNil(), "no error should be thrown")
						Expect(ok).To(BeFalse(), "permission should NOT be granted")
					}
				})
			})
		})
	})
})

var _ = Describe("When listing User", Label("app", "ListUserNames"), func() {
	var (
		db             *mocks.MockDatabase
		userHandler    u.UserHandler
		filter         *entity.UserFilter
		options        *entity.ListOptions
		name           string
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getUserFilter()
		name = "Stephen Haag"
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	When("no filters are used", func() {
		BeforeEach(func() {
			db.On("GetUserNames", mock.Anything, filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUserNames(context.Background(), filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific UserNames filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.UserFilter{
				Name: []*string{&name},
			}

			db.On("GetUserNames", mock.Anything, filter).Return([]string{name}, nil)
		})
		It("returns filtered users according to the service type", func() {
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUserNames(context.Background(), filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(name), "should only consist of name")
		})
	})
})

var _ = Describe("When listing UniqueUserID", Label("app", "ListUniqueUserIDs"), func() {
	var (
		db             *mocks.MockDatabase
		userHandler    u.UserHandler
		filter         *entity.UserFilter
		options        *entity.ListOptions
		uuid           string
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getUserFilter()
		uuid = "I978974"
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	When("no filters are used", func() {
		BeforeEach(func() {
			db.On("GetUniqueUserIDs", mock.Anything, filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUniqueUserIDs(context.Background(), filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific UniqueUserID filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.UserFilter{
				UniqueUserID: []*string{&uuid},
			}

			db.On("GetUniqueUserIDs", mock.Anything, filter).Return([]string{uuid}, nil)
		})
		It("returns filtered users according to the service type", func() {
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUniqueUserIDs(context.Background(), filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(uuid), "should only consist of UniqueUserID")
		})
	})
})
