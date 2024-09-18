// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.49

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/baseResolver"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"
)

// IssueRepository is the resolver for the issueRepository field.
func (r *issueVariantResolver) IssueRepository(ctx context.Context, obj *model.IssueVariant) (*model.IssueRepository, error) {
	childIds, err := util.ConvertStrToIntSlice([]*string{obj.IssueRepositoryID})

	if err != nil {
		logrus.WithField("obj", obj).Error("IssueVariantResolver: Error while parsing childIds'")
		return nil, err
	}

	return baseResolver.SingleIssueRepositoryBaseResolver(r.App, ctx, &model.NodeParent{
		Parent:     obj,
		ParentName: model.IssueVariantNodeName,
		ChildIds:   childIds,
	})
}

// Issue is the resolver for the issue field.
func (r *issueVariantResolver) Issue(ctx context.Context, obj *model.IssueVariant) (*model.Issue, error) {
	childIds, err := util.ConvertStrToIntSlice([]*string{obj.IssueID})

	if err != nil {
		logrus.WithField("obj", obj).Error("IssueVariantResolver: Error while parsing childIds'")
		return nil, err
	}

	return baseResolver.SingleIssueBaseResolver(r.App, ctx, &model.NodeParent{
		Parent:     obj,
		ParentName: model.IssueVariantNodeName,
		ChildIds:   childIds,
	})
}

// IssueVariant returns graph.IssueVariantResolver implementation.
func (r *Resolver) IssueVariant() graph.IssueVariantResolver { return &issueVariantResolver{r} }

type issueVariantResolver struct{ *Resolver }
