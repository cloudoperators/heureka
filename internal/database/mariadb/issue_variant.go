// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func ensureIssueVariantFilter(f *entity.IssueVariantFilter) *entity.IssueVariantFilter {
	first := 1000
	var after string
	if f == nil {
		return &entity.IssueVariantFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id:                nil,
			SecondaryName:     nil,
			IssueId:           nil,
			IssueRepositoryId: nil,
			ServiceId:         nil,
			IssueMatchId:      nil,
		}
	}

	if f.After == nil {
		f.After = &after
	}
	if f.First == nil {
		f.First = &first
	}
	return f
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

func getIssueVariantFilterString(filter *entity.IssueVariantFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "IV.issuevariant_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.SecondaryName, "IV.issuevariant_secondary_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueId, "IV.issuevariant_issue_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueRepositoryId, "IV.issuevariant_repository_id  = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "IRS.issuerepositoryservice_service_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueMatchId, "IM.issuematch_id  = ?", OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "IV.issuevariant"))

	return combineFilterQueries(fl, OP_AND)
}

func buildIssueVariantFilterParameters(filter *entity.IssueVariantFilter, withCursor bool, cursorFields []Field) []any {
	var filterParameters []any
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.SecondaryName)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueRepositoryId)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchId)
	if withCursor {
		filterParameters = append(filterParameters, GetCursorQueryParameters(filter.Paginated.First, cursorFields)...)
	}

	return filterParameters
}

func getIssueVariantUpdateFields(issueVariant *entity.IssueVariant) string {
	fl := []string{}
	if issueVariant.SecondaryName != "" {
		fl = append(fl, "issuevariant_secondary_name = :issuevariant_secondary_name")
	}
	// if rating but not vector is passed, we need to include the vector in the update in order to overwrite any existing vector
	if issueVariant.Severity.Cvss.Vector != "" || (issueVariant.Severity.Value != "" && issueVariant.Severity.Cvss.Vector == "") {
		fl = append(fl, "issuevariant_vector = :issuevariant_vector")
	}
	if issueVariant.Severity.Value != "" {
		fl = append(fl, "issuevariant_rating = :issuevariant_rating")
	}
	if issueVariant.Description != "" {
		fl = append(fl, "issuevariant_description = :issuevariant_description")
	}
	if issueVariant.ExternalUrl != "" {
		fl = append(fl, "issuevariant_external_url = :issuevariant_external_url")
	}
	if issueVariant.IssueId != 0 {
		fl = append(fl, "issuevariant_issue_id = :issuevariant_issue_id")
	}
	if issueVariant.IssueRepositoryId != 0 {
		fl = append(fl, "issuevariant_repository_id = :issuevariant_repository_id")
	}
	if issueVariant.UpdatedBy != 0 {
		fl = append(fl, "issuevariant_updated_by = :issuevariant_updated_by")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildIssueVariantStatement(baseQuery string, filter *entity.IssueVariantFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensureIssueVariantFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := getIssueVariantFilterString(filter)
	joins := s.getIssueVariantJoins(filter)
	cursorFields, err := DecodeCursor(filter.Paginated.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode IssueVariant cursor: %w", err)
	}

	cursorQuery := CreateCursorQuery("", cursorFields)
	order = GetDefaultOrder(order, entity.IssueVariantID, entity.OrderDirectionAsc)
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

	filterParameters := buildIssueVariantFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllIssueVariantIds(filter *entity.IssueVariantFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetAllIssueVariantIds",
	})

	baseQuery := `
		SELECT IV.issuevariant_id FROM IssueVariant IV 
		%s
	 	%s GROUP BY IV.issuevariant_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildIssueVariantStatement(baseQuery, filter, false, []entity.Order{}, l)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
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

	query := `
		INSERT INTO IssueVariant (
			issuevariant_issue_id,
			issuevariant_repository_id,
			issuevariant_vector,
			issuevariant_rating,
			issuevariant_secondary_name,
			issuevariant_description,
			issuevariant_external_url,
			issuevariant_created_by,
			issuevariant_updated_by
		) VALUES (
			:issuevariant_issue_id,
			:issuevariant_repository_id,
			:issuevariant_vector,
			:issuevariant_rating,
			:issuevariant_secondary_name,
			:issuevariant_description,
			:issuevariant_external_url,
			:issuevariant_created_by,
			:issuevariant_updated_by
		)
	`

	issueVariantRow := IssueVariantRow{}
	issueVariantRow.FromIssueVariant(issueVariant)

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

	updateFields := getIssueVariantUpdateFields(issueVariant)

	query := fmt.Sprintf(baseQuery, updateFields)

	ivRow := IssueVariantRow{}
	ivRow.FromIssueVariant(issueVariant)

	_, err := performExec(s, query, ivRow, l)

	return err
}

func (s *SqlDatabase) DeleteIssueVariant(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssueVariant",
	})

	query := `
		UPDATE IssueVariant SET
		issuevariant_deleted_at = NOW(),
		issuevariant_updated_by = :userId
		WHERE issuevariant_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
