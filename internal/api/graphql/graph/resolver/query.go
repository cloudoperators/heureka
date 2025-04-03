package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/baseResolver"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/samber/lo"
	"k8s.io/utils/pointer"
)

// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

func (r *queryResolver) Issues(ctx context.Context, filter *model.IssueFilter, first *int, after *string, orderBy []*model.IssueOrderBy) (*model.IssueConnection, error) {
	return baseResolver.IssueBaseResolver(r.App, ctx, filter, first, after, orderBy, nil)
}

func (r *queryResolver) IssueMatches(ctx context.Context, filter *model.IssueMatchFilter, first *int, after *string, orderBy []*model.IssueMatchOrderBy) (*model.IssueMatchConnection, error) {
	return baseResolver.IssueMatchBaseResolver(r.App, ctx, filter, first, after, orderBy, nil)
}

func (r *queryResolver) IssueMatchChanges(ctx context.Context, filter *model.IssueMatchChangeFilter, first *int, after *string) (*model.IssueMatchChangeConnection, error) {
	return baseResolver.IssueMatchChangeBaseResolver(r.App, ctx, filter, first, after, nil)
}

func (r *queryResolver) Services(ctx context.Context, filter *model.ServiceFilter, first *int, after *string, orderBy []*model.ServiceOrderBy) (*model.ServiceConnection, error) {
	return baseResolver.ServiceBaseResolver(r.App, ctx, filter, first, after, orderBy, nil)
}

func (r *queryResolver) Components(ctx context.Context, filter *model.ComponentFilter, first *int, after *string) (*model.ComponentConnection, error) {
	return baseResolver.ComponentBaseResolver(r.App, ctx, filter, first, after, nil)
}

func (r *queryResolver) ComponentVersions(ctx context.Context, filter *model.ComponentVersionFilter, first *int, after *string, orderBy []*model.ComponentVersionOrderBy) (*model.ComponentVersionConnection, error) {
	return baseResolver.ComponentVersionBaseResolver(r.App, ctx, filter, first, after, orderBy, nil)
}

func (r *queryResolver) ComponentInstances(ctx context.Context, filter *model.ComponentInstanceFilter, first *int, after *string, orderBy []*model.ComponentInstanceOrderBy) (*model.ComponentInstanceConnection, error) {
	return baseResolver.ComponentInstanceBaseResolver(r.App, ctx, filter, first, after, orderBy, nil)
}

func (r *queryResolver) Activities(ctx context.Context, filter *model.ActivityFilter, first *int, after *string) (*model.ActivityConnection, error) {
	return baseResolver.ActivityBaseResolver(r.App, ctx, filter, first, after, nil)
}

func (r *queryResolver) IssueVariants(ctx context.Context, filter *model.IssueVariantFilter, first *int, after *string) (*model.IssueVariantConnection, error) {
	return baseResolver.IssueVariantBaseResolver(r.App, ctx, filter, first, after, nil)
}

func (r *queryResolver) IssueRepositories(ctx context.Context, filter *model.IssueRepositoryFilter, first *int, after *string) (*model.IssueRepositoryConnection, error) {
	return baseResolver.IssueRepositoryBaseResolver(r.App, ctx, filter, first, after, nil)
}

func (r *queryResolver) Evidences(ctx context.Context, filter *model.EvidenceFilter, first *int, after *string) (*model.EvidenceConnection, error) {
	return baseResolver.EvidenceBaseResolver(r.App, ctx, filter, first, after, nil)
}

func (r *queryResolver) SupportGroups(ctx context.Context, filter *model.SupportGroupFilter, first *int, after *string) (*model.SupportGroupConnection, error) {
	return baseResolver.SupportGroupBaseResolver(r.App, ctx, filter, first, after, nil)
}

func (r *queryResolver) Users(ctx context.Context, filter *model.UserFilter, first *int, after *string) (*model.UserConnection, error) {
	return baseResolver.UserBaseResolver(r.App, ctx, filter, first, after, nil)
}

func (r *queryResolver) ServiceFilterValues(ctx context.Context) (*model.ServiceFilterValue, error) {
	return &model.ServiceFilterValue{}, nil
}

func (r *queryResolver) IssueMatchFilterValues(ctx context.Context) (*model.IssueMatchFilterValue, error) {
	return &model.IssueMatchFilterValue{
		Status: &model.FilterItem{
			DisplayName: &baseResolver.FilterDisplayIssueMatchStatus,
			FilterName:  &baseResolver.IssueMatchFilterStatus,
			Values:      lo.Map(model.AllIssueMatchStatusValuesOrdered, func(s model.IssueMatchStatusValues, _ int) *string { return pointer.String(s.String()) }),
		},
		IssueType: &model.FilterItem{
			DisplayName: &baseResolver.FilterDisplayIssueType,
			FilterName:  &baseResolver.IssueMatchFilterIssueType,
			Values:      lo.Map(model.AllIssueTypesOrdered, func(s model.IssueTypes, _ int) *string { return pointer.String(s.String()) }),
		},
		Severity: &model.FilterItem{
			DisplayName: &baseResolver.FilterDisplayIssueSeverity,
			FilterName:  &baseResolver.IssueMatchFilterSeverity,
			Values:      lo.Map(model.AllSeverityValuesOrdered, func(s model.SeverityValues, _ int) *string { return pointer.String(s.String()) }),
		},
	}, nil
}

func (r *queryResolver) ComponentInstanceFilterValues(ctx context.Context) (*model.ComponentInstanceFilterValue, error) {
	return &model.ComponentInstanceFilterValue{}, nil
}

func (r *queryResolver) ComponentFilterValues(ctx context.Context) (*model.ComponentFilterValue, error) {
	return &model.ComponentFilterValue{}, nil
}

func (r *queryResolver) ScannerRunTagFilterValues(ctx context.Context) ([]*string, error) {
	return baseResolver.ScannerRunTagFilterValues(r.App, ctx)
}

func (r *queryResolver) ScannerRuns(ctx context.Context, filter *model.ScannerRunFilter, first *int, after *string) (*model.ScannerRunConnection, error) {
	return baseResolver.ScannerRuns(r.App, ctx, filter, first, after)
}

func (r *queryResolver) IssueCounts(ctx context.Context, filter *model.IssueFilter) (*model.SeverityCounts, error) {
	return baseResolver.IssueCountsBaseResolver(r.App, ctx, filter, nil)
}

func (r *Resolver) Query() graph.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
