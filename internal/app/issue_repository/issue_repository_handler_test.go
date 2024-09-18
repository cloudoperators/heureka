// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_repository_test

import (
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/event"
	ir "github.com/cloudoperators/heureka/internal/app/issue_repository"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

func TestIssueRepositoryHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test IssueRepository Service")
}

var er event.EventRegistry

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
		ServiceName: []*string{&sName},
	}
}

var _ = Describe("When listing IssueRepositories", Label("app", "ListIssueRepositories"), func() {
	var (
		db                     *mocks.MockDatabase
		issueRepositoryHandler ir.IssueRepositoryHandler
		filter                 *entity.IssueRepositoryFilter
		options                *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = entity.NewListOptions()
		filter = getIssueRepositoryFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetIssueRepositories", filter).Return([]entity.IssueRepository{}, nil)
			db.On("CountIssueRepositories", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			issueRepositoryHandler = ir.NewIssueRepositoryHandler(db, er)
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
			repositories := test.NNewFakeIssueRepositories(resElements)

			var ids = lo.Map(repositories, func(ar entity.IssueRepository, _ int) int64 { return ar.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetIssueRepositories", filter).Return(repositories, nil)
			db.On("GetAllIssueRepositoryIds", filter).Return(ids, nil)
			issueRepositoryHandler = ir.NewIssueRepositoryHandler(db, er)
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
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueRepository = test.NewFakeIssueRepositoryEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueRepositoryFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("creates issueRepository", func() {
		filter.Name = []*string{&issueRepository.Name}
		db.On("CreateIssueRepository", &issueRepository).Return(&issueRepository, nil)
		db.On("GetIssueRepositories", filter).Return([]entity.IssueRepository{}, nil)
		issueRepositoryHandler = ir.NewIssueRepositoryHandler(db, er)
		newIssueRepository, err := issueRepositoryHandler.CreateIssueRepository(&issueRepository)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newIssueRepository.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newIssueRepository.Name).To(BeEquivalentTo(issueRepository.Name))
		})
	})
})

var _ = Describe("When updating IssueRepository", Label("app", "UpdateIssueRepository"), func() {
	var (
		db                     *mocks.MockDatabase
		issueRepositoryHandler ir.IssueRepositoryHandler
		issueRepository        entity.IssueRepository
		filter                 *entity.IssueRepositoryFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueRepository = test.NewFakeIssueRepositoryEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueRepositoryFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates issueRepository", func() {
		db.On("UpdateIssueRepository", &issueRepository).Return(nil)
		issueRepositoryHandler = ir.NewIssueRepositoryHandler(db, er)
		issueRepository.Name = "SecretRepository"
		filter.Id = []*int64{&issueRepository.Id}
		db.On("GetIssueRepositories", filter).Return([]entity.IssueRepository{issueRepository}, nil)
		updatedIssueRepository, err := issueRepositoryHandler.UpdateIssueRepository(&issueRepository)
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
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueRepositoryFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes issueRepository", func() {
		db.On("DeleteIssueRepository", id).Return(nil)
		issueRepositoryHandler = ir.NewIssueRepositoryHandler(db, er)
		db.On("GetIssueRepositories", filter).Return([]entity.IssueRepository{}, nil)
		err := issueRepositoryHandler.DeleteIssueRepository(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		issueRepositories, err := issueRepositoryHandler.ListIssueRepositories(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issueRepositories.Elements).To(BeEmpty(), "no error should be thrown")
	})
})
