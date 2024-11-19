// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeIssueMatchChange() entity.IssueMatchChange {
	actions := []string{"add", "remove"}
	return entity.IssueMatchChange{
		Id:         int64(gofakeit.Number(1, 10000000)),
		Action:     gofakeit.RandomString(actions),
		IssueMatch: nil,
		Activity:   nil,
		Metadata: entity.Metadata{
			CreatedAt: gofakeit.Date(),
			DeletedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
	}
}

func NNewFakeIssueMatchChanges(n int) []entity.IssueMatchChange {
	r := make([]entity.IssueMatchChange, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeIssueMatchChange()
	}
	return r
}
