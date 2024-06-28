// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func NewFakeIssueEntity() entity.Issue {
	return entity.Issue{
		Id:                int64(gofakeit.Number(1, 10000000)),
		PrimaryName:       gofakeit.Name(),
		Description:       gofakeit.AdjectiveDescriptive(),
		IssueVariants:     nil,
		IssueMatches:      nil,
		ComponentVersions: nil,
		Activity:          nil,
		CreatedAt:         gofakeit.Date(),
		DeletedAt:         gofakeit.Date(),
		UpdatedAt:         gofakeit.Date(),
	}
}

func NewFakeIssueWithAggregationsEntity() entity.IssueWithAggregations {
	return entity.IssueWithAggregations{
		IssueAggregations: entity.IssueAggregations{
			Activites:                     int64(gofakeit.Number(1, 10000000)),
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
