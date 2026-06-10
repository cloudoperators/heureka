// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package querycounter_test

import (
	"context"
	"sync"
	"testing"

	"github.com/cloudoperators/heureka/internal/database/querycounter"
)

func TestGetQueryCount_ReturnsZeroWithoutInit(t *testing.T) {
	ctx := context.Background()
	if got := querycounter.GetQueryCount(ctx); got != 0 {
		t.Errorf("GetQueryCount() = %d, want 0", got)
	}
}

func TestIncrement_NoOpWithoutInit(t *testing.T) {
	ctx := context.Background()
	// Should not panic
	querycounter.Increment(ctx)

	if got := querycounter.GetQueryCount(ctx); got != 0 {
		t.Errorf("GetQueryCount() = %d, want 0", got)
	}
}

func TestInitAndIncrement(t *testing.T) {
	ctx := querycounter.Init(context.Background())

	if got := querycounter.GetQueryCount(ctx); got != 0 {
		t.Fatalf("initial GetQueryCount() = %d, want 0", got)
	}

	querycounter.Increment(ctx)
	querycounter.Increment(ctx)
	querycounter.Increment(ctx)

	if got := querycounter.GetQueryCount(ctx); got != 3 {
		t.Errorf("GetQueryCount() = %d, want 3", got)
	}
}

func TestIncrement_ConcurrentSafety(t *testing.T) {
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
	if got := querycounter.GetQueryCount(ctx); got != expected {
		t.Errorf("GetQueryCount() = %d, want %d", got, expected)
	}
}

func TestSeparateContextsAreIndependent(t *testing.T) {
	ctx1 := querycounter.Init(context.Background())
	ctx2 := querycounter.Init(context.Background())

	querycounter.Increment(ctx1)
	querycounter.Increment(ctx1)
	querycounter.Increment(ctx2)

	if got := querycounter.GetQueryCount(ctx1); got != 2 {
		t.Errorf("ctx1 GetQueryCount() = %d, want 2", got)
	}

	if got := querycounter.GetQueryCount(ctx2); got != 1 {
		t.Errorf("ctx2 GetQueryCount() = %d, want 1", got)
	}
}
