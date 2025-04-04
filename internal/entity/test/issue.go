// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeIssueEntity() entity.Issue {
	t := gofakeit.RandomString(entity.AllIssueTypes)
	primaryName := fmt.Sprintf("CVE-%d-%d", gofakeit.Year(), gofakeit.Number(100, 9999999))
	return entity.Issue{
		Id:                int64(gofakeit.Number(1, 10000000)),
		PrimaryName:       primaryName,
		Description:       gofakeit.AdjectiveDescriptive(),
		Type:              entity.NewIssueType(t),
		IssueVariants:     nil,
		IssueMatches:      nil,
		ComponentVersions: nil,
		Activity:          nil,
		Metadata: entity.Metadata{
			CreatedAt: gofakeit.Date(),
			DeletedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
	}
}

func NewFakeIssueWithAggregationsEntity() entity.IssueWithAggregations {
	return entity.IssueWithAggregations{
		IssueAggregations: entity.IssueAggregations{
			Activities:                    int64(gofakeit.Number(1, 10000000)),
			IssueMatches:                  int64(gofakeit.Number(1, 10000000)),
			AffectedServices:              int64(gofakeit.Number(1, 10000000)),
			ComponentVersions:             int64(gofakeit.Number(1, 10000000)),
			AffectedComponentInstances:    int64(gofakeit.Number(1, 10000000)),
			EarliestTargetRemediationDate: gofakeit.Date(),
			EarliestDiscoveryDate:         gofakeit.Date(),
		},
		Issue: NewFakeIssueEntity(),
	}
}

func NNewFakeIssueEntitiesWithAggregations(n int) []entity.IssueWithAggregations {
	r := make([]entity.IssueWithAggregations, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeIssueWithAggregationsEntity()
	}
	return r
}

func NNewFakeIssueEntities(n int) []entity.Issue {
	r := make([]entity.Issue, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeIssueEntity()
	}
	return r
}

func NewFakeIssueResult() entity.IssueResult {
	issue := NewFakeIssueEntity()
	return entity.IssueResult{
		Issue: &issue,
	}
}

func NNewFakeIssueResults(n int) []entity.IssueResult {
	r := make([]entity.IssueResult, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeIssueResult()
	}
	return r
}

func NewFakeIssueResultWithAggregations() entity.IssueResult {
	issue := NewFakeIssueWithAggregationsEntity()
	return entity.IssueResult{
		Issue:             &issue.Issue,
		IssueAggregations: &issue.IssueAggregations,
	}
}

func NNewFakeIssueResultsWithAggregations(n int) []entity.IssueResult {
	r := make([]entity.IssueResult, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeIssueResultWithAggregations()
	}
	return r
}
