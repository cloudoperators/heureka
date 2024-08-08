// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.49

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/baseResolver"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/model"
	"github.wdf.sap.corp/cc/heureka/internal/util"
)

// IssueMatch is the resolver for the issueMatch field.
func (r *issueMatchChangeResolver) IssueMatch(ctx context.Context, obj *model.IssueMatchChange) (*model.IssueMatch, error) {
	childIds, err := util.ConvertStrToIntSlice([]*string{obj.IssueMatchID})

	if err != nil {
		logrus.WithField("obj", obj).Error("IssueMatchChangeResolver: Error while parsing childIds'")
		return nil, err
	}

	return baseResolver.SingleIssueMatchBaseResolver(
		r.App,
		ctx,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.IssueMatchChangeNodeName,
			ChildIds:   childIds,
		})
}

// Activity is the resolver for the activity field.
func (r *issueMatchChangeResolver) Activity(ctx context.Context, obj *model.IssueMatchChange) (*model.Activity, error) {
	childIds, err := util.ConvertStrToIntSlice([]*string{obj.ActivityID})

	if err != nil {
		logrus.WithField("obj", obj).Error("IssueMatchChangeResolver: Error while parsing childIds'")
		return nil, err
	}

	return baseResolver.SingleActivityBaseResolver(
		r.App,
		ctx,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.IssueMatchChangeNodeName,
			ChildIds:   childIds,
		})
}

// IssueMatchChange returns graph.IssueMatchChangeResolver implementation.
func (r *Resolver) IssueMatchChange() graph.IssueMatchChangeResolver {
	return &issueMatchChangeResolver{r}
}

type issueMatchChangeResolver struct{ *Resolver }
