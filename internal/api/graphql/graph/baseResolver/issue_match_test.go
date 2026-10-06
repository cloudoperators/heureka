// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	mock "github.com/stretchr/testify/mock"
	"k8s.io/utils/pointer"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/mocks"
)

func TestBaseResolver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BaseResolver Suite")
}

// gqlgenCtx returns a context enriched with minimal gqlgen operation and field
// contexts, required because IssueMatchBaseResolver calls GetPreloads which
// reads both from the context.
func gqlgenCtx() context.Context {
	ctx := graphql.WithOperationContext(context.Background(), &graphql.OperationContext{})
	ctx = graphql.WithFieldContext(ctx, &graphql.FieldContext{})

	return ctx
}

func emptyIssueMatchList() *entity.List[entity.IssueMatchResult] {
	zero := int64(0)

	return &entity.List[entity.IssueMatchResult]{
		Elements:   []entity.IssueMatchResult{},
		TotalCount: &zero,
		PageInfo:   &entity.PageInfo{},
	}
}

var _ = Describe("parseDateTimeFilter", func() {
	When("input is nil", func() {
		It("returns nil, nil", func() {
			result, err := parseDateTimeFilter(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	When("both After and Before are nil", func() {
		It("returns nil, nil", func() {
			result, err := parseDateTimeFilter(&model.DateTimeFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	When("only After is set to a valid RFC3339 string", func() {
		It("sets TimeFilter.After and leaves Before zero", func() {
			ts := "2026-01-15T10:00:00Z"
			result, err := parseDateTimeFilter(&model.DateTimeFilter{After: &ts})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.After).To(Equal(time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)))
			Expect(result.Before.IsZero()).To(BeTrue())
		})
	})

	When("only Before is set to a valid RFC3339 string", func() {
		It("sets TimeFilter.Before and leaves After zero", func() {
			ts := "2026-06-01T00:00:00Z"
			result, err := parseDateTimeFilter(&model.DateTimeFilter{Before: &ts})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Before).To(Equal(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)))
			Expect(result.After.IsZero()).To(BeTrue())
		})
	})

	When("both After and Before are valid RFC3339 strings", func() {
		It("sets both fields correctly", func() {
			afterTs := "2026-01-01T00:00:00Z"
			beforeTs := "2026-12-31T23:59:59Z"
			result, err := parseDateTimeFilter(&model.DateTimeFilter{After: &afterTs, Before: &beforeTs})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.After).To(Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
			Expect(result.Before).To(Equal(time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)))
		})
	})

	When("After is an invalid RFC3339 string", func() {
		It("returns an error", func() {
			invalid := "not-a-date"
			_, err := parseDateTimeFilter(&model.DateTimeFilter{After: &invalid})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("after"))
		})
	})

	When("Before is an invalid RFC3339 string", func() {
		It("returns an error", func() {
			invalid := "2026/01/01"
			_, err := parseDateTimeFilter(&model.DateTimeFilter{Before: &invalid})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("before"))
		})
	})
})

var _ = Describe("IssueMatchesOverdueBaseResolver", func() {
	var (
		mockApp *mocks.MockHeureka
		ctx     context.Context
	)

	BeforeEach(func() {
		mockApp = mocks.NewMockHeureka(GinkgoT())
		ctx = gqlgenCtx()
	})

	When("filter is nil", func() {
		It("sets Before to approximately now and forces the three non-remediated statuses", func() {
			beforeCall := time.Now().UTC()

			mockApp.On("ListIssueMatches", ctx, mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
				if f.TargetRemediationDate == nil {
					return false
				}

				if f.TargetRemediationDate.Before.IsZero() {
					return false
				}

				if f.TargetRemediationDate.Before.After(time.Now().UTC().Add(5 * time.Second)) {
					return false
				}

				if f.TargetRemediationDate.Before.Before(beforeCall.Add(-5 * time.Second)) {
					return false
				}

				expectedStatuses := []string{
					entity.IssueMatchStatusValuesNew.String(),
					entity.IssueMatchStatusValuesRiskAccepted.String(),
					entity.IssueMatchStatusValuesFalsePositive.String(),
				}
				if len(f.Status) != len(expectedStatuses) {
					return false
				}

				for i, s := range f.Status {
					if *s != expectedStatuses[i] {
						return false
					}
				}

				return true
			}), mock.Anything).Return(emptyIssueMatchList(), nil)

			result, err := IssueMatchesOverdueBaseResolver(mockApp, ctx, nil, nil, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})
	})

	When("filter has a pre-existing TargetRemediationDate.After", func() {
		It("preserves After but overwrites Before to approximately now", func() {
			afterTs := "2025-01-01T00:00:00Z"
			parsedAfter, _ := time.Parse(time.RFC3339, afterTs)

			mockApp.On("ListIssueMatches", ctx, mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
				if f.TargetRemediationDate == nil {
					return false
				}

				if !f.TargetRemediationDate.After.Equal(parsedAfter) {
					return false
				}

				if f.TargetRemediationDate.Before.IsZero() {
					return false
				}

				return !f.TargetRemediationDate.Before.After(time.Now().UTC().Add(5 * time.Second))
			}), mock.Anything).Return(emptyIssueMatchList(), nil)

			filter := &model.IssueMatchFilter{
				TargetRemediationDate: &model.DateTimeFilter{After: &afterTs},
			}
			result, err := IssueMatchesOverdueBaseResolver(mockApp, ctx, filter, nil, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})
	})

	When("filter has a pre-existing TargetRemediationDate.Before set to a future time", func() {
		It("overwrites Before to approximately now (ignores caller's future cutoff)", func() {
			futureTs := "2099-12-31T23:59:59Z"

			mockApp.On("ListIssueMatches", ctx, mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
				if f.TargetRemediationDate == nil {
					return false
				}
				// Before must NOT be the far-future value from caller
				parsed, _ := time.Parse(time.RFC3339, futureTs)

				return !f.TargetRemediationDate.Before.Equal(parsed) &&
					!f.TargetRemediationDate.Before.IsZero() &&
					!f.TargetRemediationDate.Before.After(time.Now().UTC().Add(5*time.Second))
			}), mock.Anything).Return(emptyIssueMatchList(), nil)

			filter := &model.IssueMatchFilter{
				TargetRemediationDate: &model.DateTimeFilter{Before: &futureTs},
			}
			result, err := IssueMatchesOverdueBaseResolver(mockApp, ctx, filter, nil, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})
	})

	When("filter has a non-nil Status list", func() {
		It("completely replaces Status with new, risk_accepted, false_positive", func() {
			mitigated := model.IssueMatchStatusValuesMitigated

			mockApp.On("ListIssueMatches", ctx, mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
				if len(f.Status) != 3 {
					return false
				}

				expected := []string{
					entity.IssueMatchStatusValuesNew.String(),
					entity.IssueMatchStatusValuesRiskAccepted.String(),
					entity.IssueMatchStatusValuesFalsePositive.String(),
				}
				for i, s := range f.Status {
					if *s != expected[i] {
						return false
					}
				}

				return true
			}), mock.Anything).Return(emptyIssueMatchList(), nil)

			filter := &model.IssueMatchFilter{
				Status: []*model.IssueMatchStatusValues{lo.ToPtr(mitigated)},
			}
			result, err := IssueMatchesOverdueBaseResolver(mockApp, ctx, filter, nil, nil, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())

			// Verify mitigated is not in the forwarded status
			mockApp.AssertExpectations(GinkgoT())
		})
	})

	When("the caller supplies an additional filter (first/after pagination)", func() {
		It("forwards pagination parameters alongside the overdue constraints", func() {
			first := 10
			after := pointer.String("cursor123")

			mockApp.On("ListIssueMatches", ctx, mock.MatchedBy(func(f *entity.IssueMatchFilter) bool {
				return f.Paginated.First != nil && *f.Paginated.First == 10 &&
					f.Paginated.After != nil && *f.Paginated.After == "cursor123"
			}), mock.Anything).Return(emptyIssueMatchList(), nil)

			result, err := IssueMatchesOverdueBaseResolver(mockApp, ctx, nil, &first, after, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})
	})
})
