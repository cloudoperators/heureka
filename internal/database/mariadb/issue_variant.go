// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) ensureIssueVariantFilter(f *entity.IssueVariantFilter) *entity.IssueVariantFilter {
	var first = 1000
	var after int64 = 0
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

func (s *SqlDatabase) getIssueVariantFilterString(filter *entity.IssueVariantFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "IV.issuevariant_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.SecondaryName, "IV.issuevariant_secondary_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueId, "IV.issuevariant_issue_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueRepositoryId, "IV.issuevariant_repository_id  = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "IRS.issuerepositoryservice_service_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueMatchId, "IM.issuematch_id  = ?", OP_OR))
	fl = append(fl, "IV.issuevariant_deleted_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) getIssueVariantUpdateFields(issueVariant *entity.IssueVariant) string {
	fl := []string{}
	if issueVariant.SecondaryName != "" {
		fl = append(fl, "issuevariant_secondary_name = :issuevariant_secondary_name")
	}
	if issueVariant.Severity.Cvss.Vector != "" {
		fl = append(fl, "issuevariant_vector = :issuevariant_vector")
	}
	if issueVariant.Severity.Value != "" {
		fl = append(fl, "issuevariant_rating = :issuevariant_rating")
	}
	if issueVariant.Description != "" {
		fl = append(fl, "issuevariant_description = :issuevariant_description")
	}
	if issueVariant.IssueId != 0 {
		fl = append(fl, "issuevariant_issue_id = :issuevariant_issue_id")
	}
	if issueVariant.IssueRepositoryId != 0 {
		fl = append(fl, "issuevariant_repository_id = :issuevariant_repository_id")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildIssueVariantStatement(baseQuery string, filter *entity.IssueVariantFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureIssueVariantFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getIssueVariantFilterString(filter)
	joins := s.getIssueVariantJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "IV.issuevariant_id > ?")

	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	// construct final query
	if withCursor {
		query = fmt.Sprintf(baseQuery, joins, whereClause, cursor.Statement)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause)
	}

	//construct prepared statement and if where clause does exist add parameters
	var stmt *sqlx.Stmt
	var err error

	stmt, err = s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			}).Error(msg)
		return nil, nil, fmt.Errorf("%s", msg)
	}

	//adding parameters
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.SecondaryName)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueRepositoryId)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchId)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllIssueVariantIds(filter *entity.IssueVariantFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetAllIssueVariantIds",
	})

	baseQuery := `
		SELECT IV.issuevariant_id FROM IssueVariant IV 
		%s
	 	%s GROUP BY IV.issuevariant_id ORDER BY IV.issuevariant_id
    `

	stmt, filterParameters, err := s.buildIssueVariantStatement(baseQuery, filter, false, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetIssueVariants(filter *entity.IssueVariantFilter) ([]entity.IssueVariant, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueVariants",
	})

	baseQuery := `
		SELECT IV.* FROM  IssueVariant IV 
		%s
		%s
		%s ORDER BY IV.issuevariant_id LIMIT ?
    `

	stmt, filterParameters, err := s.buildIssueVariantStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueVariant, e IssueVariantWithRepository) []entity.IssueVariant {
			return append(l, e.AsIssueVariantEntry())
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
    `
	stmt, filterParameters, err := s.buildIssueVariantStatement(baseQuery, filter, false, l)

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
			issuevariant_description
		) VALUES (
			:issuevariant_issue_id,
			:issuevariant_repository_id,
			:issuevariant_vector,
			:issuevariant_rating,
			:issuevariant_secondary_name,
			:issuevariant_description
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

	updateFields := s.getIssueVariantUpdateFields(issueVariant)

	query := fmt.Sprintf(baseQuery, updateFields)

	ivRow := IssueVariantRow{}
	ivRow.FromIssueVariant(issueVariant)

	_, err := performExec(s, query, ivRow, l)

	return err
}

func (s *SqlDatabase) DeleteIssueVariant(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssueVariant",
	})

	query := `
		UPDATE IssueVariant SET
		issuevariant_deleted_at = NOW()
		WHERE issuevariant_id = :id
	`

	args := map[string]interface{}{
		"id": id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
