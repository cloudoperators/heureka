// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/sirupsen/logrus"
)

func SingleComponentBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.Component, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleComponentBaseResolver")

	if parent == nil {
		return nil, NewResolverError("SingleComponentBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.ComponentFilter{
		Id: parent.ChildIds,
	}

	opt := &entity.ListOptions{}

	components, err := app.ListComponents(f, opt)
	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleComponentBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(components.Elements) > 1 {
		return nil, NewResolverError("SingleComponentBaseResolver", "Internal Error - found multiple components")
	}

	// not found
	if len(components.Elements) < 1 {
		return nil, nil
	}

	cr := components.Elements[0]
	component := model.NewComponent(cr.Component)

	return &component, nil
}

func ComponentBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentFilter, first *int, after *string, parent *model.NodeParent) (*model.ComponentConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called ComponentBaseResolver")

	if filter == nil {
		filter = &model.ComponentFilter{}
	}

	f := &entity.ComponentFilter{
		PaginatedX: entity.PaginatedX{First: first, After: after},
		CCRN:       filter.ComponentCcrn,
		State:      model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	components, err := app.ListComponents(f, opt)
	if err != nil {
		return nil, NewResolverError("ComponentBaseResolver", err.Error())
	}

	edges := []*model.ComponentEdge{}
	for _, result := range components.Elements {
		c := model.NewComponent(result.Component)
		edge := model.ComponentEdge{
			Node:   &c,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if components.TotalCount != nil {
		tc = int(*components.TotalCount)
	}

	connection := model.ComponentConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(components.PageInfo),
	}

	return &connection, nil
}

func ComponentCcrnBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentFilter) (*model.FilterItem, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called ComponentCcrnBaseResolver")

	if filter == nil {
		filter = &model.ComponentFilter{}
	}

	f := &entity.ComponentFilter{
		PaginatedX: entity.PaginatedX{},
		CCRN:       filter.ComponentCcrn,
	}

	opt := GetListOptions(requestedFields)

	names, err := app.ListComponentCcrns(f, opt)
	if err != nil {
		return nil, NewResolverError("ComponentCcrnBaseReolver", err.Error())
	}

	var pointerNames []*string

	for _, name := range names {
		pointerNames = append(pointerNames, &name)
	}

	filterItem := model.FilterItem{
		DisplayName: &FilterDisplayComponentCcrn,
		Values:      pointerNames,
	}

	return &filterItem, nil
}

func ComponentIssueCountsBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentFilter, parent *model.NodeParent) (*model.SeverityCounts, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called ComponentIssueCountsBaseResolver")

	if filter == nil {
		filter = &model.ComponentFilter{}
	}

	var componentId []*int64
	var err error
	if parent != nil {
		var pid *int64
		if parent.Parent != nil {
			parentId := parent.Parent.GetID()
			pid, err = ParseCursor(&parentId)
			if err != nil {
				return nil, ToGraphQLError(appErrors.E(appErrors.Op("ComponentIssueCountsBaseResolver"), "Issue", appErrors.InvalidArgument, "Error while parsing propagated ID"))
			}
		}

		if parent.ParentName == model.ImageNodeName {
			componentId = []*int64{pid}
		}
	}

	f := &entity.ComponentFilter{
		Id:          componentId,
		ServiceCCRN: filter.ServiceCcrn,
	}

	var severityCounts model.SeverityCounts
	counts, err := app.GetComponentVulnerabilityCounts(f)
	if err != nil {
		return nil, ToGraphQLError(err)
	}

	for _, c := range counts {
		severityCounts.Critical += int(c.Critical)
		severityCounts.High += int(c.High)
		severityCounts.Medium += int(c.Medium)
		severityCounts.Low += int(c.Low)
		severityCounts.None += int(c.None)
		severityCounts.Total += int(c.Total)
	}

	return &severityCounts, nil
}
