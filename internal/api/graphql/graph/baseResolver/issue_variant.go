// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

func SingleIssueVariantBaseResolver(
	app app.Heureka,
	ctx context.Context,
	parent *model.NodeParent,
) (*model.IssueVariant, error) {
	return singleNodeByChildIds(
		ctx,
		"SingleIssueVariantBaseResolver",
		parent,
		NewResolverError(
			"SingleIssueVariantBaseResolver",
			"Bad Request - No parent provided",
		),
		NewResolverError(
			"SingleIssueVariantBaseResolver",
			"Internal Error - found multiple variants",
		),
		func(ctx context.Context, ids []*int64) ([]entity.IssueVariantResult, error) {
			variants, err := app.ListIssueVariants(ctx, &entity.IssueVariantFilter{Id: ids}, &entity.ListOptions{})
			if err != nil {
				return nil, NewResolverError("SingleIssueVariantBaseResolver", err.Error())
			}

			return variants.Elements, nil
		},
		func(ivr entity.IssueVariantResult) model.IssueVariant {
			return model.NewIssueVariant(ivr.IssueVariant)
		},
	)
}

func IssueVariantBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.IssueVariantFilter,
	first *int,
	after *string,
	parent *model.NodeParent,
) (*model.IssueVariantConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called IssueVariantBaseResolver")

	var (
		issueId []*int64
		irId    []*int64
	)

	if parent != nil {
		parentId := parent.Parent.GetID()

		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).
				Error("IssueVariantBaseResolver: Error while parsing propagated parent ID'")

			return nil, NewResolverError(
				"IssueVariantBaseResolver",
				"Bad Request - Error while parsing propagated ID",
			)
		}

		switch parent.ParentName {
		case model.IssueNodeName, model.VulnerabilityNodeName:
			issueId = []*int64{pid}
		case model.IssueRepositoryNodeName:
			irId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.IssueVariantFilter{}
	}

	f := &entity.IssueVariantFilter{
		Paginated:         entity.Paginated{First: first, After: after},
		IssueId:           issueId,
		IssueRepositoryId: irId,
		SecondaryName:     filter.SecondaryName,
		State:             model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	variants, err := app.ListIssueVariants(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("IssueVariantBaseResolver", err.Error())
	}

	edges := buildEdges(variants.Elements, func(result entity.IssueVariantResult) *model.IssueVariantEdge {
		iv := model.NewIssueVariant(result.IssueVariant)

		return &model.IssueVariantEdge{
			Node:   &iv,
			Cursor: result.Cursor(),
		}
	})

	tc := totalCountOf(variants.TotalCount)

	connection := model.IssueVariantConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(variants.PageInfo),
	}

	return &connection, nil
}

func EffectiveIssueVariantBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.IssueVariantFilter,
	first *int,
	after *string,
	parent *model.NodeParent,
) (*model.IssueVariantConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called EffectiveIssueVariantBaseResolver")

	var imId []*int64

	if parent != nil {
		parentId := parent.Parent.GetID()

		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).
				Error("EffectiveIssueVariantBaseResolver: Error while parsing propagated parent ID'")

			return nil, NewResolverError(
				"EffectiveIssueVariantBaseResolver",
				"Bad Request - Error while parsing propagated ID",
			)
		}

		switch parent.ParentName {
		case model.IssueMatchNodeName:
			imId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.IssueVariantFilter{}
	}

	f := &entity.IssueVariantFilter{
		Paginated:    entity.Paginated{First: first, After: after},
		IssueMatchId: imId,
		State:        model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	variants, err := app.ListEffectiveIssueVariants(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("EffectiveIssueVariantBaseResolver", err.Error())
	}

	edges := buildEdges(variants.Elements, func(result entity.IssueVariantResult) *model.IssueVariantEdge {
		iv := model.NewIssueVariant(result.IssueVariant)

		return &model.IssueVariantEdge{
			Node:   &iv,
			Cursor: result.Cursor(),
		}
	})

	tc := totalCountOf(variants.TotalCount)

	connection := model.IssueVariantConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(variants.PageInfo),
	}

	return &connection, nil
}
