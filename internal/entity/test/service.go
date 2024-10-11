// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeServiceEntity() entity.Service {
	return entity.Service{
		BaseService: entity.BaseService{
			Id:           int64(gofakeit.Number(1, 10000000)),
			Name:         gofakeit.Name(),
			SupportGroup: nil,
			Activities:   nil,
			Owners:       nil,
			CreatedAt:    gofakeit.Date(),
			DeletedAt:    gofakeit.Date(),
			UpdatedAt:    gofakeit.Date(),
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
