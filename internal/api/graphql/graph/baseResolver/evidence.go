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

func EvidenceBaseResolver(app app.Heureka, ctx context.Context, filter *model.EvidenceFilter, first *int, after *string, parent *model.NodeParent) (*model.EvidenceConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called EvidenceBaseResolver")

	afterId, err := ParseCursor(after)
	if err != nil {
		logrus.WithField("after", after).Error("EvidenceBaseResolver: Error while parsing parameter 'after'")
		return nil, NewResolverError("EvidenceBaseResolver", "Bad Request - unable to parse cursor 'after'")
	}

	var activityId []*int64
	var imId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("EvidenceBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("EvidenceBaseResolver", "Bad Request - Error while parsing propagated ID")
		}

		switch parent.ParentName {
		case model.ActivityNodeName:
			activityId = []*int64{pid}
		case model.IssueMatchNodeName:
			imId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.EvidenceFilter{}
	}

	f := &entity.EvidenceFilter{
		Paginated:    entity.Paginated{First: first, After: afterId},
		ActivityId:   activityId,
		IssueMatchId: imId,
	}

	opt := GetListOptions(requestedFields)

	evidences, err := app.ListEvidences(f, opt)

	if err != nil {
		return nil, NewResolverError("EvidenceBaseResolver", err.Error())
	}

	edges := []*model.EvidenceEdge{}
	for _, result := range evidences.Elements {
		e := model.NewEvidence(result.Evidence)
		edge := model.EvidenceEdge{
			Node:   &e,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if evidences.TotalCount != nil {
		tc = int(*evidences.TotalCount)
	}

	connection := model.EvidenceConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(evidences.PageInfo),
	}

	return &connection, nil

}
