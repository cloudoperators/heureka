// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

// Characterization tests for the list base resolvers and the shared buildEdges
// helper (issue #1301, Phase 1). Run with an empty preload set (gqlTestContext),
// so the Service batch-preload and top-level issueCounts branches are inactive
// here; they are exercised by the DB-bound e2e suite. These pin the connection
// assembly that the edge-loop extraction must preserve: edge count/order, node
// IDs, the non-nil-empty edges slice, TotalCount defaulting, and the
// edges.priority gate being OFF when the field is not requested.

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mock "github.com/stretchr/testify/mock"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/mocks"
)

var _ = Describe("buildEdges", func() {
	It("returns a non-nil empty slice for empty input", func() {
		edges := buildEdges([]int{}, func(i int) *int { return &i })
		Expect(edges).NotTo(BeNil())
		Expect(edges).To(HaveLen(0))
	})

	It("maps each element in order", func() {
		edges := buildEdges([]int{1, 2, 3}, func(i int) int { return i * 10 })
		Expect(edges).To(Equal([]int{10, 20, 30}))
	})
})

var _ = Describe("ServiceBaseResolver (connection assembly)", func() {
	var (
		mockApp *mocks.MockHeureka
		ctx     context.Context
	)

	BeforeEach(func() {
		mockApp = mocks.NewMockHeureka(GinkgoT())
		ctx = gqlTestContext()
	})

	It("builds one edge per element with ids, no priority, and the given total count", func() {
		total := int64(2)
		mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
			Return(&entity.List[entity.ServiceResult]{
				TotalCount: &total,
				Elements: []entity.ServiceResult{
					{Service: &entity.Service{BaseService: entity.BaseService{Id: 1, CCRN: "svc-1"}}},
					{Service: &entity.Service{BaseService: entity.BaseService{Id: 2, CCRN: "svc-2"}}},
				},
			}, nil)

		conn, err := ServiceBaseResolver(mockApp, ctx, nil, nil, nil, nil, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(conn).NotTo(BeNil())
		Expect(conn.TotalCount).To(Equal(2))
		Expect(conn.Edges).To(HaveLen(2))
		Expect(conn.Edges[0].Node.ID).To(Equal("1"))
		Expect(conn.Edges[1].Node.ID).To(Equal("2"))
		// edges.priority not requested -> must stay nil
		Expect(conn.Edges[0].Priority).To(BeNil())
		// issueCounts not requested -> must stay nil
		Expect(conn.IssueCounts).To(BeNil())
	})

	It("returns a non-nil empty edges slice and total count 0 for no results", func() {
		mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
			Return(&entity.List[entity.ServiceResult]{Elements: []entity.ServiceResult{}}, nil)

		conn, err := ServiceBaseResolver(mockApp, ctx, nil, nil, nil, nil, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(conn.TotalCount).To(Equal(0))
		Expect(conn.Edges).NotTo(BeNil())
		Expect(conn.Edges).To(HaveLen(0))
	})

	It("wraps a list error with the resolver tag", func() {
		mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
			Return((*entity.List[entity.ServiceResult])(nil), context.DeadlineExceeded)

		conn, err := ServiceBaseResolver(mockApp, ctx, nil, nil, nil, nil, nil)

		Expect(conn).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("ServiceBaseResolver: " + context.DeadlineExceeded.Error()))
	})
})

var _ = Describe("ComponentInstanceBaseResolver (connection assembly)", func() {
	var (
		mockApp *mocks.MockHeureka
		ctx     context.Context
	)

	BeforeEach(func() {
		mockApp = mocks.NewMockHeureka(GinkgoT())
		ctx = gqlTestContext()
	})

	It("builds one edge per element with ids and the given total count", func() {
		total := int64(1)
		mockApp.On("ListComponentInstances", ctx, mock.Anything, mock.Anything).
			Return(&entity.List[entity.ComponentInstanceResult]{
				TotalCount: &total,
				Elements: []entity.ComponentInstanceResult{
					{ComponentInstance: &entity.ComponentInstance{Id: 5, CCRN: "ci-5"}},
				},
			}, nil)

		conn, err := ComponentInstanceBaseResolver(mockApp, ctx, nil, nil, nil, nil, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(conn).NotTo(BeNil())
		Expect(conn.TotalCount).To(Equal(1))
		Expect(conn.Edges).To(HaveLen(1))
		Expect(conn.Edges[0].Node.ID).To(Equal("5"))
	})
})
