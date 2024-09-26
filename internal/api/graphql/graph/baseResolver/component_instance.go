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

func SingleComponentInstanceBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.ComponentInstance, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleComponentInstanceBaseResolver")

	if parent == nil {
		return nil, NewResolverError("SingleComponentInstanceBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.ComponentInstanceFilter{
		Id: parent.ChildIds,
	}

	opt := &entity.ListOptions{}

	componentInstances, err := app.ListComponentInstances(f, opt)

	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleComponentInstanceBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(componentInstances.Elements) > 1 {
		return nil, NewResolverError("SingleComponentInstanceBaseResolver", "Internal Error - found multiple component instances")
	}

	//not found
	if len(componentInstances.Elements) < 1 {
		return nil, nil
	}

	var cir entity.ComponentInstanceResult = componentInstances.Elements[0]
	componentInstance := model.NewComponentInstance(cir.ComponentInstance)

	return &componentInstance, nil
}

func ComponentInstanceBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter, first *int, after *string, parent *model.NodeParent) (*model.ComponentInstanceConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called ComponentInstanceBaseResolver")

	afterId, err := ParseCursor(after)
	if err != nil {
		logrus.WithField("after", after).Error("ComponentInstanceBaseResolver: Error while parsing parameter 'after'")
		return nil, NewResolverError("ComponentInstanceBaseResolver", "Bad Request - unable to parse cursor 'after'")
	}

	var imId []*int64
	var serviceId []*int64
	var copmonentVersionId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("ComponentInstanceBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("ComponentInstanceBaseResolver", "Bad Request - Error while parsing propagated ID")
		}

		switch parent.ParentName {
		case model.IssueMatchNodeName:
			imId = []*int64{pid}
		case model.ServiceNodeName:
			serviceId = []*int64{pid}
		case model.ComponentVersionNodeName:
			copmonentVersionId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.ComponentInstanceFilter{}
	}

	f := &entity.ComponentInstanceFilter{
		Paginated:          entity.Paginated{First: first, After: afterId},
		IssueMatchId:       imId,
		ServiceId:          serviceId,
		ComponentVersionId: copmonentVersionId,
	}

	opt := GetListOptions(requestedFields)

	componentInstances, err := app.ListComponentInstances(f, opt)

	//@todo propper error handling
	if err != nil {
		return nil, NewResolverError("ComponentInstanceBaseResolver", err.Error())
	}

	edges := []*model.ComponentInstanceEdge{}
	for _, result := range componentInstances.Elements {
		ci := model.NewComponentInstance(result.ComponentInstance)
		edge := model.ComponentInstanceEdge{
			Node:   &ci,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if componentInstances.TotalCount != nil {
		tc = int(*componentInstances.TotalCount)
	}

	connection := model.ComponentInstanceConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(componentInstances.PageInfo),
	}

	return &connection, nil
}

func CcrnBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called CcrnBaseResolver")

	if filter == nil {
		filter = &model.ComponentInstanceFilter{}
	}

	f := &entity.ComponentInstanceFilter{
		CCRN: filter.Ccrn,
	}

	opt := GetListOptions(requestedFields)

	names, err := app.ListCcrn(f, opt)

	if err != nil {
		return nil, NewResolverError("CcrnBaseResolver", err.Error())
	}

	var pointerNames []*string

	for _, name := range names {
		pointerNames = append(pointerNames, &name)
	}

	filterItem := model.FilterItem{
		DisplayName: &FilterDisplayCcrn,
		Values:      pointerNames,
	}

	return &filterItem, nil
}
