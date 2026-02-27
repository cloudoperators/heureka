// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_repository_test

import (
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	ir "github.com/cloudoperators/heureka/internal/app/issue_repository"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	"github.com/cloudoperators/heureka/internal/openfga"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	mock "github.com/stretchr/testify/mock"
)

func TestIssueRepositoryHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test IssueRepository Service")
}

var (
	er    event.EventRegistry
	authz openfga.Authorization
)

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

func getIssueRepositoryFilter() *entity.IssueRepositoryFilter {
	sName := "SomeNotExistingService"
	return &entity.IssueRepositoryFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		Name:        nil,
		Id:          nil,
		ServiceId:   nil,
		ServiceCCRN: []*string{&sName},
	}
}

var _ = Describe("When listing IssueRepositories", Label("app", "ListIssueRepositories"), func() {
	var (
		db                     *mocks.MockDatabase
		issueRepositoryHandler ir.IssueRepositoryHandler
		filter                 *entity.IssueRepositoryFilter
		options                *entity.ListOptions
		handlerContext         common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getIssueRepositoryFilter()
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	When("the list option does include the totalCount", func() {
		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetIssueRepositories", filter, mock.Anything).Return([]entity.IssueRepositoryResult{}, nil)
			db.On("CountIssueRepositories", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			issueRepositoryHandler = ir.NewIssueRepositoryHandler(handlerContext)
			res, err := issueRepositoryHandler.ListIssueRepositories(filter, options)
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
			irResults := []entity.IssueRepositoryResult{}

			for _, ir := range test.NNewFakeIssueRepositories(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithIssueRepository([]entity.Order{}, ir))
				irResults = append(irResults, entity.IssueRepositoryResult{WithCursor: entity.WithCursor{Value: cursor}, IssueRepository: lo.ToPtr(ir)})
			}

			cursors := lo.Map(irResults, func(m entity.IssueRepositoryResult, _ int) string {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithIssueRepository([]entity.Order{}, *m.IssueRepository))
				return cursor
			})

			for i := 0; len(cursors) < dbElements; i++ {
				ir := test.NewFakeIssueRepositoryEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithIssueRepository([]entity.Order{}, ir))
				cursors = append(cursors, c)
			}

			db.On("GetIssueRepositories", filter, mock.Anything).Return(irResults, nil)
			db.On("GetAllIssueRepositoryCursors", filter, mock.Anything).Return(cursors, nil)
			issueRepositoryHandler = ir.NewIssueRepositoryHandler(handlerContext)
			res, err := issueRepositoryHandler.ListIssueRepositories(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.PageInfo.HasNextPage).To(BeEquivalentTo(hasNextPage), "correct hasNextPage indicator")
			Expect(len(res.Elements)).To(BeEquivalentTo(resElements))
			Expect(len(res.PageInfo.Pages)).To(BeEquivalentTo(int(math.Ceil(float64(dbElements)/float64(pageSize)))), "correct  number of pages")
		},
			Entry("When pageSize is 1 and the database was returning 2 elements", 1, 2, 1, true),
			Entry("When pageSize is 10 and the database was returning 9 elements", 10, 9, 9, false),
			Entry("When pageSize is 10 and the database was returning 11 elements", 10, 11, 10, true),
		)
	})
})

var _ = Describe("When creating IssueRepository", Label("app", "CreateIssueRepository"), func() {
	var (
		db                     *mocks.MockDatabase
		issueRepositoryHandler ir.IssueRepositoryHandler
		issueRepository        entity.IssueRepository
		filter                 *entity.IssueRepositoryFilter
		event                  *ir.CreateIssueRepositoryEvent
		handlerContext         common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueRepository = test.NewFakeIssueRepositoryEntity()
		first := 10
		var after string
		filter = &entity.IssueRepositoryFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
		issueRepository = test.NewFakeIssueRepositoryEntity()
		event = &ir.CreateIssueRepositoryEvent{
			IssueRepository: &issueRepository,
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	It("creates issueRepository", func() {
		filter.Name = []*string{&issueRepository.Name}
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateIssueRepository", &issueRepository).Return(&issueRepository, nil)
		db.On("GetIssueRepositories", filter, mock.Anything).Return([]entity.IssueRepositoryResult{}, nil)
		issueRepositoryHandler = ir.NewIssueRepositoryHandler(handlerContext)
		newIssueRepository, err := issueRepositoryHandler.CreateIssueRepository(common.NewAdminContext(), &issueRepository)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newIssueRepository.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newIssueRepository.Name).To(BeEquivalentTo(issueRepository.Name))
		})
	})

	Context("when services are found", func() {
		BeforeEach(func() {
			service1 := test.NewFakeServiceResult()
			service2 := test.NewFakeServiceResult()
			service1.Id = int64(1)
			service2.Id = int64(2)

			issueRepository.Id = int64(1)

			services := []entity.ServiceResult{service1, service2}
			db.On("GetServices", &entity.ServiceFilter{}, []entity.Order{}).Return(services, nil)
			db.On("AddIssueRepositoryToService", int64(1), int64(1), int64(100)).Return(nil)
			db.On("AddIssueRepositoryToService", int64(2), int64(1), int64(100)).Return(nil)
			db.On("GetDefaultIssuePriority").Return(int64(100))
		})

		It("adds the issue repository to all services", func() {
			ir.OnIssueRepositoryCreate(db, event)

			db.AssertCalled(GinkgoT(), "AddIssueRepositoryToService", int64(1), int64(1), int64(100))
			db.AssertCalled(GinkgoT(), "AddIssueRepositoryToService", int64(2), int64(1), int64(100))
		})
	})
})

var _ = Describe("When updating IssueRepository", Label("app", "UpdateIssueRepository"), func() {
	var (
		db                     *mocks.MockDatabase
		issueRepositoryHandler ir.IssueRepositoryHandler
		issueRepository        entity.IssueRepository
		filter                 *entity.IssueRepositoryFilter
		handlerContext         common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueRepository = test.NewFakeIssueRepositoryEntity()
		first := 10
		var after string
		filter = &entity.IssueRepositoryFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	It("updates issueRepository", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateIssueRepository", &issueRepository).Return(nil)
		issueRepositoryHandler = ir.NewIssueRepositoryHandler(handlerContext)
		issueRepository.Name = "SecretRepository"
		filter.Id = []*int64{&issueRepository.Id}
		db.On("GetIssueRepositories", filter, mock.Anything).Return([]entity.IssueRepositoryResult{{
			IssueRepository: &issueRepository,
		}}, nil)
		updatedIssueRepository, err := issueRepositoryHandler.UpdateIssueRepository(common.NewAdminContext(), &issueRepository)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedIssueRepository.Name).To(BeEquivalentTo(issueRepository.Name))
		})
	})
})

var _ = Describe("When deleting IssueRepository", Label("app", "DeleteIssueRepository"), func() {
	var (
		db                     *mocks.MockDatabase
		issueRepositoryHandler ir.IssueRepositoryHandler
		id                     int64
		filter                 *entity.IssueRepositoryFilter
		handlerContext         common.HandlerContext
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after string
		filter = &entity.IssueRepositoryFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
		handlerContext = common.HandlerContext{
			DB:       db,
			EventReg: er,
			Authz:    authz,
		}
	})

	It("deletes issueRepository", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteIssueRepository", id, mock.Anything).Return(nil)
		issueRepositoryHandler = ir.NewIssueRepositoryHandler(handlerContext)
		db.On("GetIssueRepositories", filter, mock.Anything).Return([]entity.IssueRepositoryResult{}, nil)
		err := issueRepositoryHandler.DeleteIssueRepository(common.NewAdminContext(), id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		issueRepositories, err := issueRepositoryHandler.ListIssueRepositories(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issueRepositories.Elements).To(BeEmpty(), "no error should be thrown")
	})
})
