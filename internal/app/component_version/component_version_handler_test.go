// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version_test

import (
	"math"
	"testing"

	cv "github.com/cloudoperators/heureka/internal/app/component_version"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	mock "github.com/stretchr/testify/mock"
)

func TestComponentVersionHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Component Version Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

func getComponentVersionFilter() *entity.ComponentVersionFilter {
	return &entity.ComponentVersionFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
	}
}

var _ = Describe("When listing ComponentVersions", Label("app", "ListComponentVersions"), func() {
	var (
		db                     *mocks.MockDatabase
		componenVersionService cv.ComponentVersionHandler
		filter                 *entity.ComponentVersionFilter
		options                *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getComponentVersionFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetComponentVersions", filter).Return([]entity.ComponentVersion{}, nil)
			db.On("CountComponentVersions", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			componenVersionService = cv.NewComponentVersionHandler(db, er)
			res, err := componenVersionService.ListComponentVersions(filter, options)
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
			componentVersions := test.NNewFakeComponentVersionEntities(resElements)

			var ids = lo.Map(componentVersions, func(cv entity.ComponentVersion, _ int) int64 { return cv.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetComponentVersions", filter).Return(componentVersions, nil)
			db.On("GetAllComponentVersionIds", filter).Return(ids, nil)
			componenVersionService = cv.NewComponentVersionHandler(db, er)
			res, err := componenVersionService.ListComponentVersions(filter, options)
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

var _ = Describe("When creating ComponentVersion", Label("app", "CreateComponentVersion"), func() {
	var (
		db                     *mocks.MockDatabase
		componenVersionService cv.ComponentVersionHandler
		componentVersion       entity.ComponentVersion
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		componentVersion = test.NewFakeComponentVersionEntity()
	})

	It("creates componentVersion", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateComponentVersion", &componentVersion).Return(&componentVersion, nil)
		componenVersionService = cv.NewComponentVersionHandler(db, er)
		newComponentVersion, err := componenVersionService.CreateComponentVersion(&componentVersion)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newComponentVersion.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newComponentVersion.Version).To(BeEquivalentTo(componentVersion.Version))
			Expect(newComponentVersion.ComponentId).To(BeEquivalentTo(componentVersion.ComponentId))
		})
	})
})

var _ = Describe("When updating ComponentVersion", Label("app", "UpdateComponentVersion"), func() {
	var (
		db                     *mocks.MockDatabase
		componenVersionService cv.ComponentVersionHandler
		componentVersion       entity.ComponentVersion
		filter                 *entity.ComponentVersionFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		componentVersion = test.NewFakeComponentVersionEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.ComponentVersionFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates componentVersion", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateComponentVersion", &componentVersion).Return(nil)
		componenVersionService = cv.NewComponentVersionHandler(db, er)
		componentVersion.Version = "7.3.3.1"
		filter.Id = []*int64{&componentVersion.Id}
		db.On("GetComponentVersions", filter).Return([]entity.ComponentVersion{componentVersion}, nil)
		updatedComponentVersion, err := componenVersionService.UpdateComponentVersion(&componentVersion)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedComponentVersion.Version).To(BeEquivalentTo(componentVersion.Version))
			Expect(updatedComponentVersion.ComponentId).To(BeEquivalentTo(componentVersion.ComponentId))
		})
	})
})

var _ = Describe("When deleting ComponentVersion", Label("app", "DeleteComponentVersion"), func() {
	var (
		db                     *mocks.MockDatabase
		componenVersionService cv.ComponentVersionHandler
		id                     int64
		filter                 *entity.ComponentVersionFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.ComponentVersionFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes componentVersion", func() {
		db.On("DeleteComponentVersion", id).Return(nil)
		componenVersionService = cv.NewComponentVersionHandler(db, er)
		db.On("GetComponentVersions", filter).Return([]entity.ComponentVersion{}, nil)
		err := componenVersionService.DeleteComponentVersion(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		componentVersions, err := componenVersionService.ListComponentVersions(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(componentVersions.Elements).To(BeEmpty(), "no error should be thrown")
	})
})
