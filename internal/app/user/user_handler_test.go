// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package user_test

import (
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
	er    event.EventRegistry
	authz openfga.Authorization
)

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
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
			db.On("GetUsers", filter).Return([]entity.UserResult{}, nil)
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
			users := []entity.UserResult{}

			for _, user := range test.NNewFakeUserEntities(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithUser([]entity.Order{}, user))
				users = append(users, entity.UserResult{WithCursor: entity.WithCursor{Value: cursor}, User: lo.ToPtr(user)})
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

			db.On("GetUsers", filter).Return(users, nil)
			db.On("GetAllUserCursors", filter, []entity.Order{}).Return(cursors, nil)
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
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateUser", &user).Return(&user, nil)
		db.On("GetUsers", filter).Return([]entity.UserResult{}, nil)
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
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateUser", &user).Return(nil)
		userHandler = u.NewUserHandler(handlerContext)
		user.Name = "Sauron"
		filter.Id = []*int64{&user.Id}
		db.On("GetUsers", filter).Return([]entity.UserResult{
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
	})

	It("deletes user", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteUser", id, mock.Anything).Return(nil)
		userHandler = u.NewUserHandler(handlerContext)
		db.On("GetUsers", filter).Return([]entity.UserResult{}, nil)
		err := userHandler.DeleteUser(common.NewAdminContext(), id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		users, err := userHandler.ListUsers(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(users.Elements).To(BeEmpty(), "no error should be thrown")
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
