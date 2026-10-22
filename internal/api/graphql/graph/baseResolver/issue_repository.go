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
)

func SingleIssueRepositoryBaseResolver(
	app app.Heureka,
	ctx context.Context,
	parent *model.NodeParent,
) (*model.IssueRepository, error) {
	return singleNodeByChildIds(
		ctx,
		"SingleIssueRepositoryBaseResolver",
		parent,
		NewResolverError(
			"SingleIssueRepositoryBaseResolver",
			"Bad Request - No parent provided",
		),
		NewResolverError(
			"SingleIssueRepositoryBaseResolver",
			"Internal Error - found multiple issue repositories",
		),
		func(ctx context.Context, ids []*int64) ([]entity.IssueRepositoryResult, error) {
			issueRepositories, err := app.ListIssueRepositories(ctx, &entity.IssueRepositoryFilter{Id: ids}, &entity.ListOptions{})
			if err != nil {
				return nil, NewResolverError("SingleIssueRepositoryBaseResolver", err.Error())
			}

			return issueRepositories.Elements, nil
		},
		func(irr entity.IssueRepositoryResult) model.IssueRepository {
			return model.NewIssueRepository(irr.IssueRepository)
		},
	)
}

func IssueRepositoryBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.IssueRepositoryFilter,
	first *int,
	after *string,
	parent *model.NodeParent,
) (*model.IssueRepositoryConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called IssueRepositoryBaseResolver")

	var serviceId []*int64

	if parent != nil {
		parentId := parent.Parent.GetID()

		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).
				Error("IssueRepositoryBaseResolver: Error while parsing propagated parent ID'")

			return nil, NewResolverError(
				"IssueRepositoryBaseResolver",
				"Bad Request - Error while parsing propagated ID",
			)
		}

		switch parent.ParentName {
		case model.ServiceNodeName:
			serviceId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.IssueRepositoryFilter{}
	}

	f := &entity.IssueRepositoryFilter{
		Paginated:   entity.Paginated{First: first, After: after},
		ServiceId:   serviceId,
		Name:        filter.Name,
		ServiceCCRN: filter.ServiceCcrn,
		State:       model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	issueRepositories, err := app.ListIssueRepositories(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("IssueRepositoryBaseResolver", err.Error())
	}

	edges := buildEdges(issueRepositories.Elements, func(result entity.IssueRepositoryResult) *model.IssueRepositoryEdge {
		ir := model.NewIssueRepository(result.IssueRepository)

		edge := &model.IssueRepositoryEdge{
			Node:   &ir,
			Cursor: result.Cursor(),
		}

		if lo.Contains(requestedFields, "edges.priority") {
			p := int(result.Priority)
			edge.Priority = &p
		}

		return edge
	})

	tc := totalCountOf(issueRepositories.TotalCount)

	connection := model.IssueRepositoryConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(issueRepositories.PageInfo),
	}

	return &connection, nil
}
