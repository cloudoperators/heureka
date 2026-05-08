// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var patchObject = DbObject[*entity.Patch]{
	FilterProperties: []*FilterProperty{
		NewFilterProperty(
			"P.patch_id = ?",
			WrapRetSlice(func(filter *entity.PatchFilter) []*int64 { return filter.Id }),
		),
		NewFilterProperty(
			"P.patch_service_id = ?",
			WrapRetSlice(func(filter *entity.PatchFilter) []*int64 { return filter.ServiceId }),
		),
		NewFilterProperty(
			"P.patch_service_name = ?",
			WrapRetSlice(func(filter *entity.PatchFilter) []*string { return filter.ServiceName }),
		),
		NewFilterProperty(
			"P.patch_component_version_id = ?",
			WrapRetSlice(
				func(filter *entity.PatchFilter) []*int64 { return filter.ComponentVersionId },
			),
		),
		NewFilterProperty(
			"P.patch_component_version_name = ?",
			WrapRetSlice(
				func(filter *entity.PatchFilter) []*string { return filter.ComponentVersionName },
			),
		),
		NewStateFilterProperty(
			"P.patch",
			WrapRetState(
				func(filter *entity.PatchFilter) []entity.StateFilterType { return filter.State },
			),
		),
	},
}

func (s *SqlDatabase) buildPatchStatement(
	ctx context.Context,
	baseQuery string,
	filter *entity.PatchFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	statement := Statement{
		Db:                 s.db,
		L:                  l,
		Obj:                &patchObject,
		BaseQuery:          baseQuery,
		Order:              NewOrder(order, entity.Order{By: entity.PatchId, Direction: entity.OrderDirectionAsc}),
		WithCursor:         withCursor,
		CheckCursorInWhere: true,
		CheckCursor:        true,
		CheckFilter:        true,
		Aggregated:         false,
	}

	return BuildStatement(statement, filter)
}

func (s *SqlDatabase) GetPatches(
	ctx context.Context,
	filter *entity.PatchFilter,
	order []entity.Order,
) ([]entity.PatchResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"order":  order,
		"event":  "database.GetPatches",
	})

	baseQuery := `
		SELECT P.* FROM Patch P
		%s
		%s
		%s
		GROUP BY P.patch_id ORDER BY %s LIMIT ?
    `

	stmt, filterParameters, err := s.buildPatchStatement(ctx, baseQuery, filter, true, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build Patch query: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	results, err := performListScan(
		ctx,
		stmt,
		filterParameters,
		l,
		func(l []entity.PatchResult, e RowComposite) []entity.PatchResult {
			p := e.AsPatch()
			cursor, _ := EncodeCursor(WithPatch(order, p))

			pr := entity.PatchResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				Patch: &p,
			}

			return append(l, pr)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get Patches: %w", err)
	}

	return results, nil
}

func (s *SqlDatabase) CountPatches(ctx context.Context, filter *entity.PatchFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "database.CountPatches",
		"filter": filter,
	})

	baseQuery := `
		SELECT count(distinct P.patch_id) FROM Patch P
		%s
		%s
        ORDER BY %s
	`

	stmt, filterParameters, err := s.buildPatchStatement(
		ctx,
		baseQuery,
		filter,
		false,
		[]entity.Order{},
		l,
	)
	if err != nil {
		return -1, fmt.Errorf("failed to build Patch count query: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	count, err := performCountScan(ctx, stmt, filterParameters, l)
	if err != nil {
		return -1, fmt.Errorf("failed to count Patches: %w", err)
	}

	return count, nil
}

func (s *SqlDatabase) GetAllPatchCursors(
	ctx context.Context,
	filter *entity.PatchFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllPatchCursors",
	})

	baseQuery := `
		SELECT P.* FROM Patch P
		%s
	    %s GROUP BY P.patch_id ORDER BY %s
    `

	filter = EnsureFilter(filter)

	stmt, filterParameters, err := s.buildPatchStatement(ctx, baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build Patch cursor query: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	rows, err := performListScan(
		ctx,
		stmt,
		filterParameters,
		l,
		func(l []RowComposite, e RowComposite) []RowComposite {
			return append(l, e)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get Patch cursors: %w", err)
	}

	return lo.Map(rows, func(row RowComposite, _ int) string {
		r := row.AsPatch()

		cursor, _ := EncodeCursor(WithPatch(order, r))

		return cursor
	}), nil
}
