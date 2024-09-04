// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package activity_test

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/activity"
	a "github.wdf.sap.corp/cc/heureka/internal/app/activity"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"math"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/internal/entity/test"
	"github.wdf.sap.corp/cc/heureka/internal/mocks"
)

func TestActivityHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Activity Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

func activityFilter() *entity.ActivityFilter {
	return &entity.ActivityFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		ServiceId: nil,
	}
}

var _ = Describe("When listing Activities", Label("app", "ListActivities"), func() {
	var (
		db              *mocks.MockDatabase
		activityHandler activity.ActivityHandler
		filter          *entity.ActivityFilter
		options         *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = activityFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetActivities", filter).Return([]entity.Activity{}, nil)
			db.On("CountActivities", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			activityHandler = a.NewActivityHandler(db, er)
			res, err := activityHandler.ListActivities(filter, options)
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
			activityHandler = a.NewActivityHandler(db, er)
			res, err := activityHandler.ListActivities(filter, options)
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
		db              *mocks.MockDatabase
		activityHandler a.ActivityHandler
		activity        entity.Activity
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		activity = test.NewFakeActivityEntity()
	})

	It("creates activity", func() {
		db.On("CreateActivity", &activity).Return(&activity, nil)
		activityHandler = a.NewActivityHandler(db, er)
		newActivity, err := activityHandler.CreateActivity(&activity)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newActivity.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
		})
	})
})

var _ = Describe("When updating Activity", Label("app", "UpdateService"), func() {
	var (
		db              *mocks.MockDatabase
		activityHandler a.ActivityHandler
		activity        entity.Activity
		filter          *entity.ActivityFilter
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
		activityHandler = a.NewActivityHandler(db, er)
		if activity.Status.String() == entity.ActivityStatusValuesOpen.String() {
			activity.Status = entity.ActivityStatusValuesInProgress
		} else {
			activity.Status = entity.ActivityStatusValuesOpen
		}
		filter.Id = []*int64{&activity.Id}
		db.On("GetActivities", filter).Return([]entity.Activity{activity}, nil)
		updatedActivity, err := activityHandler.UpdateActivity(&activity)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedActivity.Status.String()).To(BeEquivalentTo(activity.Status.String()))
		})
	})
})

var _ = Describe("When deleting Activity", Label("app", "DeleteActivity"), func() {
	var (
		db              *mocks.MockDatabase
		activityHandler a.ActivityHandler
		id              int64
		filter          *entity.ActivityFilter
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
		activityHandler = a.NewActivityHandler(db, er)
		db.On("GetActivities", filter).Return([]entity.Activity{}, nil)
		err := activityHandler.DeleteActivity(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		activities, err := activityHandler.ListActivities(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(activities.Elements).To(BeEmpty(), "no error should be thrown")
	})
})

var _ = Describe("When modifying relationship of Service and Activity", Label("app", "ServiceActivityRelationship"), func() {
	var (
		db              *mocks.MockDatabase
		activityHandler a.ActivityHandler
		service         entity.Service
		activity        entity.Activity
		filter          *entity.ActivityFilter
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
		activityHandler = a.NewActivityHandler(db, er)
		activity, err := activityHandler.AddServiceToActivity(activity.Id, service.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(activity).NotTo(BeNil(), "activity should be returned")
	})

	It("removes service from activity", func() {
		db.On("RemoveServiceFromActivity", activity.Id, service.Id).Return(nil)
		db.On("GetActivities", filter).Return([]entity.Activity{activity}, nil)
		activityHandler = a.NewActivityHandler(db, er)
		activity, err := activityHandler.RemoveServiceFromActivity(activity.Id, service.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(activity).NotTo(BeNil(), "activity should be returned")
	})
})

var _ = Describe("When modifying relationship of Issue and Activity", Label("app", "IssueActivityRelationship"), func() {
	var (
		db              *mocks.MockDatabase
		activityHandler a.ActivityHandler
		issue           entity.Issue
		activity        entity.Activity
		filter          *entity.ActivityFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issue = test.NewFakeIssueEntity()
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

	It("adds issue to activity", func() {
		db.On("AddIssueToActivity", activity.Id, issue.Id).Return(nil)
		db.On("GetActivities", filter).Return([]entity.Activity{activity}, nil)
		activityHandler = a.NewActivityHandler(db, er)
		activity, err := activityHandler.AddIssueToActivity(activity.Id, issue.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(activity).NotTo(BeNil(), "activity should be returned")
	})

	It("removes issue from activity", func() {
		db.On("RemoveIssueFromActivity", activity.Id, issue.Id).Return(nil)
		db.On("GetActivities", filter).Return([]entity.Activity{activity}, nil)
		activityHandler = a.NewActivityHandler(db, er)
		activity, err := activityHandler.RemoveIssueFromActivity(activity.Id, issue.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(activity).NotTo(BeNil(), "activity should be returned")
	})
})
