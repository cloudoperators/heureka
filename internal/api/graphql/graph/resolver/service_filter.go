// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.49

import (
	"context"

	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/baseResolver"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/model"
)

// ServiceName is the resolver for the serviceName field.
func (r *serviceFilterValueResolver) ServiceName(ctx context.Context, obj *model.ServiceFilterValue, filter *model.ServiceFilter) (*model.FilterItem, error) {
	return baseResolver.ServiceNameBaseResolver(r.App, ctx, filter)
}

// UniqueUserID is the resolver for the uniqueUserId field.
func (r *serviceFilterValueResolver) UniqueUserID(ctx context.Context, obj *model.ServiceFilterValue, filter *model.UserFilter) (*model.FilterItem, error) {
	return baseResolver.UniqueUserIDBaseResolver(r.App, ctx, filter)
}

// UserName is the resolver for the userName field.
func (r *serviceFilterValueResolver) UserName(ctx context.Context, obj *model.ServiceFilterValue, filter *model.UserFilter) (*model.FilterItem, error) {
	return baseResolver.UserNameBaseResolver(r.App, ctx, filter)
}

// SupportGroupName is the resolver for the supportGroupName field.
func (r *serviceFilterValueResolver) SupportGroupName(ctx context.Context, obj *model.ServiceFilterValue, filter *model.SupportGroupFilter) (*model.FilterItem, error) {
	return baseResolver.SupportGroupNameBaseResolver(r.App, ctx, filter)
}

// ServiceFilterValue returns graph.ServiceFilterValueResolver implementation.
func (r *Resolver) ServiceFilterValue() graph.ServiceFilterValueResolver {
	return &serviceFilterValueResolver{r}
}

type serviceFilterValueResolver struct{ *Resolver }
