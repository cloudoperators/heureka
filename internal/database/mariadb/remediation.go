// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var remediationObject = DbObject[*entity.Remediation, *entity.RemediationFilter, entity.RemediationResult]{
	Prefix:       "remediation",
	TableName:    "Remediation",
	TableKey:     "R",
	DefaultOrder: entity.Order{By: entity.RemediationId, Direction: entity.OrderDirectionAsc},
	Properties: []*Property[*entity.Remediation]{
		NewProperty("remediation_description", func(r *entity.Remediation) (any, bool) { return r.Description, r.Description != "" }),
		NewProperty("remediation_type", func(r *entity.Remediation) (any, bool) {
			return r.Type, r.Type != "" && r.Type != entity.RemediationTypeUnknown
		}),
		NewProperty("remediation_url", func(r *entity.Remediation) (any, bool) { return r.URL, r.URL != "" }),
		NewProperty("remediation_severity", func(r *entity.Remediation) (any, bool) {
			return r.Severity, r.Severity != "" && r.Severity != entity.SeverityValuesUnknown
		}),
		NewProperty("remediation_remediation_date", func(r *entity.Remediation) (any, bool) { return r.RemediationDate, !r.RemediationDate.IsZero() }),
		NewProperty("remediation_expiration_date", func(r *entity.Remediation) (any, bool) { return r.ExpirationDate, !r.ExpirationDate.IsZero() }),
		NewProperty("remediation_service", func(r *entity.Remediation) (any, bool) { return r.Service, r.Service != "" }),
		NewProperty("remediation_service_id", func(r *entity.Remediation) (any, bool) { return r.ServiceId, r.ServiceId != 0 }),
		NewProperty("remediation_component", func(r *entity.Remediation) (any, bool) { return r.Component, r.Component != "" }),
		NewProperty("remediation_component_id", func(r *entity.Remediation) (any, bool) {
			return sql.NullInt64{Int64: r.ComponentId, Valid: IsValidId(r.ComponentId)}, r.ComponentId != 0
		}),
		NewProperty("remediation_issue", func(r *entity.Remediation) (any, bool) { return r.Issue, r.Issue != "" }),
		NewProperty("remediation_issue_id", func(r *entity.Remediation) (any, bool) { return r.IssueId, r.IssueId != 0 }),
		NewProperty("remediation_remediated_by", func(r *entity.Remediation) (any, bool) { return r.RemediatedBy, NoUpdate }),
		NewProperty("remediation_remediated_by_id", func(r *entity.Remediation) (any, bool) {
			return sql.NullInt64{Int64: r.RemediatedById, Valid: IsValidId(r.RemediatedById)}, NoUpdate
		}),
		NewProperty("remediation_created_by", func(r *entity.Remediation) (any, bool) { return r.CreatedBy, NoUpdate }),
		NewProperty("remediation_updated_by", func(r *entity.Remediation) (any, bool) { return r.UpdatedBy, r.UpdatedBy != 0 }),
	},
	FilterProperties: []*FilterProperty[*entity.RemediationFilter]{
		NewFilterProperty("R.remediation_id = ?", func(filter *entity.RemediationFilter) any { return filter.Id }),
		NewFilterProperty("R.remediation_severity = ?", func(filter *entity.RemediationFilter) any { return filter.Severity }),
		NewFilterProperty("R.remediation_type = ?", func(filter *entity.RemediationFilter) any { return filter.Type }),
		NewFilterProperty("R.remediation_url = ?", func(filter *entity.RemediationFilter) any { return filter.URL }),
		NewFilterProperty("R.remediation_service = ?", func(filter *entity.RemediationFilter) any { return filter.Service }),
		NewFilterProperty("R.remediation_service_id = ?", func(filter *entity.RemediationFilter) any { return filter.ServiceId }),
		NewFilterProperty("R.remediation_component = ?", func(filter *entity.RemediationFilter) any { return filter.Component }),
		NewFilterProperty("R.remediation_component_id = ?", func(filter *entity.RemediationFilter) any { return filter.ComponentId }),
		NewFilterProperty("R.remediation_issue = ?", func(filter *entity.RemediationFilter) any { return filter.Issue }),
		NewFilterProperty("R.remediation_issue_id = ?", func(filter *entity.RemediationFilter) any { return filter.IssueId }),
		NewFilterProperty("R.remediation_issue LIKE Concat('%',?,'%')", func(filter *entity.RemediationFilter) any { return filter.Search }),
		NewStateFilterProperty("R.remediation", func(filter *entity.RemediationFilter) any { return filter.State }),
	},
	GetItemAppender: func(l []entity.RemediationResult, e RowComposite, order []entity.Order) []entity.RemediationResult {
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
}

func (s *SqlDatabase) buildRemediationStatement(
	ctx context.Context,
	baseQuery sq.SelectBuilder,
	filter *entity.RemediationFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	statement := Statement[*entity.RemediationFilter]{
		Db:         s.db,
		L:          l,
		Obj:        &remediationObject,
		BaseQuery:  baseQuery,
		Order:      NewOrder(order, remediationObject.DefaultOrder),
		WithCursor: withCursor,
	}

	return BuildStatement(ctx, statement, filter)
}

func (s *SqlDatabase) GetRemediations(
	ctx context.Context,
	filter *entity.RemediationFilter,
	order []entity.Order,
) ([]entity.RemediationResult, error) {
	return remediationObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountRemediations(ctx context.Context, filter *entity.RemediationFilter) (int64, error) {
	return remediationObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) GetAllRemediationCursors(
	ctx context.Context,
	filter *entity.RemediationFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllRemediationCursors",
	})

	baseQuery := sq.Select("R.*").From("Remediation R").GroupBy("R.remediation_id")

	stmt, filterParameters, err := s.buildRemediationStatement(ctx, baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build Remediation cursor query: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error during close stmt: %s", err)
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
		return nil, fmt.Errorf("failed to get Remediation cursors: %w", err)
	}

	return lo.Map(rows, func(row RowComposite, _ int) string {
		r := row.AsRemediation()

		cursor, _ := EncodeCursor(WithRemediation(order, r))

		return cursor
	}), nil
}

func (s *SqlDatabase) CreateRemediation(
	remediation *entity.Remediation,
) (*entity.Remediation, error) {
	return remediationObject.Create(s.db, remediation)
}

func (s *SqlDatabase) UpdateRemediation(remediation *entity.Remediation) error {
	return remediationObject.Update(s.db, remediation)
}

func (s *SqlDatabase) DeleteRemediation(id int64, userId int64) error {
	return remediationObject.Delete(s.db, id, userId)
}
