package component_instance_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	ci "github.wdf.sap.corp/cc/heureka/internal/app/component_instance"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/internal/entity/test"
	"github.wdf.sap.corp/cc/heureka/internal/mocks"
	"math"
	"testing"
)

func TestComponentInstanceService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Component Instance Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	er = event.NewEventRegistry()
})

func componentInstanceFilter() *entity.ComponentInstanceFilter {
	return &entity.ComponentInstanceFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		IssueMatchId: nil,
	}
}

func componentInstanceListOptions() *entity.ListOptions {
	return &entity.ListOptions{
		ShowTotalCount:      false,
		ShowPageInfo:        false,
		IncludeAggregations: false,
	}
}

var _ = Describe("When listing Component Instances", Label("app", "ListComponentInstances"), func() {
	var (
		db                       *mocks.MockDatabase
		componentInstanceService ci.ComponentInstanceService
		filter                   *entity.ComponentInstanceFilter
		options                  *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = componentInstanceListOptions()
		filter = componentInstanceFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetComponentInstances", filter).Return([]entity.ComponentInstance{}, nil)
			db.On("CountComponentInstances", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			componentInstanceService = ci.NewComponentInstanceService(db, er)
			res, err := componentInstanceService.ListComponentInstances(filter, options)
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
			componentInstances := test.NNewFakeComponentInstances(resElements)

			var ids = lo.Map(componentInstances, func(ci entity.ComponentInstance, _ int) int64 { return ci.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetComponentInstances", filter).Return(componentInstances, nil)
			db.On("GetAllComponentInstanceIds", filter).Return(ids, nil)
			componentInstanceService = ci.NewComponentInstanceService(db, er)
			res, err := componentInstanceService.ListComponentInstances(filter, options)
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

var _ = Describe("When creating ComponentInstance", Label("app", "CreateComponentInstance"), func() {
	var (
		db                       *mocks.MockDatabase
		componentInstanceService ci.ComponentInstanceService
		componentInstance        entity.ComponentInstance
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		componentInstance = test.NewFakeComponentInstanceEntity()
	})

	It("creates componentInstance", func() {
		db.On("CreateComponentInstance", &componentInstance).Return(&componentInstance, nil)
		componentInstanceService = ci.NewComponentInstanceService(db, er)
		newComponentInstance, err := componentInstanceService.CreateComponentInstance(&componentInstance)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newComponentInstance.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newComponentInstance.CCRN).To(BeEquivalentTo(componentInstance.CCRN))
			Expect(newComponentInstance.Count).To(BeEquivalentTo(componentInstance.Count))
			Expect(newComponentInstance.ComponentVersionId).To(BeEquivalentTo(componentInstance.ComponentVersionId))
			Expect(newComponentInstance.ServiceId).To(BeEquivalentTo(componentInstance.ServiceId))
		})
	})
})

var _ = Describe("When updating ComponentInstance", Label("app", "UpdateComponentInstance"), func() {
	var (
		db                       *mocks.MockDatabase
		componentInstanceService ci.ComponentInstanceService
		componentInstance        entity.ComponentInstance
		filter                   *entity.ComponentInstanceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		componentInstance = test.NewFakeComponentInstanceEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.ComponentInstanceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates componentInstance", func() {
		db.On("UpdateComponentInstance", &componentInstance).Return(nil)
		componentInstanceService = ci.NewComponentInstanceService(db, er)
		componentInstance.CCRN = "NewCCRN"
		filter.Id = []*int64{&componentInstance.Id}
		db.On("GetComponentInstances", filter).Return([]entity.ComponentInstance{componentInstance}, nil)
		updatedComponentInstance, err := componentInstanceService.UpdateComponentInstance(&componentInstance)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedComponentInstance.CCRN).To(BeEquivalentTo(componentInstance.CCRN))
			Expect(updatedComponentInstance.Count).To(BeEquivalentTo(componentInstance.Count))
			Expect(updatedComponentInstance.ComponentVersionId).To(BeEquivalentTo(componentInstance.ComponentVersionId))
			Expect(updatedComponentInstance.ServiceId).To(BeEquivalentTo(componentInstance.ServiceId))
		})
	})
})

var _ = Describe("When deleting ComponentInstance", Label("app", "DeleteComponentInstance"), func() {
	var (
		db                       *mocks.MockDatabase
		componentInstanceService ci.ComponentInstanceService
		id                       int64
		filter                   *entity.ComponentInstanceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.ComponentInstanceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes componentInstance", func() {
		db.On("DeleteComponentInstance", id).Return(nil)
		componentInstanceService = ci.NewComponentInstanceService(db, er)
		db.On("GetComponentInstances", filter).Return([]entity.ComponentInstance{}, nil)
		err := componentInstanceService.DeleteComponentInstance(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		componentInstances, err := componentInstanceService.ListComponentInstances(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(componentInstances.Elements).To(BeEmpty(), "no error should be thrown")
	})
})
