// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func NewFakeUserEntity() entity.User {
	sapId := fmt.Sprintf("I%d", gofakeit.IntRange(100000, 999999))
	return entity.User{
		Id:        int64(gofakeit.Number(1, 10000000)),
		Name:      gofakeit.Name(),
		SapID:     sapId,
		CreatedAt: gofakeit.Date(),
		DeletedAt: gofakeit.Date(),
		UpdatedAt: gofakeit.Date(),
	}
}

func NNewFakeUserEntities(n int) []entity.User {
	r := make([]entity.User, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeUserEntity()
	}
	return r
}
