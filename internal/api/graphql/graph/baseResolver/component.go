// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/model"
	"github.wdf.sap.corp/cc/heureka/internal/app"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
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

	//not found
	if len(components.Elements) < 1 {
		return nil, nil
	}

	var cr entity.ComponentResult = components.Elements[0]
	component := model.NewComponent(cr.Component)

	return &component, nil
}

func ComponentBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentFilter, first *int, after *string, parent *model.NodeParent) (*model.ComponentConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called ComponentBaseResolver")

	afterId, err := ParseCursor(after)
	if err != nil {
		logrus.WithField("after", after).Error("ComponentBaseResolver: Error while parsing parameter 'after'")
		return nil, NewResolverError("ComponentBaseResolver", "Bad Request - unable to parse cursor 'after'")
	}

	if filter == nil {
		filter = &model.ComponentFilter{}
	}

	f := &entity.ComponentFilter{
		Paginated: entity.Paginated{First: first, After: afterId},
		Name:      filter.ComponentName,
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

func ComponentNameBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentFilter) (*model.FilterItem, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called IssueNameBaseResolver")

	if filter == nil {
		filter = &model.ComponentFilter{}
	}

	f := &entity.ComponentFilter{
		Paginated: entity.Paginated{},
		Name:      filter.ComponentName,
	}

	opt := GetListOptions(requestedFields)

	names, err := app.ListComponentNames(f, opt)

	if err != nil {
		return nil, NewResolverError("ComponentNameBaseReolver", err.Error())
	}

	var pointerNames []*string

	for _, name := range names {
		pointerNames = append(pointerNames, &name)
	}

	filterItem := model.FilterItem{
		DisplayName: &FilterDisplayComponentName,
		Values:      pointerNames,
	}

	return &filterItem, nil
}
