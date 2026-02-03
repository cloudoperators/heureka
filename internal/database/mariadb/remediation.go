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

func buildRemediationFilterParameters(filter *entity.RemediationFilter, withCursor bool, cursorFields []Field) []interface{} {
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.Severity)
	filterParameters = buildQueryParameters(filterParameters, filter.Type)
	filterParameters = buildQueryParameters(filterParameters, filter.Service)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	filterParameters = buildQueryParameters(filterParameters, filter.Component)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentId)
	filterParameters = buildQueryParameters(filterParameters, filter.Issue)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	filterParameters = buildQueryParameters(filterParameters, filter.Search)
	if withCursor {
		filterParameters = append(filterParameters, GetCursorQueryParameters(filter.PaginatedX.First, cursorFields)...)
	}
	return filterParameters
}

func ensureRemediationFilter(filter *entity.RemediationFilter) *entity.RemediationFilter {
	var first int = 1000
	var after string = ""
	if filter == nil {
		filter = &entity.RemediationFilter{
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

func getRemediationUpdateFields(remediation *entity.Remediation) string {
	fl := []string{}
	if remediation.Description != "" {
		fl = append(fl, "remediation_description = :remediation_description")
	}
	if remediation.Type != "" && remediation.Type != entity.RemediationTypeUnknown {
		fl = append(fl, "remediation_type = :remediation_type")
	}
	if remediation.Severity != "" && remediation.Severity != entity.SeverityValuesUnknown {
		fl = append(fl, "remediation_severity = :remediation_severity")
	}
	if !remediation.RemediationDate.IsZero() {
		fl = append(fl, "remediation_remediation_date = :remediation_remediation_date")
	}
	if !remediation.ExpirationDate.IsZero() {
		fl = append(fl, "remediation_expiration_date = :remediation_expiration_date")
	}
	if remediation.Service != "" {
		fl = append(fl, "remediation_service = :remediation_service")
	}
	if remediation.ServiceId != 0 {
		fl = append(fl, "remediation_service_id = :remediation_service_id")
	}
	if remediation.Component != "" {
		fl = append(fl, "remediation_component = :remediation_component")
	}
	if remediation.ComponentId != 0 {
		fl = append(fl, "remediation_component_id = :remediation_component_id")
	}
	if remediation.Issue != "" {
		fl = append(fl, "remediation_issue = :remediation_issue")
	}
	if remediation.IssueId != 0 {
		fl = append(fl, "remediation_issue_id = :remediation_issue_id")
	}
	if remediation.UpdatedBy != 0 {
		fl = append(fl, "remediation_updated_by = :remediation_updated_by")
	}
	return strings.Join(fl, ", ")
}

func getRemediationFilterString(filter *entity.RemediationFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "R.remediation_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Severity, "R.remediation_severity = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Type, "R.remediation_type = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Service, "R.remediation_service = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "R.remediation_service_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Issue, "R.remediation_issue = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueId, "R.remediation_issue_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Component, "R.remediation_component = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentId, "R.remediation_component_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Search, "R.remediation_issue LIKE Concat('%',?,'%')", OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "R.remediation"))
	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) buildRemediationStatement(baseQuery string, filter *entity.RemediationFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensureRemediationFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := getRemediationFilterString(filter)
	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode Remediation cursor: %w", err)
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.RemediationId, entity.OrderDirectionAsc)
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
		return nil, nil, fmt.Errorf("failed to prepare Remediation statement: %w", err)
	}

	filterParameters := buildRemediationFilterParameters(filter, withCursor, cursorFields)

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

func (s *SqlDatabase) CreateRemediation(remediation *entity.Remediation) (*entity.Remediation, error) {
	l := logrus.WithFields(logrus.Fields{
		"remediation": remediation,
		"event":       "database.CreateRemediation",
	})

	query := `
		INSERT INTO Remediation (
			remediation_description,
			remediation_type,
			remediation_severity,
			remediation_remediation_date,
			remediation_expiration_date,
			remediation_service,
			remediation_service_id,
			remediation_component,
			remediation_component_id,
			remediation_issue,
			remediation_issue_id,
			remediation_remediated_by,
			remediation_remediated_by_id,
			remediation_created_by,
			remediation_updated_by
		) VALUES (
			:remediation_description,
			:remediation_type,
			:remediation_severity,
			:remediation_remediation_date,
			:remediation_expiration_date,
			:remediation_service,
			:remediation_service_id,
			:remediation_component,
			:remediation_component_id,
			:remediation_issue,
			:remediation_issue_id,
			:remediation_remediated_by,
			:remediation_remediated_by_id,
			:remediation_created_by,
			:remediation_updated_by
		)
	`

	remediationRow := RemediationRow{}
	remediationRow.FromRemediation(remediation)

	id, err := performInsert(s, query, remediationRow, l)
	if err != nil {
		return nil, fmt.Errorf("failed to create Remediation: %w", err)
	}

	remediation.Id = id

	return remediation, nil
}

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

	updateFields := getRemediationUpdateFields(remediation)

	query := fmt.Sprintf(baseQuery, updateFields)

	remediationRow := RemediationRow{}
	remediationRow.FromRemediation(remediation)

	_, err := performExec(s, query, remediationRow, l)
	if err != nil {
		return fmt.Errorf("failed to update Remediation: %w", err)
	}

	return nil
}

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
