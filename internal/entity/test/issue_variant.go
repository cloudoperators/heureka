// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeIssueVariantEntity() entity.IssueVariant {
	vector := test.GenerateRandomCVSS31Vector()
	severity := entity.NewSeverity(vector)
	return entity.IssueVariant{
		Id:                int64(gofakeit.Number(1, 10000000)),
		SecondaryName:     gofakeit.Noun(),
		Description:       gofakeit.Sentence(10),
		Severity:          severity,
		IssueId:           0,
		Issue:             nil,
		IssueRepositoryId: 0,
		IssueRepository:   nil,
		CreatedAt:         gofakeit.Date(),
		DeletedAt:         gofakeit.Date(),
		UpdatedAt:         gofakeit.Date(),
	}
}

func NNewFakeIssueVariants(n int) []entity.IssueVariant {
	r := make([]entity.IssueVariant, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeIssueVariantEntity()
	}
	return r
}
