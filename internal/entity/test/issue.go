// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeIssueEntity() entity.Issue {
	issueType := gofakeit.RandomString(entity.AllIssueTypes)
	primaryName := fmt.Sprintf("CVE-%d-%d", gofakeit.Year(), gofakeit.Number(100, 9999999))
	knownExploited := gofakeit.Bool()

	var addedDate, dueDate *time.Time

	if knownExploited {
		added := gofakeit.DateRange(
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		).UTC().Truncate(time.Second)
		due := added.Add(90 * 24 * time.Hour)
		addedDate = &added
		dueDate = &due
	}

	return entity.Issue{
		Id:                      int64(gofakeit.Number(1, 10000000)),
		PrimaryName:             primaryName,
		Description:             gofakeit.AdjectiveDescriptive(),
		Type:                    entity.NewIssueType(issueType),
		KnownExploited:          knownExploited,
		KnownExploitedAddedDate: addedDate,
		KnownExploitedDueDate:   dueDate,
		IssueVariants:           nil,
		IssueMatches:            nil,
		ComponentVersions:       nil,
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
	for i := range n {
		r[i] = NewFakeIssueWithAggregationsEntity()
	}

	return r
}

func NNewFakeIssueEntities(n int) []entity.Issue {
	r := make([]entity.Issue, n)
	for i := range n {
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
	for i := range n {
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
	for i := range n {
		r[i] = NewFakeIssueResultWithAggregations()
	}

	return r
}
