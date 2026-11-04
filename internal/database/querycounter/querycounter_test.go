// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package querycounter_test

import (
	"context"
	"sync"

	"github.com/cloudoperators/heureka/internal/database/querycounter"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Query Counter", Label("QueryCount"), func() {
	When("GetQueryCount without init", func() {
		It("Returns zero", func() {
			ctx := context.Background()
			Expect(querycounter.GetQueryCount(ctx)).To(BeEquivalentTo(0))
		})
	})
	When("GetQueryCount with increment without init", func() {
		It("Returns zero", func() {
			ctx := context.Background()
			// Should not panic
			querycounter.Increment(ctx)
			Expect(querycounter.GetQueryCount(ctx)).To(BeEquivalentTo(0))
		})
	})
	When("GetQueryCount with increments after init", func() {
		It("Returns number equal to number of increments after init", func() {
			ctx := querycounter.Init(context.Background())

			Expect(querycounter.GetQueryCount(ctx)).To(BeEquivalentTo(0))

			querycounter.Increment(ctx)
			querycounter.Increment(ctx)
			querycounter.Increment(ctx)
			Expect(querycounter.GetQueryCount(ctx)).To(BeEquivalentTo(3))
		})
	})
	When("GetQueryCount with concurrent increments", func() {
		It("Returns number equal to number of increments ensuring concurrent safety", func() {
			ctx := querycounter.Init(context.Background())

			const goroutines = 100

			const incrementsPerGoroutine = 1000

			var wg sync.WaitGroup
			wg.Add(goroutines)

			for range goroutines {
				go func() {
					defer wg.Done()

					for range incrementsPerGoroutine {
						querycounter.Increment(ctx)
					}
				}()
			}

			wg.Wait()

			expected := goroutines * incrementsPerGoroutine
			Expect(querycounter.GetQueryCount(ctx)).To(BeEquivalentTo(expected))
		})
	})
	When("GetQueryCount with two separate contexts", func() {
		It("Returns number equal to number of increments per specific context", func() {
			ctx1 := querycounter.Init(context.Background())
			ctx2 := querycounter.Init(context.Background())

			querycounter.Increment(ctx1)
			querycounter.Increment(ctx1)
			querycounter.Increment(ctx2)

			Expect(querycounter.GetQueryCount(ctx1)).To(BeEquivalentTo(2))
			Expect(querycounter.GetQueryCount(ctx2)).To(BeEquivalentTo(1))
		})
	})
})
