// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

// Characterization tests for the shared single-node base resolver skeleton
// (issue #1301, Phase 1). These pin the three control-flow branches
// (no-parent, multiple-results, found/not-found) AND the exact client-visible
// error strings for BOTH coexisting error styles, so the generic
// singleNodeByChildIds helper is proven behavior-preserving:
//   - SingleServiceBaseResolver uses the old NewResolverError style
//     ("<Op>: <msg>", resolver name leaked),
//   - SingleIssueBaseResolver uses the structured ToGraphQLError style
//     (sanitized message; note the current E(...) arg order makes the
//     "No parent provided" string an ID, so InvalidArgument sanitizes to
//     "Invalid issue" and Internal to the generic internal message).

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mock "github.com/stretchr/testify/mock"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/mocks"
)

// gqlTestContext builds the minimal gqlgen operation+field context that
// GetPreloads requires. A zero-value OperationContext has a nil-safe
// collect-fields cache and a zero FieldContext has nil Selections, so field
// collection returns an empty set (no preloads) without panicking.
func gqlTestContext() context.Context {
	ctx := graphql.WithOperationContext(context.Background(), &graphql.OperationContext{})
	ctx = graphql.WithFieldContext(ctx, &graphql.FieldContext{})

	return ctx
}

func int64Ptr(i int64) *int64 { return &i }

// newParent builds a NodeParent carrying only ChildIds, which is all the
// Single* resolvers read.
func newParent(childIds []*int64) *model.NodeParent {
	return &model.NodeParent{ChildIds: childIds}
}

var _ = Describe("SingleServiceBaseResolver (old error style)", func() {
	var (
		mockApp *mocks.MockHeureka
		ctx     context.Context
	)

	BeforeEach(func() {
		mockApp = mocks.NewMockHeureka(GinkgoT())
		ctx = gqlTestContext()
	})

	It("returns a tagged bad-request error when parent is nil", func() {
		result, err := SingleServiceBaseResolver(mockApp, ctx, nil)

		Expect(result).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("SingleServiceBaseResolver: Bad Request - No parent provided"))
	})

	It("returns a tagged internal error when multiple services are found", func() {
		mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
			Return(&entity.List[entity.ServiceResult]{
				Elements: []entity.ServiceResult{
					{Service: &entity.Service{BaseService: entity.BaseService{Id: 1}}},
					{Service: &entity.Service{BaseService: entity.BaseService{Id: 2}}},
				},
			}, nil)

		result, err := SingleServiceBaseResolver(mockApp, ctx, newParent([]*int64{int64Ptr(1)}))

		Expect(result).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("SingleServiceBaseResolver: Internal Error - found multiple services"))
	})

	It("wraps a list error with the resolver tag", func() {
		mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
			Return((*entity.List[entity.ServiceResult])(nil), errors.New("db down"))

		result, err := SingleServiceBaseResolver(mockApp, ctx, newParent([]*int64{int64Ptr(1)}))

		Expect(result).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("SingleServiceBaseResolver: db down"))
	})

	It("returns nil, nil when no service is found", func() {
		mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
			Return(&entity.List[entity.ServiceResult]{Elements: []entity.ServiceResult{}}, nil)

		result, err := SingleServiceBaseResolver(mockApp, ctx, newParent([]*int64{int64Ptr(1)}))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
	})

	It("returns the single service on exactly one result", func() {
		mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
			Return(&entity.List[entity.ServiceResult]{
				Elements: []entity.ServiceResult{
					{Service: &entity.Service{BaseService: entity.BaseService{Id: 7, CCRN: "svc"}}},
				},
			}, nil)

		result, err := SingleServiceBaseResolver(mockApp, ctx, newParent([]*int64{int64Ptr(1)}))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		Expect(result.ID).To(Equal("7"))
	})
})

var _ = Describe("SingleIssueBaseResolver (structured error style)", func() {
	var (
		mockApp *mocks.MockHeureka
		ctx     context.Context
	)

	BeforeEach(func() {
		mockApp = mocks.NewMockHeureka(GinkgoT())
		ctx = gqlTestContext()
	})

	It("sanitizes the no-parent error to 'Invalid issue' (current behavior)", func() {
		result, err := SingleIssueBaseResolver(mockApp, ctx, nil)

		Expect(result).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Invalid issue"))
	})

	It("returns the generic internal message when multiple issues are found", func() {
		mockApp.On("ListIssues", ctx, mock.Anything, mock.Anything).
			Return(issueListOf(1, 2), nil)

		result, err := SingleIssueBaseResolver(mockApp, ctx, newParent([]*int64{int64Ptr(1)}))

		Expect(result).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("An internal server error occurred"))
	})

	It("returns nil, nil when no issue is found", func() {
		mockApp.On("ListIssues", ctx, mock.Anything, mock.Anything).
			Return(issueListOf(), nil)

		result, err := SingleIssueBaseResolver(mockApp, ctx, newParent([]*int64{int64Ptr(1)}))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
	})

	It("returns the single issue on exactly one result", func() {
		mockApp.On("ListIssues", ctx, mock.Anything, mock.Anything).
			Return(issueListOf(42), nil)

		result, err := SingleIssueBaseResolver(mockApp, ctx, newParent([]*int64{int64Ptr(1)}))

		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		Expect(result.ID).To(Equal("42"))
	})
})

func issueListOf(ids ...int64) *entity.IssueList {
	elements := make([]entity.IssueResult, 0, len(ids))
	for _, id := range ids {
		elements = append(elements, entity.IssueResult{Issue: &entity.Issue{Id: id}})
	}

	return &entity.IssueList{List: &entity.List[entity.IssueResult]{Elements: elements}}
}
