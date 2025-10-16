// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package user_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	u "github.com/cloudoperators/heureka/internal/app/user"
	"github.com/cloudoperators/heureka/internal/cache"
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

var er event.EventRegistry
var authz openfga.Authorization
var handlerContext common.HandlerContext
var cfg *util.Config
var enableLogs bool

var _ = BeforeSuite(func() {
	cfg = &util.Config{
		AuthzModelFilePath:    "./internal/openfga/model/model.fga",
		AuthzOpenFgaApiUrl:    "http://localhost:8080",
		AuthzOpenFgaStoreName: "heureka-store",
		CurrentUser:           "testuser",
		AuthTokenSecret:       "testkey",
		AuthzOpenFgaApiToken:  "testkey",
	}
	enableLogs := false
	db := mocks.NewMockDatabase(GinkgoT())
	authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
	er = event.NewEventRegistry(db, authz)
	handlerContext = common.HandlerContext{
		DB:       db,
		EventReg: er,
		Cache:    cache.NewNoCache(),
		Authz:    authz,
	}
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
		db             *mocks.MockDatabase
		userHandler    u.UserHandler
		filter         *entity.UserFilter
		options        *entity.ListOptions
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getUserFilter()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetUsers", filter).Return([]entity.User{}, nil)
			db.On("CountUsers", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUsers(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.TotalCount).Should(BeEquivalentTo(int64(1337)), "return correct Totalcount")
		})
	})

	When("the list option does include the PageInfo", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
		})
		DescribeTable("pagination information is correct", func(pageSize int, dbElements int, resElements int, hasNextPage bool) {
			filter.First = &pageSize
			users := test.NNewFakeUserEntities(resElements)

			var ids = lo.Map(users, func(u entity.User, _ int) int64 { return u.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetUsers", filter).Return(users, nil)
			db.On("GetAllUserIds", filter).Return(ids, nil)
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUsers(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.PageInfo.HasNextPage).To(BeEquivalentTo(hasNextPage), "correct hasNextPage indicator")
			Expect(len(res.Elements)).To(BeEquivalentTo(resElements))
			Expect(len(res.PageInfo.Pages)).To(BeEquivalentTo(int(math.Ceil(float64(dbElements)/float64(pageSize)))), "correct  number of pages")
		},
			Entry("When  pageSize is 1 and the database was returning 2 elements", 1, 2, 1, true),
			Entry("When  pageSize is 10 and the database was returning 9 elements", 10, 9, 9, false),
			Entry("When  pageSize is 10 and the database was returning 11 elements", 10, 11, 10, true),
		)
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
		var after int64
		after = 0
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
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateUser", &user).Return(&user, nil)
		db.On("GetUsers", filter).Return([]entity.User{}, nil)
		userHandler = u.NewUserHandler(handlerContext)
		newUser, err := userHandler.CreateUser(&user)
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
		var after int64
		after = 0
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
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateUser", &user).Return(nil)
		userHandler = u.NewUserHandler(handlerContext)
		user.Name = "Sauron"
		filter.Id = []*int64{&user.Id}
		db.On("GetUsers", filter).Return([]entity.User{user}, nil)
		updatedUser, err := userHandler.UpdateUser(&user)
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
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
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

	It("deletes user", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteUser", id, mock.Anything).Return(nil)
		userHandler = u.NewUserHandler(handlerContext)
		db.On("GetUsers", filter).Return([]entity.User{}, nil)
		err := userHandler.DeleteUser(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		users, err := userHandler.ListUsers(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(users.Elements).To(BeEmpty(), "no error should be thrown")
	})

	Context("when handling a DeleteUserEvent", func() {
		BeforeEach(func() {
			db.On("GetDefaultIssuePriority").Return(int64(100))
			db.On("GetDefaultRepositoryName").Return("nvd")
		})

		Context("when new user is deleted", func() {
			It("should delete tuples related to that user in openfga", func() {
				// Test OnUserDeleteAuthz against all possible relations
				authz := openfga.NewAuthorizationHandler(cfg, enableLogs)
				userFake := test.NewFakeUserEntity()
				deleteEvent := &u.DeleteUserEvent{
					UserID: userFake.Id,
				}
				userId := openfga.UserId(strconv.FormatInt(deleteEvent.UserID, 10))

				relations := []openfga.RelationInput{
					{ // user - role
						UserType:   "user",
						UserId:     userId,
						ObjectId:   "roleID",
						ObjectType: "role",
						Relation:   "admin",
					},
					{ // user - service
						UserType:   "user",
						UserId:     userId,
						ObjectId:   "serviceID",
						ObjectType: "service",
						Relation:   "member",
					},
					{ // user - component_instance
						UserType:   "user",
						UserId:     userId,
						ObjectId:   "componentInstanceID",
						ObjectType: "component_instance",
						Relation:   "can_view",
					},
					{ // user - support_group
						UserType:   "user",
						UserId:     userId,
						ObjectId:   "supportGroupID",
						ObjectType: "support_group",
						Relation:   "member",
					},
					{ // user - issue_match
						UserType:   "user",
						UserId:     userId,
						ObjectId:   "issueMatchID",
						ObjectType: "issue_match",
						Relation:   "can_view",
					},
					{ // user - component_version
						UserType:   "user",
						UserId:     userId,
						ObjectId:   "componentVersionID",
						ObjectType: "component_version",
						Relation:   "can_view",
					},
					{ // user - component
						UserType:   "user",
						UserId:     userId,
						ObjectId:   "componentID",
						ObjectType: "component",
						Relation:   "can_view",
					},
				}

				for _, rel := range relations {
					authz.AddRelation(rel)
				}

				var event event.Event = deleteEvent
				// Simulate event
				u.OnUserDeleteAuthz(db, event, authz)

				remaining, err := authz.ListRelations(relations)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(remaining).To(BeEmpty(), "no relations should remain after deletion")
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
			db.On("GetUserNames", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUserNames(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific UserNames filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.UserFilter{
				Name: []*string{&name},
			}

			db.On("GetUserNames", filter).Return([]string{name}, nil)
		})
		It("returns filtered users according to the service type", func() {
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUserNames(filter, options)
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
			db.On("GetUniqueUserIDs", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUniqueUserIDs(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific UniqueUserID filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.UserFilter{
				UniqueUserID: []*string{&uuid},
			}

			db.On("GetUniqueUserIDs", filter).Return([]string{uuid}, nil)
		})
		It("returns filtered users according to the service type", func() {
			userHandler = u.NewUserHandler(handlerContext)
			res, err := userHandler.ListUniqueUserIDs(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(uuid), "should only consist of UniqueUserID")
		})
	})
})
