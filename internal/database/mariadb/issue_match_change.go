// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

func ensureIssueMatchChangeFilter(f *entity.IssueMatchChangeFilter) *entity.IssueMatchChangeFilter {
	if f != nil {
		return f
	}

	first := 1000
	var after int64 = 0
	return &entity.IssueMatchChangeFilter{
		Paginated: entity.Paginated{
			First: &first,
			After: &after,
		},
		Id:           nil,
		ActivityId:   nil,
		IssueMatchId: nil,
		Action:       nil,
	}
}

func getIssueMatchChangeFilterString(filter *entity.IssueMatchChangeFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "IMC.issuematchchange_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ActivityId, "IMC.issuematchchange_activity_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueMatchId, "IMC.issuematchchange_issue_match_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Action, "IMC.issuematchchange_action = ?", OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "IMC.issuematchchange"))

	return combineFilterQueries(fl, OP_AND)
}

func getIssueMatchChangeUpdateFields(imc *entity.IssueMatchChange) string {
	fl := []string{}
	if imc.Action != "" {
		fl = append(fl, "issuematchchange_action = :issuematchchange_action")
	}
	if imc.UpdatedBy != 0 {
		fl = append(fl, "issuematchchange_updated_by = :issuematchchange_updated_by")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildIssueMatchChangeStatement(baseQuery string, filter *entity.IssueMatchChangeFilter, withCursor bool, l *logrus.Entry) (Stmt, []interface{}, error) {
	var query string
	filter = ensureIssueMatchChangeFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := getIssueMatchChangeFilterString(filter)
	cursor := getCursor(filter.Paginated, filterStr, "IMC.issuematchchange_id > ?")

	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	// construct final query
	if withCursor {
		query = fmt.Sprintf(baseQuery, whereClause, cursor.Statement)
	} else {
		query = fmt.Sprintf(baseQuery, whereClause)
	}

	// construct prepared statement and if where clause does exist add parameters
	stmt, err := s.db.Preparex(query)
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

	// adding parameters
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.ActivityId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchId)
	filterParameters = buildQueryParameters(filterParameters, filter.Action)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllIssueMatchChangeIds(filter *entity.IssueMatchChangeFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueMatchChangeIds",
	})

	baseQuery := `
		SELECT IMC.issuematchchange_id FROM IssueMatchChange IMC 
		%s GROUP BY IMC.issuematchchange_id ORDER BY IMC.issuematchchange_id
    `

	stmt, filterParameters, err := s.buildIssueMatchChangeStatement(baseQuery, filter, false, l)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetIssueMatchChanges(filter *entity.IssueMatchChangeFilter) ([]entity.IssueMatchChange, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueMatchChanges",
	})

	baseQuery := `
		SELECT IMC.* FROM IssueMatchChange IMC
		%s %s GROUP BY IMC.issuematchchange_id ORDER BY IMC.issuematchchange_id LIMIT ?
	`

	filter = ensureIssueMatchChangeFilter(filter)
	baseQuery = fmt.Sprintf(baseQuery, "%s", "%s")

	stmt, filterParameters, err := s.buildIssueMatchChangeStatement(baseQuery, filter, true, l)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueMatchChange, e IssueMatchChangeRow) []entity.IssueMatchChange {
			return append(l, e.AsIssueMatchChange())
		},
	)
}

func (s *SqlDatabase) CountIssueMatchChanges(filter *entity.IssueMatchChangeFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssueMatchChanges",
	})

	baseQuery := `
		SELECT count(distinct IMC.issuematchchange_id) FROM IssueMatchChange IMC 
		%s
	`
	stmt, filterParameters, err := s.buildIssueMatchChangeStatement(baseQuery, filter, false, l)
	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateIssueMatchChange(imc *entity.IssueMatchChange) (*entity.IssueMatchChange, error) {
	l := logrus.WithFields(logrus.Fields{
		"issueMatchChange": imc,
		"event":            "database.CreateIssueMatchchange",
	})

	query := `
		INSERT INTO IssueMatchChange (
			issuematchchange_action,
			issuematchchange_activity_id,
			issuematchchange_issue_match_id,
			issuematchchange_created_by,
			issuematchchange_updated_by
		) VALUES (
			:issuematchchange_action,
			:issuematchchange_activity_id,
			:issuematchchange_issue_match_id,
			:issuematchchange_created_by,
			:issuematchchange_updated_by
		)
	`

	imcRow := IssueMatchChangeRow{}
	imcRow.FromIssueMatchChange(imc)

	id, err := performInsert(s, query, imcRow, l)
	if err != nil {
		return nil, err
	}

	imc.Id = id

	return imc, nil
}

func (s *SqlDatabase) UpdateIssueMatchChange(imc *entity.IssueMatchChange) error {
	l := logrus.WithFields(logrus.Fields{
		"issueMatchChange": imc,
		"event":            "database.UpdateIssueMatchChange",
	})

	baseQuery := `
		UPDATE IssueMatchChange SET
		%s
		WHERE issuematchchange_id = :issuematchchange_id
	`

	updateFields := getIssueMatchChangeUpdateFields(imc)

	query := fmt.Sprintf(baseQuery, updateFields)

	imcRow := IssueMatchChangeRow{}
	imcRow.FromIssueMatchChange(imc)

	_, err := performExec(s, query, imcRow, l)

	return err
}

func (s *SqlDatabase) DeleteIssueMatchChange(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssueMatchChange",
	})

	query := `
		UPDATE IssueMatchChange SET
		issuematchchange_deleted_at = NOW(),
		issuematchchange_updated_by = :userId
		WHERE issuematchchange_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
