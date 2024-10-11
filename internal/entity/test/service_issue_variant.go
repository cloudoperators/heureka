package test

import (
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeServiceIssueVariantEntity(prio int64, issueId *int64) entity.ServiceIssueVariant {
	return entity.ServiceIssueVariant{
		IssueVariant: NewFakeIssueVariantEntity(issueId),
		ServiceId:    0,
		Priority:     prio,
	}
}

func NNewFakeServiceIssueVariantEntity(n int, prio int64, issueId *int64) []entity.ServiceIssueVariant {
	r := make([]entity.ServiceIssueVariant, n)
	for i := 0; i < n; i++ {
		r[i] = NewFakeServiceIssueVariantEntity(prio, issueId)
	}
	return r
}
