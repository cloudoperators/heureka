// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var issueVariantObject = DbObject{
	Prefix:    "issuevariant",
	TableName: "IssueVariant",
	Properties: []*Property{
		NewProperty("issuevariant_issue_id", WrapChecker(func(iv *entity.IssueVariant) bool { return iv.IssueId != 0 })),
		NewProperty("issuevariant_repository_id", WrapChecker(func(iv *entity.IssueVariant) bool { return iv.IssueRepositoryId != 0 })),
		// if rating but not vector is passed, we need to include the vector in the update in order to overwrite any existing vector
		NewProperty("issuevariant_vector", WrapChecker(func(iv *entity.IssueVariant) bool {
			return iv.Severity.Cvss.Vector != "" || (iv.Severity.Value != "" && iv.Severity.Cvss.Vector == "")
		})),
		NewProperty("issuevariant_rating", WrapChecker(func(iv *entity.IssueVariant) bool { return iv.Severity.Value != "" })),
		NewProperty("issuevariant_secondary_name", WrapChecker(func(iv *entity.IssueVariant) bool { return iv.SecondaryName != "" })),
		NewProperty("issuevariant_description", WrapChecker(func(iv *entity.IssueVariant) bool { return iv.Description != "" })),
		NewProperty("issuevariant_external_url", WrapChecker(func(iv *entity.IssueVariant) bool { return iv.ExternalUrl != "" })),
		NewImmutableProperty("issuevariant_created_by"),
		NewProperty("issuevariant_updated_by", WrapChecker(func(iv *entity.IssueVariant) bool { return iv.UpdatedBy != 0 })),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty("IV.issuevariant_id = ?", WrapRetSlice(func(filter *entity.IssueVariantFilter) []*int64 { return filter.Id })),
		NewFilterProperty("IV.issuevariant_secondary_name = ?", WrapRetSlice(func(filter *entity.IssueVariantFilter) []*string { return filter.SecondaryName })),
		NewFilterProperty("IV.issuevariant_issue_id = ?", WrapRetSlice(func(filter *entity.IssueVariantFilter) []*int64 { return filter.IssueId })),
		NewFilterProperty("IV.issuevariant_repository_id = ?", WrapRetSlice(func(filter *entity.IssueVariantFilter) []*int64 { return filter.IssueRepositoryId })),
		NewFilterProperty("IRS.issuerepositoryservice_service_id = ?", WrapRetSlice(func(filter *entity.IssueVariantFilter) []*int64 { return filter.ServiceId })),
		NewFilterProperty("IM.issuematch_id = ?", WrapRetSlice(func(filter *entity.IssueVariantFilter) []*int64 { return filter.IssueMatchId })),
		NewStateFilterProperty("IV.issuevariant", WrapRetState(func(filter *entity.IssueVariantFilter) []entity.StateFilterType { return filter.State })),
	},
}

func ensureIssueVariantFilter(filter *entity.IssueVariantFilter) *entity.IssueVariantFilter {
	if filter == nil {
		filter = &entity.IssueVariantFilter{}
	}
	return EnsurePagination(filter)
}

func (s *SqlDatabase) getIssueVariantJoins(filter *entity.IssueVariantFilter) string {
	joins := "INNER JOIN IssueRepository IR on IV.issuevariant_repository_id = IR.issuerepository_id"

	if len(filter.ServiceId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			INNER JOIN IssueRepositoryService IRS on IRS.issuerepositoryservice_issue_repository_id = IR.issuerepository_id
		`)
	}

	if len(filter.IssueId) > 0 || len(filter.IssueMatchId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			INNER JOIN Issue I on IV.issuevariant_issue_id = I.issue_id
		`)
	}

	if len(filter.IssueMatchId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			INNER JOIN IssueMatch IM on IM.issuematch_issue_id = I.issue_id
		`)
	}

	return joins
}

func (s *SqlDatabase) buildIssueVariantStatement(baseQuery string, filter *entity.IssueVariantFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensureIssueVariantFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	joins := s.getIssueVariantJoins(filter)
	cursorFields, err := DecodeCursor(filter.Paginated.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode IssueVariant cursor: %w", err)
	}

	cursorQuery := CreateCursorQuery("", cursorFields)
	order = GetDefaultOrder(order, entity.IssueVariantID, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

	filterStr := issueVariantObject.GetFilterQuery(filter)
	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && withCursor && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	var query string
	if withCursor {
		query = fmt.Sprintf(baseQuery, joins, whereClause, cursorQuery, orderStr)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause, orderStr)
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
		return nil, nil, fmt.Errorf("failed to prepare IssueVariant statement: %w", err)
	}

	filterParameters := issueVariantObject.GetFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllIssueVariantCursors(filter *entity.IssueVariantFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllIssueVariantCursors",
	})

	baseQuery := `
		SELECT IV.* FROM  IssueVariant IV 
		%s
		%s GROUP BY IV.issuevariant_id ORDER BY %s
	`

	filter = ensureIssueVariantFilter(filter)
	stmt, filterParameters, err := s.buildIssueVariantStatement(baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build IssueVariant cursor query: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error while close statement: %s", err.Error())
		}
	}()

	rows, err := performListScan(
		stmt,
		filterParameters,
		l,
		func(l []IssueVariantRow, e IssueVariantRow) []IssueVariantRow {
			return append(l, e)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get IssueVariant cursors: %w", err)
	}

	return lo.Map(rows, func(row IssueVariantRow, _ int) string {
		iv := row.AsIssueVariant(&entity.IssueRepository{})

		cursor, _ := EncodeCursor(WithIssueVariant(order, iv))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetIssueVariants(filter *entity.IssueVariantFilter, order []entity.Order) ([]entity.IssueVariantResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueVariants",
	})

	baseQuery := `
		SELECT IV.* FROM  IssueVariant IV 
		%s
		%s
		%s ORDER BY %s LIMIT ?
    `

	stmt, filterParameters, err := s.buildIssueVariantStatement(baseQuery, filter, true, order, l)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error while close statement: %s", err.Error())
		}
	}()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueVariantResult, e IssueVariantRow) []entity.IssueVariantResult {
			iv := e.AsIssueVariant(&entity.IssueRepository{})
			cursor, _ := EncodeCursor(WithIssueVariant(order, iv))

			ivr := entity.IssueVariantResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				IssueVariant: &iv,
			}

			return append(l, ivr)
		},
	)
}

func (s *SqlDatabase) CountIssueVariants(filter *entity.IssueVariantFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssueVariants",
	})

	baseQuery := `
		SELECT count(distinct IV.issuevariant_id) FROM  IssueVariant IV 
		%s
		%s
		ORDER BY %s
    `
	stmt, filterParameters, err := s.buildIssueVariantStatement(baseQuery, filter, false, []entity.Order{}, l)
	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(
		stmt,
		filterParameters,
		l,
	)
}

func (s *SqlDatabase) CreateIssueVariant(issueVariant *entity.IssueVariant) (*entity.IssueVariant, error) {
	l := logrus.WithFields(logrus.Fields{
		"issueVariant": issueVariant,
		"event":        "database.CreateIssueVariant",
	})

	issueVariantRow := IssueVariantRow{}
	issueVariantRow.FromIssueVariant(issueVariant)

	query := issueVariantObject.InsertQuery()
	id, err := performInsert(s, query, issueVariantRow, l)
	if err != nil {
		return nil, err
	}

	issueVariant.Id = id

	return issueVariant, nil
}

func (s *SqlDatabase) UpdateIssueVariant(issueVariant *entity.IssueVariant) error {
	l := logrus.WithFields(logrus.Fields{
		"issueVariant": issueVariant,
		"event":        "database.UpdateIssueVariant",
	})

	baseQuery := `
		UPDATE IssueVariant SET
		%s
		WHERE issuevariant_id = :issuevariant_id
	`

	updateFields := issueVariantObject.GetUpdateFields(issueVariant)

	query := fmt.Sprintf(baseQuery, updateFields)

	ivRow := IssueVariantRow{}
	ivRow.FromIssueVariant(issueVariant)

	_, err := performExec(s, query, ivRow, l)

	return err
}

func (s *SqlDatabase) DeleteIssueVariant(id int64, userId int64) error {
	return issueVariantObject.Delete(s.db, id, userId)
}
