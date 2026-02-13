// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"k8s.io/utils/ptr"
)

func SingleIssueMatchBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.IssueMatch, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleIssueMatchBaseResolver")

	if parent == nil {
		return nil, NewResolverError("SingleIssueMatchBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.IssueMatchFilter{
		Id: parent.ChildIds,
	}

	opt := entity.NewListOptions()

	issueMatches, err := app.ListIssueMatches(f, opt)
	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleIssueMatchBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(issueMatches.Elements) > 1 {
		return nil, NewResolverError("SingleIssueMatchBaseResolver", "Internal Error - found multiple IssueMatches")
	}

	// not found
	if len(issueMatches.Elements) < 1 {
		return nil, nil
	}

	imr := issueMatches.Elements[0]
	issueMatch := model.NewIssueMatch(imr.IssueMatch)

	return &issueMatch, nil
}

func IssueMatchBaseResolver(app app.Heureka, ctx context.Context, filter *model.IssueMatchFilter, first *int, after *string, orderBy []*model.IssueMatchOrderBy, parent *model.NodeParent) (*model.IssueMatchConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called IssueMatchBaseResolver")

	var eId []*int64
	var ciId []*int64
	var issueId []*int64
	var serviceId []*int64
	var imIds []*int64
	var err error
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			return nil, NewResolverError("IssueMatchBaseResolver", "Bad Request - Error while parsing parent ID")
		}

		switch parent.ParentName {
		case model.EvidenceNodeName:
			eId = []*int64{pid}
		case model.ComponentInstanceNodeName:
			ciId = []*int64{pid}
		case model.IssueNodeName, model.VulnerabilityNodeName:
			issueId = []*int64{pid}
		case model.ServiceNodeName:
			serviceId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.IssueMatchFilter{}
	}

	if filter.ID != nil {
		imIds, err = util.ConvertStrToIntSlice(filter.ID)
		if err != nil {
			return nil, NewResolverError("IssueMatchBaseResolver", "Bad Request - unable to parse filter, the value of the filter ID is invalid")
		}
	}

	if filter.ComponentInstanceID != nil {
		ciId, err = util.ConvertStrToIntSlice(filter.ComponentInstanceID)
		if err != nil {
			return nil, NewResolverError("IssueMatchBaseResolver", "Bad Request - unable to parse filter, the value of the filter ComponentInstanceId is invalid")
		}
	}

	f := &entity.IssueMatchFilter{
		Id:                       imIds,
		PaginatedX:               entity.PaginatedX{First: first, After: after},
		ServiceCCRN:              filter.ServiceCcrn,
		Status:                   lo.Map(filter.Status, func(item *model.IssueMatchStatusValues, _ int) *string { return ptr.To(item.String()) }),
		SeverityValue:            lo.Map(filter.Severity, func(item *model.SeverityValues, _ int) *string { return ptr.To(item.String()) }),
		SupportGroupCCRN:         filter.SupportGroupCcrn,
		IssueId:                  issueId,
		EvidenceId:               eId,
		ServiceId:                serviceId,
		ComponentInstanceId:      ciId,
		Search:                   filter.Search,
		ComponentCCRN:            filter.ComponentCcrn,
		PrimaryName:              filter.PrimaryName,
		IssueType:                lo.Map(filter.IssueType, func(item *model.IssueTypes, _ int) *string { return ptr.To(item.String()) }),
		ServiceOwnerUsername:     filter.ServiceOwnerUsername,
		ServiceOwnerUniqueUserId: filter.ServiceOwnerUniqueUserID,
		State:                    model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)
	for _, o := range orderBy {
		opt.Order = append(opt.Order, o.ToOrderEntity())
	}

	issueMatches, err := app.ListIssueMatches(f, opt)
	if err != nil {
		return nil, NewResolverError("IssueMatchBaseResolver", err.Error())
	}

	edges := []*model.IssueMatchEdge{}
	for _, result := range issueMatches.Elements {
		im := model.NewIssueMatch(result.IssueMatch)
		edge := model.IssueMatchEdge{
			Node:   &im,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if issueMatches.TotalCount != nil {
		tc = int(*issueMatches.TotalCount)
	}

	connection := model.IssueMatchConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(issueMatches.PageInfo),
	}

	return &connection, nil
}
