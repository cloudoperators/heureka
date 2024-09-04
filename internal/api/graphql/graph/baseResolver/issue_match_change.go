// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"k8s.io/utils/pointer"
)

func IssueMatchChangeBaseResolver(app app.Heureka, ctx context.Context, filter *model.IssueMatchChangeFilter, first *int, after *string, parent *model.NodeParent) (*model.IssueMatchChangeConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called IssueMatchChangeBaseResolver")

	afterId, err := ParseCursor(after)
	if err != nil {
		logrus.WithField("after", after).Error("IssueMatchChangeBaseResolver: Error while parsing parameter 'after'")
		return nil, NewResolverError("IssueMatchChangeBaseResolver", "Bad Request - unable to parse cursor 'after'")
	}

	var aId []*int64
	var imId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("IssueMatchChangeBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("IssueMatchChangeBaseResolver", "Bad Request - Error while parsing propagated ID")
		}

		switch parent.ParentName {
		case model.ActivityNodeName:
			aId = []*int64{pid}
		case model.IssueMatchNodeName:
			imId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.IssueMatchChangeFilter{}
	}

	f := &entity.IssueMatchChangeFilter{
		Paginated:    entity.Paginated{First: first, After: afterId},
		Action:       lo.Map(filter.Action, func(item *model.IssueMatchChangeActions, _ int) *string { return pointer.String(item.String()) }),
		ActivityId:   aId,
		IssueMatchId: imId,
	}

	opt := GetListOptions(requestedFields)

	issueMatchChanges, err := app.ListIssueMatchChanges(f, opt)

	if err != nil {
		return nil, NewResolverError("IssueMatchChangeBaseResolver", err.Error())
	}

	edges := []*model.IssueMatchChangeEdge{}
	for _, result := range issueMatchChanges.Elements {
		imc := model.NewIssueMatchChange(result.IssueMatchChange)
		edge := model.IssueMatchChangeEdge{
			Node:   &imc,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if issueMatchChanges.TotalCount != nil {
		tc = int(*issueMatchChanges.TotalCount)
	}

	connection := model.IssueMatchChangeConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(issueMatchChanges.PageInfo),
	}

	return &connection, nil

}
