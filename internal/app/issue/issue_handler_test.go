// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0
package issue_test

import (
	"errors"
	"math"
	"testing"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/app/issue"
	appIssue "github.com/cloudoperators/heureka/internal/app/issue"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/entity/test"
	"github.com/cloudoperators/heureka/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mock "github.com/stretchr/testify/mock"
)

func TestIssueHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Issue Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	db := mocks.NewMockDatabase(GinkgoT())
	er = event.NewEventRegistry(db)
})

func getIssueFilter() *entity.IssueFilter {
	return &entity.IssueFilter{
		PaginatedX: entity.PaginatedX{
			First: nil,
			After: nil,
		},
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
			db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{}, nil)
			db.On("CountIssueTypes", filter).Return(issueTypeCounts, nil)
		})

		It("shows the total count in the results", func() {
			issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
			res, err := issueHandler.ListIssues(filter, options)
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
			issues := []entity.IssueResult{}
			for _, i := range test.NNewFakeIssueEntities(resElements) {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithIssue([]entity.Order{}, i, 0))
				issues = append(issues, entity.IssueResult{WithCursor: entity.WithCursor{Value: cursor}, Issue: lo.ToPtr(i)})
			}

			var cursors = lo.Map(issues, func(ir entity.IssueResult, _ int) string {
				cursor, _ := mariadb.EncodeCursor(mariadb.WithIssue([]entity.Order{}, *ir.Issue, 0))
				return cursor
			})

			var i int64 = 0
			for len(cursors) < dbElements {
				i++
				issue := test.NewFakeIssueEntity()
				c, _ := mariadb.EncodeCursor(mariadb.WithIssue([]entity.Order{}, issue, 0))
				cursors = append(cursors, c)
			}
			db.On("GetIssues", filter, []entity.Order{}).Return(issues, nil)
			db.On("GetAllIssueCursors", filter, []entity.Order{}).Return(cursors, nil)
			db.On("CountIssueTypes", filter).Return(issueTypeCounts, nil)
			issueHandler = appIssue.NewIssueHandler(db, er, cache.NewNoCache())
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
				db.On("GetIssuesWithAggregations", filter, []entity.Order{}).Return([]entity.IssueResult{}, nil)
			})

			It("should return an empty result", func() {
				issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")

			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				db.On("GetIssuesWithAggregations", filter, []entity.Order{}).Return(test.NNewFakeIssueResultsWithAggregations(10), nil)
			})
			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(10), "return 10 results")
			})
		})
		Context("and the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetIssuesWithAggregations", filter, []entity.Order{}).Return([]entity.IssueResult{}, errors.New("some error"))
			})

			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
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
				db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{}, nil)
			})
			It("should return an empty result", func() {

				issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")

			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				db.On("GetIssues", filter, []entity.Order{}).Return(test.NNewFakeIssueResults(15), nil)
			})
			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
				res, err := issueHandler.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(15), "return 15 results")
			})
		})

		Context("and  the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{}, errors.New("some error"))
			})

			It("should return the expected issues in the result", func() {
				issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
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
		after := ""
		filter = &entity.IssueFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
	})

	It("creates issue", func() {
		filter.PrimaryName = []*string{&issueEntity.PrimaryName}
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("CreateIssue", &issueEntity).Return(&issueEntity, nil)
		db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{}, nil)
		issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
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
		issueResult  entity.IssueResult
		filter       *entity.IssueFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueResult = test.NewFakeIssueResult()
		first := 10
		after := ""
		filter = &entity.IssueFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
	})

	It("updates issueEntity", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("UpdateIssue", issueResult.Issue).Return(nil)
		issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
		issueResult.Issue.Description = "New Description"
		filter.Id = []*int64{&issueResult.Issue.Id}
		db.On("GetIssues", filter, []entity.Order{}).Return([]entity.IssueResult{issueResult}, nil)
		updatedIssue, err := issueHandler.UpdateIssue(issueResult.Issue)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedIssue.PrimaryName).To(BeEquivalentTo(issueResult.Issue.PrimaryName))
			Expect(updatedIssue.Description).To(BeEquivalentTo(issueResult.Issue.Description))
			Expect(updatedIssue.Type.String()).To(BeEquivalentTo(issueResult.Issue.Type.String()))
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
		after := ""
		filter = &entity.IssueFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
	})

	It("deletes issue", func() {
		db.On("GetAllUserIds", mock.Anything).Return([]int64{}, nil)
		db.On("DeleteIssue", id, mock.Anything).Return(nil)
		issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
		db.On("GetIssues", mock.Anything, []entity.Order{}).Return([]entity.IssueResult{}, nil)
		err := issueHandler.DeleteIssue(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		lo := entity.IssueListOptions{
			ListOptions: *entity.NewListOptions(),
		}
		issues, err := issueHandler.ListIssues(filter, &lo)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issues.Elements).To(BeEmpty(), "no error should be thrown")
	})
})

var _ = Describe("When modifying relationship of ComponentVersion and Issue", Label("app", "ComponentVersionIssueRelationship"), func() {
	var (
		db               *mocks.MockDatabase
		issueHandler     issue.IssueHandler
		issueResult      entity.IssueResult
		componentVersion entity.ComponentVersion
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issueResult = test.NewFakeIssueResult()
		componentVersion = test.NewFakeComponentVersionEntity()
	})

	It("adds componentVersion to issueEntity", func() {
		db.On("AddComponentVersionToIssue", issueResult.Issue.Id, componentVersion.Id).Return(nil)
		db.On("GetIssues", mock.Anything, mock.Anything).Return([]entity.IssueResult{issueResult}, nil)
		issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
		issue, err := issueHandler.AddComponentVersionToIssue(issueResult.Issue.Id, componentVersion.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issue).NotTo(BeNil(), "issueEntity should be returned")
	})

	It("removes componentVersion from issueEntity", func() {
		db.On("RemoveComponentVersionFromIssue", issueResult.Issue.Id, componentVersion.Id).Return(nil)
		db.On("GetIssues", mock.Anything, mock.Anything).Return([]entity.IssueResult{issueResult}, nil)
		issueHandler = issue.NewIssueHandler(db, er, cache.NewNoCache())
		issue, err := issueHandler.RemoveComponentVersionFromIssue(issueResult.Issue.Id, componentVersion.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issue).NotTo(BeNil(), "issueEntity should be returned")
	})
})
