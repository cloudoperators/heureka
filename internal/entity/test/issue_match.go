// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/database/mariadb/test"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeIssueMatch() entity.IssueMatch {
	v := test.GenerateRandomCVSS31Vector()
	s := entity.NewSeverity(v)
	return entity.IssueMatch{
		Id:                    int64(gofakeit.Number(1, 10000000)),
		Status:                NewRandomIssueStatus(),
		Severity:              s,
		UserId:                0,
		User:                  nil,
		Evidences:             nil,
		ComponentInstanceId:   0,
		ComponentInstance:     nil,
		IssueId:               0,
		Issue:                 nil,
		RemediationDate:       gofakeit.Date(),
		TargetRemediationDate: gofakeit.Date(),
		CreatedAt:             gofakeit.Date(),
		DeletedAt:             gofakeit.Date(),
		UpdatedAt:             gofakeit.Date(),
	}
}

func NNewFakeIssueMatches(n int) []entity.IssueMatch {
	r := make([]entity.IssueMatch, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeIssueMatch()
	}
	return r
}

func NewRandomIssueStatus() entity.IssueMatchStatusValue {
	value := gofakeit.RandomString(entity.AllIssueMatchStatusValues)
	return entity.NewIssueMatchStatusValue(value)
}
