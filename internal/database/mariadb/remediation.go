// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var remediationObject = DbObject[*entity.Remediation]{
	Prefix:    "remediation",
	TableName: "Remediation",
	Properties: []*Property{
		NewProperty("remediation_description", WrapAccess(func(r *entity.Remediation) (string, bool) { return r.Description, r.Description != "" })),
		NewProperty("remediation_type", WrapAccess(func(r *entity.Remediation) (entity.RemediationType, bool) {
			return r.Type, r.Type != "" && r.Type != entity.RemediationTypeUnknown
		})),
		NewProperty("remediation_severity", WrapAccess(func(r *entity.Remediation) (entity.SeverityValues, bool) {
			return r.Severity, r.Severity != "" && r.Severity != entity.SeverityValuesUnknown
		})),
		NewProperty("remediation_remediation_date", WrapAccess(func(r *entity.Remediation) (time.Time, bool) { return r.RemediationDate, !r.RemediationDate.IsZero() })),
		NewProperty("remediation_expiration_date", WrapAccess(func(r *entity.Remediation) (time.Time, bool) { return r.ExpirationDate, !r.ExpirationDate.IsZero() })),
		NewProperty("remediation_service", WrapAccess(func(r *entity.Remediation) (string, bool) { return r.Service, r.Service != "" })),
		NewProperty("remediation_service_id", WrapAccess(func(r *entity.Remediation) (int64, bool) { return r.ServiceId, r.ServiceId != 0 })),
		NewProperty("remediation_component", WrapAccess(func(r *entity.Remediation) (string, bool) { return r.Component, r.Component != "" })),
		NewProperty("remediation_component_id", WrapAccess(func(r *entity.Remediation) (int64, bool) { return r.ComponentId, r.ComponentId != 0 })),
		NewProperty("remediation_issue", WrapAccess(func(r *entity.Remediation) (string, bool) { return r.Issue, r.Issue != "" })),
		NewProperty("remediation_issue_id", WrapAccess(func(r *entity.Remediation) (int64, bool) { return r.IssueId, r.IssueId != 0 })),
		NewProperty("remediation_remediated_by", WrapAccess(func(r *entity.Remediation) (string, bool) { return r.RemediatedBy, NoUpdate })),
		NewProperty("remediation_remediated_by_id", WrapAccess(func(r *entity.Remediation) (int64, bool) { return r.RemediatedById, NoUpdate })),
		NewProperty("remediation_created_by", WrapAccess(func(r *entity.Remediation) (int64, bool) { return r.CreatedBy, NoUpdate })),
		NewProperty("remediation_updated_by", WrapAccess(func(r *entity.Remediation) (int64, bool) { return r.UpdatedBy, r.UpdatedBy != 0 })),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty("R.remediation_id = ?", WrapRetSlice(func(filter *entity.RemediationFilter) []*int64 { return filter.Id })),
		NewFilterProperty("R.remediation_severity = ?", WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Severity })),
		NewFilterProperty("R.remediation_type = ?", WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Type })),
		NewFilterProperty("R.remediation_service = ?", WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Service })),
		NewFilterProperty("R.remediation_service_id = ?", WrapRetSlice(func(filter *entity.RemediationFilter) []*int64 { return filter.ServiceId })),
		NewFilterProperty("R.remediation_component = ?", WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Component })),
		NewFilterProperty("R.remediation_component_id = ?", WrapRetSlice(func(filter *entity.RemediationFilter) []*int64 { return filter.ComponentId })),
		NewFilterProperty("R.remediation_issue = ?", WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Issue })),
		NewFilterProperty("R.remediation_issue_id = ?", WrapRetSlice(func(filter *entity.RemediationFilter) []*int64 { return filter.IssueId })),
		NewFilterProperty("R.remediation_issue LIKE Concat('%',?,'%')", WrapRetSlice(func(filter *entity.RemediationFilter) []*string { return filter.Search })),
		NewStateFilterProperty("R.remediation", WrapRetState(func(filter *entity.RemediationFilter) []entity.StateFilterType { return filter.State })),
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

	cursorFields, err := DecodeCursor(filter.Paginated.After)
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

func (s *SqlDatabase) CreateRemediation(remediation *entity.Remediation) (*entity.Remediation, error) {
	return remediationObject.Create(s.db, remediation)
}

func (s *SqlDatabase) UpdateRemediation(remediation *entity.Remediation) error {
	return remediationObject.Update(s.db, remediation)
}

func (s *SqlDatabase) DeleteRemediation(id int64, userId int64) error {
	return remediationObject.Delete(s.db, id, userId)
}
