// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

// // SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// // SPDX-License-Identifier: Apache-2.0
package severity_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/app/issue_repository"
	"github.wdf.sap.corp/cc/heureka/internal/app/issue_variant"
	ss "github.wdf.sap.corp/cc/heureka/internal/app/severity"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
	"github.wdf.sap.corp/cc/heureka/internal/entity/test"
	"github.wdf.sap.corp/cc/heureka/internal/mocks"
	"testing"
)

func TestSeverityService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Severity Service Test Suite")
}

var er event.EventRegistry

var _ = BeforeSuite(func() {
	er = event.NewEventRegistry()
})

func severityFilter() *entity.SeverityFilter {
	return &entity.SeverityFilter{
		IssueMatchId: nil,
	}
}

var _ = Describe("When get Severity", Label("app", "GetSeverity"), func() {
	var (
		db               *mocks.MockDatabase
		ivs              issue_variant.IssueVariantService
		rs               issue_repository.IssueRepositoryService
		severityService  ss.SeverityService
		sFilter          *entity.SeverityFilter
		ivFilter         *entity.IssueVariantFilter
		irFilter         *entity.IssueRepositoryFilter
		issueVariants    []entity.IssueVariant
		repositories     []entity.IssueRepository
		maxSeverityScore float64
	)

	BeforeEach(func() {
		db = mocks.NewMockDatabase(GinkgoT())
		sFilter = severityFilter()
		ivFilter = entity.NewIssueVariantFilter()
		first := 10
		ivFilter.First = &first
		var after int64 = 0
		ivFilter.After = &after
		irFilter = entity.NewIssueRepositoryFilter()
		irFilter.First = &first
		irFilter.After = &after
		rs = issue_repository.NewIssueRepositoryService(db, er)
		ivs = issue_variant.NewIssueVariantService(db, er, rs)
	})

	Context("issue repositories have different priority", func() {
		BeforeEach(func() {
			issueVariants = test.NNewFakeIssueVariants(25)
			repositories = test.NNewFakeIssueRepositories(2)
			repositories[0].Priority = 1
			repositories[1].Priority = 2
			for i := range issueVariants {
				issueVariants[i].IssueRepositoryId = repositories[i%2].Id
			}
			irFilter.Id = lo.Map(issueVariants, func(item entity.IssueVariant, _ int) *int64 {
				return &item.IssueRepositoryId
			})
			db.On("GetIssueVariants", ivFilter).Return(issueVariants, nil)
			db.On("GetIssueRepositories", irFilter).Return(repositories, nil)
		})
		When("higher priority issue variant has highest severity score", func() {
			BeforeEach(func() {
				maxSeverityScore = 90000.0
				issueVariants[1].Severity.Score = maxSeverityScore
			})
			It("returns severity value", func() {
				severityService = ss.NewSeverityService(db, er, ivs)
				severity, err := severityService.GetSeverity(sFilter)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(severity).ToNot((BeNil()), "severity should exist.")
				Expect(severity.Score).To(BeEquivalentTo(maxSeverityScore), "severity score is correct.")
			})
		})
		When("lower priority issueVariant has highest score", func() {
			BeforeEach(func() {
				maxSeverityScore = 90000.0
				issueVariants[0].Severity.Score = maxSeverityScore
				issueVariants[1].Severity.Score = maxSeverityScore - 1
			})
			It("returns severity value", func() {
				severityService = ss.NewSeverityService(db, er, ivs)
				severity, err := severityService.GetSeverity(sFilter)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(severity).ToNot((BeNil()), "severity should exist.")
				Expect(severity.Score).To(BeEquivalentTo(maxSeverityScore-1), "severity score is correct.")
			})
		})
	})
	Context("issue repositories have same priority", func() {
		BeforeEach(func() {
			issueVariants = test.NNewFakeIssueVariants(25)
			repositories = test.NNewFakeIssueRepositories(2)
			repositories[0].Priority = 1
			repositories[1].Priority = 1
			for i := range issueVariants {
				issueVariants[i].IssueRepositoryId = repositories[i%2].Id
			}
			irFilter.Id = lo.Map(issueVariants, func(item entity.IssueVariant, _ int) *int64 {
				return &item.IssueRepositoryId
			})
			db.On("GetIssueVariants", ivFilter).Return(issueVariants, nil)
			db.On("GetIssueRepositories", irFilter).Return(repositories, nil)
		})
		When("issueVariants have different severity score", func() {
			BeforeEach(func() {
				maxSeverityScore = 90000.0
				issueVariants[0].Severity.Score = maxSeverityScore
				issueVariants[1].Severity.Score = maxSeverityScore - 1
			})
			It("return severity value", func() {
				severityService = ss.NewSeverityService(db, er, ivs)
				severity, err := severityService.GetSeverity(sFilter)
				Expect(err).To(BeNil(), "no error should be thrown")
				Expect(severity).ToNot((BeNil()), "severity should exist.")
				Expect(severity.Score).To(BeEquivalentTo(maxSeverityScore), "severity score ist correct.")
			})
		})
	})
})
