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

func SupportGroupBaseResolver(app app.Heureka, ctx context.Context, filter *model.SupportGroupFilter, first *int, after *string, parent *model.NodeParent) (*model.SupportGroupConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SupportGroupBaseResolver")

	afterId, err := ParseCursor(after)
	if err != nil {
		logrus.WithField("after", after).Error("SupportGroupBaseResolver: Error while parsing parameter 'after'")
		return nil, NewResolverError("SupportGroupBaseResolver", "Bad Request - unable to parse cursor 'after'")
	}

	var serviceId []*int64
	var userId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("SupportGroupBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("SupportGroupBaseResolver", "Bad Request - Error while parsing propagated ID")
		}

		switch parent.ParentName {
		case model.ServiceNodeName:
			serviceId = []*int64{pid}
		case model.UserNodeName:
			userId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.SupportGroupFilter{}
	}

	f := &entity.SupportGroupFilter{
		Paginated: entity.Paginated{First: first, After: afterId},
		ServiceId: serviceId,
		UserId:    userId,
		Name:      filter.SupportGroupName,
	}

	opt := GetListOptions(requestedFields)

	supportGroups, err := app.ListSupportGroups(f, opt)

	if err != nil {
		return nil, NewResolverError("SupportGroupBaseResolver", err.Error())
	}

	edges := []*model.SupportGroupEdge{}
	for _, result := range supportGroups.Elements {
		sg := model.NewSupportGroup(result.SupportGroup)
		edge := model.SupportGroupEdge{
			Node:   &sg,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if supportGroups.TotalCount != nil {
		tc = int(*supportGroups.TotalCount)
	}

	connection := model.SupportGroupConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(supportGroups.PageInfo),
	}

	return &connection, nil
}

func SupportGroupNameBaseResolver(app app.Heureka, ctx context.Context, filter *model.SupportGroupFilter) (*model.FilterItem, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called SupportGroupNameBaseResolver")
	var err error

	if filter == nil {
		filter = &model.SupportGroupFilter{}
	}
	var userIds []*int64

	if len(filter.UserIds) > 0 {
		userIds, err = util.ConvertStrToIntSlice(filter.UserIds)

		if err != nil {
			logrus.WithField("Filter", filter).Error("SupportGroupNameBaseResolver: Error while parsing 'UserIds'")
			return nil, NewResolverError("SupportGroupNameBaseResolver", "Bad Request - unable to parse 'UserIds'")
		}
	}

	f := &entity.SupportGroupFilter{
		Paginated: entity.Paginated{},
		UserId:    userIds,
		Name:      filter.SupportGroupName,
	}

	opt := GetListOptions(requestedFields)

	names, err := app.ListSupportGroupNames(f, opt)

	if err != nil {
		return nil, NewResolverError("SupportGroupNameBaseResolver", err.Error())
	}

	var pointerNames []*string

	for _, name := range names {
		pointerNames = append(pointerNames, &name)
	}

	filterItem := model.FilterItem{
		DisplayName: &FilterDisplaySupportGroupName,
		Values:      pointerNames,
	}

	return &filterItem, nil
}
