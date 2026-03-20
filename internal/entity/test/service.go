// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeServiceEntity() entity.Service {
	return entity.Service{
		BaseService: entity.BaseService{
			Id:           int64(gofakeit.Number(1, 10000000)),
			CCRN:         gofakeit.Name(),
			Domain:       strings.ToLower(gofakeit.SongName()),
			Region:       gofakeit.RandomString([]string{"test-de-1", "test-de-2", "test-us-1", "test-jp-2", "test-jp-1"}),
			SupportGroup: nil,
			Owners:       nil,
			Metadata: entity.Metadata{
				CreatedAt: gofakeit.Date(),
				DeletedAt: gofakeit.Date(),
				UpdatedAt: gofakeit.Date(),
			},
		},
	}
}

func NewFakeServiceWithAggregationsEntity() entity.ServiceWithAggregations {
	return entity.ServiceWithAggregations{
		ServiceAggregations: entity.ServiceAggregations{
			IssueMatches:       int64(gofakeit.Number(1, 10000000)),
			ComponentInstances: int64(gofakeit.Number(1, 10000000)),
		},
		Service: NewFakeServiceEntity(),
	}
}

func NNewFakeServiceEntitiesWithAggregations(n int) []entity.ServiceWithAggregations {
	r := make([]entity.ServiceWithAggregations, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeServiceWithAggregationsEntity()
	}
	return r
}

func NNewFakeServiceEntities(n int) []entity.Service {
	r := make([]entity.Service, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeServiceEntity()
	}
	return r
}

func NewFakeServiceResult() entity.ServiceResult {
	service := NewFakeServiceEntity()
	return entity.ServiceResult{
		Service: &service,
	}
}
