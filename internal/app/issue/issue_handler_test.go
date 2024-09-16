// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0
package issue_test

import (
	"errors"
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/app/issue"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
)

func TestIssueHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Issue Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	er = event.NewEventRegistry()
})

func getIssueFilter() *entity.IssueFilter {
	serviceName := "SomeNotExistingService"
	return &entity.IssueFilter{
		Paginated: entity.Paginated{
			First: nil,
			After: nil,
		},
		ServiceName:                     []*string{&serviceName},
		Id:                              nil,
		IssueMatchStatus:                nil,
		IssueMatchDiscoveryDate:         nil,
		IssueMatchTargetRemediationDate: nil,
	}
}

func getIssueListOptions() *entity.IssueListOptions {
	listOptions := entity.NewListOptions()
	return &entity.IssueListOptions{
		ListOptions:         *listOptions,
		ShowIssueTypeCounts: false,
	}
}

func getIssueTypeCounts() *entity.IssueTypeCounts {
	return &entity.IssueTypeCounts{
		VulnerabilityCount:   1000,
		PolicyViolationCount: 300,
		SecurityEventCount:   37,
	}
}

var _ = Describe("When listing Issues", Label("app", "ListIssues"), func() {
	var (
		db              *mocks.MockDatabase
		issueHandler    issue.IssueHandler
		filter          *entity.IssueFilter
		options         *entity.IssueListOptions
		issueTypeCounts *entity.IssueTypeCounts
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = getIssueListOptions()
		filter = getIssueFilter()
		issueTypeCounts = getIssueTypeCounts()

	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetIssues", filter).Return([]entity.Issue{}, nil)
			db.On("CountIssueTypes", filter).Return(issueTypeCounts, nil)
		})

		It("shows the total count in the results", func() {
			issueHandler = issue.NewIssueHandler(db, er)
			res, err := issueHandler.ListIssues(filter, options)
			Expect(err).To(BeNil(), "no error should be thrown")
			Expect(*res.TotalCount).Should(BeEquivalentTo(int64(1337)), "return correct Totalcount")
		})
	})

	When("the list option does include the PageInfo", func() {
		BeforeEach(func() {
			options.ShowPageInfo = true
			db.On("CountIssueTypes", filter).Return(issueTypeCounts, nil)
		})
		DescribeTable("pagination information is correct", func(pageSize int, dbElements int, resElements int, hasNextPage bool) {
			filter.First = &pageSize
			matches := test.NNewFakeIssueEntities(resElements)

			var ids = lo.Map(matches, func(m entity.Issue, _ int) int64 { return m.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetIssues", filter).Return(matches, nil)
			db.On("GetAllIssueIds", filter).Return(ids, nil)
			issueHandler = issue.NewIssueHandler(db, er)
			res, err := issueHandler.ListIssues(filter, options)
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

	When("the list options does include aggregations", func() {
		BeforeEach(func() {
			options.IncludeAggregations = true
		})
		Context("and the given filter does not have any matches in the database", func() {

			BeforeEach(func() {
				db.On("GetIssuesWithAggregations", filter).Return([]entity.IssueWithAggregations{}, nil)
			})

			It("should return an empty result", func() {

				issueHandler = issue.NewIssueHandler(db, er)
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")

			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				db.On("GetIssuesWithAggregations", filter).Return(test.NNewFakeIssueEntitiesWithAggregations(10), nil)
			})
			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(db, er)
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(10), "return 10 results")
			})
		})
		Context("and the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetIssuesWithAggregations", filter).Return([]entity.IssueWithAggregations{}, errors.New("some error"))
			})

			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(db, er)
				_, err := issueHandler.ListIssues(filter, options)
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
				db.On("GetIssues", filter).Return([]entity.Issue{}, nil)
			})
			It("should return an empty result", func() {

				issueHandler = issue.NewIssueHandler(db, er)
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")

			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				db.On("GetIssues", filter).Return(test.NNewFakeIssueEntities(15), nil)
			})
			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(db, er)
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(15), "return 15 results")
			})
		})

		Context("and  the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetIssues", filter).Return([]entity.Issue{}, errors.New("some error"))
			})

			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(db, er)
				_, err := issueHandler.ListIssues(filter, options)
				Expect(err).Error()
				Expect(err.Error()).ToNot(BeEquivalentTo("some error"), "error gets not passed through")
			})
		})
	})
})

var _ = Describe("When creating Issue", Label("app", "CreateIssue"), func() {
	var (
		db           *mocks.MockDatabase
		issueHandler issue.IssueHandler
		issueEntity  entity.Issue
		filter       *entity.IssueFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueEntity = test.NewFakeIssueEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("creates issue", func() {
		filter.PrimaryName = []*string{&issueEntity.PrimaryName}
		db.On("CreateIssue", &issueEntity).Return(&issueEntity, nil)
		db.On("GetIssues", filter).Return([]entity.Issue{}, nil)
		issueHandler = issue.NewIssueHandler(db, er)
		newIssue, err := issueHandler.CreateIssue(&issueEntity)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newIssue.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newIssue.PrimaryName).To(BeEquivalentTo(issueEntity.PrimaryName))
			Expect(newIssue.Description).To(BeEquivalentTo(issueEntity.Description))
			Expect(newIssue.Type.String()).To(BeEquivalentTo(issueEntity.Type.String()))
		})
	})
})

var _ = Describe("When updating Issue", Label("app", "UpdateIssue"), func() {
	var (
		db           *mocks.MockDatabase
		issueHandler issue.IssueHandler
		issueEntity  entity.Issue
		filter       *entity.IssueFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueEntity = test.NewFakeIssueEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates issueEntity", func() {
		db.On("UpdateIssue", &issueEntity).Return(nil)
		issueHandler = issue.NewIssueHandler(db, er)
		issueEntity.Description = "New Description"
		filter.Id = []*int64{&issueEntity.Id}
		db.On("GetIssues", filter).Return([]entity.Issue{issueEntity}, nil)
		updatedIssue, err := issueHandler.UpdateIssue(&issueEntity)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedIssue.PrimaryName).To(BeEquivalentTo(issueEntity.PrimaryName))
			Expect(updatedIssue.Description).To(BeEquivalentTo(issueEntity.Description))
			Expect(updatedIssue.Type.String()).To(BeEquivalentTo(issueEntity.Type.String()))
		})
	})
})

var _ = Describe("When deleting Issue", Label("app", "DeleteIssue"), func() {
	var (
		db           *mocks.MockDatabase
		issueHandler issue.IssueHandler
		id           int64
		filter       *entity.IssueFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		id = 1
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes issue", func() {
		db.On("DeleteIssue", id).Return(nil)
		issueHandler = issue.NewIssueHandler(db, er)
		db.On("GetIssues", filter).Return([]entity.Issue{}, nil)
		err := issueHandler.DeleteIssue(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		issues, err := issueHandler.ListIssues(filter, &entity.IssueListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issues.Elements).To(BeEmpty(), "no error should be thrown")
	})
})

var _ = Describe("When modifying relationship of ComponentVersion and Issue", Label("app", "ComponentVersionIssueRelationship"), func() {
	var (
		db               *mocks.MockDatabase
		issueHandler     issue.IssueHandler
		issueEntity      entity.Issue
		componentVersion entity.ComponentVersion
		filter           *entity.IssueFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueEntity = test.NewFakeIssueEntity()
		componentVersion = test.NewFakeComponentVersionEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id: []*int64{&issueEntity.Id},
		}
	})

	It("adds componentVersion to issueEntity", func() {
		db.On("AddComponentVersionToIssue", issueEntity.Id, componentVersion.Id).Return(nil)
		db.On("GetIssues", filter).Return([]entity.Issue{issueEntity}, nil)
		issueHandler = issue.NewIssueHandler(db, er)
		issue, err := issueHandler.AddComponentVersionToIssue(issueEntity.Id, componentVersion.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issue).NotTo(BeNil(), "issueEntity should be returned")
	})

	It("removes componentVersion from issueEntity", func() {
		db.On("RemoveComponentVersionFromIssue", issueEntity.Id, componentVersion.Id).Return(nil)
		db.On("GetIssues", filter).Return([]entity.Issue{issueEntity}, nil)
		issueHandler = issue.NewIssueHandler(db, er)
		issue, err := issueHandler.RemoveComponentVersionFromIssue(issueEntity.Id, componentVersion.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issue).NotTo(BeNil(), "issueEntity should be returned")
	})
})
