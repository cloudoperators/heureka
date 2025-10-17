// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeEvidenceEntity() entity.Evidence {
	vector := test.GenerateRandomCVSS31Vector()
	severity := entity.NewSeverity(vector)
	t := gofakeit.RandomString(entity.AllEvidenceTypeValues)
	return entity.Evidence{
		Id:          int64(gofakeit.Number(1, 10000000)),
		Description: gofakeit.Sentence(),
		RaaEnd:      gofakeit.Date(),
		Type:        entity.NewEvidenceTypeValue(t),
		Severity:    severity,
		Metadata: entity.Metadata{
			CreatedAt: gofakeit.Date(),
			DeletedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
	}
}

func NNewFakeEvidences(n int) []entity.Evidence {
	r := make([]entity.Evidence, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeEvidenceEntity()
	}
	return r
}
