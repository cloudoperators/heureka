// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
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

// ImageBaseResolver retrieves a list of images based for a specific service.
// It's designed for the Image List View in the UI
// - Default ordering is by Vulnerability Count (descending) and Repository (ascending)
func ImageBaseResolver(app app.Heureka, ctx context.Context, filter *model.ImageFilter, first *int, after *string) (*model.ImageConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called ImageBaseResolver")

	if filter == nil {
		filter = &model.ImageFilter{}
	}

	f := &entity.ComponentFilter{
		PaginatedX:                 entity.PaginatedX{First: first, After: after},
		ServiceCCRN:                filter.Service,
		ComponentVersionRepository: filter.Repository,
	}

	opt := GetListOptions(requestedFields)
	// Set default ordering
	if lo.Contains(requestedFields, "edges.node.vulnerabilities") {
		opt.Order = append(opt.Order, entity.Order{By: entity.CriticalCount, Direction: entity.OrderDirectionDesc})
		opt.Order = append(opt.Order, entity.Order{By: entity.HighCount, Direction: entity.OrderDirectionDesc})
		opt.Order = append(opt.Order, entity.Order{By: entity.MediumCount, Direction: entity.OrderDirectionDesc})
		opt.Order = append(opt.Order, entity.Order{By: entity.LowCount, Direction: entity.OrderDirectionDesc})
		opt.Order = append(opt.Order, entity.Order{By: entity.NoneCount, Direction: entity.OrderDirectionDesc})
	}
	opt.Order = append(opt.Order, entity.Order{
		By:        entity.ComponentVersionRepository,
		Direction: entity.OrderDirectionAsc,
	})

	components, err := app.ListComponents(f, opt)

	if err != nil {
		return nil, NewResolverError("ImageBaseResolver", err.Error())
	}

	edges := []*model.ImageEdge{}
	for _, result := range components.Elements {
		image := model.NewImage(result.Component)
		edge := model.ImageEdge{
			Node:   &image,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	totalCount := 0
	if components.TotalCount != nil {
		totalCount = int(*components.TotalCount)
	}

	connection := model.ImageConnection{
		TotalCount: totalCount,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(components.PageInfo),
	}

	if lo.Contains(requestedFields, "counts") {
		icFilter := &entity.ComponentFilter{
			ServiceCCRN: filter.Service,
		}
		counts, err := app.GetComponentVulnerabilityCounts(icFilter)
		if err != nil {
			return nil, NewResolverError("ImageBaseResolver", err.Error())
		}
		var severityCounts model.SeverityCounts
		for _, c := range counts {
			severityCounts.Critical += int(c.Critical)
			severityCounts.High += int(c.High)
			severityCounts.Medium += int(c.Medium)
			severityCounts.Low += int(c.Low)
			severityCounts.None += int(c.None)
			severityCounts.Total += int(c.Total)
		}
		connection.Counts = &severityCounts
	}

	return &connection, nil
}
