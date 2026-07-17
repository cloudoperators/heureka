// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"
	"strconv"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func SIEMAlertCommentBaseResolver(
	a app.Heureka,
	ctx context.Context,
	filter *model.SIEMAlertCommentFilter,
	first *int,
	after *string,
	parent *model.NodeParent,
) (*model.SIEMAlertCommentConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SIEMAlertCommentBaseResolver")

	var issueMatchId int64

	if parent != nil && parent.Parent != nil {
		if node, ok := parent.Parent.(*model.SIEMAlertNode); ok {
			id, err := strconv.ParseInt(node.ID, 10, 64)
			if err == nil {
				issueMatchId = id
			}
		}
	}

	f := model.NewCommentFilterEntity(filter, issueMatchId)
	f.Paginated = entity.Paginated{First: first, After: after}

	opt := GetListOptions(requestedFields)

	comments, err := a.ListComments(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("SIEMAlertCommentBaseResolver", err.Error())
	}

	edges := []*model.SIEMAlertCommentEdge{}

	for _, result := range comments.Elements {
		node := model.NewSIEMAlertCommentNode(result.Comment)
		cursor := result.Cursor()
		edge := model.SIEMAlertCommentEdge{
			Node:   &node,
			Cursor: cursor,
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if comments.TotalCount != nil {
		tc = int(lo.FromPtr(comments.TotalCount))
	}

	connection := model.SIEMAlertCommentConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(comments.PageInfo),
	}

	return &connection, nil
}
