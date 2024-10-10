// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.54

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/baseResolver"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
)

// ServiceCcrn is the resolver for the serviceCcrn field.
func (r *serviceFilterValueResolver) ServiceCcrn(ctx context.Context, obj *model.ServiceFilterValue, filter *model.ServiceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.ServiceCcrnBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ServiceFilterServiceCcrn
	return item, err
}

// UniqueUserID is the resolver for the uniqueUserId field.
func (r *serviceFilterValueResolver) UniqueUserID(ctx context.Context, obj *model.ServiceFilterValue, filter *model.UserFilter) (*model.FilterItem, error) {
	item, err := baseResolver.UniqueUserIDBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ServiceFilterUniqueUserId
	return item, err
}

// UserName is the resolver for the userName field.
func (r *serviceFilterValueResolver) UserName(ctx context.Context, obj *model.ServiceFilterValue, filter *model.UserFilter) (*model.FilterItem, error) {
	item, err := baseResolver.UserNameBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ServiceFilterUserName
	return item, err
}

// SupportGroupName is the resolver for the supportGroupName field.
func (r *serviceFilterValueResolver) SupportGroupName(ctx context.Context, obj *model.ServiceFilterValue, filter *model.SupportGroupFilter) (*model.FilterItem, error) {
	item, err := baseResolver.SupportGroupNameBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ServiceFilterSupportGroupName
	return item, err
}

// ServiceFilterValue returns graph.ServiceFilterValueResolver implementation.
func (r *Resolver) ServiceFilterValue() graph.ServiceFilterValueResolver {
	return &serviceFilterValueResolver{r}
}

type serviceFilterValueResolver struct{ *Resolver }
