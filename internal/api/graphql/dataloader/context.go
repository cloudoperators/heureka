// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package dataloader

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cloudoperators/heureka/internal/app"
)

type contextKey struct{}

func ToContext(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, contextKey{}, loaders)
}

func FromContext(ctx context.Context) *Loaders {
	return ctx.Value(contextKey{}).(*Loaders)
}

func Middleware(a app.Heureka) func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		ctx = ToContext(ctx, NewLoaders(a))
		return next(ctx)
	}
}
