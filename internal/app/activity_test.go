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

func activityFilter() *entity.ActivityFilter {
	return &entity.ActivityFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		ServiceId: nil,
	}
}

func activityListOptions() *entity.ListOptions {
	return &entity.ListOptions{
		ShowTotalCount:      false,
		ShowPageInfo:        false,
		IncludeAggregations: false,
	}
}

var _ = Describe("When listing Activities", Label("app", "ListActivities"), func() {
	var (
		db      *mocks.MockDatabase
		heureka app.Heureka
		filter  *entity.ActivityFilter
		options *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = activityListOptions()
		filter = activityFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetActivities", filter).Return([]entity.Activity{}, nil)
			db.On("CountActivities", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			heureka = app.NewHeurekaApp(db)
			res, err := heureka.ListActivities(filter, options)
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
			activities := test.NNewFakeActivities(resElements)

			var ids = lo.Map(activities, func(a entity.Activity, _ int) int64 { return a.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetActivities", filter).Return(activities, nil)
			db.On("GetAllActivityIds", filter).Return(ids, nil)
			heureka = app.NewHeurekaApp(db)
			res, err := heureka.ListActivities(filter, options)
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

var _ = Describe("When creating Activity", Label("app", "CreateActivity"), func() {
	var (
		db       *mocks.MockDatabase
		heureka  app.Heureka
		activity entity.Activity
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		activity = test.NewFakeActivityEntity()
	})

	It("creates activity", func() {
		db.On("CreateActivity", &activity).Return(&activity, nil)
		heureka = app.NewHeurekaApp(db)
		newActivity, err := heureka.CreateActivity(&activity)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newActivity.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
		})
	})
})

var _ = Describe("When updating Activity", Label("app", "UpdateService"), func() {
	var (
		db       *mocks.MockDatabase
		heureka  app.Heureka
		activity entity.Activity
		filter   *entity.ActivityFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		activity = test.NewFakeActivityEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.ActivityFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates activity", func() {
		db.On("UpdateActivity", &activity).Return(nil)
		heureka = app.NewHeurekaApp(db)
		if activity.Status.String() == entity.ActivityStatusValuesOpen.String() {
			activity.Status = entity.ActivityStatusValuesInProgress
		} else {
			activity.Status = entity.ActivityStatusValuesOpen
		}
		filter.Id = []*int64{&activity.Id}
		db.On("GetActivities", filter).Return([]entity.Activity{activity}, nil)
		updatedActivity, err := heureka.UpdateActivity(&activity)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedActivity.Status.String()).To(BeEquivalentTo(activity.Status.String()))
		})
	})
})

var _ = Describe("When deleting Activity", Label("app", "DeleteActivity"), func() {
	var (
		db      *mocks.MockDatabase
		heureka app.Heureka
		id      int64
		filter  *entity.ActivityFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.ActivityFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes activity", func() {
		db.On("DeleteActivity", id).Return(nil)
		heureka = app.NewHeurekaApp(db)
		db.On("GetActivities", filter).Return([]entity.Activity{}, nil)
		err := heureka.DeleteActivity(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		activities, err := heureka.ListActivities(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(activities.Elements).To(BeEmpty(), "no error should be thrown")
	})
})

var _ = Describe("When modifying Service and Activity", Label("app", "ServiceActivity"), func() {
	var (
		db       *mocks.MockDatabase
		heureka  app.Heureka
		service  entity.Service
		activity entity.Activity
		filter   *entity.ActivityFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		service = test.NewFakeServiceEntity()
		activity = test.NewFakeActivityEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.ActivityFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id: []*int64{&activity.Id},
		}
	})

	It("adds service to activity", func() {
		db.On("AddServiceToActivity", activity.Id, service.Id).Return(nil)
		db.On("GetActivities", filter).Return([]entity.Activity{activity}, nil)
		heureka = app.NewHeurekaApp(db)
		activity, err := heureka.AddServiceToActivity(activity.Id, service.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(activity).NotTo(BeNil(), "activity should be returned")
	})

	It("removes service from activity", func() {
		db.On("RemoveServiceFromActivity", activity.Id, service.Id).Return(nil)
		db.On("GetActivities", filter).Return([]entity.Activity{activity}, nil)
		heureka = app.NewHeurekaApp(db)
		activity, err := heureka.RemoveServiceFromActivity(activity.Id, service.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(activity).NotTo(BeNil(), "activity should be returned")
	})
})
