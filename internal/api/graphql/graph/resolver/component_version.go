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
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"
)

// Component is the resolver for the component field.
func (r *componentVersionResolver) Component(ctx context.Context, obj *model.ComponentVersion) (*model.Component, error) {
	childIds, err := util.ConvertStrToIntSlice([]*string{obj.ComponentID})

	if err != nil {
		logrus.WithField("obj", obj).Error("ComponentVersionResolver: Error while parsing childIds'")
		return nil, err
	}

	return baseResolver.SingleComponentBaseResolver(
		r.App,
		ctx,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.IssueMatchNodeName,
			ChildIds:   childIds,
		})
}

// Issues is the resolver for the issues field.
func (r *componentVersionResolver) Issues(ctx context.Context, obj *model.ComponentVersion, first *int, after *string) (*model.IssueConnection, error) {
	return baseResolver.IssueBaseResolver(r.App, ctx, nil, first, after, &model.NodeParent{
		Parent:     obj,
		ParentName: model.ComponentVersionNodeName,
	})
}

// ComponentInstances is the resolver for the componentInstances field.
func (r *componentVersionResolver) ComponentInstances(ctx context.Context, obj *model.ComponentVersion, first *int, after *string) (*model.ComponentInstanceConnection, error) {
	return baseResolver.ComponentInstanceBaseResolver(r.App, ctx, nil, first, after, &model.NodeParent{
		Parent:     obj,
		ParentName: model.ComponentVersionNodeName,
	})
}

// ComponentVersion returns graph.ComponentVersionResolver implementation.
func (r *Resolver) ComponentVersion() graph.ComponentVersionResolver {
	return &componentVersionResolver{r}
}

type componentVersionResolver struct{ *Resolver }
