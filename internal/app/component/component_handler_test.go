// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_test

import (
	"math"
	"strconv"
	"testing"

	c "github.com/cloudoperators/heureka/internal/app/component"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	mock "github.com/stretchr/testify/mock"
)

func TestComponentHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Component Service Test Suite")
}

var er event.EventRegistry
var authz openfga.Authorization

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db, authz)
})

func getComponentFilter() *entity.ComponentFilter {
	cCCRN := "SomeNotExistingComponent"
	return &entity.ComponentFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		CCRN: []*string{&cCCRN},
	}
}

var _ = Describe("When listing Components", Label("app", "ListComponents"), func() {
	var (
		db               *mocks.MockDatabase
		componentHandler c.ComponentHandler
		filter           *entity.ComponentFilter
		options          *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getComponentFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetComponents", filter).Return([]entity.Component{}, nil)
			db.On("CountComponents", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			componentHandler = c.NewComponentHandler(db, er, cache.NewNoCache(), authz)
			res, err := componentHandler.ListComponents(filter, options)
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
			components := test.NNewFakeComponentEntities(resElements)

			var ids = lo.Map(components, func(c entity.Component, _ int) int64 { return c.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetComponents", filter).Return(components, nil)
			db.On("GetAllComponentIds", filter).Return(ids, nil)
			componentHandler = c.NewComponentHandler(db, er, cache.NewNoCache(), authz)
			res, err := componentHandler.ListComponents(filter, options)
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

var _ = Describe("When creating Component", Label("app", "CreateComponent"), func() {
	var (
		db               *mocks.MockDatabase
		componentHandler c.ComponentHandler
		component        entity.Component
		filter           *entity.ComponentFilter
		authz            openfga.Authorization
		cfg              *util.Config
		enableLogs       bool
		userFieldName    string
		userId           string
		resourceId       string
		resourceType     string
		permission       string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		component = test.NewFakeComponentEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.ComponentFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}

		// setup authz testing
		userFieldName = "role"
		userId = "testuser"
		resourceId = ""
		resourceType = "component"
		permission = "role"

		cfg = &util.Config{
			AuthzEnabled:      true,
			CurrentUser:       userId,
			AuthModelFilePath: "../../../internal/openfga/model/model.fga",
			OpenFGApiUrl:      "http://localhost:8080",
		}
	})

	It("creates component", func() {
		filter.CCRN = []*string{&component.CCRN}
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateComponent", &component).Return(&component, nil)
		db.On("GetComponents", filter).Return([]entity.Component{}, nil)
		componentHandler = c.NewComponentHandler(db, er, cache.NewNoCache(), authz)
		newComponent, err := componentHandler.CreateComponent(&component)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newComponent.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newComponent.CCRN).To(BeEquivalentTo(component.CCRN))
			Expect(newComponent.Type).To(BeEquivalentTo(component.Type))
		})
	})

	Context("when handling a CreateComponentEvent", func() {
		BeforeEach(func() {
			db.On("GetDefaultIssuePriority").Return(int64(100))
			db.On("GetDefaultRepositoryName").Return("nvd")
		})

		Context("when new component is created", func() {
			It("should add user resource relationship tuple in openfga", func() {
				authz := openfga.NewAuthorizationHandler(cfg, enableLogs)

				compFake := test.NewFakeComponentEntity()
				createEvent := &c.CreateComponentEvent{
					Component: &compFake,
				}

				// Use type assertion to convert a CreateServiceEvent into an Event
				var event event.Event = createEvent
				resourceId = strconv.FormatInt(createEvent.Component.Id, 10)

				// Simulate event
				c.OnComponentCreateAuthz(db, event, authz)

				ok, err := authz.CheckPermission(userFieldName, userId, resourceId, resourceType, permission)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(ok).To(BeTrue(), "permission should be granted")
			})
		})
	})
})

var _ = Describe("When updating Component", Label("app", "UpdateComponent"), func() {
	var (
		db               *mocks.MockDatabase
		componentHandler c.ComponentHandler
		component        entity.Component
		filter           *entity.ComponentFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		component = test.NewFakeComponentEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.ComponentFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates component", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateComponent", &component).Return(nil)
		componentHandler = c.NewComponentHandler(db, er, cache.NewNoCache(), authz)
		component.CCRN = "NewComponent"
		filter.Id = []*int64{&component.Id}
		db.On("GetComponents", filter).Return([]entity.Component{component}, nil)
		updatedComponent, err := componentHandler.UpdateComponent(&component)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedComponent.CCRN).To(BeEquivalentTo(component.CCRN))
			Expect(updatedComponent.Type).To(BeEquivalentTo(component.Type))
		})
	})
})

var _ = Describe("When deleting Component", Label("app", "DeleteComponent"), func() {
	var (
		db               *mocks.MockDatabase
		componentHandler c.ComponentHandler
		id               int64
		filter           *entity.ComponentFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.ComponentFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes component", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteComponent", id, mock.Anything).Return(nil)
		componentHandler = c.NewComponentHandler(db, er, cache.NewNoCache(), authz)
		db.On("GetComponents", filter).Return([]entity.Component{}, nil)
		err := componentHandler.DeleteComponent(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		components, err := componentHandler.ListComponents(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(components.Elements).To(BeEmpty(), "no error should be thrown")
	})
})
