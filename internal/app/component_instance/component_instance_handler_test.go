// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance_test

import (
	"math"
	"testing"

	ci "github.com/cloudoperators/heureka/internal/app/component_instance"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	dbtest "github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	mock "github.com/stretchr/testify/mock"
)

func TestComponentInstanceHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Component Instance Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

func componentInstanceFilter() *entity.ComponentInstanceFilter {
	return &entity.ComponentInstanceFilter{
		PaginatedX: entity.PaginatedX{
			First: nil,
			After: nil,
		},
		IssueMatchId: nil,
		CCRN:         nil,
	}
}

var _ = Describe("When listing Component Instances", Label("app", "ListComponentInstances"), func() {
	var (
		db                       *mocks.MockDatabase
		componentInstanceHandler ci.ComponentInstanceHandler
		filter                   *entity.ComponentInstanceFilter
		options                  *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = componentInstanceFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetComponentInstances", filter, []entity.Order{}).Return([]entity.ComponentInstanceResult{}, nil)
			db.On("CountComponentInstances", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			componentInstanceHandler = ci.NewComponentInstanceHandler(db, er)
			res, err := componentInstanceHandler.ListComponentInstances(filter, options)
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
			componentInstances := []entity.ComponentInstanceResult{}
			for _, ci := range test.NNewFakeComponentInstances(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithComponentInstance([]entity.Order{}, ci))
				componentInstances = append(componentInstances, entity.ComponentInstanceResult{WithCursor: entity.WithCursor{Value: cursor}, ComponentInstance: lo.ToPtr(ci)})
			}

			var cursors = lo.Map(componentInstances, func(m entity.ComponentInstanceResult, _ int) string {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithComponentInstance([]entity.Order{}, *m.ComponentInstance))
				return cursor
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				componentInstance := test.NewFakeComponentInstanceEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithComponentInstance([]entity.Order{}, componentInstance))
				cursors = append(cursors, c)
			}
			db.On("GetComponentInstances", filter, []entity.Order{}).Return(componentInstances, nil)
			db.On("GetAllComponentInstanceCursors", filter, []entity.Order{}).Return(cursors, nil)
			componentInstanceHandler = ci.NewComponentInstanceHandler(db, er)
			res, err := componentInstanceHandler.ListComponentInstances(filter, options)
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
		componentInstanceHandler ci.ComponentInstanceHandler
		componentInstance        entity.ComponentInstance
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		componentInstance = test.NewFakeComponentInstanceEntity()
	})

	It("creates componentInstance", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateComponentInstance", &componentInstance).Return(&componentInstance, nil)
		db.On("CreateScannerRunComponentInstanceTracker", componentInstance.Id, "").Return(nil)
		componentInstanceHandler = ci.NewComponentInstanceHandler(db, er)
		newComponentInstance, err := componentInstanceHandler.CreateComponentInstance(&componentInstance, "")
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newComponentInstance.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newComponentInstance.CCRN).To(BeEquivalentTo(componentInstance.CCRN))
			Expect(newComponentInstance.Region).To(BeEquivalentTo(componentInstance.Region))
			Expect(newComponentInstance.Cluster).To(BeEquivalentTo(componentInstance.Cluster))
			Expect(newComponentInstance.Namespace).To(BeEquivalentTo(componentInstance.Namespace))
			Expect(newComponentInstance.Domain).To(BeEquivalentTo(componentInstance.Domain))
			Expect(newComponentInstance.Project).To(BeEquivalentTo(componentInstance.Project))
			Expect(newComponentInstance.Pod).To(BeEquivalentTo(componentInstance.Pod))
			Expect(newComponentInstance.Container).To(BeEquivalentTo(componentInstance.Container))
			Expect(newComponentInstance.Type).To(BeEquivalentTo(componentInstance.Type))
			Expect(newComponentInstance.Context).To(BeEquivalentTo(componentInstance.Context))
			Expect(newComponentInstance.Count).To(BeEquivalentTo(componentInstance.Count))
			Expect(newComponentInstance.ComponentVersionId).To(BeEquivalentTo(componentInstance.ComponentVersionId))
			Expect(newComponentInstance.ServiceId).To(BeEquivalentTo(componentInstance.ServiceId))
		})
	})
})

var _ = Describe("When updating ComponentInstance", Label("app", "UpdateComponentInstance"), func() {
	var (
		db                       *mocks.MockDatabase
		componentInstanceHandler ci.ComponentInstanceHandler
		componentInstance        entity.ComponentInstanceResult
		filter                   *entity.ComponentInstanceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		componentInstance = test.NewFakeComponentInstanceResult()
		first := 10
		after := ""
		filter = &entity.ComponentInstanceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates componentInstance", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateComponentInstance", componentInstance.ComponentInstance).Return(nil)
		componentInstanceHandler = ci.NewComponentInstanceHandler(db, er)
		componentInstance.Region = "NewRegion"
		componentInstance.Cluster = "NewCluster"
		componentInstance.Namespace = "NewNamespace"
		componentInstance.Domain = "NewDomain"
		componentInstance.Project = "NewProject"
		componentInstance.Pod = "NewPod"
		componentInstance.Container = "NewContainer"
		componentInstance.Type = "Server"
		componentInstance.Context = &entity.Json{"my_ip": "192.168.0.0"}
		componentInstance.CCRN = dbtest.GenerateFakeCcrn(componentInstance.Cluster, componentInstance.Namespace)
		filter.Id = []*int64{&componentInstance.Id}
		db.On("GetComponentInstances", filter, []entity.Order{}).Return([]entity.ComponentInstanceResult{componentInstance}, nil)
		updatedComponentInstance, err := componentInstanceHandler.UpdateComponentInstance(componentInstance.ComponentInstance)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedComponentInstance.CCRN).To(BeEquivalentTo(componentInstance.CCRN))
			Expect(updatedComponentInstance.Region).To(BeEquivalentTo(componentInstance.Region))
			Expect(updatedComponentInstance.Cluster).To(BeEquivalentTo(componentInstance.Cluster))
			Expect(updatedComponentInstance.Namespace).To(BeEquivalentTo(componentInstance.Namespace))
			Expect(updatedComponentInstance.Domain).To(BeEquivalentTo(componentInstance.Domain))
			Expect(updatedComponentInstance.Project).To(BeEquivalentTo(componentInstance.Project))
			Expect(updatedComponentInstance.Pod).To(BeEquivalentTo(componentInstance.Pod))
			Expect(updatedComponentInstance.Container).To(BeEquivalentTo(componentInstance.Container))
			Expect(updatedComponentInstance.Type).To(BeEquivalentTo(componentInstance.Type))
			Expect(updatedComponentInstance.Context).To(BeEquivalentTo(componentInstance.Context))
			Expect(updatedComponentInstance.Count).To(BeEquivalentTo(componentInstance.Count))
			Expect(updatedComponentInstance.ComponentVersionId).To(BeEquivalentTo(componentInstance.ComponentVersionId))
			Expect(updatedComponentInstance.ServiceId).To(BeEquivalentTo(componentInstance.ServiceId))
		})
	})
})

var _ = Describe("When deleting ComponentInstance", Label("app", "DeleteComponentInstance"), func() {
	var (
		db                       *mocks.MockDatabase
		componentInstanceHandler ci.ComponentInstanceHandler
		id                       int64
		filter                   *entity.ComponentInstanceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		after := ""
		filter = &entity.ComponentInstanceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes componentInstance", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteComponentInstance", id, mock.Anything).Return(nil)
		componentInstanceHandler = ci.NewComponentInstanceHandler(db, er)
		db.On("GetComponentInstances", filter, []entity.Order{}).Return([]entity.ComponentInstanceResult{}, nil)
		err := componentInstanceHandler.DeleteComponentInstance(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		lo := entity.NewListOptions()
		componentInstances, err := componentInstanceHandler.ListComponentInstances(filter, lo)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(componentInstances.Elements).To(BeEmpty(), "no error should be thrown")
	})
})

var _ = Describe("When listing CCRN", Label("app", "ListCcrn"), func() {
	var (
		db                       *mocks.MockDatabase
		componentInstanceHandler ci.ComponentInstanceHandler
		filter                   *entity.ComponentInstanceFilter
		options                  *entity.ListOptions
		CCRN                     string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = componentInstanceFilter()
		CCRN = "ca9d963d-b441-4167-b08d-086e76186653"
	})

	When("no filters are used", func() {

		BeforeEach(func() {
			db.On("GetCcrn", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			componentInstanceHandler = ci.NewComponentInstanceHandler(db, er)
			res, err := componentInstanceHandler.ListCcrns(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific CCRN filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.ComponentInstanceFilter{
				CCRN: []*string{&CCRN},
			}

			db.On("GetCcrn", filter).Return([]string{CCRN}, nil)
		})
		It("returns filtered CCRN according to the CCRN type", func() {
			componentInstanceHandler = ci.NewComponentInstanceHandler(db, er)
			res, err := componentInstanceHandler.ListCcrns(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(CCRN), "should only consist of CCRN")
		})
	})
})
