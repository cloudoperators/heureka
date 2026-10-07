// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

// Unit tests for the pure, context-free helpers extracted from the list base
// resolvers (issue #1301, Phase 1 Tier-1). These pin the behavior that the
// per-entity list resolvers depend on — especially the exact Service severity
// column ordering and the Issue severity tiebreaker — so later consolidation
// onto a generic list engine stays behavior-preserving.

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/entity"
)

func TestBaseResolver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BaseResolver Suite")
}

var _ = Describe("totalCountOf", func() {
	It("returns 0 for a nil pointer", func() {
		Expect(totalCountOf(nil)).To(Equal(0))
	})

	It("returns the dereferenced value for a non-nil pointer", func() {
		v := int64(42)
		Expect(totalCountOf(&v)).To(Equal(42))
	})

	It("returns 0 for a zero value", func() {
		v := int64(0)
		Expect(totalCountOf(&v)).To(Equal(0))
	})
})

var _ = Describe("toFilterItem", func() {
	display := "Service"

	It("returns nil Values for an empty input (preserving prior behavior)", func() {
		item := toFilterItem(nil, &display)
		Expect(item).NotTo(BeNil())
		Expect(item.DisplayName).To(Equal(&display))
		Expect(item.Values).To(BeNil())
	})

	It("maps each name to a distinct pointer in order", func() {
		item := toFilterItem([]string{"a", "b", "c"}, &display)
		Expect(item.Values).To(HaveLen(3))
		Expect(*item.Values[0]).To(Equal("a"))
		Expect(*item.Values[1]).To(Equal("b"))
		Expect(*item.Values[2]).To(Equal("c"))
		// Each element must point to its own backing value, not a shared loop var.
		Expect(item.Values[0]).NotTo(Equal(item.Values[1]))
		Expect(item.Values[1]).NotTo(Equal(item.Values[2]))
	})
})

var _ = Describe("appendServiceSeverityOrder", func() {
	It("appends exactly Critical, High, Medium, Low, None, ServiceId in the given direction", func() {
		dir := entity.OrderDirectionDesc
		order := appendServiceSeverityOrder(nil, dir)

		Expect(order).To(Equal([]entity.Order{
			{By: entity.CriticalCount, Direction: dir},
			{By: entity.HighCount, Direction: dir},
			{By: entity.MediumCount, Direction: dir},
			{By: entity.LowCount, Direction: dir},
			{By: entity.NoneCount, Direction: dir},
			{By: entity.ServiceId, Direction: dir},
		}))
	})

	It("appends onto an existing order slice without dropping prior entries", func() {
		dir := entity.OrderDirectionAsc
		existing := []entity.Order{{By: entity.ServiceCcrn, Direction: dir}}
		order := appendServiceSeverityOrder(existing, dir)

		Expect(order).To(HaveLen(7))
		Expect(order[0]).To(Equal(entity.Order{By: entity.ServiceCcrn, Direction: dir}))
		Expect(order[6]).To(Equal(entity.Order{By: entity.ServiceId, Direction: dir}))
	})
})

var _ = Describe("appendIssueSeverityOrder", func() {
	It("appends the mapped severity order followed by the IssueId tiebreaker", func() {
		field := model.IssueOrderByFieldSeverity
		dir := model.OrderDirectionDesc
		o := &model.IssueOrderBy{By: &field, Direction: &dir}

		order := appendIssueSeverityOrder(nil, o)

		Expect(order).To(HaveLen(2))
		Expect(order[0]).To(Equal(o.ToOrderEntity()))
		Expect(order[1]).To(Equal(entity.Order{
			By:        entity.IssueId,
			Direction: dir.ToOrderDirectionEntity(),
		}))
	})
})
