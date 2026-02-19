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
)

func ImageVersionBaseResolver(app app.Heureka, ctx context.Context, filter *model.ImageVersionFilter, first *int, after *string, parent *model.NodeParent) (*model.ImageVersionConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called ImageVersionBaseResolver")

	if filter == nil {
		filter = &model.ImageVersionFilter{}
	}

	var issueId []*int64
	var componentId []*int64
	var err error
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("ImageVersionBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("ImageVersionBaseResolver", "Bad Request - Error while parsing propagated ID")
		}

		switch parent.ParentName {
		case model.ImageNodeName:
			componentId = []*int64{pid}
		}
	}

	f := &entity.ComponentVersionFilter{
		PaginatedX:    entity.PaginatedX{First: first, After: after},
		IssueId:       issueId,
		ComponentId:   componentId,
		ComponentCCRN: filter.Image,
		ServiceCCRN:   filter.Service,
		Version:       filter.Version,
		State:         model.GetStateFilterType(filter.State),
		EndOfLife:     filter.EndOfLife,
	}

	opt := GetListOptions(requestedFields)
	opt.Order = append(opt.Order, entity.Order{By: entity.CriticalCount, Direction: entity.OrderDirectionDesc})
	opt.Order = append(opt.Order, entity.Order{By: entity.HighCount, Direction: entity.OrderDirectionDesc})
	opt.Order = append(opt.Order, entity.Order{By: entity.MediumCount, Direction: entity.OrderDirectionDesc})
	opt.Order = append(opt.Order, entity.Order{By: entity.LowCount, Direction: entity.OrderDirectionDesc})
	opt.Order = append(opt.Order, entity.Order{By: entity.NoneCount, Direction: entity.OrderDirectionDesc})
	opt.Order = append(opt.Order, entity.Order{By: entity.ComponentVersionRepository, Direction: entity.OrderDirectionAsc})

	componentVersions, err := app.ListComponentVersions(f, opt)
	if err != nil {
		return nil, NewResolverError("ImageVersionBaseResolver", err.Error())
	}

	edges := []*model.ImageVersionEdge{}
	for _, result := range componentVersions.Elements {
		iv := model.NewImageVersion(result.ComponentVersion)
		edge := model.ImageVersionEdge{
			Node:   &iv,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if componentVersions.TotalCount != nil {
		tc = int(*componentVersions.TotalCount)
	}

	connection := model.ImageVersionConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(componentVersions.PageInfo),
	}

	if lo.Contains(requestedFields, "counts") {
		cvIds := lo.Map(componentVersions.Elements, func(e entity.ComponentVersionResult, _ int) *int64 {
			return &e.ComponentVersion.Id
		})

		icFilter := &entity.IssueFilter{
			ServiceCCRN:        filter.Service,
			ComponentVersionId: cvIds,
		}

		counts, err := app.GetIssueSeverityCounts(icFilter)
		if err != nil {
			return nil, NewResolverError("ImageVersionBaseResolver", err.Error())
		}

		var severityCounts model.SeverityCounts
		severityCounts.Critical = int(counts.Critical)
		severityCounts.High = int(counts.High)
		severityCounts.Medium = int(counts.Medium)
		severityCounts.Low = int(counts.Low)
		severityCounts.None = int(counts.None)
		severityCounts.Total = int(counts.Total)

		connection.Counts = &severityCounts
	}

	return &connection, nil
}
