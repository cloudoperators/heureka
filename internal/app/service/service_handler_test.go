// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"errors"
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	s "github.com/cloudoperators/heureka/internal/app/service"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
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

func TestServiceHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Service Test Suite")
}

var handlerContext common.HandlerContext
var cfg *util.Config

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
		Cache:    cache.NewNoCache(),
		Authz:    authz,
	}
	handlerContext.Authz.RemoveAllRelations()
})

func getServiceFilter() *entity.ServiceFilter {
	sgName := "SomeNotExistingSupportGroup"
	return &entity.ServiceFilter{
		PaginatedX: entity.PaginatedX{
			First: nil,
			After: nil,
		},
		CCRN:             nil,
		Id:               nil,
		SupportGroupCCRN: []*string{&sgName},
	}
}

var _ = Describe("When listing Services", Label("app", "ListServices"), func() {
	var (
		er             event.EventRegistry
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		filter         *entity.ServiceFilter
		options        *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		options = entity.NewListOptions()
		filter = getServiceFilter()
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
			db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{}, nil)
			db.On("CountServices", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			serviceHandler = s.NewServiceHandler(handlerContext)
			res, err := serviceHandler.ListServices(filter, options)
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
			services := []entity.ServiceResult{}
			for _, s := range test.NNewFakeServiceEntities(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithService([]entity.Order{}, s, entity.IssueSeverityCounts{}))
				services = append(services, entity.ServiceResult{WithCursor: entity.WithCursor{Value: cursor}, Service: lo.ToPtr(s)})
			}

			var cursors = lo.Map(services, func(m entity.ServiceResult, _ int) string {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithService([]entity.Order{}, *m.Service, entity.IssueSeverityCounts{}))
				return cursor
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				service := test.NewFakeServiceEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithService([]entity.Order{}, service, entity.IssueSeverityCounts{}))
				cursors = append(cursors, c)
			}
			db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
			db.On("GetServices", filter, []entity.Order{}).Return(services, nil)
			db.On("GetAllServiceCursors", filter, []entity.Order{}).Return(cursors, nil)
			serviceHandler = s.NewServiceHandler(handlerContext)
			res, err := serviceHandler.ListServices(filter, options)
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
	When("the list options does include aggregations", func() {
		BeforeEach(func() {
			options.IncludeAggregations = true
		})
		Context("and the given filter does not have any matches in the database", func() {

			BeforeEach(func() {
				db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
				db.On("GetServicesWithAggregations", filter, []entity.Order{}).Return([]entity.ServiceResult{}, nil)
			})

			It("should return an empty result", func() {
				serviceHandler = s.NewServiceHandler(handlerContext)
				res, err := serviceHandler.ListServices(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")

			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				services := []entity.ServiceResult{}
				for _, s := range test.NNewFakeServiceEntitiesWithAggregations(10) {
					services = append(services, entity.ServiceResult{Service: &s.Service})
				}
				db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
				db.On("GetServicesWithAggregations", filter, []entity.Order{}).Return(services, nil)
			})
			It("should return the expected services in the result", func() {
				serviceHandler = s.NewServiceHandler(handlerContext)
				res, err := serviceHandler.ListServices(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(10), "return 10 results")
			})
		})
		Context("and the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
				db.On("GetServicesWithAggregations", filter, []entity.Order{}).Return([]entity.ServiceResult{}, errors.New("some error"))
			})

			It("should return the expected services in the result", func() {
				serviceHandler = s.NewServiceHandler(handlerContext)
				_, err := serviceHandler.ListServices(filter, options)
				Expect(err).Error()
				Expect(err.Error()).ToNot(BeEquivalentTo("some error"), "error gets not passed through")
			})
		})
	})
	When("the list options does NOT include aggregations", func() {

		BeforeEach(func() {
			options.IncludeAggregations = false
		})

		Context("and the given filter does not have any matches in the database", func() {

			BeforeEach(func() {
				db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
				db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{}, nil)
			})
			It("should return an empty result", func() {

				serviceHandler = s.NewServiceHandler(handlerContext)
				res, err := serviceHandler.ListServices(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")

			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				services := []entity.ServiceResult{}
				for _, s := range test.NNewFakeServiceEntitiesWithAggregations(15) {
					services = append(services, entity.ServiceResult{Service: &s.Service})
				}
				db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
				db.On("GetServices", filter, []entity.Order{}).Return(services, nil)
			})
			It("should return the expected services in the result", func() {
				serviceHandler = s.NewServiceHandler(handlerContext)
				res, err := serviceHandler.ListServices(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(15), "return 15 results")
			})
		})

		Context("and the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
				db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{}, errors.New("some error"))
			})

			It("should return the expected services in the result", func() {
				serviceHandler = s.NewServiceHandler(handlerContext)
				_, err := serviceHandler.ListServices(filter, options)
				Expect(err).Error()
				Expect(err.Error()).ToNot(BeEquivalentTo("some error"), "error gets not passed through")
			})
		})
	})
})

var _ = Describe("When creating Service", Label("app", "CreateService"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		service        entity.Service
		filter         *entity.ServiceFilter
		r              openfga.RelationInput
		er             event.EventRegistry
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		handlerContext.Authz.RemoveAllRelations()
		service = test.NewFakeServiceEntity()
		first := 10
		after := ""
		filter = &entity.ServiceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}

		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("creates service", func() {
		filter.CCRN = []*string{&service.CCRN}
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateService", &service).Return(&service, nil)
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{}, nil)

		serviceHandler = s.NewServiceHandler(handlerContext)
		newService, err := serviceHandler.CreateService(&service)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newService.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newService.CCRN).To(BeEquivalentTo(service.CCRN))
		})
	})

	Context("when handling a CreateServiceEvent", func() {
		BeforeEach(func() {
			db.On("GetDefaultIssuePriority").Return(int64(100))
			db.On("GetDefaultRepositoryName").Return("nvd")
		})
		Context("that is valid", func() {
			It("should add the default issue repository to the service", func() {
				srv := test.NewFakeServiceEntity()
				createEvent := &s.CreateServiceEvent{
					Service: &srv,
				}

				// Use type assertion to convert a CreateServiceEvent into an Event
				var event event.Event = createEvent

				// Create IssueRepository
				defaultRepoName := "nvd"
				defaultPrio := 100
				repo := test.NewFakeIssueRepositoryEntity()
				repo.Id = 456
				repo.Name = defaultRepoName

				db.On("GetIssueRepositories", &entity.IssueRepositoryFilter{
					Name: []*string{&defaultRepoName},
				}).Return([]entity.IssueRepository{repo}, nil)
				db.On("AddIssueRepositoryToService", createEvent.Service.Id, repo.Id, int64(defaultPrio)).Return(nil)

				// Simulate event
				s.OnServiceCreate(db, event, handlerContext.Authz)

				// Check AddIssueRepositoryToService was called
				db.AssertCalled(GinkgoT(), "AddIssueRepositoryToService", createEvent.Service.Id, repo.Id, int64(defaultPrio))
			})
		})

		Context("that as an invalid event", func() {
			It("should not perform any database operations", func() {
				invalidEvent := &s.UpdateServiceEvent{}

				// Use type assertion to convert
				var event event.Event = invalidEvent

				s.OnServiceCreate(db, event, handlerContext.Authz)

				// These functions should not be called in case of a different event
				db.AssertNotCalled(GinkgoT(), "GetIssueRepositories")
				db.AssertNotCalled(GinkgoT(), "AddIssueRepositoryToService")
			})

		})

		Context("when no issue repository is found", func() {
			It("should not add any repository to the service", func() {
				srv := test.NewFakeServiceEntity()
				createEvent := &s.CreateServiceEvent{
					Service: &srv,
				}

				// Use type assertion to convert a CreateServiceEvent into an Event
				var event event.Event = createEvent

				defaultRepoName := "nvd"
				db.On("GetIssueRepositories", &entity.IssueRepositoryFilter{
					Name: []*string{&defaultRepoName},
				}).Return([]entity.IssueRepository{}, nil)

				s.OnServiceCreate(db, event, handlerContext.Authz)

				db.AssertNotCalled(GinkgoT(), "AddIssueRepositoryToService")
				// TODO: we could also check for the error message here
			})
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
			authEnabled := false
			cfg = common.GetTestConfig(authEnabled)
			enableLogs := false
			handlerContext.Authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
		})

		Context("when handling a CreateServiceEvent authz", func() {
			Context("when new service is created", func() {
				It("should add user resource relationship tuple in openfga", func() {
					srv := test.NewFakeServiceEntity()
					createEvent := &s.CreateServiceEvent{
						Service: &srv,
					}

					r = openfga.RelationInput{
						UserType:   openfga.TypeRole,
						UserId:     "0",
						ObjectId:   openfga.IDService,
						ObjectType: openfga.TypeService,
						Relation:   openfga.RelRole,
					}

					// Use type assertion to convert a CreateServiceEvent into an Event
					var event event.Event = createEvent
					r.ObjectId = openfga.ObjectIdFromInt(createEvent.Service.Id)
					// Simulate event
					s.OnServiceCreateAuthz(db, event, handlerContext.Authz)

					ok, err := handlerContext.Authz.CheckPermission(r)
					Expect(err).To(BeNil(), "no error should be thrown")
					Expect(ok).To(BeTrue(), "permission should be granted")
				})
			})
		})
	})
})

var _ = Describe("When updating Service", Label("app", "UpdateService"), func() {
	var (
		er             event.EventRegistry
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		service        entity.ServiceResult
		filter         *entity.ServiceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		service = test.NewFakeServiceResult()
		first := 10
		after := ""
		filter = &entity.ServiceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("updates service", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateService", service.Service).Return(nil)
		serviceHandler = s.NewServiceHandler(handlerContext)
		service.CCRN = "SecretService"
		filter.Id = []*int64{&service.Id}
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{service}, nil)
		updatedService, err := serviceHandler.UpdateService(service.Service)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedService.CCRN).To(BeEquivalentTo(service.CCRN))
		})
	})
})

var _ = Describe("When deleting Service", Label("app", "DeleteService"), func() {
	var (
		er             event.EventRegistry
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		id             int64
		filter         *entity.ServiceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		id = 1
		first := 10
		after := ""
		filter = &entity.ServiceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("deletes service", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteService", id, mock.Anything).Return(nil)
		serviceHandler = s.NewServiceHandler(handlerContext)
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{}, nil)
		err := serviceHandler.DeleteService(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		lo := entity.NewListOptions()
		services, err := serviceHandler.ListServices(filter, lo)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(services.Elements).To(BeEmpty(), "no services should be found")
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

		Context("when handling a DeleteServiceEvent", func() {
			Context("when new service is deleted", func() {
				It("should delete tuples related to that service in openfga", func() {
					// Test OnServiceDeleteAuthz against all possible relations
					srv := test.NewFakeServiceEntity()
					deleteEvent := &s.DeleteServiceEvent{
						ServiceID: srv.Id,
					}
					objectId := openfga.ObjectIdFromInt(deleteEvent.ServiceID)
					userId := openfga.UserIdFromInt(deleteEvent.ServiceID)

					relations := []openfga.RelationInput{
						{ // user - service: a user can view the service
							UserType:   openfga.TypeUser,
							UserId:     openfga.IDUser,
							ObjectId:   objectId,
							ObjectType: openfga.TypeService,
							Relation:   openfga.RelCanView,
						},
						{ // role - service: a role is assigned to the service
							UserType:   openfga.TypeRole,
							UserId:     openfga.IDRole,
							ObjectId:   objectId,
							ObjectType: openfga.TypeService,
							Relation:   openfga.RelRole,
						},
						{ // support group - service: a support group is related to the service
							UserType:   openfga.TypeSupportGroup,
							UserId:     openfga.IDSupportGroup,
							ObjectId:   objectId,
							ObjectType: openfga.TypeService,
							Relation:   openfga.RelSupportGroup,
						},
						{ // service - component_instance: a service is related to a component instance
							UserType:   openfga.TypeService,
							UserId:     userId,
							ObjectId:   openfga.IDComponentInstance,
							ObjectType: openfga.RelComponentInstance,
							Relation:   openfga.RelRelatedService,
						},
					}

					handlerContext.Authz.AddRelationBulk(relations)

					// get the number of relations before deletion
					relCountBefore := 0
					for _, r := range relations {
						relations, err := handlerContext.Authz.ListRelations(r)
						Expect(err).To(BeNil(), "no error should be thrown")
						relCountBefore += len(relations)
					}
					Expect(relCountBefore).To(Equal(len(relations)), "all relations should exist before deletion")

					// check that relations were created
					for _, r := range relations {
						ok, err := handlerContext.Authz.CheckPermission(r)
						Expect(err).To(BeNil(), "no error should be thrown")
						Expect(ok).To(BeTrue(), "permission should be granted")
					}

					var event event.Event = deleteEvent
					// Simulate event
					s.OnServiceDeleteAuthz(db, event, handlerContext.Authz)

					// get the number of relations after deletion
					relCountAfter := 0
					for _, r := range relations {
						relations, err := handlerContext.Authz.ListRelations(r)
						Expect(err).To(BeNil(), "no error should be thrown")
						relCountAfter += len(relations)
					}
					Expect(relCountAfter < relCountBefore).To(BeTrue(), "less relations after deletion")
					Expect(relCountAfter).To(BeEquivalentTo(0), "no relations should exist after deletion")

					// verify that relations were deleted
					for _, r := range relations {
						ok, err := handlerContext.Authz.CheckPermission(r)
						Expect(err).To(BeNil(), "no error should be thrown")
						Expect(ok).To(BeFalse(), "permission should NOT be granted")
					}
				})
			})
		})
	})
})

var _ = Describe("When modifying owner and Service", Label("app", "OwnerService"), func() {
	var (
		db             *mocks.MockDatabase
		er             event.EventRegistry
		serviceHandler s.ServiceHandler
		service        entity.ServiceResult
		owner          entity.User
		filter         *entity.ServiceFilter
		r              openfga.RelationInput
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		service = test.NewFakeServiceResult()
		owner = test.NewFakeUserEntity()
		first := 10
		after := ""
		filter = &entity.ServiceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			Id: []*int64{&service.Id},
		}
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("adds owner to service", func() {
		db.On("AddOwnerToService", service.Id, owner.Id).Return(nil)
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{service}, nil)
		serviceHandler = s.NewServiceHandler(handlerContext)
		service, err := serviceHandler.AddOwnerToService(service.Id, owner.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})

	It("removes owner from service", func() {
		db.On("RemoveOwnerFromService", service.Id, owner.Id).Return(nil)
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{service}, nil)
		serviceHandler = s.NewServiceHandler(handlerContext)
		service, err := serviceHandler.RemoveOwnerFromService(service.Id, owner.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
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

		Context("when handling an AddOwnerToServiceEvent", func() {
			It("should add the owner-service relation tuple in openfga", func() {
				serviceFake := test.NewFakeServiceResult()
				ownerFake := test.NewFakeUserEntity()
				addEvent := &s.AddOwnerToServiceEvent{
					ServiceID: serviceFake.Id,
					OwnerID:   ownerFake.Id,
				}
				r = openfga.RelationInput{
					UserType:   openfga.TypeUser,
					UserId:     "",
					ObjectType: openfga.TypeService,
					ObjectId:   "",
					Relation:   openfga.RelOwner,
				}

				var event event.Event = addEvent
				s.OnAddOwnerToService(db, event, handlerContext.Authz)

				r.ObjectId = openfga.ObjectIdFromInt(addEvent.ServiceID)
				r.UserId = openfga.UserIdFromInt(addEvent.OwnerID)
				ok, err := handlerContext.Authz.CheckPermission(r)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(ok).To(BeTrue(), "permission should be granted")
			})
		})

		Context("when handling a RemoveOwnerFromServiceEvent", func() {
			It("should remove the owner-service relation tuple in openfga", func() {
				serviceFake := test.NewFakeServiceResult()
				ownerFake := test.NewFakeUserEntity()
				removeEvent := &s.RemoveOwnerFromServiceEvent{
					ServiceID: serviceFake.Id,
					OwnerID:   ownerFake.Id,
				}
				serviceId := openfga.ObjectIdFromInt(removeEvent.ServiceID)
				ownerId := openfga.UserIdFromInt(removeEvent.OwnerID)

				rel := openfga.RelationInput{
					UserType:   openfga.TypeUser,
					UserId:     ownerId,
					ObjectType: openfga.TypeService,
					ObjectId:   serviceId,
					Relation:   openfga.RelOwner,
				}

				handlerContext.Authz.AddRelationBulk([]openfga.RelationInput{rel})

				var event event.Event = removeEvent
				s.OnRemoveOwnerFromService(db, event, handlerContext.Authz)

				remaining, err := handlerContext.Authz.ListRelations(rel)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(remaining).To(BeEmpty(), "relation should not exist after removal")
			})
		})

	})
})

var _ = Describe("When modifying relationship of issueRepository and Service", Label("app", "IssueRepositoryHandlerRelationship"), func() {
	var (
		er              event.EventRegistry
		db              *mocks.MockDatabase
		serviceHandler  s.ServiceHandler
		service         entity.ServiceResult
		issueRepository entity.IssueRepository
		filter          *entity.ServiceFilter
		priority        int64
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		service = test.NewFakeServiceResult()
		issueRepository = test.NewFakeIssueRepositoryEntity()
		first := 10
		after := ""
		filter = &entity.ServiceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			Id: []*int64{&service.Id},
		}
		priority = 1
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	It("adds issueRepository to service", func() {
		db.On("AddIssueRepositoryToService", service.Id, issueRepository.Id, priority).Return(nil)
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{service}, nil)
		serviceHandler = s.NewServiceHandler(handlerContext)
		service, err := serviceHandler.AddIssueRepositoryToService(service.Id, issueRepository.Id, priority)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})

	It("removes issueRepository from service", func() {
		db.On("RemoveIssueRepositoryFromService", service.Id, issueRepository.Id).Return(nil)
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{service}, nil)
		serviceHandler = s.NewServiceHandler(handlerContext)
		service, err := serviceHandler.RemoveIssueRepositoryFromService(service.Id, issueRepository.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})
})

var _ = Describe("When listing serviceCcrns", Label("app", "ListServicesCcrns"), func() {
	var (
		er             event.EventRegistry
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		filter         *entity.ServiceFilter
		options        *entity.ListOptions
		name           string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		options = entity.NewListOptions()
		filter = getServiceFilter()
		name = "f1"
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("no filters are used", func() {

		BeforeEach(func() {
			db.On("GetServiceCcrns", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			serviceHandler = s.NewServiceHandler(handlerContext)
			res, err := serviceHandler.ListServiceCcrns(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific serviceCcrns filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.ServiceFilter{
				CCRN: []*string{&name},
			}

			db.On("GetServiceCcrns", filter).Return([]string{name}, nil)
		})
		It("returns filtered services according to the service type", func() {
			serviceHandler = s.NewServiceHandler(handlerContext)
			res, err := serviceHandler.ListServiceCcrns(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(name), "should only consist of serviceCcrn")
		})
	})
})

var _ = Describe("When listing serviceDomains", Label("app", "ListServicesDomains"), func() {
	var (
		er             event.EventRegistry
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		filter         *entity.ServiceFilter
		options        *entity.ListOptions
		domain         string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		options = entity.NewListOptions()
		filter = getServiceFilter()
		domain = "f1"
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("no filters are used", func() {

		BeforeEach(func() {
			db.On("GetServiceDomains", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			serviceHandler = s.NewServiceHandler(handlerContext)
			res, err := serviceHandler.ListServiceDomains(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific serviceDomains filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.ServiceFilter{
				Domain: []*string{&domain},
			}

			db.On("GetServiceDomains", filter).Return([]string{domain}, nil)
		})
		It("returns filtered services according to the service type", func() {
			serviceHandler = s.NewServiceHandler(handlerContext)
			res, err := serviceHandler.ListServiceDomains(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(domain), "should only consist of domain")
		})
	})
})

var _ = Describe("When listing serviceRegions", Label("app", "ListServiceRegions"), func() {
	var (
		er             event.EventRegistry
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		filter         *entity.ServiceFilter
		options        *entity.ListOptions
		region         string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		er = event.NewEventRegistry(db, handlerContext.Authz)
		options = entity.NewListOptions()
		filter = getServiceFilter()
		region = "f1"
		handlerContext.DB = db
		handlerContext.EventReg = er
	})

	When("no filters are used", func() {

		BeforeEach(func() {
			db.On("GetServiceRegions", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			serviceHandler = s.NewServiceHandler(handlerContext)
			res, err := serviceHandler.ListServiceRegions(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific serviceRegions filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.ServiceFilter{
				Region: []*string{&region},
			}

			db.On("GetServiceRegions", filter).Return([]string{region}, nil)
		})
		It("returns filtered services according to the service type", func() {
			serviceHandler = s.NewServiceHandler(handlerContext)
			res, err := serviceHandler.ListServiceRegions(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(region), "should only consist of region")
		})
	})
})
