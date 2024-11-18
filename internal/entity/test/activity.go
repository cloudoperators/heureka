// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeActivityEntity() entity.Activity {
	return entity.Activity{
		Id:        int64(gofakeit.Number(1, 10000000)),
		Status:    entity.NewActivityStatusValue(gofakeit.RandomString(entity.AllActivityStatusValues)),
		Issues:    nil,
		Evidences: nil,
		Service:   nil,
		Metadata: entity.Metadata{
			CreatedAt: gofakeit.Date(),
			DeletedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
	}
}

func NNewFakeActivities(n int) []entity.Activity {
	r := make([]entity.Activity, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeActivityEntity()
	}
	return r
}
