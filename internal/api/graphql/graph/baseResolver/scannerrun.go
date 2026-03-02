// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"
	"fmt"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"k8s.io/utils/pointer"
)

func ScannerRunTagFilterValues(app app.Heureka, ctx context.Context) ([]*string, error) {
	tags, err := app.GetScannerRunTags()
	if err != nil {
		return nil, err
	}

	res := make([]*string, len(tags))
	for _, tag := range tags {
		res = append(res, pointer.String(tag))
	}

	return res, nil
}

func ScannerRuns(app app.Heureka, ctx context.Context, filter *model.ScannerRunFilter, first *int, after *string) (*model.ScannerRunConnection, error) {
	requestedFields := GetPreloads(ctx)
	listOptions := GetListOptions(requestedFields)

	value, err := ParseCursor(after)
	if err != nil {
		return nil, err
	}

	var efilter entity.ScannerRunFilter

	if filter == nil {
		efilter = entity.ScannerRunFilter{
			Paginated: entity.Paginated{
				First: first,
				After: after,
			},
			Tag:       nil,
			Completed: false,
		}
	} else {
		efilter = entity.ScannerRunFilter{
			Paginated: entity.Paginated{
				First: first,
				After: after,
			},
			Tag:       filter.Tag,
			Completed: filter.Completed,
		}
	}

	scannerRuns, err := app.GetScannerRuns(&efilter, listOptions)
	if err != nil {
		return nil, err
	}

	totalCount, err := app.CountScannerRuns(&efilter)
	if err != nil {
		return nil, err
	}

	hasNext := false
	if first == nil {
		first = pointer.Int(0)
	}

	if totalCount > *first+int(*value) {
		hasNext = true
	}

	if after == nil {
		after = pointer.String("")
	}
	hasPrevious := len(*after) > 0

	var edges []*model.ScannerRunEdge

	for _, scannerRun := range scannerRuns {
		srm := model.NewScannerRun(&scannerRun)
		cursor := fmt.Sprintf("%d", scannerRun.RunID)
		edge := model.ScannerRunEdge{
			Node: &srm,

			Cursor: &cursor,
		}
		edges = append(edges, &edge)
	}

	src := model.ScannerRunConnection{
		TotalCount: totalCount,
		PageInfo: &model.PageInfo{
			HasNextPage:     &hasNext,
			HasPreviousPage: &hasPrevious,
		},
		Edges: edges,
	}

	return &src, nil
}
