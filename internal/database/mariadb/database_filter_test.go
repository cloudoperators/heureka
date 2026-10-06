// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("buildTimeRangeFilterQuery", func() {
	const col = "IM.issuematch_target_remediation_date"

	When("filter slice is nil or empty", func() {
		It("returns an empty string for nil", func() {
			Expect(buildTimeRangeFilterQuery(nil, col)).To(Equal(""))
		})
		It("returns an empty string for empty slice", func() {
			Expect(buildTimeRangeFilterQuery([]any{}, col)).To(Equal(""))
		})
	})

	When("filter[0] is nil", func() {
		It("returns an empty string", func() {
			Expect(buildTimeRangeFilterQuery([]any{nil}, col)).To(Equal(""))
		})
	})

	When("filter[0] is a *TimeFilter with only After set", func() {
		It("returns a single >= clause", func() {
			tf := &entity.TimeFilter{After: time.Now()}
			result := buildTimeRangeFilterQuery([]any{tf}, col)
			Expect(result).To(Equal(fmt.Sprintf("%s >= ?", col)))
		})
	})

	When("filter[0] is a *TimeFilter with only Before set", func() {
		It("returns a single <= clause", func() {
			tf := &entity.TimeFilter{Before: time.Now()}
			result := buildTimeRangeFilterQuery([]any{tf}, col)
			Expect(result).To(Equal(fmt.Sprintf("%s <= ?", col)))
		})
	})

	When("filter[0] is a *TimeFilter with both After and Before set", func() {
		It("returns a >= AND <= clause", func() {
			tf := &entity.TimeFilter{
				After:  time.Now().Add(-24 * time.Hour),
				Before: time.Now(),
			}
			result := buildTimeRangeFilterQuery([]any{tf}, col)
			Expect(result).To(Equal(fmt.Sprintf("%s >= ? AND %s <= ?", col, col)))
		})
	})

	When("filter[0] is a *TimeFilter with both fields at zero value", func() {
		It("returns an empty string", func() {
			tf := &entity.TimeFilter{}
			result := buildTimeRangeFilterQuery([]any{tf}, col)
			Expect(result).To(Equal(""))
		})
	})
})
