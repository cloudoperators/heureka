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
	"github.com/samber/lo"
	"k8s.io/utils/pointer"
)

// Issues is the resolver for the Issues field.
func (r *queryResolver) Issues(ctx context.Context, filter *model.IssueFilter, first *int, after *string) (*model.IssueConnection, error) {
	return baseResolver.IssueBaseResolver(r.App, ctx, filter, first, after, nil)
}

// IssueMatches is the resolver for the IssueMatches field.
func (r *queryResolver) IssueMatches(ctx context.Context, filter *model.IssueMatchFilter, first *int, after *string) (*model.IssueMatchConnection, error) {
	return baseResolver.IssueMatchBaseResolver(r.App, ctx, filter, first, after, nil)
}

// IssueMatchChanges is the resolver for the IssueMatchChanges field.
func (r *queryResolver) IssueMatchChanges(ctx context.Context, filter *model.IssueMatchChangeFilter, first *int, after *string) (*model.IssueMatchChangeConnection, error) {
	return baseResolver.IssueMatchChangeBaseResolver(r.App, ctx, filter, first, after, nil)
}

// Services is the resolver for the Services field.
func (r *queryResolver) Services(ctx context.Context, filter *model.ServiceFilter, first *int, after *string) (*model.ServiceConnection, error) {
	return baseResolver.ServiceBaseResolver(r.App, ctx, filter, first, after, nil)
}

// Components is the resolver for the Components field.
func (r *queryResolver) Components(ctx context.Context, filter *model.ComponentFilter, first *int, after *string) (*model.ComponentConnection, error) {
	return baseResolver.ComponentBaseResolver(r.App, ctx, filter, first, after, nil)
}

// ComponentVersions is the resolver for the ComponentVersions field.
func (r *queryResolver) ComponentVersions(ctx context.Context, filter *model.ComponentVersionFilter, first *int, after *string) (*model.ComponentVersionConnection, error) {
	return baseResolver.ComponentVersionBaseResolver(r.App, ctx, filter, first, after, nil)
}

// ComponentInstances is the resolver for the ComponentInstances field.
func (r *queryResolver) ComponentInstances(ctx context.Context, filter *model.ComponentInstanceFilter, first *int, after *string) (*model.ComponentInstanceConnection, error) {
	return baseResolver.ComponentInstanceBaseResolver(r.App, ctx, filter, first, after, nil)
}

// Activities is the resolver for the Activities field.
func (r *queryResolver) Activities(ctx context.Context, filter *model.ActivityFilter, first *int, after *string) (*model.ActivityConnection, error) {
	return baseResolver.ActivityBaseResolver(r.App, ctx, filter, first, after, nil)
}

// IssueVariants is the resolver for the IssueVariants field.
func (r *queryResolver) IssueVariants(ctx context.Context, filter *model.IssueVariantFilter, first *int, after *string) (*model.IssueVariantConnection, error) {
	return baseResolver.IssueVariantBaseResolver(r.App, ctx, filter, first, after, nil)
}

// IssueRepositories is the resolver for the IssueRepositories field.
func (r *queryResolver) IssueRepositories(ctx context.Context, filter *model.IssueRepositoryFilter, first *int, after *string) (*model.IssueRepositoryConnection, error) {
	return baseResolver.IssueRepositoryBaseResolver(r.App, ctx, filter, first, after, nil)
}

// Evidences is the resolver for the Evidences field.
func (r *queryResolver) Evidences(ctx context.Context, filter *model.EvidenceFilter, first *int, after *string) (*model.EvidenceConnection, error) {
	return baseResolver.EvidenceBaseResolver(r.App, ctx, filter, first, after, nil)
}

// SupportGroups is the resolver for the SupportGroups field.
func (r *queryResolver) SupportGroups(ctx context.Context, filter *model.SupportGroupFilter, first *int, after *string) (*model.SupportGroupConnection, error) {
	return baseResolver.SupportGroupBaseResolver(r.App, ctx, filter, first, after, nil)
}

// Users is the resolver for the Users field.
func (r *queryResolver) Users(ctx context.Context, filter *model.UserFilter, first *int, after *string) (*model.UserConnection, error) {
	return baseResolver.UserBaseResolver(r.App, ctx, filter, first, after, nil)
}

// ServiceFilterValues is the resolver for the ServiceFilterValues field.
func (r *queryResolver) ServiceFilterValues(ctx context.Context) (*model.ServiceFilterValue, error) {
	return &model.ServiceFilterValue{}, nil
}

// IssueMatchFilterValues is the resolver for the IssueMatchFilterValues field.
func (r *queryResolver) IssueMatchFilterValues(ctx context.Context) (*model.IssueMatchFilterValue, error) {
	return &model.IssueMatchFilterValue{
		Status: &model.FilterItem{
			DisplayName: &baseResolver.FilterDisplayIssueMatchStatus,
			FilterName:  &baseResolver.IssueMatchFilterStatus,
			Values:      lo.Map(model.AllIssueMatchStatusValues, func(item model.IssueMatchStatusValues, _ int) *string { return pointer.String(item.String()) }),
		},
		IssueType: &model.FilterItem{
			DisplayName: &baseResolver.FilterDisplayIssueType,
			FilterName:  &baseResolver.IssueMatchFilterIssueType,
			Values:      lo.Map(model.AllIssueTypes, func(item model.IssueTypes, _ int) *string { return pointer.String(item.String()) }),
		},
		Severity: &model.FilterItem{
			DisplayName: &baseResolver.FilterDisplayIssueSeverity,
			FilterName:  &baseResolver.IssueMatchFilterSeverity,
			Values:      lo.Map(model.AllSeverityValues, func(item model.SeverityValues, _ int) *string { return pointer.String(item.String()) }),
		},
	}, nil
}

// Query returns graph.QueryResolver implementation.
func (r *Resolver) Query() graph.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
