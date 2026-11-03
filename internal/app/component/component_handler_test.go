// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_test

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	c "github.com/cloudoperators/heureka/internal/app/component"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
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

func expectPermissionDenied(err error) {
	Expect(err).ToNot(BeNil(), "an error should be thrown")

	var appErr *appErrors.Error
	Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
	Expect(appErr.Code).To(Equal(appErrors.PermissionDenied), "should be PermissionDenied error")
}

var (
	handlerContext common.HandlerContext
	cfg            *util.Config
)

var _ = BeforeSuite(func() {
	authEnabled := false
	cfg = common.GetTestConfig(authEnabled)
	enableLogs := false
	db := mocks.NewMockDatabase(GinkgoT())
	authz := openfga.NewAuthorizationHandler(cfg, enableLogs)
	er := event.NewEventRegistry(db, authz)
	handlerContext = common.HandlerContext{
		DB:       db,
		EventReg: er,
		Cache:    nil,
		Authz:    authz,
	}
	handlerContext.Authz.RemoveAllRelations()
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
		er               event.EventRegistry
		db               *mocks.MockDatabase
		componentHandler c.ComponentHandler
		ctx              context.Context
		filter           *entity.ComponentFilter
		options          *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		options = entity.NewListOptions()
		filter = getComponentFilter()
		ctx = common.NewAdminContext()

		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("the list option does include the totalCount", func() {
		BeforeEach(func() {
			options.ShowTotalCount = true

			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
			db.On("GetComponents", mock.Anything, filter, []entity.Order{}).Return([]entity.ComponentResult{}, nil)
			db.On("CountComponents", mock.Anything, filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			componentHandler = c.NewComponentHandler(handlerContext)
			res, err := componentHandler.ListComponents(ctx, filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.TotalCount).Should(BeEquivalentTo(int64(1337)), "return correct Totalcount")
		})
	})

	When("the list option does include the PageInfo", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
		})
		DescribeTable(
			"pagination information is correct",
			func(pageSize int, dbElements int, resElements int, hasNextPage bool) {
				filter.First = &pageSize
				components := []entity.ComponentResult{}

				for _, c := range test.NNewFakeComponentEntities(resElements) {
					cursor, _ := mariadb.EncodeCursor(
						mariadb.WithComponent([]entity.Order{}, c, entity.IssueSeverityCounts{}),
					)
					components = append(
						components,
						entity.ComponentResult{
							WithCursor: entity.WithCursor{Value: cursor},
							Component:  new(c),
						},
					)
				}

				cursors := lo.Map(components, func(m entity.ComponentResult, _ int) string {
					cursor, _ := mariadb.EncodeCursor(
						mariadb.WithComponent(
							[]entity.Order{},
							*m.Component,
							entity.IssueSeverityCounts{},
						),
					)

					return cursor
				})

				var i int64 = 0
				for len(cursors) < dbElements {
					i++
					component := test.NewFakeComponentEntity()
					c, _ := mariadb.EncodeCursor(
						mariadb.WithComponent(
							[]entity.Order{},
							component,
							entity.IssueSeverityCounts{},
						),
					)
					cursors = append(cursors, c)
				}

				db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
				db.On("GetComponents", mock.Anything, filter, []entity.Order{}).Return(components, nil)
				db.On("GetAllComponentCursors", mock.Anything, filter, []entity.Order{}).Return(cursors, nil)

				componentHandler = c.NewComponentHandler(handlerContext)
				res, err := componentHandler.ListComponents(ctx, filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(
					*res.PageInfo.HasNextPage,
				).To(BeEquivalentTo(hasNextPage), "correct hasNextPage indicator")
				Expect(len(res.Elements)).To(BeEquivalentTo(resElements))
				Expect(
					len(res.PageInfo.Pages),
				).To(BeEquivalentTo(int(math.Ceil(float64(dbElements)/float64(pageSize)))), "correct  number of pages")
			},
			Entry("When  pageSize is 1 and the database was returning 2 elements", 1, 2, 1, true),
			Entry(
				"When  pageSize is 10 and the database was returning 9 elements",
				10,
				9,
				9,
				false,
			),
			Entry(
				"When  pageSize is 10 and the database was returning 11 elements",
				10,
				11,
				10,
				true,
			),
		)
	})

	Context("when authz is enabled", func() {
		BeforeEach(func() {
			authEnabled := true
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		AfterEach(func() {
			authEnabled := false
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		Context("and the user has no access to any components", func() {
			BeforeEach(func() {
				compIds := int64(-1)
				filter.Id = []*int64{&compIds}

				db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
				db.On("GetComponents", mock.Anything, filter, []entity.Order{}).
					Return([]entity.ComponentResult{}, nil)
			})

			It("should return no components", func() {
				componentHandler = c.NewComponentHandler(handlerContext)
				res, err := componentHandler.ListComponents(ctx, filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return 0 results")
			})
		})

		Context("and the filter includes component IDs the user has access to", func() {
			var component entity.Component

			BeforeEach(func() {
				userId := int64(123)
				systemUserId := int64(1)
				component = test.NewFakeComponentEntity()
				filter.Id = []*int64{&component.Id}

				db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
				db.On("GetComponents", mock.Anything, filter, []entity.Order{}).
					Return([]entity.ComponentResult{{Component: &component}}, nil)

				relations := []openfga.RelationInput{
					{ // create component
						UserType:   openfga.TypeRole,
						UserId:     openfga.UserIdFromInt(systemUserId),
						Relation:   openfga.RelRole,
						ObjectType: openfga.TypeComponent,
						ObjectId:   openfga.ObjectIdFromInt(component.Id),
					},
					{ // give user read permission to component
						UserType:   openfga.TypeUser,
						UserId:     openfga.UserIdFromInt(userId),
						Relation:   openfga.RelCanView,
						ObjectType: openfga.TypeComponent,
						ObjectId:   openfga.ObjectIdFromInt(component.Id),
					},
				}

				err := handlerContext.Authz.AddRelationBulk(relations)
				Expect(err).To(BeNil(), "no error should be thrown when adding relations")
			})

			It("should return the expected components in the result", func() {
				componentHandler = c.NewComponentHandler(handlerContext)
				res, err := componentHandler.ListComponents(ctx, filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(1), "return 1 result")
				Expect(
					res.Elements[0].CCRN,
				).To(BeEquivalentTo(component.CCRN))
				// check that the returned component is the expected one
			})
		})
	})
})

var _ = Describe("When creating Component", Label("app", "CreateComponent"), func() {
	var (
		er               event.EventRegistry
		db               *mocks.MockDatabase
		componentHandler c.ComponentHandler
		component        entity.Component
		filter           *entity.ComponentFilter
		r                openfga.RelationInput
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		component = test.NewFakeComponentEntity()

		handlerContext.Authz.RemoveAllRelations()

		first := 10
		after := ""
		filter = &entity.ComponentFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}

		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("creates component", func() {
		filter.CCRN = []*string{&component.CCRN}

		db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
		db.On("CreateComponent", &component).Return(&component, nil)
		db.On("GetComponents", mock.Anything, filter, []entity.Order{}).Return([]entity.ComponentResult{}, nil)

		componentHandler = c.NewComponentHandler(handlerContext)
		newComponent, err := componentHandler.CreateComponent(common.NewAdminContext(), &component)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newComponent.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newComponent.CCRN).To(BeEquivalentTo(component.CCRN))
			Expect(newComponent.Type).To(BeEquivalentTo(component.Type))
		})
	})

	Context("when authz is enabled", func() {
		BeforeEach(func() {
			authEnabled := true
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		AfterEach(func() {
			// Reset authz to disabled after finishing tests
			authEnabled := false
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		Context("when handling a CreateComponentEvent", func() {
			Context("when new component is created", func() {
				It("should add user resource relationship tuple in openfga", func() {
					compFake := test.NewFakeComponentEntity()
					createEvent := &c.CreateComponentEvent{
						Component: &compFake,
					}

					r = openfga.RelationInput{
						UserType:   openfga.TypeRole,
						UserId:     "0",
						ObjectType: openfga.TypeComponent,
						Relation:   openfga.RelRole,
					}

					// Use type assertion to convert a CreateServiceEvent into an Event
					var event event.Event = createEvent

					r.ObjectId = openfga.ObjectIdFromInt(createEvent.Component.Id)
					// Simulate event
					c.OnComponentCreateAuthz(db, event, handlerContext.Authz)

					ok, err := handlerContext.Authz.CheckPermission(r)
					Expect(err).To(BeNil(), "no error should be thrown")
					Expect(ok).To(BeTrue(), "permission should be granted")
				})
			})
		})

		Context("and checking write permission for CreateComponent", func() {
			It("denies creation when the user lacks admin permission on the system role", func() {
				userId := int64(9201)
				db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{userId}, nil)

				newComp := test.NewFakeComponentEntity()
				componentHandler = c.NewComponentHandler(handlerContext)
				_, err := componentHandler.CreateComponent(common.NewAdminContext(), &newComp)

				expectPermissionDenied(err)
			})

			It("allows creation when the user has admin permission on the system role", func() {
				userId := int64(9201)
				db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{userId}, nil)

				err := handlerContext.Authz.AddRelation(openfga.RelationInput{
					UserType:   openfga.TypeUser,
					UserId:     openfga.UserIdFromInt(userId),
					Relation:   openfga.RelAdmin,
					ObjectType: openfga.TypeRole,
					ObjectId:   openfga.SystemRoleId,
				})
				Expect(err).To(BeNil(), "no error should be thrown when adding relation")

				newComp := test.NewFakeComponentEntity()
				filter.CCRN = []*string{&newComp.CCRN}
				db.On("CreateComponent", &newComp).Return(&newComp, nil)
				db.On("GetComponents", mock.Anything, filter, []entity.Order{}).Return([]entity.ComponentResult{}, nil)

				componentHandler = c.NewComponentHandler(handlerContext)
				newComponent, err := componentHandler.CreateComponent(common.NewAdminContext(), &newComp)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(newComponent).NotTo(BeNil(), "component should be returned")
			})
		})
	})
})

var _ = Describe("When updating Component", Label("app", "UpdateComponent"), func() {
	var (
		er               event.EventRegistry
		db               *mocks.MockDatabase
		componentHandler c.ComponentHandler
		component        entity.ComponentResult
		filter           *entity.ComponentFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		component = test.NewFakeComponentResult()

		handlerContext.Authz.RemoveAllRelations()

		first := 10
		after := ""
		filter = &entity.ComponentFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}

		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("updates component", func() {
		db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
		db.On("UpdateComponent", component.Component).Return(nil)

		componentHandler = c.NewComponentHandler(handlerContext)
		component.CCRN = "NewComponent"
		filter.Id = []*int64{&component.Id}
		db.On("GetComponents", mock.Anything, filter, []entity.Order{}).
			Return([]entity.ComponentResult{component}, nil)
		updatedComponent, err := componentHandler.UpdateComponent(
			common.NewAdminContext(),
			component.Component,
		)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedComponent.CCRN).To(BeEquivalentTo(component.CCRN))
			Expect(updatedComponent.Type).To(BeEquivalentTo(component.Type))
		})
	})

	Context("when authz is enabled", func() {
		BeforeEach(func() {
			authEnabled := true
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
			handlerContext.Authz.RemoveAllRelations()
		})

		AfterEach(func() {
			authEnabled := false
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		It("denies the update when the user lacks can_write permission on the component", func() {
			userId := int64(9202)
			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{userId}, nil)

			componentHandler = c.NewComponentHandler(handlerContext)
			_, err := componentHandler.UpdateComponent(common.NewAdminContext(), component.Component)

			expectPermissionDenied(err)
		})

		It("allows the update when the user has can_write permission on the component", func() {
			userId := int64(9202)
			db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{userId}, nil)
			db.On("UpdateComponent", component.Component).Return(nil)

			err := handlerContext.Authz.AddRelation(openfga.RelationInput{
				UserType:   openfga.TypeUser,
				UserId:     openfga.UserIdFromInt(userId),
				Relation:   openfga.RelCanWrite,
				ObjectType: openfga.TypeComponent,
				ObjectId:   openfga.ObjectIdFromInt(component.Id),
			})
			Expect(err).To(BeNil(), "no error should be thrown when adding relation")

			filter.Id = []*int64{&component.Id}
			db.On("GetComponents", mock.Anything, filter, []entity.Order{}).Return([]entity.ComponentResult{component}, nil)

			componentHandler = c.NewComponentHandler(handlerContext)
			updatedComponent, err := componentHandler.UpdateComponent(common.NewAdminContext(), component.Component)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(updatedComponent).NotTo(BeNil(), "component should be returned")
		})
	})
})

var _ = Describe("When deleting Component", Label("app", "DeleteComponent"), func() {
	var (
		er               event.EventRegistry
		db               *mocks.MockDatabase
		componentHandler c.ComponentHandler
		id               int64
		filter           *entity.ComponentFilter
		ctx              context.Context
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		handlerContext.Authz.RemoveAllRelations()

		id = 1
		first := 10
		after := ""
		filter = &entity.ComponentFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
		ctx = common.NewAdminContext()

		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("deletes component", func() {
		db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{}, nil)
		db.On("DeleteComponent", id, mock.Anything).Return(nil)

		componentHandler = c.NewComponentHandler(handlerContext)

		db.On("GetComponents", mock.Anything, filter, []entity.Order{}).Return([]entity.ComponentResult{}, nil)

		err := componentHandler.DeleteComponent(common.NewAdminContext(), id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		lo := entity.NewListOptions()
		components, err := componentHandler.ListComponents(ctx, filter, lo)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(components.Elements).To(BeEmpty(), "no error should be thrown")
	})

	Context("when authz is enabled", func() {
		BeforeEach(func() {
			authEnabled := true
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		AfterEach(func() {
			// Reset authz to disabled after finishing tests
			authEnabled := false
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		Context("when handling an DeleteComponentEvent", func() {
			Context("when new component is deleted", func() {
				It("should delete tuples related to that component in openfga", func() {
					// Test OnComponentDeleteAuthz against all possible relations
					compFake := test.NewFakeComponentEntity()
					deleteEvent := &c.DeleteComponentEvent{
						ComponentID: compFake.Id,
					}
					objectId := openfga.ObjectIdFromInt(deleteEvent.ComponentID)
					relations := []openfga.RelationInput{
						{ // role - component: a role is assigned to the component
							UserType:   openfga.TypeRole,
							UserId:     openfga.IDRole,
							ObjectId:   objectId,
							ObjectType: openfga.TypeComponent,
							Relation:   openfga.RelRole,
						},
						{ // component_version - component: a component version is related to the component
							UserType:   openfga.TypeComponentVersion,
							UserId:     openfga.IDComponentVersion,
							ObjectId:   objectId,
							ObjectType: openfga.TypeComponent,
							Relation:   openfga.RelComponentVersion,
						},
						{ // user - component: a user can view the component
							UserType:   openfga.TypeUser,
							UserId:     openfga.IDUser,
							ObjectId:   objectId,
							ObjectType: openfga.TypeComponent,
							Relation:   openfga.RelCanView,
						},
					}

					handlerContext.Authz.AddRelationBulk(relations)

					// get the number of relations before deletion
					relCountBefore := 0

					for _, r := range relations {
						relationsList, err := handlerContext.Authz.ListRelations(r)
						Expect(err).To(BeNil(), "no error should be thrown")

						relCountBefore += len(relationsList)
					}

					relationsCountBefore := relCountBefore
					Expect(
						relationsCountBefore,
					).To(BeEquivalentTo(len(relations)), "all relations should exist before deletion")

					// check that relations were created
					for _, r := range relations {
						ok, err := handlerContext.Authz.CheckPermission(r)
						Expect(err).To(BeNil(), "no error should be thrown")
						Expect(ok).To(BeTrue(), "permission should be granted")
					}

					var event event.Event = deleteEvent
					c.OnComponentDeleteAuthz(db, event, handlerContext.Authz)

					// get the number of relations after deletion
					relCountAfter := 0

					for _, r := range relations {
						relationsList, err := handlerContext.Authz.ListRelations(r)
						Expect(err).To(BeNil(), "no error should be thrown")

						relCountAfter += len(relationsList)
					}

					relationsCountAfter := relCountAfter
					Expect(
						relationsCountAfter < relationsCountBefore,
					).To(BeTrue(), "less relations after deletion")
					Expect(
						relationsCountAfter,
					).To(BeEquivalentTo(0), "no relations should exist after deletion")

					// verify that relations were deleted
					for _, r := range relations {
						ok, err := handlerContext.Authz.CheckPermission(r)
						Expect(err).To(BeNil(), "no error should be thrown")
						Expect(ok).To(BeFalse(), "permission should NOT be granted")
					}
				})
			})
		})

		Context("and checking write permission for DeleteComponent", func() {
			It("denies deletion when the user lacks can_write permission on the component", func() {
				userId := int64(9203)
				db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{userId}, nil)

				componentHandler = c.NewComponentHandler(handlerContext)
				err := componentHandler.DeleteComponent(common.NewAdminContext(), id)

				expectPermissionDenied(err)
			})

			It("allows deletion when the user has can_write permission on the component", func() {
				userId := int64(9203)
				db.On("GetAllUserIds", mock.Anything, mock.Anything).Return([]int64{userId}, nil)
				db.On("DeleteComponent", id, mock.Anything).Return(nil)

				err := handlerContext.Authz.AddRelation(openfga.RelationInput{
					UserType:   openfga.TypeUser,
					UserId:     openfga.UserIdFromInt(userId),
					Relation:   openfga.RelCanWrite,
					ObjectType: openfga.TypeComponent,
					ObjectId:   openfga.ObjectIdFromInt(id),
				})
				Expect(err).To(BeNil(), "no error should be thrown when adding relation")

				componentHandler = c.NewComponentHandler(handlerContext)
				err = componentHandler.DeleteComponent(common.NewAdminContext(), id)
				Expect(err).To(BeNil(), "no error should be thrown")
			})
		})
	})
})
