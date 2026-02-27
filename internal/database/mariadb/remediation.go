// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var remediationObject = DbObject{
	Properties: []PropertySpec{
		Property{Name: "remediation_description", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return r.Description != "" })},
		Property{Name: "remediation_type", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return r.Type != "" && r.Type != entity.RemediationTypeUnknown })},
		Property{Name: "remediation_severity", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool {
			return r.Severity != "" && r.Severity != entity.SeverityValuesUnknown
		})},
		Property{Name: "remediation_remediation_date", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return !r.RemediationDate.IsZero() })},
		Property{Name: "remediation_expiration_date", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return !r.ExpirationDate.IsZero() })},
		Property{Name: "remediation_service", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return r.Service != "" })},
		Property{Name: "remediation_service_id", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return r.ServiceId != 0 })},
		Property{Name: "remediation_component", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return r.Component != "" })},
		Property{Name: "remediation_component_id", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return r.ComponentId != 0 })},
		Property{Name: "remediation_issue", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return r.Issue != "" })},
		Property{Name: "remediation_issue_id", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return r.IssueId != 0 })},
		Property{Name: "remediation_remediated_by"},
		Property{Name: "remediation_remediated_by_id"},
		Property{Name: "remediation_created_by"},
		Property{Name: "remediation_updated_by", IsUpdatePresent: WrapChecker(func(r *entity.Remediation) bool { return r.UpdatedBy != 0 })},
	},
	FilterProperties: []FilterPropertySpec{
		FilterProperty{Query: "R.remediation_id = ?", Param: WrapRetSlice(func(filter *entity.RemediationFilter) []*int64 { return filter.Id })},
		FilterProperty{Query: "R.remediation_severity = ?", Param: WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Severity })},
		FilterProperty{Query: "R.remediation_type = ?", Param: WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Type })},
		FilterProperty{Query: "R.remediation_service = ?", Param: WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Service })},
		FilterProperty{Query: "R.remediation_service_id = ?", Param: WrapRetSlice(func(filter *entity.RemediationFilter) []*int64 { return filter.ServiceId })},
		FilterProperty{Query: "R.remediation_component = ?", Param: WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Component })},
		FilterProperty{Query: "R.remediation_component_id = ?", Param: WrapRetSlice(func(filter *entity.RemediationFilter) []*int64 { return filter.ComponentId })},
		FilterProperty{Query: "R.remediation_issue = ?", Param: WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Issue })},
		FilterProperty{Query: "R.remediation_issue_id = ?", Param: WrapRetSlice(func(filter *entity.RemediationFilter) []*int64 { return filter.IssueId })},
		FilterProperty{Query: "R.remediation_issue LIKE Concat('%',?,'%')", Param: WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Search })},
		StateFilterProperty{Prefix: "R.remediation", Param: WrapRetState(func(filter *entity.RemediationFilter) []entity.StateFilterType { return filter.State })},
	},
}

func ensureRemediationFilter(filter *entity.RemediationFilter) *entity.RemediationFilter {
	if filter == nil {
		filter = &entity.RemediationFilter{}
	}
	return EnsurePagination(filter)
}

func (s *SqlDatabase) buildRemediationStatement(baseQuery string, filter *entity.RemediationFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensureRemediationFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode Remediation cursor: %w", err)
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.RemediationId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

	filterStr := remediationObject.GetFilterQuery(filter)
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
		return nil, nil, fmt.Errorf("failed to prepare Remediation statement: %w", err)
	}

	filterParameters := remediationObject.GetFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetRemediations(filter *entity.RemediationFilter, order []entity.Order) ([]entity.RemediationResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"order":  order,
		"event":  "database.GetRemediations",
	})

	baseQuery := `
		SELECT R.* FROM Remediation R
		%s
		%s
		GROUP BY R.remediation_id ORDER BY %s LIMIT ?
    `

	stmt, filterParameters, err := s.buildRemediationStatement(baseQuery, filter, true, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build Remediation query: %w", err)
	}

	defer stmt.Close()

	results, err := performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.RemediationResult, e RowComposite) []entity.RemediationResult {
			r := e.AsRemediation()
			cursor, _ := EncodeCursor(WithRemediation(order, r))

			rr := entity.RemediationResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				Remediation: &r,
			}
			return append(l, rr)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get Remediations: %w", err)
	}

	return results, nil
}

func (s *SqlDatabase) CountRemediations(filter *entity.RemediationFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "database.CountRemediations",
		"filter": filter,
	})

	baseQuery := `
		SELECT count(distinct R.remediation_id) FROM Remediation R
		%s
        ORDER BY %s
	`
	stmt, filterParameters, err := s.buildRemediationStatement(baseQuery, filter, false, []entity.Order{}, l)
	if err != nil {
		return -1, fmt.Errorf("failed to build Remediation count query: %w", err)
	}

	defer stmt.Close()

	count, err := performCountScan(stmt, filterParameters, l)
	if err != nil {
		return -1, fmt.Errorf("failed to count Remediations: %w", err)
	}

	return count, nil
}

func (s *SqlDatabase) GetAllRemediationCursors(filter *entity.RemediationFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllRemediationCursors",
	})

	baseQuery := `
		SELECT R.* FROM Remediation R 
	    %s GROUP BY R.remediation_id ORDER BY %s
    `

	filter = ensureRemediationFilter(filter)
	stmt, filterParameters, err := s.buildRemediationStatement(baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build Remediation cursor query: %w", err)
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
		return nil, fmt.Errorf("failed to get Remediation cursors: %w", err)
	}

	return lo.Map(rows, func(row RowComposite, _ int) string {
		r := row.AsRemediation()

		cursor, _ := EncodeCursor(WithRemediation(order, r))

		return cursor
	}), nil
}

// This function can be extracted to generic function, but we would need to change
// all functions like FromRemediation etc functions to interface (FromRemediation->EntityToRow, AsRemediation->RowToEntity)
func (s *SqlDatabase) CreateRemediation(remediation *entity.Remediation) (*entity.Remediation, error) {
	l := logrus.WithFields(logrus.Fields{
		"remediation": remediation,
		"event":       "database.CreateRemediation",
	})

	remediationRow := RemediationRow{}
	remediationRow.FromRemediation(remediation)

	query := remediationObject.InsertQuery("Remediation")
	id, err := performInsert(s, query, remediationRow, l)
	if err != nil {
		return nil, fmt.Errorf("failed to create Remediation: %w", err)
	}

	remediation.Id = id

	return remediation, nil
}

// This function can be extracted to generic function, but we would need to change
// all functions like FromRemediation etc functions to interface (FromRemediation->EntityToRow, AsRemediation->RowToEntity)
func (s *SqlDatabase) UpdateRemediation(remediation *entity.Remediation) error {
	l := logrus.WithFields(logrus.Fields{
		"remediation": remediation,
		"event":       "database.UpdateRemediation",
	})

	baseQuery := `
		UPDATE Remediation SET
		%s
		WHERE remediation_id = :remediation_id
	`

	updateFields := remediationObject.GetUpdateFields(remediation)
	query := fmt.Sprintf(baseQuery, updateFields)

	remediationRow := RemediationRow{}
	remediationRow.FromRemediation(remediation)

	_, err := performExec(s, query, remediationRow, l)
	if err != nil {
		return fmt.Errorf("failed to update Remediation: %w", err)
	}

	return nil
}

// This function can be extracted to generic function when CreateRemediation and UpdateRemediation will be generic
func (s *SqlDatabase) DeleteRemediation(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteRemediation",
	})

	query := `
		UPDATE Remediation SET
		remediation_deleted_at = NOW(),
		remediation_updated_by = :userId
		WHERE remediation_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)
	if err != nil {
		return fmt.Errorf("failed to delete Remediation: %w", err)
	}

	return nil
}
