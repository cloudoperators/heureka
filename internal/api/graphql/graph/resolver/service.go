// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.55

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/baseResolver"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
)

// Owners is the resolver for the owners field.
func (r *serviceResolver) Owners(ctx context.Context, obj *model.Service, filter *model.UserFilter, first *int, after *string) (*model.UserConnection, error) {
	return baseResolver.UserBaseResolver(r.App, ctx, filter, first, after,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.ServiceNodeName,
		})
}

// SupportGroups is the resolver for the supportGroups field.
func (r *serviceResolver) SupportGroups(ctx context.Context, obj *model.Service, filter *model.SupportGroupFilter, first *int, after *string) (*model.SupportGroupConnection, error) {
	return baseResolver.SupportGroupBaseResolver(r.App, ctx, filter, first, after,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.ServiceNodeName,
		})
}

// Activities is the resolver for the activities field.
func (r *serviceResolver) Activities(ctx context.Context, obj *model.Service, filter *model.ActivityFilter, first *int, after *string) (*model.ActivityConnection, error) {
	return baseResolver.ActivityBaseResolver(r.App, ctx, filter, first, after,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.ServiceNodeName,
		})
}

// IssueRepositories is the resolver for the issueRepositories field.
func (r *serviceResolver) IssueRepositories(ctx context.Context, obj *model.Service, filter *model.IssueRepositoryFilter, first *int, after *string) (*model.IssueRepositoryConnection, error) {
	return baseResolver.IssueRepositoryBaseResolver(r.App, ctx, filter, first, after,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.ServiceNodeName,
		})
}

// ComponentInstances is the resolver for the componentInstances field.
func (r *serviceResolver) ComponentInstances(ctx context.Context, obj *model.Service, filter *model.ComponentInstanceFilter, first *int, after *string) (*model.ComponentInstanceConnection, error) {
	return baseResolver.ComponentInstanceBaseResolver(r.App, ctx, filter, first, after,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.ServiceNodeName,
		})
}

// Service returns graph.ServiceResolver implementation.
func (r *Resolver) Service() graph.ServiceResolver { return &serviceResolver{r} }

type serviceResolver struct{ *Resolver }
