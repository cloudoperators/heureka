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

func SingleIssueRepositoryBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.IssueRepository, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleIssueRepositoryBaseResolver")

	if parent == nil {
		return nil, NewResolverError("SingleIssueRepositoryBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.IssueRepositoryFilter{
		Id: parent.ChildIds,
	}

	opt := &entity.ListOptions{}

	issueRepositories, err := app.ListIssueRepositories(f, opt)

	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleIssueRepositoryBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(issueRepositories.Elements) > 1 {
		return nil, NewResolverError("SingleIssueRepositoryBaseResolver", "Internal Error - found multiple issue repositories")
	}

	//not found
	if len(issueRepositories.Elements) < 1 {
		return nil, nil
	}

	var irr entity.IssueRepositoryResult = issueRepositories.Elements[0]
	issueRepository := model.NewIssueRepository(irr.IssueRepository)

	return &issueRepository, nil
}

func IssueRepositoryBaseResolver(app app.Heureka, ctx context.Context, filter *model.IssueRepositoryFilter, first *int, after *string, parent *model.NodeParent) (*model.IssueRepositoryConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called IssueRepositoryBaseResolver")

	afterId, err := ParseCursor(after)
	if err != nil {
		logrus.WithField("after", after).Error("IssueRepositoryBaseResolver: Error while parsing parameter 'after'")
		return nil, NewResolverError("IssueRepositoryBaseResolver", "Bad Request - unable to parse cursor 'after'")
	}

	var serviceId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("IssueRepositoryBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("IssueRepositoryBaseResolver", "Bad Request - Error while parsing propagated ID")
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
		Paginated:   entity.Paginated{First: first, After: afterId},
		ServiceId:   serviceId,
		Name:        filter.Name,
		ServiceCCRN: filter.ServiceCcrn,
	}

	opt := GetListOptions(requestedFields)

	issueRepositories, err := app.ListIssueRepositories(f, opt)

	if err != nil {
		return nil, NewResolverError("IssueRepositoryBaseResolver", err.Error())
	}

	edges := []*model.IssueRepositoryEdge{}
	for _, result := range issueRepositories.Elements {
		ir := model.NewIssueRepository(result.IssueRepository)

		edge := model.IssueRepositoryEdge{
			Node:   &ir,
			Cursor: result.Cursor(),
		}

		if lo.Contains(requestedFields, "edges.priority") {
			p := int(result.IssueRepositoryService.Priority)
			edge.Priority = &p
		}

		edges = append(edges, &edge)
	}

	tc := 0
	if issueRepositories.TotalCount != nil {
		tc = int(*issueRepositories.TotalCount)
	}

	connection := model.IssueRepositoryConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(issueRepositories.PageInfo),
	}

	return &connection, nil

}
