// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakePatchEntity() entity.Patch {
	return entity.Patch{
		Id:                   int64(gofakeit.Number(1, 10000000)),
		ServiceId:            int64(gofakeit.Number(1, 10000000)),
		ServiceName:          gofakeit.AppName(),
		ComponentVersionId:   int64(gofakeit.Number(1, 10000000)),
		ComponentVersionName: gofakeit.AppName(),
		Metadata: entity.Metadata{
			CreatedAt: gofakeit.Date(),
			DeletedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
	}
}

func NNewFakePatches(n int) []entity.Patch {
	p := make([]entity.Patch, n)
	for i := 0; i < n; i++ {
		p[i] = NewFakePatchEntity()
	}
	return p
}

func NewFakePatchResult() entity.PatchResult {
	patch := NewFakePatchEntity()
	return entity.PatchResult{
		Patch: &patch,
	}
}
