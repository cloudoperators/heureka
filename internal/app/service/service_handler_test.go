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

var er event.EventRegistry
var authz openfga.Authorization

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
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
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		filter         *entity.ServiceFilter
		options        *entity.ListOptions
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getServiceFilter()
		cache := cache.NewNoCache()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Cache:    cache,
			Authz:    authz,
		}
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
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
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		service = test.NewFakeServiceEntity()
		first := 10
		after := ""
		filter = &entity.ServiceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		cache := cache.NewNoCache()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Cache:    cache,
			Authz:    authz,
		}
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
				s.OnServiceCreate(db, event)

				// Check AddIssueRepositoryToService was called
				db.AssertCalled(GinkgoT(), "AddIssueRepositoryToService", createEvent.Service.Id, repo.Id, int64(defaultPrio))
			})
		})

		Context("that as an invalid event", func() {
			It("should not perform any database operations", func() {
				invalidEvent := &s.UpdateServiceEvent{}

				// Use type assertion to convert
				var event event.Event = invalidEvent

				s.OnServiceCreate(db, event)

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

				s.OnServiceCreate(db, event)

				db.AssertNotCalled(GinkgoT(), "AddIssueRepositoryToService")
				// TODO: we could also check for the error message here
			})
		})
	})
})

var _ = Describe("When updating Service", Label("app", "UpdateService"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		service        entity.ServiceResult
		filter         *entity.ServiceFilter
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		service = test.NewFakeServiceResult()
		first := 10
		after := ""
		filter = &entity.ServiceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		cache := cache.NewNoCache()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Cache:    cache,
			Authz:    authz,
		}
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
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		id             int64
		filter         *entity.ServiceFilter
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		after := ""
		filter = &entity.ServiceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
		cache := cache.NewNoCache()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Cache:    cache,
			Authz:    authz,
		}
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
})

var _ = Describe("When modifying owner and Service", Label("app", "OwnerService"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		service        entity.ServiceResult
		owner          entity.User
		filter         *entity.ServiceFilter
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
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
		cache := cache.NewNoCache()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Cache:    cache,
			Authz:    authz,
		}
	})

	It("adds owner to service", func() {
		db.On("AddOwnerToService", service.Id, owner.Id).Return(nil)
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{service}, nil)
		serviceHandler = s.NewServiceHandler(handlerContext)
		service, err := serviceHandler.AddOwnerToService(service.Id, owner.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})

	It("removes owner from service", func() {
		db.On("RemoveOwnerFromService", service.Id, owner.Id).Return(nil)
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{service}, nil)
		serviceHandler = s.NewServiceHandler(handlerContext)
		service, err := serviceHandler.RemoveOwnerFromService(service.Id, owner.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})
})

var _ = Describe("When modifying relationship of issueRepository and Service", Label("app", "IssueRepositoryHandlerRelationship"), func() {
	var (
		db              *mocks.MockDatabase
		serviceHandler  s.ServiceHandler
		service         entity.ServiceResult
		issueRepository entity.IssueRepository
		filter          *entity.ServiceFilter
		priority        int64
		handlerContext  common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
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
		cache := cache.NewNoCache()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Cache:    cache,
			Authz:    authz,
		}
	})

	It("adds issueRepository to service", func() {
		db.On("AddIssueRepositoryToService", service.Id, issueRepository.Id, priority).Return(nil)
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{service}, nil)
		serviceHandler = s.NewServiceHandler(handlerContext)
		service, err := serviceHandler.AddIssueRepositoryToService(service.Id, issueRepository.Id, priority)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})

	It("removes issueRepository from service", func() {
		db.On("RemoveIssueRepositoryFromService", service.Id, issueRepository.Id).Return(nil)
		db.On("GetServices", filter, []entity.Order{}).Return([]entity.ServiceResult{service}, nil)
		serviceHandler = s.NewServiceHandler(handlerContext)
		service, err := serviceHandler.RemoveIssueRepositoryFromService(service.Id, issueRepository.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})
})

var _ = Describe("When listing serviceCcrns", Label("app", "ListServicesCcrns"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		filter         *entity.ServiceFilter
		options        *entity.ListOptions
		name           string
		handlerContext common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getServiceFilter()
		name = "f1"
		cache := cache.NewNoCache()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Cache:    cache,
			Authz:    authz,
		}
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
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		filter         *entity.ServiceFilter
		options        *entity.ListOptions
		domain         string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getServiceFilter()
		domain = "f1"
	})

	When("no filters are used", func() {

		BeforeEach(func() {
			db.On("GetServiceDomains", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			serviceHandler = s.NewServiceHandler(db, er, cache.NewNoCache())
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
			serviceHandler = s.NewServiceHandler(db, er, cache.NewNoCache())
			res, err := serviceHandler.ListServiceDomains(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(domain), "should only consist of domain")
		})
	})
})

var _ = Describe("When listing serviceRegions", Label("app", "ListServiceRegions"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		filter         *entity.ServiceFilter
		options        *entity.ListOptions
		region         string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getServiceFilter()
		region = "f1"
	})

	When("no filters are used", func() {

		BeforeEach(func() {
			db.On("GetServiceRegions", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			serviceHandler = s.NewServiceHandler(db, er, cache.NewNoCache())
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
			serviceHandler = s.NewServiceHandler(db, er, cache.NewNoCache())
			res, err := serviceHandler.ListServiceRegions(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(region), "should only consist of region")
		})
	})
})
