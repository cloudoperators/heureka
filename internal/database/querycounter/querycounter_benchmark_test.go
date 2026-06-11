// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package querycounter_test

import (
	"context"
	"testing"

	"github.com/cloudoperators/heureka/internal/database/querycounter"
)

func BenchmarkIncrement(b *testing.B) {
	ctx := querycounter.Init(context.Background())

	b.ResetTimer()

	for range b.N {
		querycounter.Increment(ctx)
	}
}

func BenchmarkIncrement_NoInit(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()

	for range b.N {
		querycounter.Increment(ctx)
	}
}

func BenchmarkGetQueryCount(b *testing.B) {
	ctx := querycounter.Init(context.Background())
	querycounter.Increment(ctx)
	b.ResetTimer()

	for range b.N {
		querycounter.GetQueryCount(ctx)
	}
}
