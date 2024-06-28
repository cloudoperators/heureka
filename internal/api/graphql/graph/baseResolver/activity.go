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

func SingleActivityBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.Activity, error) {

	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleActivityBaseResolver")

	if parent == nil {
		return nil, NewResolverError("SingleActivityBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.ActivityFilter{
		Id: parent.ChildIds,
	}

	opt := &entity.ListOptions{}

	activities, err := app.ListActivities(f, opt)

	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleActivityBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(activities.Elements) > 1 {
		return nil, NewResolverError("SingleActivityBaseResolver", "Internal Error - found multiple activities")
	}

	//not found
	if len(activities.Elements) < 1 {
		return nil, nil
	}

	var activityResult = activities.Elements[0]
	activity := model.NewActivity(activityResult.Activity)

	return &activity, nil
}

func ActivityBaseResolver(app app.Heureka, ctx context.Context, filter *model.ActivityFilter, first *int, after *string, parent *model.NodeParent) (*model.ActivityConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called ActivityBaseResolver")

	afterId, err := ParseCursor(after)
	if err != nil {
		logrus.WithField("after", after).Error("ActivityBaseResolver: Error while parsing parameter 'after'")
		return nil, NewResolverError("ActivityBaseResolver", "Bad Request - unable to parse cursor 'after'")
	}

	var sId []*int64
	var issueId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("ActivityBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("ActivityBaseResolver", "Bad Request - Error while parsing propagated ID")
		}

		switch parent.ParentName {
		case model.ServiceNodeName:
			sId = []*int64{pid}
		case model.IssueNodeName:
			issueId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.ActivityFilter{}
	}

	f := &entity.ActivityFilter{
		Paginated:   entity.Paginated{First: first, After: afterId},
		ServiceName: filter.ServiceName,
		ServiceId:   sId,
		IssueId:     issueId,
	}

	opt := GetListOptions(requestedFields)

	activities, err := app.ListActivities(f, opt)

	if err != nil {
		return nil, NewResolverError("ActivityBaseResolver", err.Error())
	}

	edges := []*model.ActivityEdge{}
	for _, result := range activities.Elements {
		a := model.NewActivity(result.Activity)
		edge := model.ActivityEdge{
			Node:   &a,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if activities.TotalCount != nil {
		tc = int(*activities.TotalCount)
	}

	connection := model.ActivityConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(activities.PageInfo),
	}

	return &connection, nil

}
