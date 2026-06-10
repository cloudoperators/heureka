// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package querycounter

import (
	"context"
	"sync/atomic"
)

type contextKey struct{}

// Init stores a new atomic counter in the context and returns the new context.
// Call this once per request (e.g., in middleware).
func Init(ctx context.Context) context.Context {
	counter := new(atomic.Int64)
	return context.WithValue(ctx, contextKey{}, counter)
}

// Increment atomically increments the query counter stored in ctx.
// If no counter was initialized (e.g., in tests without middleware), this is a no-op.
func Increment(ctx context.Context) {
	if c := fromContext(ctx); c != nil {
		c.Add(1)
	}
}

// GetQueryCount returns the current value of the query counter from the context.
// Returns 0 if no counter was initialized.
func GetQueryCount(ctx context.Context) int {
	if c := fromContext(ctx); c != nil {
		return int(c.Load())
	}

	return 0
}

func fromContext(ctx context.Context) *atomic.Int64 {
	v := ctx.Value(contextKey{})
	if v == nil {
		return nil
	}

	return v.(*atomic.Int64)
}
