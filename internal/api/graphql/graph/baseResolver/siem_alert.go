// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"k8s.io/utils/ptr"
)

func SIEMAlertBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.SIEMAlertFilter,
	first *int,
	after *string,
	orderBy []*model.SIEMAlertOrderBy,
) (*model.SIEMAlertConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called SIEMAlertBaseResolver")

	if filter == nil {
		filter = &model.SIEMAlertFilter{}
	}

	securityEvent := entity.IssueTypeSecurityEvent.String()

	f := &entity.IssueMatchFilter{
		Paginated:                entity.Paginated{First: first, After: after},
		IssueType:                []*string{&securityEvent},
		IncludeComponentInstance: true,
		IncludeService:           true,
		ServiceCCRN:              filter.Service,
		SupportGroupCCRN:         filter.SupportGroup,
		Region:                   filter.Region,
		SeverityValue: lo.FilterMap(
			filter.Severity,
			func(item *model.SeverityValues, _ int) (*string, bool) {
				if item == nil {
					return nil, false
				}

				return ptr.To(item.String()), true
			},
		),
		Status: lo.FilterMap(
			filter.Status,
			func(item *model.IssueMatchStatusValues, _ int) (*string, bool) {
				if item == nil {
					return nil, false
				}

				return ptr.To(item.String()), true
			},
		),
	}

	opt := GetListOptions(requestedFields)
	for _, o := range orderBy {
		opt.Order = append(opt.Order, o.ToOrderEntity())
	}

	issueMatches, err := app.ListIssueMatches(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("SIEMAlertBaseResolver", err.Error())
	}

	edges := []*model.SIEMAlertEdge{}

	for _, result := range issueMatches.Elements {
		node := model.NewSIEMAlertNode(result.IssueMatch)
		edge := model.SIEMAlertEdge{
			Node:   &node,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if issueMatches.TotalCount != nil {
		tc = int(*issueMatches.TotalCount)
	}

	connection := model.SIEMAlertConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(issueMatches.PageInfo),
	}

	return &connection, nil
}
