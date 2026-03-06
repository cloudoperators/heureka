// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var patchObject = DbObject{
	Properties: []*Property{},
	FilterProperties: []*FilterProperty{
		NewFilterProperty("P.patch_id = ?", WrapRetSlice(func(filter *entity.PatchFilter) []*int64 { return filter.Id })),
		NewFilterProperty("P.patch_service_id = ?", WrapRetSlice(func(filter *entity.PatchFilter) []*int64 { return filter.ServiceId })),
		NewFilterProperty("P.patch_service_name = ?", WrapRetSlice(func(filter *entity.PatchFilter) []*string { return filter.ServiceName })),
		NewFilterProperty("P.patch_component_version_id = ?", WrapRetSlice(func(filter *entity.PatchFilter) []*int64 { return filter.ComponentVersionId })),
		NewFilterProperty("P.patch_component_version_name = ?", WrapRetSlice(func(filter *entity.PatchFilter) []*string { return filter.ComponentVersionName })),
		NewStateFilterProperty("P.patch", WrapRetState(func(filter *entity.PatchFilter) []entity.StateFilterType { return filter.State })),
	},
}

func ensurePatchFilter(filter *entity.PatchFilter) *entity.PatchFilter {
	if filter == nil {
		filter = &entity.PatchFilter{}
	}
	return EnsurePagination(filter)
}

func (s *SqlDatabase) buildPatchStatement(baseQuery string, filter *entity.PatchFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensurePatchFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	cursorFields, err := DecodeCursor(filter.Paginated.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode Patch cursor: %w", err)
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.PatchId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

	filterStr := patchObject.GetFilterQuery(filter)
	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && withCursor && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	var query string
	if withCursor {
		query = fmt.Sprintf(baseQuery, whereClause, cursorQuery, orderStr)
	} else {
		query = fmt.Sprintf(baseQuery, whereClause, orderStr)
	}

	stmt, err := s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			}).Error(msg)
		return nil, nil, fmt.Errorf("failed to prepare Patch statement: %w", err)
	}

	filterParameters := patchObject.GetFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetPatches(filter *entity.PatchFilter, order []entity.Order) ([]entity.PatchResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"order":  order,
		"event":  "database.GetPatches",
	})

	baseQuery := `
		SELECT P.* FROM Patch P
		%s
		%s
		GROUP BY P.patch_id ORDER BY %s LIMIT ?
    `

	stmt, filterParameters, err := s.buildPatchStatement(baseQuery, filter, true, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build Patch query: %w", err)
	}

	defer stmt.Close()

	results, err := performListScan(
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

func (s *SqlDatabase) CountPatches(filter *entity.PatchFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "database.CountPatches",
		"filter": filter,
	})

	baseQuery := `
		SELECT count(distinct P.patch_id) FROM Patch P
		%s
        ORDER BY %s
	`
	stmt, filterParameters, err := s.buildPatchStatement(baseQuery, filter, false, []entity.Order{}, l)
	if err != nil {
		return -1, fmt.Errorf("failed to build Patch count query: %w", err)
	}

	defer stmt.Close()

	count, err := performCountScan(stmt, filterParameters, l)
	if err != nil {
		return -1, fmt.Errorf("failed to count Patches: %w", err)
	}

	return count, nil
}

func (s *SqlDatabase) GetAllPatchCursors(filter *entity.PatchFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllPatchCursors",
	})

	baseQuery := `
		SELECT P.* FROM Patch P
	    %s GROUP BY P.patch_id ORDER BY %s
    `

	filter = ensurePatchFilter(filter)
	stmt, filterParameters, err := s.buildPatchStatement(baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build Patch cursor query: %w", err)
	}

	defer stmt.Close()

	rows, err := performListScan(
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
