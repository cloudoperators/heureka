// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeSupportGroupEntity() entity.SupportGroup {
	return entity.SupportGroup{
		Id:   int64(gofakeit.Number(1, 10000000)),
		CCRN: gofakeit.AppName(),
		Metadata: entity.Metadata{
			CreatedAt: gofakeit.Date(),
			DeletedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
	}
}

func NNewFakeSupportGroupEntities(n int) []entity.SupportGroup {
	r := make([]entity.SupportGroup, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeSupportGroupEntity()
	}
	return r
}
