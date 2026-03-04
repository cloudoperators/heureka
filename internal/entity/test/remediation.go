// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeRemediationEntity() entity.Remediation {
	t := gofakeit.RandomString(entity.AllRemediationTypes)
	s := gofakeit.RandomString(entity.AllSeverityValuesString)
	return entity.Remediation{
		Id:              int64(gofakeit.Number(1, 10000000)),
		Description:     gofakeit.Sentence(10),
		Severity:        entity.NewSeverityValues(s),
		RemediationDate: gofakeit.Date(),
		ExpirationDate:  gofakeit.Date(),
		Type:            entity.NewRemediationType(t),
		Service:         gofakeit.AppName(),
		ServiceId:       int64(gofakeit.Number(1, 10000000)),
		Issue:           gofakeit.HackerPhrase(),
		IssueId:         int64(gofakeit.Number(1, 10000000)),
		Component:       gofakeit.AppName(),
		ComponentId:     int64(gofakeit.Number(1, 10000000)),
		RemediatedById:  int64(gofakeit.Number(1, 10000000)),
		RemediatedBy:    gofakeit.Name(),
		Metadata: entity.Metadata{
			CreatedAt: gofakeit.Date(),
			DeletedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
	}
}

func NNewFakeRemediations(n int) []entity.Remediation {
	r := make([]entity.Remediation, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeRemediationEntity()
	}
	return r
}

func NewFakeRemediationResult() entity.RemediationResult {
	remediation := NewFakeRemediationEntity()
	return entity.RemediationResult{
		Remediation: &remediation,
	}
}
