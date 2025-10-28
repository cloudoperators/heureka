// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance_test

import (
	"errors"
	"os"

	"math"
	"strconv"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	ci "github.com/cloudoperators/heureka/internal/app/component_instance"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	dbtest "github.com/cloudoperators/heureka/internal/database/mariadb/test"
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

func TestComponentInstanceHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Component Instance Service Test Suite")
}

var handlerContext common.HandlerContext
var cfg *util.Config

var _ = BeforeSuite(func() {
	modelFilePath := "./../../openfga/model/model.fga"

	cfg = &util.Config{
		AuthzOpenFgaApiUrl:    os.Getenv("AUTHZ_FGA_API_URL"),
		AuthzOpenFgaApiToken:  os.Getenv("AUTHZ_FGA_API_TOKEN"),
		AuthzOpenFgaStoreName: os.Getenv("AUTHZ_FGA_STORE_NAME"),
		AuthzModelFilePath:    modelFilePath,
		CurrentUser:           "testuser",
	}
	enableLogs := false
	authz := openfga.NewAuthorizationHandler(cfg, enableLogs)
	handlerContext = common.HandlerContext{
		Cache: cache.NewNoCache(),
		Authz: authz,
	}
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
		er                       event.EventRegistry
		db                       *mocks.MockDatabase
		componentInstanceHandler ci.ComponentInstanceHandler
		filter                   *entity.ComponentInstanceFilter
		options                  *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)

		options = entity.NewListOptions()
		filter = componentInstanceFilter()
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetComponentInstances", filter, []entity.Order{}).Return([]entity.ComponentInstanceResult{}, nil)
			db.On("CountComponentInstances", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
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
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
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

	Context("when GetComponentInstances fails", func() {
		It("should return Internal error", func() {
			// Mock database error
			dbError := errors.New("database connection failed")
			db.On("GetComponentInstances", filter, []entity.Order{}).Return([]entity.ComponentInstanceResult{}, dbError)

			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			result, err := componentInstanceHandler.ListComponentInstances(filter, options)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			// Verify it's our structured error with correct code
			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("ComponentInstances"), "should reference ComponentInstances entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("componentInstanceHandler.ListComponentInstances"), "should include operation")
			Expect(appErr.Err.Error()).To(ContainSubstring("database connection failed"), "should contain original error message")
		})
	})

	Context("when GetAllComponentInstanceCursors fails", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
			filter.First = lo.ToPtr(10)
		})

		It("should return Internal error", func() {
			componentInstances := []entity.ComponentInstanceResult{}
			for _, ci := range test.NNewFakeComponentInstances(5) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithComponentInstance([]entity.Order{}, ci))
				componentInstances = append(componentInstances, entity.ComponentInstanceResult{
					WithCursor:        entity.WithCursor{Value: cursor},
					ComponentInstance: lo.ToPtr(ci),
				})
			}

			db.On("GetComponentInstances", filter, []entity.Order{}).Return(componentInstances, nil)
			cursorsError := errors.New("cursor database error")
			db.On("GetAllComponentInstanceCursors", filter, []entity.Order{}).Return([]string{}, cursorsError)

			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			result, err := componentInstanceHandler.ListComponentInstances(filter, options)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("ComponentInstanceCursors"), "should reference ComponentInstanceCursors entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("componentInstanceHandler.ListComponentInstances"), "should include operation")
		})
	})

})

var _ = Describe("When creating ComponentInstance", Label("app", "CreateComponentInstance"), func() {
	var (
		er                       event.EventRegistry
		db                       *mocks.MockDatabase
		componentInstanceHandler ci.ComponentInstanceHandler
		componentInstance        entity.ComponentInstance
		p                        openfga.PermissionInput
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		componentInstance = test.NewFakeComponentInstanceEntity()

		p = openfga.PermissionInput{
			UserType:   "role",
			UserId:     "testuser",
			ObjectId:   "test_component_instance",
			ObjectType: "component_instance",
			Relation:   "role",
		}

		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	Context("with valid input", func() {
		It("creates componentInstance", func() {
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil)
			db.On("CreateComponentInstance", mock.AnythingOfType("*entity.ComponentInstance")).Return(&componentInstance, nil)

			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			// Ensure type is allowed if ParentId is set
			componentInstance.Type = "RecordSet"
			componentInstance.ParentId = 1234
			newComponentInstance, err := componentInstanceHandler.CreateComponentInstance(&componentInstance, nil)
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
				Expect(newComponentInstance.ParentId).To(BeEquivalentTo(componentInstance.ParentId))
			})
		})
	})

	Context("when handling a CreateComponentInstanceEvent", func() {
		BeforeEach(func() {
			db.On("GetDefaultIssuePriority").Return(int64(100))
			db.On("GetDefaultRepositoryName").Return("nvd")
		})

		Context("when new component instance is created", func() {
			It("should add user resource relationship tuple in openfga", func() {
				ciFake := test.NewFakeComponentInstanceEntity()
				createEvent := &ci.CreateComponentInstanceEvent{
					ComponentInstance: &ciFake,
				}

				// Use type assertion to convert a CreateServiceEvent into an Event
				var event event.Event = createEvent
				resourceId := strconv.FormatInt(createEvent.ComponentInstance.Id, 10)
				p.ObjectId = openfga.ObjectId(resourceId)
				// Simulate event
				ci.OnComponentInstanceCreateAuthz(db, event, handlerContext.Authz)

				ok, err := handlerContext.Authz.CheckPermission(p)
				Expect(err).To(BeNil(), "no error should be thrown")
				if cfg.AuthzOpenFgaApiUrl != "" {
					Expect(ok).To(BeTrue(), "permission should be granted")
				} else {
					Expect(ok).To(BeFalse(), "permission should not be granted when no AuthzOpenFgaApiUrl is set")
				}
			})
		})
	})
})

var _ = Describe("When updating ComponentInstance", Label("app", "UpdateComponentInstance"), func() {
	var (
		er                       event.EventRegistry
		db                       *mocks.MockDatabase
		componentInstanceHandler ci.ComponentInstanceHandler
		componentInstance        entity.ComponentInstanceResult
		filter                   *entity.ComponentInstanceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		componentInstance = test.NewFakeComponentInstanceResult()
		first := 10
		after := ""
		filter = &entity.ComponentInstanceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		handlerContext.DB = db
		handlerContext.EventReg = er
	})
	Context("with valid input", func() {
		It("updates componentInstance", func() {
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil) // Changed: return actual user ID
			db.On("UpdateComponentInstance", componentInstance.ComponentInstance).Return(nil)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstance.Region = "NewRegion"
			componentInstance.Cluster = "NewCluster"
			componentInstance.Namespace = "NewNamespace"
			componentInstance.Domain = "NewDomain"
			componentInstance.Project = "NewProject"
			componentInstance.Pod = "NewPod"
			componentInstance.Container = "NewContainer"
			componentInstance.Type = "RecordSet"
			componentInstance.Context = &entity.Json{"my_ip": "192.168.0.0"}
			componentInstance.ParentId = 1234
			componentInstance.CCRN = dbtest.GenerateFakeCcrn(componentInstance.Cluster, componentInstance.Namespace)
			filter.Id = []*int64{&componentInstance.Id}
			db.On("GetComponentInstances", filter, []entity.Order{}).Return([]entity.ComponentInstanceResult{componentInstance}, nil)
			updatedComponentInstance, err := componentInstanceHandler.UpdateComponentInstance(componentInstance.ComponentInstance, nil)
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
				Expect(updatedComponentInstance.ParentId).To(BeEquivalentTo(componentInstance.ParentId))
			})
		})
	})

	Context("when handling an UpdateComponentInstanceEvent", func() {
		BeforeEach(func() {
			db.On("GetDefaultIssuePriority").Return(int64(100))
			db.On("GetDefaultRepositoryName").Return("nvd")
		})

		It("should update service and component_version relations in openfga", func() {
			ciFake := test.NewFakeComponentInstanceEntity()
			oldServiceId := int64(111)
			newServiceId := int64(222)
			oldComponentVersionId := int64(333)
			newComponentVersionId := int64(444)

			// Add initial relations
			initialServiceRelation := openfga.RelationInput{
				UserType:   "service",
				UserId:     openfga.UserId(strconv.FormatInt(oldServiceId, 10)),
				Relation:   "related_service",
				ObjectType: "component_instance",
				ObjectId:   openfga.ObjectId(strconv.FormatInt(ciFake.Id, 10)),
			}
			initialComponentVersionRelation := openfga.RelationInput{
				UserType:   "component_instance",
				UserId:     openfga.UserId(strconv.FormatInt(ciFake.Id, 10)),
				Relation:   "component_instance",
				ObjectType: "component_version",
				ObjectId:   openfga.ObjectId(strconv.FormatInt(oldComponentVersionId, 10)),
			}
			handlerContext.Authz.AddRelation(initialServiceRelation)
			handlerContext.Authz.AddRelation(initialComponentVersionRelation)

			// Fake update event with new service and component_version ids
			ciFake.ServiceId = newServiceId
			ciFake.ComponentVersionId = newComponentVersionId
			updateEvent := &ci.UpdateComponentInstanceEvent{
				ComponentInstance: &ciFake,
			}
			var event event.Event = updateEvent

			// Simulate event
			ci.OnComponentInstanceUpdateAuthz(db, event, handlerContext.Authz)

			// Check that the old relations are gone
			remainingOldService, err := handlerContext.Authz.ListRelations([]openfga.RelationInput{initialServiceRelation})
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remainingOldService).To(BeEmpty(), "old service relation should be removed")

			remainingOldComponentVersion, err := handlerContext.Authz.ListRelations([]openfga.RelationInput{initialComponentVersionRelation})
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remainingOldComponentVersion).To(BeEmpty(), "old component_version relation should be removed")

			// Check that the new relations exist
			newServiceRelation := openfga.RelationInput{
				UserType:   "service",
				UserId:     openfga.UserId(strconv.FormatInt(newServiceId, 10)),
				Relation:   "related_service",
				ObjectType: "component_instance",
				ObjectId:   openfga.ObjectId(strconv.FormatInt(ciFake.Id, 10)),
			}
			newComponentVersionRelation := openfga.RelationInput{
				UserType:   "component_instance",
				UserId:     openfga.UserId(strconv.FormatInt(ciFake.Id, 10)),
				Relation:   "component_instance",
				ObjectType: "component_version",
				ObjectId:   openfga.ObjectId(strconv.FormatInt(newComponentVersionId, 10)),
			}
			remainingNewService, err := handlerContext.Authz.ListRelations([]openfga.RelationInput{newServiceRelation})
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remainingNewService).NotTo(BeEmpty(), "new service relation should exist")

			remainingNewComponentVersion, err := handlerContext.Authz.ListRelations([]openfga.RelationInput{newComponentVersionRelation})
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remainingNewComponentVersion).NotTo(BeEmpty(), "new component_version relation should exist")
		})
	})
})

var _ = Describe("When deleting ComponentInstance", Label("app", "DeleteComponentInstance"), func() {
	var (
		er                       event.EventRegistry
		db                       *mocks.MockDatabase
		componentInstanceHandler ci.ComponentInstanceHandler
		id                       int64
		filter                   *entity.ComponentInstanceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		id = 1
		first := 10
		after := ""
		filter = &entity.ComponentInstanceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	Context("with valid input", func() {
		It("deletes componentInstance", func() {
			db.On("GetAllUserIds", mock.Anything).Return([]int64{123}, nil) // Changed: return actual user ID
			db.On("DeleteComponentInstance", id, int64(123)).Return(nil)    // Changed: specify exact user ID
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			db.On("GetComponentInstances", filter, []entity.Order{}).Return([]entity.ComponentInstanceResult{}, nil)
			err := componentInstanceHandler.DeleteComponentInstance(id)
			Expect(err).To(BeNil(), "no error should be thrown")

			filter.Id = []*int64{&id}
			lo := entity.NewListOptions()
			componentInstances, err := componentInstanceHandler.ListComponentInstances(filter, lo)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(componentInstances.Elements).To(BeEmpty(), "component instance should be deleted")
		})

		Context("when handling a DeleteComponentInstanceEvent", func() {
			BeforeEach(func() {
				db.On("GetDefaultIssuePriority").Return(int64(100))
				db.On("GetDefaultRepositoryName").Return("nvd")
			})

			Context("when new component instance is deleted", func() {
				It("should delete tuples related to that component instance in openfga", func() {
					// Test OnComponentInstanceDeleteAuthz against all possible relations
					ciFake := test.NewFakeComponentInstanceEntity()
					deleteEvent := &ci.DeleteComponentInstanceEvent{
						ComponentInstanceID: ciFake.Id,
					}
					objectId := openfga.ObjectId(strconv.FormatInt(deleteEvent.ComponentInstanceID, 10))
					userId := openfga.UserId(strconv.FormatInt(deleteEvent.ComponentInstanceID, 10))
					relations := []openfga.RelationInput{
						{ // role - component_instance: a role is assigned to the component instance
							UserType:   "role",
							UserId:     "roleID",
							ObjectId:   objectId,
							ObjectType: "component_instance",
							Relation:   "role",
						},
						{ // service - component_instance: a service is related to the component instance
							UserType:   "service",
							UserId:     "serviceID",
							ObjectId:   objectId,
							ObjectType: "component_instance",
							Relation:   "related_service",
						},
						{ // user - component_instance: a user can view the component instance
							UserType:   "user",
							UserId:     "userID",
							ObjectId:   objectId,
							ObjectType: "component_instance",
							Relation:   "can_view",
						},
						{ // component_instance - component_version: a component instance is related to a component version
							UserType:   "component_instance",
							UserId:     userId,
							ObjectId:   "cvID",
							ObjectType: "component_version",
							Relation:   "component_instance",
						},
						{ // component_instance - issue_match: a component instance is related to an issue match
							UserType:   "component_instance",
							UserId:     userId,
							ObjectId:   "issueMatchID",
							ObjectType: "issue_match",
							Relation:   "component_instance",
						},
					}

					for _, rel := range relations {
						handlerContext.Authz.AddRelation(rel)
					}

					var event event.Event = deleteEvent
					// Simulate event
					ci.OnComponentInstanceDeleteAuthz(db, event, handlerContext.Authz)

					remaining, err := handlerContext.Authz.ListRelations(relations)
					Expect(err).To(BeNil(), "no error should be thrown")
					Expect(remaining).To(BeEmpty(), "no relations should remain after deletion")
				})
			})
		})
	})
})

var _ = Describe("When listing CCRN", Label("app", "ListCcrn"), func() {
	var (
		er                       event.EventRegistry
		db                       *mocks.MockDatabase
		componentInstanceHandler ci.ComponentInstanceHandler
		filter                   *entity.ComponentInstanceFilter
		options                  *entity.ListOptions
		CCRN                     string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		options = entity.NewListOptions()
		filter = componentInstanceFilter()
		CCRN = "ca9d963d-b441-4167-b08d-086e76186653"
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("no filters are used", func() {
		BeforeEach(func() {
			db.On("GetCcrn", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
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
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			res, err := componentInstanceHandler.ListCcrns(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(CCRN), "should only consist of CCRN")
		})
	})

	// NEW: Add error handling test case
	Context("when database operation fails", func() {
		It("should return Internal error", func() {
			// Mock database error
			dbError := errors.New("database connection failed")
			db.On("GetCcrn", filter).Return([]string{}, dbError)

			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			componentInstanceHandler = ci.NewComponentInstanceHandler(handlerContext)
			result, err := componentInstanceHandler.ListCcrns(filter, options)

			Expect(result).To(BeNil(), "no result should be returned")
			Expect(err).ToNot(BeNil(), "error should be returned")

			// Verify it's our structured error with correct code
			var appErr *appErrors.Error
			Expect(errors.As(err, &appErr)).To(BeTrue(), "should be application error")
			Expect(appErr.Code).To(Equal(appErrors.Internal), "should be Internal error")
			Expect(appErr.Entity).To(Equal("ComponentInstanceCcrns"), "should reference ComponentInstanceCcrns entity")
			Expect(appErr.ID).To(Equal(""), "should have empty ID for list operation")
			Expect(appErr.Op).To(Equal("componentInstanceHandler.ListCcrns"), "should include operation")
			Expect(appErr.Err.Error()).To(ContainSubstring("database connection failed"), "should contain original error message")
		})
	})
})
