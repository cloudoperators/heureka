// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"
)

func SingleComponentVersionBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.ComponentVersion, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleComponentVersionBaseResolver")

	if parent == nil {
		return nil, NewResolverError("SingleComponentVersionBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.ComponentVersionFilter{
		Id: parent.ChildIds,
	}

	opt := &entity.ListOptions{}

	componentVersions, err := app.ListComponentVersions(f, opt)

	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleComponentVersionBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(componentVersions.Elements) > 1 {
		return nil, NewResolverError("SingleComponentVersionBaseResolver", "Internal Error - found multiple component versions")
	}

	//not found
	if len(componentVersions.Elements) < 1 {
		return nil, nil
	}

	var cvr entity.ComponentVersionResult = componentVersions.Elements[0]
	componentVersion := model.NewComponentVersion(cvr.ComponentVersion)

	return &componentVersion, nil
}

func ComponentVersionBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentVersionFilter, first *int, after *string, orderBy []*model.ComponentVersionOrderBy, parent *model.NodeParent) (*model.ComponentVersionConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called ComponentVersionBaseResolver")

	if filter == nil {
		filter = &model.ComponentVersionFilter{}
	}

	var issueId []*int64
	var componentId []*int64
	var err error
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("ComponentVersionBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("ComponentVersionBaseResolver", "Bad Request - Error while parsing propagated ID")
		}

		switch parent.ParentName {
		case model.IssueNodeName:
			issueId = []*int64{pid}
		case model.ComponentNodeName:
			componentId = []*int64{pid}
		}
	} else {
		componentId, err = util.ConvertStrToIntSlice(filter.ComponentID)

		if err != nil {
			return nil, NewResolverError("ComponentVersionBaseResolver", "Bad Request - Error while parsing filter component ID")
		}
	}

	serviceIds, err := util.ConvertStrToIntSlice(filter.ServiceID)

	if err != nil {
		return nil, NewResolverError("ComponentVersionBaseResolver", "Bad Request - Error while parsing filter service ID")
	}

	repositoryIds, err := util.ConvertStrToIntSlice(filter.IssueRepositoryID)

	if err != nil {
		return nil, NewResolverError("ComponentVersionBaseResolver", "Bad Request - Error while parsing filter issue repository ID")
	}

	f := &entity.ComponentVersionFilter{
		PaginatedX:        entity.PaginatedX{First: first, After: after},
		IssueId:           issueId,
		ComponentId:       componentId,
		ComponentCCRN:     filter.ComponentCcrn,
		ServiceCCRN:       filter.ServiceCcrn,
		ServiceId:         serviceIds,
		IssueRepositoryId: repositoryIds,
		Version:           filter.Version,
		State:             model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)
	for _, o := range orderBy {
		if *o.By == model.ComponentVersionOrderByFieldSeverity {
			opt.Order = append(opt.Order, entity.Order{By: entity.CriticalCount, Direction: o.Direction.ToOrderDirectionEntity()})
			opt.Order = append(opt.Order, entity.Order{By: entity.HighCount, Direction: o.Direction.ToOrderDirectionEntity()})
			opt.Order = append(opt.Order, entity.Order{By: entity.MediumCount, Direction: o.Direction.ToOrderDirectionEntity()})
			opt.Order = append(opt.Order, entity.Order{By: entity.LowCount, Direction: o.Direction.ToOrderDirectionEntity()})
			opt.Order = append(opt.Order, entity.Order{By: entity.NoneCount, Direction: o.Direction.ToOrderDirectionEntity()})
			opt.Order = append(opt.Order, entity.Order{By: entity.ComponentVersionId, Direction: o.Direction.ToOrderDirectionEntity()})
		} else {
			opt.Order = append(opt.Order, o.ToOrderEntity())
		}
	}

	componentVersions, err := app.ListComponentVersions(f, opt)

	//@todo propper error handling
	if err != nil {
		return nil, NewResolverError("ComponentVersionBaseResolver", err.Error())
	}

	edges := []*model.ComponentVersionEdge{}
	for _, result := range componentVersions.Elements {
		cv := model.NewComponentVersion(result.ComponentVersion)
		edge := model.ComponentVersionEdge{
			Node:   &cv,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if componentVersions.TotalCount != nil {
		tc = int(*componentVersions.TotalCount)
	}

	connection := model.ComponentVersionConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(componentVersions.PageInfo),
	}

	return &connection, nil
}
