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

// SupportGroups is the resolver for the supportGroups field.
func (r *userResolver) SupportGroups(ctx context.Context, obj *model.User, filter *model.SupportGroupFilter, first *int, after *string) (*model.SupportGroupConnection, error) {
	return baseResolver.SupportGroupBaseResolver(r.App, ctx, filter, first, after, &model.NodeParent{
		Parent:     obj,
		ParentName: model.UserNodeName,
	})
}

// Services is the resolver for the services field.
func (r *userResolver) Services(ctx context.Context, obj *model.User, filter *model.ServiceFilter, first *int, after *string) (*model.ServiceConnection, error) {
	return baseResolver.ServiceBaseResolver(r.App, ctx, filter, first, after, &model.NodeParent{
		Parent:     obj,
		ParentName: model.UserNodeName,
	})
}

// User returns graph.UserResolver implementation.
func (r *Resolver) User() graph.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
