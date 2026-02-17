// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
)

type BatchLimiter struct {
	batchLimit int
}

func NewBatchLimiterMiddleware(batchLimit int) BatchLimiter {
	return BatchLimiter{
		batchLimit: batchLimit,
	}
}

func (m *BatchLimiter) Middleware() func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		if len(graphql.GetOperationContext(ctx).Operation.SelectionSet) > m.batchLimit {
			return graphql.OneShot(graphql.ErrorResponse(ctx, "the limit for sending batches has been exceeded"))
		}

		return next(ctx)
	}
}
