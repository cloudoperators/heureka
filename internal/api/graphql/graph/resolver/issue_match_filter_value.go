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

// PrimaryName is the resolver for the primaryName field.
func (r *issueMatchFilterValueResolver) PrimaryName(ctx context.Context, obj *model.IssueMatchFilterValue, filter *model.IssueFilter) (*model.FilterItem, error) {
	item, err := baseResolver.IssueNameBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.IssueMatchFilterPrimaryName
	return item, nil
}

// AffectedService is the resolver for the affectedService field.
func (r *issueMatchFilterValueResolver) AffectedService(ctx context.Context, obj *model.IssueMatchFilterValue, filter *model.ServiceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.ServiceCcrnBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.IssueMatchFilterAffectedService
	return item, nil
}

// ComponentName is the resolver for the componentName field.
func (r *issueMatchFilterValueResolver) ComponentName(ctx context.Context, obj *model.IssueMatchFilterValue, filter *model.ComponentFilter) (*model.FilterItem, error) {
	item, err := baseResolver.ComponentNameBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.IssueMatchFilterComponentName
	return item, nil
}

// SupportGroupName is the resolver for the supportGroupName field.
func (r *issueMatchFilterValueResolver) SupportGroupName(ctx context.Context, obj *model.IssueMatchFilterValue, filter *model.SupportGroupFilter) (*model.FilterItem, error) {
	item, err := baseResolver.SupportGroupNameBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.IssueMatchFilterSupportGroupName
	return item, nil
}

// IssueMatchFilterValue returns graph.IssueMatchFilterValueResolver implementation.
func (r *Resolver) IssueMatchFilterValue() graph.IssueMatchFilterValueResolver {
	return &issueMatchFilterValueResolver{r}
}

type issueMatchFilterValueResolver struct{ *Resolver }
