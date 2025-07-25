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

// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

func (r *issueMatchResolver) EffectiveIssueVariants(ctx context.Context, obj *model.IssueMatch, filter *model.IssueVariantFilter, first *int, after *string) (*model.IssueVariantConnection, error) {
	return baseResolver.EffectiveIssueVariantBaseResolver(
		r.App,
		ctx,
		filter,
		first,
		after,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.IssueMatchNodeName,
		})
}

func (r *issueMatchResolver) Evidences(ctx context.Context, obj *model.IssueMatch, filter *model.EvidenceFilter, first *int, after *string) (*model.EvidenceConnection, error) {
	return baseResolver.EvidenceBaseResolver(
		r.App,
		ctx,
		filter,
		first,
		after,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.IssueMatchNodeName,
		})
}

func (r *issueMatchResolver) Issue(ctx context.Context, obj *model.IssueMatch) (*model.Issue, error) {
	childIds, err := util.ConvertStrToIntSlice([]*string{obj.IssueID})

	if err != nil {
		logrus.WithField("obj", obj).Error("IssueMatchResolver: Error while parsing childIds'")
		return nil, err
	}

	return baseResolver.SingleIssueBaseResolver(
		r.App,
		ctx,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.IssueMatchNodeName,
			ChildIds:   childIds,
		})
}

func (r *issueMatchResolver) ComponentInstance(ctx context.Context, obj *model.IssueMatch) (*model.ComponentInstance, error) {
	childIds, err := util.ConvertStrToIntSlice([]*string{obj.ComponentInstanceID})

	if err != nil {
		logrus.WithField("obj", obj).Error("IssueMatchResolver: Error while parsing childIds'")
		return nil, err
	}

	return baseResolver.SingleComponentInstanceBaseResolver(
		r.App,
		ctx,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.IssueMatchNodeName,
			ChildIds:   childIds,
		})
}

func (r *issueMatchResolver) IssueMatchChanges(ctx context.Context, obj *model.IssueMatch, filter *model.IssueMatchChangeFilter, first *int, after *string) (*model.IssueMatchChangeConnection, error) {
	return baseResolver.IssueMatchChangeBaseResolver(
		r.App,
		ctx,
		filter,
		first,
		after,
		&model.NodeParent{
			Parent:     obj,
			ParentName: model.IssueMatchNodeName,
		})
}

func (r *Resolver) IssueMatch() graph.IssueMatchResolver { return &issueMatchResolver{r} }

type issueMatchResolver struct{ *Resolver }
