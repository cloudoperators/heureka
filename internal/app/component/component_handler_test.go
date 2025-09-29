// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	c "github.com/cloudoperators/heureka/internal/app/component"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
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

func TestComponentHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Component Service Test Suite")
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
		AuthTokenSecret:       "key1",
		AuthzOpenFgaApiToken:  "key1",
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

func getComponentFilter() *entity.ComponentFilter {
	cCCRN := "SomeNotExistingComponent"
	return &entity.ComponentFilter{
		PaginatedX: entity.PaginatedX{
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

		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Cache:    cache.NewNoCache(),
			Authz:    authz,
		}
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetComponents", filter, []entity.Order{}).Return([]entity.ComponentResult{}, nil)
			db.On("CountComponents", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			componentHandler = c.NewComponentHandler(handlerContext)
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
			components := []entity.ComponentResult{}
			for _, c := range test.NNewFakeComponentEntities(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithComponent([]entity.Order{}, c, entity.ComponentVersion{}, entity.IssueSeverityCounts{}))
				components = append(components, entity.ComponentResult{WithCursor: entity.WithCursor{Value: cursor}, Component: lo.ToPtr(c)})
			}

			var cursors = lo.Map(components, func(m entity.ComponentResult, _ int) string {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithComponent([]entity.Order{}, *m.Component, entity.ComponentVersion{}, entity.IssueSeverityCounts{}))
				return cursor
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				component := test.NewFakeComponentEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithComponent([]entity.Order{}, component, entity.ComponentVersion{}, entity.IssueSeverityCounts{}))
				cursors = append(cursors, c)
			}
			db.On("GetComponents", filter, []entity.Order{}).Return(components, nil)
			db.On("GetAllComponentCursors", filter, []entity.Order{}).Return(cursors, nil)
			componentHandler = c.NewComponentHandler(handlerContext)
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
		p                openfga.PermissionInput
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		component = test.NewFakeComponentEntity()
		first := 10
		after := ""
		filter = &entity.ComponentFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}

		p = openfga.PermissionInput{
			UserType:   "role",
			UserId:     "testuser",
			ObjectId:   "testcomponent",
			ObjectType: "component",
			Relation:   "role",
		}

		handlerContext.DB = db
		cfg.CurrentUser = handlerContext.Authz.GetCurrentUser()
	})

	It("creates component", func() {
		filter.CCRN = []*string{&component.CCRN}
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateComponent", &component).Return(&component, nil)
		db.On("GetComponents", filter, []entity.Order{}).Return([]entity.ComponentResult{}, nil)
		componentHandler = c.NewComponentHandler(handlerContext)
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
				resourceId := strconv.FormatInt(createEvent.Component.Id, 10)
				p.ObjectId = openfga.ObjectId(resourceId)
				// Simulate event
				c.OnComponentCreateAuthz(db, event, authz)

				ok, err := authz.CheckPermission(p)
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
		component        entity.ComponentResult
		filter           *entity.ComponentFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		component = test.NewFakeComponentResult()
		first := 10
		after := ""
		filter = &entity.ComponentFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}

		handlerContext.DB = db
		cfg.CurrentUser = handlerContext.Authz.GetCurrentUser()
	})

	It("updates component", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateComponent", component.Component).Return(nil)
		componentHandler = c.NewComponentHandler(handlerContext)
		component.CCRN = "NewComponent"
		filter.Id = []*int64{&component.Id}
		db.On("GetComponents", filter, []entity.Order{}).Return([]entity.ComponentResult{component}, nil)
		updatedComponent, err := componentHandler.UpdateComponent(component.Component)
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
		p                openfga.PermissionInput
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		after := ""
		filter = &entity.ComponentFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}

		p = openfga.PermissionInput{
			UserType:   "role",
			UserId:     "testuser",
			ObjectId:   "testcomponent",
			ObjectType: "component",
			Relation:   "role",
		}

		handlerContext.DB = db
		cfg.CurrentUser = handlerContext.Authz.GetCurrentUser()
	})

	It("deletes component", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteComponent", id, mock.Anything).Return(nil)
		componentHandler = c.NewComponentHandler(handlerContext)
		db.On("GetComponents", filter, []entity.Order{}).Return([]entity.ComponentResult{}, nil)
		err := componentHandler.DeleteComponent(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		lo := entity.NewListOptions()
		components, err := componentHandler.ListComponents(filter, lo)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(components.Elements).To(BeEmpty(), "no error should be thrown")
	})

	Context("when handling an DeleteComponentEvent", func() {
		BeforeEach(func() {
			db.On("GetDefaultIssuePriority").Return(int64(100))
			db.On("GetDefaultRepositoryName").Return("nvd")
			p = openfga.PermissionInput{
				UserType:   "role",
				UserId:     "testuser",
				ObjectId:   "testcomponent",
				ObjectType: "component",
				Relation:   "role",
			}
		})

		Context("when new component is deleted", func() {
			It("should delete user resource relationship tuple in openfga", func() {
				authz := openfga.NewAuthorizationHandler(cfg, enableLogs)

				compFake := test.NewFakeComponentEntity()
				deleteEvent := &c.DeleteComponentEvent{
					ComponentID: compFake.Id,
				}

				// Use type assertion to convert a CreateServiceEvent into an Event
				var event event.Event = deleteEvent
				resourceId := strconv.FormatInt(deleteEvent.ComponentID, 10)
				p.ObjectId = openfga.ObjectId(resourceId)
				// Simulate event
				c.OnComponentDeleteAuthz(db, event, authz)

				ok, err := authz.CheckPermission(p)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(ok).To(BeFalse(), "permission should be granted")
			})
		})
	})

})
