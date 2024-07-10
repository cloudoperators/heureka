// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app_test

import (
	"errors"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.wdf.sap.corp/cc/heureka/internal/app"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/internal/entity/test"
	"github.wdf.sap.corp/cc/heureka/internal/mocks"
)

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

var _ = Describe("When listing Issues", Label("app", "ListIssues"), func() {
	var (
		db      *mocks.MockDatabase
		heureka app.Heureka
		filter  *entity.IssueFilter
		options *entity.ListOptions
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		options = getListOptions()
		filter = getIssueFilter()
	})

	When("the list option does include the totalCount", func() {

		BeforeEach(func() {
			options.ShowTotalCount = true
			db.On("GetIssues", filter).Return([]entity.Issue{}, nil)
			db.On("CountIssues", filter).Return(int64(1337), nil)
		})

		It("shows the total count in the results", func() {
			heureka = app.NewHeurekaApp(db)
			res, err := heureka.ListIssues(filter, options)
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
			matches := test.NNewFakeIssueEntities(resElements)

			var ids = lo.Map(matches, func(m entity.Issue, _ int) int64 { return m.Id })
			var i int64 = 0
			for len(ids) < dbElements {
				i++
				ids = append(ids, i)
			}
			db.On("GetIssues", filter).Return(matches, nil)
			db.On("GetAllIssueIds", filter).Return(ids, nil)
			heureka = app.NewHeurekaApp(db)
			res, err := heureka.ListIssues(filter, options)
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

				heureka = app.NewHeurekaApp(db)
				res, err := heureka.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")

			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				db.On("GetIssuesWithAggregations", filter).Return(test.NNewFakeIssueEntitiesWithAggregations(10), nil)
			})
			It("should return the expected issues in the result", func() {
				heureka = app.NewHeurekaApp(db)
				res, err := heureka.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(10), "return 10 results")
			})
		})
		Context("and the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetIssuesWithAggregations", filter).Return([]entity.IssueWithAggregations{}, errors.New("some error"))
			})

			It("should return the expected issues in the result", func() {
				heureka = app.NewHeurekaApp(db)
				_, err := heureka.ListIssues(filter, options)
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

				heureka = app.NewHeurekaApp(db)
				res, err := heureka.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(0), "return no results")

			})
		})
		Context("and the filter does have results in the database", func() {
			BeforeEach(func() {
				db.On("GetIssues", filter).Return(test.NNewFakeIssueEntities(15), nil)
			})
			It("should return the expected issues in the result", func() {
				heureka = app.NewHeurekaApp(db)
				res, err := heureka.ListIssues(filter, options)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(len(res.Elements)).Should(BeEquivalentTo(15), "return 15 results")
			})
		})

		Context("and  the database operations throw an error", func() {
			BeforeEach(func() {
				db.On("GetIssues", filter).Return([]entity.Issue{}, errors.New("some error"))
			})

			It("should return the expected issues in the result", func() {
				heureka = app.NewHeurekaApp(db)
				_, err := heureka.ListIssues(filter, options)
				Expect(err).Error()
				Expect(err.Error()).ToNot(BeEquivalentTo("some error"), "error gets not passed through")
			})
		})
	})
})

var _ = Describe("When creating Issue", Label("app", "CreateIssue"), func() {
	var (
		db      *mocks.MockDatabase
		heureka app.Heureka
		issue   entity.Issue
		filter  *entity.IssueFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issue = test.NewFakeIssueEntity()
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
		filter.PrimaryName = []*string{&issue.PrimaryName}
		db.On("CreateIssue", &issue).Return(&issue, nil)
		db.On("GetIssues", filter).Return([]entity.Issue{}, nil)
		heureka = app.NewHeurekaApp(db)
		newIssue, err := heureka.CreateIssue(&issue)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(newIssue.Id).NotTo(BeEquivalentTo(0))
		By("setting fields", func() {
			Expect(newIssue.PrimaryName).To(BeEquivalentTo(issue.PrimaryName))
			Expect(newIssue.Description).To(BeEquivalentTo(issue.Description))
			Expect(newIssue.Type.String()).To(BeEquivalentTo(issue.Type.String()))
		})
	})
})

var _ = Describe("When updating Issue", Label("app", "UpdateIssue"), func() {
	var (
		db      *mocks.MockDatabase
		heureka app.Heureka
		issue   entity.Issue
		filter  *entity.IssueFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issue = test.NewFakeIssueEntity()
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

	It("updates issue", func() {
		db.On("UpdateIssue", &issue).Return(nil)
		heureka = app.NewHeurekaApp(db)
		issue.Description = "New Description"
		filter.Id = []*int64{&issue.Id}
		db.On("GetIssues", filter).Return([]entity.Issue{issue}, nil)
		updatedIssue, err := heureka.UpdateIssue(&issue)
		Expect(err).To(BeNil(), "no error should be thrown")
		By("setting fields", func() {
			Expect(updatedIssue.PrimaryName).To(BeEquivalentTo(issue.PrimaryName))
			Expect(updatedIssue.Description).To(BeEquivalentTo(issue.Description))
			Expect(updatedIssue.Type.String()).To(BeEquivalentTo(issue.Type.String()))
		})
	})
})

var _ = Describe("When deleting Issue", Label("app", "DeleteIssue"), func() {
	var (
		db      *mocks.MockDatabase
		heureka app.Heureka
		id      int64
		filter  *entity.IssueFilter
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
		heureka = app.NewHeurekaApp(db)
		db.On("GetIssues", filter).Return([]entity.Issue{}, nil)
		err := heureka.DeleteIssue(id)
		Expect(err).To(BeNil(), "no error should be thrown")

		filter.Id = []*int64{&id}
		issues, err := heureka.ListIssues(filter, &entity.ListOptions{})
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issues.Elements).To(BeEmpty(), "no error should be thrown")
	})
})

var _ = Describe("When modifying ComponentVersion and Issue", Label("app", "ComponentVersionIssue"), func() {
	var (
		db               *mocks.MockDatabase
		heureka          app.Heureka
		issue            entity.Issue
		componentVersion entity.ComponentVersion
		filter           *entity.IssueFilter
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		issue = test.NewFakeIssueEntity()
		componentVersion = test.NewFakeComponentVersionEntity()
		first := 10
		var after int64
		after = 0
		filter = &entity.IssueFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id: []*int64{&issue.Id},
		}
	})

	It("adds componentVersion to issue", func() {
		db.On("AddComponentVersionToIssue", issue.Id, componentVersion.Id).Return(nil)
		db.On("GetIssues", filter).Return([]entity.Issue{issue}, nil)
		heureka = app.NewHeurekaApp(db)
		issue, err := heureka.AddComponentVersionToIssue(issue.Id, componentVersion.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issue).NotTo(BeNil(), "issue should be returned")
	})

	It("removes componentVersion from issue", func() {
		db.On("RemoveComponentVersionFromIssue", issue.Id, componentVersion.Id).Return(nil)
		db.On("GetIssues", filter).Return([]entity.Issue{issue}, nil)
		heureka = app.NewHeurekaApp(db)
		issue, err := heureka.RemoveComponentVersionFromIssue(issue.Id, componentVersion.Id)
		Expect(err).To(BeNil(), "no error should be thrown")
		Expect(issue).NotTo(BeNil(), "issue should be returned")
	})
})
