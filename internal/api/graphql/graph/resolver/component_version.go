package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/baseResolver"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"
)

// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

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

func (r *componentVersionResolver) Issues(ctx context.Context, obj *model.ComponentVersion, first *int, after *string) (*model.IssueConnection, error) {
	return baseResolver.IssueBaseResolver(r.App, ctx, nil, first, after, &model.NodeParent{
		Parent:     obj,
		ParentName: model.ComponentVersionNodeName,
	})
}

func (r *componentVersionResolver) ComponentInstances(ctx context.Context, obj *model.ComponentVersion, filter *model.ComponentInstanceFilter, first *int, after *string, orderBy []*model.ComponentInstanceOrderBy) (*model.ComponentInstanceConnection, error) {
	return baseResolver.ComponentInstanceBaseResolver(r.App, ctx, filter, first, after, orderBy, &model.NodeParent{
		Parent:     obj,
		ParentName: model.ComponentVersionNodeName,
	})
}

func (r *componentVersionResolver) IssueCounts(ctx context.Context, obj *model.ComponentVersion, filter *model.IssueFilter) (*model.SeverityCounts, error) {
	return baseResolver.IssueCountsBaseResolver(r.App, ctx, filter, &model.NodeParent{
		Parent:     obj,
		ParentName: model.ComponentVersionNodeName,
	})
}

func (r *Resolver) ComponentVersion() graph.ComponentVersionResolver {
	return &componentVersionResolver{r}
}

type componentVersionResolver struct{ *Resolver }
