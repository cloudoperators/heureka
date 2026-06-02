// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var patchObject = DbObject[*entity.Patch, *entity.PatchFilter, entity.PatchResult]{
	Prefix:       "patch",
	TableName:    "Patch",
	TableKey:     "P",
	DefaultOrder: entity.Order{By: entity.PatchId, Direction: entity.OrderDirectionAsc},
	FilterProperties: []*FilterProperty[*entity.PatchFilter]{
		NewFilterProperty("P.patch_id = ?", func(filter *entity.PatchFilter) any { return filter.Id }),
		NewFilterProperty("P.patch_service_id = ?", func(filter *entity.PatchFilter) any { return filter.ServiceId }),
		NewFilterProperty("P.patch_service_name = ?", func(filter *entity.PatchFilter) any { return filter.ServiceName }),
		NewFilterProperty("P.patch_component_version_id = ?", func(filter *entity.PatchFilter) any { return filter.ComponentVersionId }),
		NewFilterProperty("P.patch_component_version_name = ?", func(filter *entity.PatchFilter) any { return filter.ComponentVersionName }),
		NewStateFilterProperty("P.patch", func(filter *entity.PatchFilter) any { return filter.State }),
	},
	GetItemAppender: func(l []entity.PatchResult, e RowComposite, order []entity.Order) []entity.PatchResult {
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
}

func (s *SqlDatabase) buildPatchStatement(
	ctx context.Context,
	baseQuery sq.SelectBuilder,
	filter *entity.PatchFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	statement := Statement[*entity.PatchFilter]{
		Db:         s.db,
		L:          l,
		Obj:        &patchObject,
		BaseQuery:  baseQuery,
		Order:      NewOrder(order, patchObject.DefaultOrder),
		WithCursor: withCursor,
	}

	return BuildStatement(ctx, statement, filter)
}

func (s *SqlDatabase) GetPatches(
	ctx context.Context,
	filter *entity.PatchFilter,
	order []entity.Order,
) ([]entity.PatchResult, error) {
	return patchObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountPatches(ctx context.Context, filter *entity.PatchFilter) (int64, error) {
	return patchObject.Count(ctx, s.db, filter)
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

	baseQuery := sq.Select("P.*").From("Patch P").GroupBy("P.patch_id")

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
