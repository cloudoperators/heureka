// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeComponentEntity() entity.Component {
	return entity.Component{
		Id:        int64(gofakeit.Number(1, 10000000)),
		CCRN:      gofakeit.Name(),
		Type:      gofakeit.Word(),
		CreatedAt: gofakeit.Date(),
		DeletedAt: gofakeit.Date(),
		UpdatedAt: gofakeit.Date(),
	}
}

func NNewFakeComponentEntities(n int) []entity.Component {
	r := make([]entity.Component, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeComponentEntity()
	}
	return r
}
