// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeComponentInstanceEntity() entity.ComponentInstance {
	return entity.ComponentInstance{
		Id:                 int64(gofakeit.Number(1, 10000000)),
		CCRN:               gofakeit.URL(),
		Count:              int16(gofakeit.Number(1, 100)),
		ComponentVersion:   nil,
		ComponentVersionId: int64(gofakeit.Number(1, 10000000)),
		Service:            nil,
		ServiceId:          int64(gofakeit.Number(1, 10000000)),
		CreatedAt:          gofakeit.Date(),
		DeletedAt:          gofakeit.Date(),
		UpdatedAt:          gofakeit.Date(),
	}
}

func NNewFakeComponentInstances(n int) []entity.ComponentInstance {
	r := make([]entity.ComponentInstance, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeComponentInstanceEntity()
	}
	return r
}
