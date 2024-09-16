// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeIssueRepositoryEntity() entity.IssueRepository {
	return entity.IssueRepository{
		BaseIssueRepository: entity.BaseIssueRepository{
			Id:        int64(gofakeit.Number(1, 10000000)),
			Name:      gofakeit.Noun(),
			Url:       gofakeit.URL(),
			CreatedAt: gofakeit.Date(),
			DeletedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
		IssueRepositoryService: entity.IssueRepositoryService{
			Priority: int64(gofakeit.Number(1, 10)),
		},
	}
}

func NNewFakeIssueRepositories(n int) []entity.IssueRepository {
	r := make([]entity.IssueRepository, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeIssueRepositoryEntity()
	}
	return r
}
