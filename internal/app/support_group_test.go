// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app_test

import (
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.wdf.sap.corp/cc/heureka/internal/app"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/internal/entity/test"
	"github.wdf.sap.corp/cc/heureka/internal/mocks"
)

func getSupportGroupFilter() *entity.SupportGroupFilter {
	return &entity.SupportGroupFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
	}
}

var _ = Describe("When listing SupportGroups", Label("app", "ListSupportGroups"), func() {
	var (
		db      *mocks.MockDatabase
		heureka app.Heureka
		filter  *entity.SupportGroupFilter
		options *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = getListOptions()
		filter = getSupportGroupFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetSupportGroups", filter).Return([]entity.SupportGroup{}, nil)
			db.On("CountSupportGroups", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			heureka = app.NewHeurekaApp(db)
			res, err := heureka.ListSupportGroups(filter, options)
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
			supportGroups := test.NNewFakeSupportGroupEntities(resElements)

			var ids = lo.Map(supportGroups, func(s entity.SupportGroup, _ int) int64 { return s.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetSupportGroups", filter).Return(supportGroups, nil)
			db.On("GetAllSupportGroupIds", filter).Return(ids, nil)
			heureka = app.NewHeurekaApp(db)
			res, err := heureka.ListSupportGroups(filter, options)
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
		db           *mocks.MockDatabase
		heureka      app.Heureka
		supportGroup entity.SupportGroup
		filter       *entity.SupportGroupFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		supportGroup = test.NewFakeSupportGroupEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.SupportGroupFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("creates supportGroup", func() {
		filter.Name = []*string{&supportGroup.Name}
		db.On("CreateSupportGroup", &supportGroup).Return(&supportGroup, nil)
		db.On("GetSupportGroups", filter).Return([]entity.SupportGroup{}, nil)
		heureka = app.NewHeurekaApp(db)
		newSupportGroup, err := heureka.CreateSupportGroup(&supportGroup)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newSupportGroup.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newSupportGroup.Name).To(BeEquivalentTo(supportGroup.Name))
		})
	})
})

var _ = Describe("When updating SupportGroup", Label("app", "UpdateSupportGroup"), func() {
	var (
		db           *mocks.MockDatabase
		heureka      app.Heureka
		supportGroup entity.SupportGroup
		filter       *entity.SupportGroupFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		supportGroup = test.NewFakeSupportGroupEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.SupportGroupFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates supportGroup", func() {
		db.On("UpdateSupportGroup", &supportGroup).Return(nil)
		heureka = app.NewHeurekaApp(db)
		supportGroup.Name = "Team Alone"
		filter.Id = []*int64{&supportGroup.Id}
		db.On("GetSupportGroups", filter).Return([]entity.SupportGroup{supportGroup}, nil)
		updatedSupportGroup, err := heureka.UpdateSupportGroup(&supportGroup)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedSupportGroup.Name).To(BeEquivalentTo(supportGroup.Name))
		})
	})
})

var _ = Describe("When deleting SupportGroup", Label("app", "DeleteSupportGroup"), func() {
	var (
		db      *mocks.MockDatabase
		heureka app.Heureka
		id      int64
		filter  *entity.SupportGroupFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.SupportGroupFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes supportGroup", func() {
		db.On("DeleteSupportGroup", id).Return(nil)
		heureka = app.NewHeurekaApp(db)
		db.On("GetSupportGroups", filter).Return([]entity.SupportGroup{}, nil)
		err := heureka.DeleteSupportGroup(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		supportGroups, err := heureka.ListSupportGroups(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroups.Elements).To(BeEmpty(), "no error should be thrown")
	})
})

var _ = Describe("When modifying Service and SupportGroup", Label("app", "ServiceSupportGroup"), func() {
	var (
		db           *mocks.MockDatabase
		heureka      app.Heureka
		service      entity.Service
		supportGroup entity.SupportGroup
		filter       *entity.SupportGroupFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		service = test.NewFakeServiceEntity()
		supportGroup = test.NewFakeSupportGroupEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.SupportGroupFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id: []*int64{&supportGroup.Id},
		}
	})

	It("adds service to supportGroup", func() {
		db.On("AddServiceToSupportGroup", supportGroup.Id, service.Id).Return(nil)
		db.On("GetSupportGroups", filter).Return([]entity.SupportGroup{supportGroup}, nil)
		heureka = app.NewHeurekaApp(db)
		supportGroup, err := heureka.AddServiceToSupportGroup(supportGroup.Id, service.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroup).NotTo(BeNil(), "supportGroup should be returned")
	})

	It("removes service from supportGroup", func() {
		db.On("RemoveServiceFromSupportGroup", supportGroup.Id, service.Id).Return(nil)
		db.On("GetSupportGroups", filter).Return([]entity.SupportGroup{supportGroup}, nil)
		heureka = app.NewHeurekaApp(db)
		supportGroup, err := heureka.RemoveServiceFromSupportGroup(supportGroup.Id, service.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(supportGroup).NotTo(BeNil(), "supportGroup should be returned")
	})
})
