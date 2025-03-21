// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeComponentVersionEntity() entity.ComponentVersion {
	return entity.ComponentVersion{
		Id:                 int64(gofakeit.Number(1, 10000000)),
		Version:            gofakeit.Regex("^sha:[a-fA-F0-9]{64}$"),
		Tag:                gofakeit.AppVersion(),
		ComponentId:        0,
		ComponentInstances: nil,
		Issues:             nil,
		Metadata: entity.Metadata{
			CreatedAt: gofakeit.Date(),
			DeletedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
	}
}

func NNewFakeComponentVersionEntities(n int) []entity.ComponentVersion {
	r := make([]entity.ComponentVersion, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeComponentVersionEntity()
	}
	return r
}
