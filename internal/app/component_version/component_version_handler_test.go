// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version_test

import (
	"math"
	"strconv"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	cv "github.com/cloudoperators/heureka/internal/app/component_version"
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

func TestComponentVersionHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Component Version Service Test Suite")
}

var handlerContext common.HandlerContext
var cfg *util.Config

var _ = BeforeSuite(func() {
	cfg = common.GetTestConfig()
	enableLogs := false
	authz := openfga.NewAuthorizationHandler(cfg, enableLogs)
	handlerContext = common.HandlerContext{
		Cache: cache.NewNoCache(),
		Authz: authz,
	}
})

func getComponentVersionFilter() *entity.ComponentVersionFilter {
	return &entity.ComponentVersionFilter{
		PaginatedX: entity.PaginatedX{
			First: nil,
			After: nil,
		},
	}
}

var _ = Describe("When listing ComponentVersions", Label("app", "ListComponentVersions"), func() {
	var (
		er        event.EventRegistry
		db        *mocks.MockDatabase
		cvHandler cv.ComponentVersionHandler
		filter    *entity.ComponentVersionFilter
		options   *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)

		options = entity.NewListOptions()
		filter = getComponentVersionFilter()
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetComponentVersions", filter, []entity.Order{}).Return([]entity.ComponentVersionResult{}, nil)
			db.On("CountComponentVersions", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			cvHandler = cv.NewComponentVersionHandler(handlerContext)
			res, err := cvHandler.ListComponentVersions(filter, options)
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
			componentVersions := []entity.ComponentVersionResult{}
			for _, cv := range test.NNewFakeComponentVersionEntities(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithComponentVersion([]entity.Order{}, cv, entity.IssueSeverityCounts{}))
				componentVersions = append(componentVersions, entity.ComponentVersionResult{WithCursor: entity.WithCursor{Value: cursor}, ComponentVersion: lo.ToPtr(cv)})
			}

			var cursors = lo.Map(componentVersions, func(m entity.ComponentVersionResult, _ int) string {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithComponentVersion([]entity.Order{}, *m.ComponentVersion, entity.IssueSeverityCounts{}))
				return cursor
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				componentVersion := test.NewFakeComponentVersionEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithComponentVersion([]entity.Order{}, componentVersion, entity.IssueSeverityCounts{}))
				cursors = append(cursors, c)
			}
			db.On("GetComponentVersions", filter, []entity.Order{}).Return(componentVersions, nil)
			db.On("GetAllComponentVersionCursors", filter, []entity.Order{}).Return(cursors, nil)
			cvHandler = cv.NewComponentVersionHandler(handlerContext)
			res, err := cvHandler.ListComponentVersions(filter, options)
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
	When("filtering by tag", func() {
		It("filters results correctly", func() {
			// Create test data with a specific tag
			testTag := "test-filter-tag"
			componentVersions := test.NNewFakeComponentVersionResults(3)
			for i := range componentVersions {
				componentVersions[i].Tag = testTag
			}

			// Set up the filter
			tagFilter := getComponentVersionFilter()
			tagFilter.Tag = []*string{&testTag}

			// Mock database calls
			db.On("GetComponentVersions", tagFilter, []entity.Order{}).Return(componentVersions, nil)
			if options.ShowTotalCount {
				db.On("CountComponentVersions", tagFilter).Return(int64(len(componentVersions)), nil)
			}

			// Execute the handler
			cvHandler = cv.NewComponentVersionHandler(handlerContext)
			result, err := cvHandler.ListComponentVersions(tagFilter, options)

			// Verify results
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(len(result.Elements)).To(Equal(len(componentVersions)))

			// Verify all results have the correct tag
			for _, element := range result.Elements {
				Expect(element.ComponentVersion.Tag).To(Equal(testTag))
			}
		})
	})
	When("filtering by repository", func() {
		It("filters results correctly", func() {
			// Create test data with a specific repository
			testRepo := "test-filter-repo"
			componentVersions := test.NNewFakeComponentVersionResults(3)
			for i := range componentVersions {
				componentVersions[i].Repository = testRepo
			}

			// Set up the filter
			repoFilter := getComponentVersionFilter()
			repoFilter.Repository = []*string{&testRepo}

			// Mock database calls
			db.On("GetComponentVersions", repoFilter, []entity.Order{}).Return(componentVersions, nil)
			if options.ShowTotalCount {
				db.On("CountComponentVersions", repoFilter).Return(int64(len(componentVersions)), nil)
			}

			// Execute the handler
			cvHandler = cv.NewComponentVersionHandler(handlerContext)
			result, err := cvHandler.ListComponentVersions(repoFilter, options)

			// Verify results
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(len(result.Elements)).To(Equal(len(componentVersions)))

			// Verify all results have the correct repository
			for _, element := range result.Elements {
				Expect(element.ComponentVersion.Repository).To(Equal(testRepo))
			}
		})
	})
	When("filtering by organization", func() {
		It("filters results correctly", func() {
			// Create test data with a specific organization
			testOrg := "test-filter-org"
			componentVersions := test.NNewFakeComponentVersionResults(3)
			for i := range componentVersions {
				componentVersions[i].Organization = testOrg
			}

			// Set up the filter
			orgFilter := getComponentVersionFilter()
			orgFilter.Organization = []*string{&testOrg}

			// Mock database calls
			db.On("GetComponentVersions", orgFilter, []entity.Order{}).Return(componentVersions, nil)
			if options.ShowTotalCount {
				db.On("CountComponentVersions", orgFilter).Return(int64(len(componentVersions)), nil)
			}

			// Execute the handler
			cvHandler = cv.NewComponentVersionHandler(handlerContext)
			result, err := cvHandler.ListComponentVersions(orgFilter, options)

			// Verify results
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(len(result.Elements)).To(Equal(len(componentVersions)))

			// Verify all results have the correct organization
			for _, element := range result.Elements {
				Expect(element.ComponentVersion.Organization).To(Equal(testOrg))
			}
		})
	})
})

var _ = Describe("When creating ComponentVersion", Label("app", "CreateComponentVersion"), func() {
	var (
		er                     event.EventRegistry
		db                     *mocks.MockDatabase
		componenVersionService cv.ComponentVersionHandler
		componentVersion       entity.ComponentVersion
		p                      openfga.PermissionInput
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		componentVersion = test.NewFakeComponentVersionEntity()

		p = openfga.PermissionInput{
			UserType:   openfga.TypeRole,
			UserId:     "0",
			ObjectId:   "",
			ObjectType: openfga.TypeComponentVersion,
			Relation:   openfga.TypeRole,
		}

		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("creates componentVersion", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateComponentVersion", &componentVersion).Return(&componentVersion, nil)
		componenVersionService = cv.NewComponentVersionHandler(handlerContext)
		newComponentVersion, err := componenVersionService.CreateComponentVersion(&componentVersion)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newComponentVersion.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newComponentVersion.Version).To(BeEquivalentTo(componentVersion.Version))
			Expect(newComponentVersion.ComponentId).To(BeEquivalentTo(componentVersion.ComponentId))
			Expect(newComponentVersion.Tag).To(BeEquivalentTo(componentVersion.Tag))
		})
	})

	Context("when handling a CreateComponentInstanceEvent", func() {
		Context("when new component instance is created", func() {
			It("should add user resource relationship tuple in openfga", func() {
				cvFake := test.NewFakeComponentVersionEntity()
				createEvent := &cv.CreateComponentVersionEvent{
					ComponentVersion: &cvFake,
				}

				// Use type assertion to convert a CreateServiceEvent into an Event
				var event event.Event = createEvent
				resourceId := strconv.FormatInt(createEvent.ComponentVersion.Id, 10)
				p.ObjectId = openfga.ObjectId(resourceId)
				// Simulate event
				cv.OnComponentVersionCreateAuthz(db, event, handlerContext.Authz)

				ok, err := handlerContext.Authz.CheckPermission(p)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(ok).To(BeTrue(), "permission should be granted")
			})
		})
	})
})

var _ = Describe("When updating ComponentVersion", Label("app", "UpdateComponentVersion"), func() {
	var (
		er                     event.EventRegistry
		db                     *mocks.MockDatabase
		componenVersionService cv.ComponentVersionHandler
		componentVersion       entity.ComponentVersionResult
		filter                 *entity.ComponentVersionFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		componentVersion = test.NewFakeComponentVersionResult()
		first := 10
		after := ""
		filter = &entity.ComponentVersionFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("updates componentVersion", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateComponentVersion", componentVersion.ComponentVersion).Return(nil)
		componenVersionService = cv.NewComponentVersionHandler(handlerContext)
		componentVersion.Version = "7.3.3.1"
		componentVersion.Tag = "updated-tag"
		filter.Id = []*int64{&componentVersion.Id}
		db.On("GetComponentVersions", filter, []entity.Order{}).Return([]entity.ComponentVersionResult{componentVersion}, nil)
		updatedComponentVersion, err := componenVersionService.UpdateComponentVersion(componentVersion.ComponentVersion)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedComponentVersion.Version).To(BeEquivalentTo(componentVersion.Version))
			Expect(updatedComponentVersion.ComponentId).To(BeEquivalentTo(componentVersion.ComponentId))
			Expect(updatedComponentVersion.Tag).To(BeEquivalentTo(componentVersion.Tag))
		})
	})

	Context("when handling an UpdateComponentVersionEvent", func() {
		It("should update the component relation tuple in openfga", func() {
			cvFake := test.NewFakeComponentVersionEntity()
			oldComponentId := int64(12345)
			newComponentId := int64(67890)

			// Add an initial relation: component_version -> old component
			initialRelation := openfga.RelationInput{
				UserType:   "component_version",
				UserId:     openfga.UserIdFromInt(cvFake.Id),
				Relation:   "component_version",
				ObjectType: "component",
				ObjectId:   openfga.ObjectIdFromInt(oldComponentId),
			}
			// Bulk add instead of single add
			handlerContext.Authz.AddRelationBulk([]openfga.RelationInput{initialRelation})

			// Prepare the update event with the new component id
			cvFake.ComponentId = newComponentId
			updateEvent := &cv.UpdateComponentVersionEvent{
				ComponentVersion: &cvFake,
			}
			var event event.Event = updateEvent

			// Simulate event
			cv.OnComponentVersionUpdateAuthz(db, event, handlerContext.Authz)

			// Check that the old relation is gone
			remainingOld, err := handlerContext.Authz.ListRelations(initialRelation)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remainingOld).To(BeEmpty(), "old relation should be removed")

			// Check that the new relation exists
			newRelation := openfga.RelationInput{
				UserType:   openfga.TypeComponentVersion,
				UserId:     openfga.UserIdFromInt(cvFake.Id),
				Relation:   openfga.RelComponentVersion,
				ObjectType: openfga.TypeComponent,
				ObjectId:   openfga.ObjectIdFromInt(newComponentId),
			}
			remainingNew, err := handlerContext.Authz.ListRelations(newRelation)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(remainingNew).NotTo(BeEmpty(), "new relation should exist")
		})
	})
})

var _ = Describe("When deleting ComponentVersion", Label("app", "DeleteComponentVersion"), func() {
	var (
		er                     event.EventRegistry
		db                     *mocks.MockDatabase
		componenVersionService cv.ComponentVersionHandler
		id                     int64
		filter                 *entity.ComponentVersionFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		id = 1
		first := 10
		after := ""
		filter = &entity.ComponentVersionFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("deletes componentVersion", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteComponentVersion", id, mock.Anything).Return(nil)
		componenVersionService = cv.NewComponentVersionHandler(handlerContext)
		db.On("GetComponentVersions", filter, []entity.Order{}).Return([]entity.ComponentVersionResult{}, nil)
		err := componenVersionService.DeleteComponentVersion(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		lo := entity.NewListOptions()
		componentVersions, err := componenVersionService.ListComponentVersions(filter, lo)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(componentVersions.Elements).To(BeEmpty(), "no error should be thrown")
	})

	Context("when handling a DeleteComponentVersionEvent", func() {
		Context("when new component version is deleted", func() {
			It("should delete tuples related to that component version in openfga", func() {
				// Test OnComponentVersionDeleteAuthz against all possible relations
				cvFake := test.NewFakeComponentVersionEntity()
				deleteEvent := &cv.DeleteComponentVersionEvent{
					ComponentVersionID: cvFake.Id,
				}
				objectId := openfga.ObjectIdFromInt(deleteEvent.ComponentVersionID)
				userId := openfga.UserIdFromInt(deleteEvent.ComponentVersionID)
				relations := []openfga.RelationInput{
					{ // user - component_version: a user can view the component version
						UserType:   openfga.TypeUser,
						UserId:     openfga.IDUser,
						ObjectId:   objectId,
						ObjectType: openfga.TypeComponentVersion,
						Relation:   openfga.RelCanView,
					},
					{ // component_instance - component_version: a component instance is related to the component version
						UserType:   openfga.TypeComponentInstance,
						UserId:     openfga.IDComponentInstance,
						ObjectId:   objectId,
						ObjectType: openfga.TypeComponentVersion,
						Relation:   openfga.RelComponentInstance,
					},
					{ // role - component_version: a role is assigned to the component version
						UserType:   openfga.TypeRole,
						UserId:     openfga.IDRole,
						ObjectId:   objectId,
						ObjectType: openfga.TypeComponentVersion,
						Relation:   openfga.RelRole,
					},
					{ // component_version - component: a component version is related to a component
						UserType:   openfga.TypeComponentVersion,
						UserId:     userId,
						ObjectId:   openfga.IDComponent,
						ObjectType: openfga.TypeComponent,
						Relation:   openfga.RelComponentVersion,
					},
				}

				handlerContext.Authz.AddRelationBulk(relations)

				// get the number of relations before deletion
				relCountBefore := 0
				for _, r := range relations {
					relations, err := handlerContext.Authz.ListRelations(r)
					if err != nil {
						Expect(err).To(BeNil(), "no error should be thrown")
					}
					relCountBefore += len(relations)
				}
				Expect(relCountBefore).To(Equal(len(relations)), "all relations should exist before deletion")

				// check that relations were created
				for _, r := range relations {
					ok, err := handlerContext.Authz.CheckPermission(openfga.PermissionInput{
						UserType:   r.UserType,
						UserId:     r.UserId,
						ObjectType: r.ObjectType,
						ObjectId:   r.ObjectId,
						Relation:   r.Relation,
					})
					Expect(err).To(BeNil(), "no error should be thrown")
					Expect(ok).To(BeTrue(), "permission should be granted")
				}

				var event event.Event = deleteEvent
				// Simulate event
				cv.OnComponentVersionDeleteAuthz(db, event, handlerContext.Authz)

				// get the number of relations after deletion
				relCountAfter := 0
				for _, r := range relations {
					relations, err := handlerContext.Authz.ListRelations(r)
					if err != nil {
						Expect(err).To(BeNil(), "no error should be thrown")
					}
					relCountAfter += len(relations)
				}
				Expect(relCountAfter < relCountBefore).To(BeTrue(), "less relations after deletion")
				Expect(relCountAfter).To(BeEquivalentTo(0), "no relations should exist after deletion")

				// verify that relations were deleted
				for _, r := range relations {
					ok, err := handlerContext.Authz.CheckPermission(openfga.PermissionInput{
						UserType:   r.UserType,
						UserId:     r.UserId,
						ObjectType: r.ObjectType,
						ObjectId:   r.ObjectId,
						Relation:   r.Relation,
					})
					Expect(err).To(BeNil(), "no error should be thrown")
					Expect(ok).To(BeFalse(), "permission should NOT be granted")
				}
			})
		})
	})
})
