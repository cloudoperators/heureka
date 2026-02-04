// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func buildPatchFilterParameters(filter *entity.PatchFilter, withCursor bool, cursorFields []Field) []interface{} {
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceName)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionName)
	if withCursor {
		filterParameters = append(filterParameters, GetCursorQueryParameters(filter.PaginatedX.First, cursorFields)...)
	}

	return filterParameters
}

func ensurePatchFilter(filter *entity.PatchFilter) *entity.PatchFilter {
	var first int = 1000
	var after string = ""
	if filter == nil {
		filter = &entity.PatchFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
		}
	}
	if filter.First == nil {
		filter.First = &first
	}
	if filter.After == nil {
		filter.After = &after
	}
	return filter
}

func getPatchUpdateFields(patch *entity.Patch) string {
	fl := []string{}
	if patch.Id != 0 {
		fl = append(fl, "patch_id = :patch_id")
	}
	if patch.ServiceId != 0 {
		fl = append(fl, "patch_service_id = :patch_service_id")
	}
	if patch.ServiceName != "" {
		fl = append(fl, "patch_service_name = :patch_service_name")
	}
	if patch.ComponentVersionId != 0 {
		fl = append(fl, "patch_component_version_id = :patch_component_version_id")
	}
	if patch.ComponentVersionName != "" {
		fl = append(fl, "patch_component_version_name = :patch_component_version_name")
	}
	if patch.UpdatedBy != 0 {
		fl = append(fl, "patch_updated_by = :patch_updated_by")
	}
	return strings.Join(fl, ", ")
}

func getPatchFilterString(filter *entity.PatchFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "P.patch_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "P.patch_service_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceName, "P.patch_service_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentVersionId, "P.patch_component_version_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentVersionName, "P.patch_component_version_name = ?", OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "P.patch"))
	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) buildPatchStatement(baseQuery string, filter *entity.PatchFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensurePatchFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := getPatchFilterString(filter)
	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode Patch cursor: %w", err)
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.PatchId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

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

	filterParameters := buildPatchFilterParameters(filter, withCursor, cursorFields)

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
