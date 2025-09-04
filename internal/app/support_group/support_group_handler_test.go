// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group_test

import (
	"context"
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/event"
	sg "github.com/cloudoperators/heureka/internal/app/support_group"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	mock "github.com/stretchr/testify/mock"
)

func TestSupportGroupHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Support Group Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

func getSupportGroupFilter() *entity.SupportGroupFilter {
	return &entity.SupportGroupFilter{
		PaginatedX: entity.PaginatedX{
			First: nil,
			After: nil,
		},
	}
}

var _ = Describe("When listing SupportGroups", Label("app", "ListSupportGroups"), func() {
	var (
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		filter              *entity.SupportGroupFilter
		options             *entity.ListOptions
		order               []entity.Order
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getSupportGroupFilter()
		order = []entity.Order{}
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{}, nil)
			db.On("CountSupportGroups", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			supportGroupHandler = sg.NewSupportGroupHandler(db, er)
			res, err := supportGroupHandler.ListSupportGroups(context.TODO(), filter, options)
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
			var supportGroups []entity.SupportGroupResult
			for _, sg := range test.NNewFakeSupportGroupEntities(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithSupportGroup(order, sg))
				supportGroups = append(supportGroups, entity.SupportGroupResult{
					WithCursor:   entity.WithCursor{Value: cursor},
					SupportGroup: &sg,
				})
			}

			var cursors = lo.Map(supportGroups, func(s entity.SupportGroupResult, _ int) string {
				return s.Value
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				supportGroup := test.NewFakeSupportGroupEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithSupportGroup([]entity.Order{}, supportGroup))
				cursors = append(cursors, c)
			}
			db.On("GetSupportGroups", filter, order).Return(supportGroups, nil)
			db.On("GetAllSupportGroupCursors", filter, order).Return(cursors, nil)
			supportGroupHandler = sg.NewSupportGroupHandler(db, er)
			res, err := supportGroupHandler.ListSupportGroups(context.TODO(), filter, options)
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

var _ = Describe("When creating SupportGroup", Label("app", "CreateSupportGroup"), func() {
	var (
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		supportGroup        entity.SupportGroup
		filter              *entity.SupportGroupFilter
		order               []entity.Order
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		supportGroup = test.NewFakeSupportGroupEntity()
		order = []entity.Order{}
		first := 10
		after := ""
		filter = &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
	})

	It("creates supportGroup", func() {
		filter.CCRN = []*string{&supportGroup.CCRN}
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateSupportGroup", &supportGroup).Return(&supportGroup, nil)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{}, nil)
		supportGroupHandler = sg.NewSupportGroupHandler(db, er)
		newSupportGroup, err := supportGroupHandler.CreateSupportGroup(&supportGroup)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newSupportGroup.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newSupportGroup.CCRN).To(BeEquivalentTo(supportGroup.CCRN))
		})
	})
})

var _ = Describe("When updating SupportGroup", Label("app", "UpdateSupportGroup"), func() {
	var (
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		supportGroup        entity.SupportGroupResult
		filter              *entity.SupportGroupFilter
		order               []entity.Order
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		supportGroup = test.NewFakeSupportGroupResult()
		first := 10
		after := ""
		order = []entity.Order{}
		filter = &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates supportGroup", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateSupportGroup", supportGroup.SupportGroup).Return(nil)
		supportGroupHandler = sg.NewSupportGroupHandler(db, er)
		supportGroup.CCRN = "Team Alone"
		filter.Id = []*int64{&supportGroup.Id}
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{supportGroup}, nil)
		updatedSupportGroup, err := supportGroupHandler.UpdateSupportGroup(supportGroup.SupportGroup)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedSupportGroup.CCRN).To(BeEquivalentTo(supportGroup.CCRN))
		})
	})
})

var _ = Describe("When deleting SupportGroup", Label("app", "DeleteSupportGroup"), func() {
	var (
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		id                  int64
		filter              *entity.SupportGroupFilter
		order               []entity.Order
		listOptions         *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		listOptions = entity.NewListOptions()
		id = 1
		first := 10
		after := ""
		order = []entity.Order{}
		filter = &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes supportGroup", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteSupportGroup", id, mock.Anything).Return(nil)
		supportGroupHandler = sg.NewSupportGroupHandler(db, er)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{}, nil)
		err := supportGroupHandler.DeleteSupportGroup(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		supportGroups, err := supportGroupHandler.ListSupportGroups(context.TODO(), filter, listOptions)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroups.Elements).To(BeEmpty(), "no error should be thrown")
	})
})

var _ = Describe("When modifying relationship of Service and SupportGroup", Label("app", "ServiceSupportGroupRelationship"), func() {
	var (
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		service             entity.Service
		supportGroup        entity.SupportGroupResult
		filter              *entity.SupportGroupFilter
		order               []entity.Order
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		service = test.NewFakeServiceEntity()
		supportGroup = test.NewFakeSupportGroupResult()
		order = []entity.Order{}
		first := 10
		after := ""
		filter = &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			Id: []*int64{&supportGroup.Id},
		}
	})

	It("adds service to supportGroup", func() {
		db.On("AddServiceToSupportGroup", supportGroup.Id, service.Id).Return(nil)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{supportGroup}, nil)
		supportGroupHandler = sg.NewSupportGroupHandler(db, er)
		supportGroup, err := supportGroupHandler.AddServiceToSupportGroup(supportGroup.Id, service.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroup).NotTo(BeNil(), "supportGroup should be returned")
	})

	It("removes service from supportGroup", func() {
		db.On("RemoveServiceFromSupportGroup", supportGroup.Id, service.Id).Return(nil)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{supportGroup}, nil)
		supportGroupHandler = sg.NewSupportGroupHandler(db, er)
		supportGroup, err := supportGroupHandler.RemoveServiceFromSupportGroup(supportGroup.Id, service.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroup).NotTo(BeNil(), "supportGroup should be returned")
	})
})

var _ = Describe("When modifying relationship of User and SupportGroup", Label("app", "UserSupportGroupRelationship"), func() {
	var (
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		user                entity.User
		supportGroup        entity.SupportGroupResult
		filter              *entity.SupportGroupFilter
		order               []entity.Order
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		user = test.NewFakeUserEntity()
		supportGroup = test.NewFakeSupportGroupResult()
		first := 10
		after := ""
		order = []entity.Order{}
		filter = &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			Id: []*int64{&supportGroup.Id},
		}
	})

	It("adds user to supportGroup", func() {
		db.On("AddUserToSupportGroup", supportGroup.Id, user.Id).Return(nil)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{supportGroup}, nil)
		supportGroupHandler = sg.NewSupportGroupHandler(db, er)
		supportGroup, err := supportGroupHandler.AddUserToSupportGroup(supportGroup.Id, user.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroup).NotTo(BeNil(), "supportGroup should be returned")
	})

	It("removes user from supportGroup", func() {
		db.On("RemoveUserFromSupportGroup", supportGroup.Id, user.Id).Return(nil)
		db.On("GetSupportGroups", filter, order).Return([]entity.SupportGroupResult{supportGroup}, nil)
		supportGroupHandler = sg.NewSupportGroupHandler(db, er)
		supportGroup, err := supportGroupHandler.RemoveUserFromSupportGroup(supportGroup.Id, user.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroup).NotTo(BeNil(), "supportGroup should be returned")
	})
})

var _ = Describe("When listing supportGroupCcrns", Label("app", "ListSupportGroupCcrns"), func() {
	var (
		db                  *mocks.MockDatabase
		supportGroupHandler sg.SupportGroupHandler
		filter              *entity.SupportGroupFilter
		options             *entity.ListOptions
		ccrn                string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getSupportGroupFilter()
		ccrn = "src"
	})

	When("no filters are used", func() {

		BeforeEach(func() {
			db.On("GetSupportGroupCcrns", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			supportGroupHandler = sg.NewSupportGroupHandler(db, er)
			res, err := supportGroupHandler.ListSupportGroupCcrns(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific supportGroupCcrns filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.SupportGroupFilter{
				CCRN: []*string{&ccrn},
			}

			db.On("GetSupportGroupCcrns", filter).Return([]string{ccrn}, nil)
		})
		It("returns filtered userGroups according to the service type", func() {
			supportGroupHandler = sg.NewSupportGroupHandler(db, er)
			res, err := supportGroupHandler.ListSupportGroupCcrns(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(ccrn), "should only consist of supportGroup")
		})
	})
})
