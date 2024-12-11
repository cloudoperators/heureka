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

func SingleIssueVariantBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.IssueVariant, error) {

	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleIssueVariantBaseResolver")

	if parent == nil {
		return nil, NewResolverError("SingleIssueVariantBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.IssueVariantFilter{
		Id: parent.ChildIds,
	}

	opt := &entity.ListOptions{}

	variants, err := app.ListIssueVariants(f, opt)

	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleIssueVariantBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(variants.Elements) > 1 {
		return nil, NewResolverError("SingleIssueVariantBaseResolver", "Internal Error - found multiple variants")
	}

	//not found
	if len(variants.Elements) < 1 {
		return nil, nil
	}

	var ivr entity.IssueVariantResult = variants.Elements[0]
	variant := model.NewIssueVariant(ivr.IssueVariant)

	return &variant, nil
}

func IssueVariantBaseResolver(app app.Heureka, ctx context.Context, filter *model.IssueVariantFilter, first *int, after *string, parent *model.NodeParent) (*model.IssueVariantConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called IssueVariantBaseResolver")

	afterId, err := ParseCursor(after)
	if err != nil {
		logrus.WithField("after", after).Error("IssueVariantBaseResolver: Error while parsing parameter 'after'")
		return nil, NewResolverError("IssueVariantBaseResolver", "Bad Request - unable to parse cursor 'after'")
	}

	var issueId []*int64
	var irId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("IssueVariantBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("IssueVariantBaseResolver", "Bad Request - Error while parsing propagated ID")
		}

		switch parent.ParentName {
		case model.IssueNodeName:
			issueId = []*int64{pid}
		case model.IssueRepositoryNodeName:
			irId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.IssueVariantFilter{}
	}

	f := &entity.IssueVariantFilter{
		Paginated:         entity.Paginated{First: first, After: afterId},
		IssueId:           issueId,
		IssueRepositoryId: irId,
		SecondaryName:     filter.SecondaryName,
		State:             entity.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	variants, err := app.ListIssueVariants(f, opt)

	if err != nil {
		return nil, NewResolverError("IssueVariantBaseResolver", err.Error())
	}

	edges := []*model.IssueVariantEdge{}
	for _, result := range variants.Elements {
		iv := model.NewIssueVariant(result.IssueVariant)
		edge := model.IssueVariantEdge{
			Node:   &iv,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if variants.TotalCount != nil {
		tc = int(*variants.TotalCount)
	}

	connection := model.IssueVariantConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(variants.PageInfo),
	}

	return &connection, nil

}

func EffectiveIssueVariantBaseResolver(app app.Heureka, ctx context.Context, filter *model.IssueVariantFilter, first *int, after *string, parent *model.NodeParent) (*model.IssueVariantConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called EffectiveIssueVariantBaseResolver")

	afterId, err := ParseCursor(after)
	if err != nil {
		logrus.WithField("after", after).Error("EffectiveIssueVariantBaseResolver: Error while parsing parameter 'after'")
		return nil, NewResolverError("EffectiveIssueVariantBaseResolver", "Bad Request - unable to parse cursor 'after'")
	}

	var imId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("EffectiveIssueVariantBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("EffectiveIssueVariantBaseResolver", "Bad Request - Error while parsing propagated ID")
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
		Paginated:    entity.Paginated{First: first, After: afterId},
		IssueMatchId: imId,
		State:        entity.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	variants, err := app.ListEffectiveIssueVariants(f, opt)

	if err != nil {
		return nil, NewResolverError("EffectiveIssueVariantBaseResolver", err.Error())
	}

	edges := []*model.IssueVariantEdge{}
	for _, result := range variants.Elements {
		iv := model.NewIssueVariant(result.IssueVariant)
		edge := model.IssueVariantEdge{
			Node:   &iv,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if variants.TotalCount != nil {
		tc = int(*variants.TotalCount)
	}

	connection := model.IssueVariantConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(variants.PageInfo),
	}

	return &connection, nil

}
