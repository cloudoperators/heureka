// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"k8s.io/utils/pointer"
)

func RemediationBaseResolver(app app.Heureka, ctx context.Context, filter *model.RemediationFilter, first *int, after *string, orderBy []*model.RemediationOrderBy, parent *model.NodeParent) (*model.RemediationConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called RemediationBaseResolver")

	var serviceId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("RemediationBaseResolver: Error while parsing propagated parent ID'")
			return nil, ToGraphQLError(appErrors.E(appErrors.Op("RemediationBaseResolver"), "Remediation", appErrors.InvalidArgument, "Error while parsing propagated ID"))
		}

		switch parent.ParentName {
		case model.ServiceNodeName:
			serviceId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.RemediationFilter{}
	}

	f := &entity.RemediationFilter{
		PaginatedX: entity.PaginatedX{First: first, After: after},
		Service:    filter.Service,
		Component:  filter.Image,
		Issue:      filter.Vulnerability,
		Type:       lo.Map(filter.Type, func(item *model.RemediationTypeValues, _ int) *string { return pointer.String(item.String()) }),
		ServiceId:  serviceId,
		State:      model.GetStateFilterType(filter.State),
		Search:     filter.Search,
	}

	opt := GetListOptions(requestedFields)
	for _, o := range orderBy {
		opt.Order = append(opt.Order, o.ToOrderEntity())
		opt.Order = append(opt.Order, entity.Order{By: entity.RemediationId, Direction: o.Direction.ToOrderDirectionEntity()})
	}
	remediations, err := app.ListRemediations(f, opt)
	if err != nil {
		return nil, ToGraphQLError(err)
	}

	edges := []*model.RemediationEdge{}
	for _, result := range remediations.Elements {
		ci := model.NewRemediation(result.Remediation)
		edge := model.RemediationEdge{
			Node:   &ci,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if remediations.TotalCount != nil {
		tc = int(*remediations.TotalCount)
	}

	connection := model.RemediationConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(remediations.PageInfo),
	}

	return &connection, nil
}
