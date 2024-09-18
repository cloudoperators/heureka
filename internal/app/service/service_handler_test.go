// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/event"
	s "github.com/cloudoperators/heureka/internal/app/service"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

func TestServiceHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

func GetListOptions() *entity.ListOptions {
	return &entity.ListOptions{
		ShowTotalCount:      false,
		ShowPageInfo:        false,
		IncludeAggregations: false,
	}
}

func getServiceFilter() *entity.ServiceFilter {
	sgName := "SomeNotExistingSupportGroup"
	return &entity.ServiceFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		Name:             nil,
		Id:               nil,
		SupportGroupName: []*string{&sgName},
	}
}

var _ = Describe("When listing Services", Label("app", "ListServices"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		filter         *entity.ServiceFilter
		options        *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = GetListOptions()
		filter = getServiceFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetServices", filter).Return([]entity.Service{}, nil)
			db.On("CountServices", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			serviceHandler = s.NewServiceHandler(db, er)
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
			services := test.NNewFakeServiceEntities(resElements)

			var ids = lo.Map(services, func(s entity.Service, _ int) int64 { return s.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetServices", filter).Return(services, nil)
			db.On("GetAllServiceIds", filter).Return(ids, nil)
			serviceHandler = s.NewServiceHandler(db, er)
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
})

var _ = Describe("When creating Service", Label("app", "CreateService"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		service        entity.Service
		filter         *entity.ServiceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		service = test.NewFakeServiceEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.ServiceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("creates service", func() {
		filter.Name = []*string{&service.Name}
		db.On("CreateService", &service).Return(&service, nil)
		db.On("GetServices", filter).Return([]entity.Service{}, nil)
		serviceHandler = s.NewServiceHandler(db, er)
		newService, err := serviceHandler.CreateService(&service)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newService.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newService.Name).To(BeEquivalentTo(service.Name))
		})
	})
})

var _ = Describe("When updating Service", Label("app", "UpdateService"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		service        entity.Service
		filter         *entity.ServiceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		service = test.NewFakeServiceEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.ServiceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates service", func() {
		db.On("UpdateService", &service).Return(nil)
		serviceHandler = s.NewServiceHandler(db, er)
		service.Name = "SecretService"
		filter.Id = []*int64{&service.Id}
		db.On("GetServices", filter).Return([]entity.Service{service}, nil)
		updatedService, err := serviceHandler.UpdateService(&service)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedService.Name).To(BeEquivalentTo(service.Name))
		})
	})
})

var _ = Describe("When deleting Service", Label("app", "DeleteService"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		id             int64
		filter         *entity.ServiceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.ServiceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes service", func() {
		db.On("DeleteService", id).Return(nil)
		serviceHandler = s.NewServiceHandler(db, er)
		db.On("GetServices", filter).Return([]entity.Service{}, nil)
		err := serviceHandler.DeleteService(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		services, err := serviceHandler.ListServices(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(services.Elements).To(BeEmpty(), "no services should be found")
	})
})

var _ = Describe("When modifying owner and Service", Label("app", "OwnerService"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		service        entity.Service
		owner          entity.User
		filter         *entity.ServiceFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		service = test.NewFakeServiceEntity()
		owner = test.NewFakeUserEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.ServiceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id: []*int64{&service.Id},
		}
	})

	It("adds owner to service", func() {
		db.On("AddOwnerToService", service.Id, owner.Id).Return(nil)
		db.On("GetServices", filter).Return([]entity.Service{service}, nil)
		serviceHandler = s.NewServiceHandler(db, er)
		service, err := serviceHandler.AddOwnerToService(service.Id, owner.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})

	It("removes owner from service", func() {
		db.On("RemoveOwnerFromService", service.Id, owner.Id).Return(nil)
		db.On("GetServices", filter).Return([]entity.Service{service}, nil)
		serviceHandler = s.NewServiceHandler(db, er)
		service, err := serviceHandler.RemoveOwnerFromService(service.Id, owner.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})
})

var _ = Describe("When modifying relationship of issueRepository and Service", Label("app", "IssueRepositoryHandlerRelationship"), func() {
	var (
		db              *mocks.MockDatabase
		serviceHandler  s.ServiceHandler
		service         entity.Service
		issueRepository entity.IssueRepository
		filter          *entity.ServiceFilter
		priority        int64
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		service = test.NewFakeServiceEntity()
		issueRepository = test.NewFakeIssueRepositoryEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.ServiceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id: []*int64{&service.Id},
		}
		priority = 1
	})

	It("adds issueRepository to service", func() {
		db.On("AddIssueRepositoryToService", service.Id, issueRepository.Id, priority).Return(nil)
		db.On("GetServices", filter).Return([]entity.Service{service}, nil)
		serviceHandler = s.NewServiceHandler(db, er)
		service, err := serviceHandler.AddIssueRepositoryToService(service.Id, issueRepository.Id, priority)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})

	It("removes issueRepository from service", func() {
		db.On("RemoveIssueRepositoryFromService", service.Id, issueRepository.Id).Return(nil)
		db.On("GetServices", filter).Return([]entity.Service{service}, nil)
		serviceHandler = s.NewServiceHandler(db, er)
		service, err := serviceHandler.RemoveIssueRepositoryFromService(service.Id, issueRepository.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(service).NotTo(BeNil(), "service should be returned")
	})
})

var _ = Describe("When listing serviceNames", Label("app", "ListServicesNames"), func() {
	var (
		db             *mocks.MockDatabase
		serviceHandler s.ServiceHandler
		filter         *entity.ServiceFilter
		options        *entity.ListOptions
		name           string
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = GetListOptions()
		filter = getServiceFilter()
		name = "f1"
	})

	When("no filters are used", func() {

		BeforeEach(func() {
			db.On("GetServiceNames", filter).Return([]string{}, nil)
		})

		It("it return the results", func() {
			serviceHandler = s.NewServiceHandler(db, er)
			res, err := serviceHandler.ListServiceNames(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(BeEmpty(), "return correct result")
		})
	})
	When("specific serviceNames filter is applied", func() {
		BeforeEach(func() {
			filter = &entity.ServiceFilter{
				Name: []*string{&name},
			}

			db.On("GetServiceNames", filter).Return([]string{name}, nil)
		})
		It("returns filtered services according to the service type", func() {
			serviceHandler = s.NewServiceHandler(db, er)
			res, err := serviceHandler.ListServiceNames(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(res).Should(ConsistOf(name), "should only consist of serviceName")
		})
	})
})
