// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"github.com/samber/lo"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) getIssueRepositoryFilterString(filter *entity.IssueRepositoryFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Name, "IR.issuerepository_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Id, "IR.issuerepository_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceCCRN, "S.service_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "IRS.issuerepositoryservice_service_id = ?", OP_OR))
	fl = append(fl, "IR.issuerepository_deleted_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) getIssueRepositoryJoins(filter *entity.IssueRepositoryFilter) string {
	joins := ""
	if len(filter.ServiceId) > 0 || len(filter.ServiceCCRN) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueRepositoryService IRS on IR.issuerepository_id = IRS.issuerepositoryservice_issue_repository_id
		`)
	}
	if len(filter.ServiceCCRN) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN Service S on S.service_id = IRS.issuerepositoryservice_service_id
		`)
	}
	return joins
}

func (s *SqlDatabase) getIssueRepositoryUpdateFields(issueRepository *entity.IssueRepository) string {
	fl := []string{}
	if issueRepository.Name != "" {
		fl = append(fl, "issuerepository_name = :issuerepository_name")
	}
	if issueRepository.Url != "" {
		fl = append(fl, "issuerepository_url = :issuerepository_url")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) getIssueRepositoryColumns(filter *entity.IssueRepositoryFilter) string {
	columns := "IR.*"
	if len(filter.ServiceId) > 0 || len(filter.ServiceCCRN) > 0 {
		columns = fmt.Sprintf("%s, %s", columns, "IRS.*")
	}
	return columns
}

func (s *SqlDatabase) ensureIssueRepositoryFilter(f *entity.IssueRepositoryFilter) *entity.IssueRepositoryFilter {
	var first int = 1000
	var after int64 = 0
	if f == nil {
		return &entity.IssueRepositoryFilter{
			Paginated: entity.Paginated{
				First:  &first,
				After:  &after,
				Cursor: lo.ToPtr(""),
			},
			ServiceId:   nil,
			Name:        nil,
			Id:          nil,
			ServiceCCRN: nil,
		}
	}
	if f.First == nil {
		f.First = &first
	}
	if f.After == nil {
		f.After = &after
	}
	return f
}

func (s *SqlDatabase) buildIssueRepositoryStatement(baseQuery string, filter *entity.IssueRepositoryFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureIssueRepositoryFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getIssueRepositoryFilterString(filter)
	joins := s.getIssueRepositoryJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "IR.issuerepository_id > ?")

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
	filterParameters = buildQueryParameters(filterParameters, filter.Name)
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceCCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllIssueRepositoryIds(filter *entity.IssueRepositoryFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueRepositoryIds",
	})

	baseQuery := `
		SELECT IR.issuerepository_id FROM IssueRepository IR 
		%s
	 	%s GROUP BY IR.issuerepository_id ORDER BY IR.issuerepository_id
    `

	stmt, filterParameters, err := s.buildIssueRepositoryStatement(baseQuery, filter, false, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetIssueRepositories(filter *entity.IssueRepositoryFilter) ([]entity.IssueRepository, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueRepositories",
	})

	baseQuery := `
		SELECT %s FROM IssueRepository IR 
		%s
		%s
		%s GROUP BY IR.issuerepository_id ORDER BY IR.issuerepository_id LIMIT ?
    `

	filter = s.ensureIssueRepositoryFilter(filter)
	columns := s.getIssueRepositoryColumns(filter)
	baseQuery = fmt.Sprintf(baseQuery, columns, "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildIssueRepositoryStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueRepository, e IssueRepositoryRow) []entity.IssueRepository {
			return append(l, e.AsIssueRepository())
		},
	)
}

func (s *SqlDatabase) CountIssueRepositories(filter *entity.IssueRepositoryFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssueRepositories",
	})

	baseQuery := `
		SELECT count(distinct IR.issuerepository_id) FROM IssueRepository IR 
		%s
		%s
	`
	stmt, filterParameters, err := s.buildIssueRepositoryStatement(baseQuery, filter, false, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateIssueRepository(issueRepository *entity.IssueRepository) (*entity.IssueRepository, error) {
	l := logrus.WithFields(logrus.Fields{
		"issueRepository": issueRepository,
		"event":           "database.CreateIssueRepository",
	})

	query := `
		INSERT INTO IssueRepository (
			issuerepository_name,
			issuerepository_url
		) VALUES (
			:issuerepository_name,
			:issuerepository_url
		)
	`

	issueRepositoryRow := IssueRepositoryRow{}
	issueRepositoryRow.FromIssueRepository(issueRepository)

	id, err := performInsert(s, query, issueRepositoryRow, l)

	if err != nil {
		return nil, err
	}

	issueRepository.Id = id

	return issueRepository, nil
}

func (s *SqlDatabase) UpdateIssueRepository(issueRepository *entity.IssueRepository) error {
	l := logrus.WithFields(logrus.Fields{
		"issueRepository": issueRepository,
		"event":           "database.UpdateIssueRepository",
	})

	baseQuery := `
		UPDATE IssueRepository SET
		%s
		WHERE issuerepository_id = :issuerepository_id
	`

	updateFields := s.getIssueRepositoryUpdateFields(issueRepository)

	query := fmt.Sprintf(baseQuery, updateFields)

	issueRepositoryRow := IssueRepositoryRow{}
	issueRepositoryRow.FromIssueRepository(issueRepository)

	_, err := performExec(s, query, issueRepositoryRow, l)

	return err
}

func (s *SqlDatabase) DeleteIssueRepository(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssueRepository",
	})

	query := `
		UPDATE IssueRepository SET
		issuerepository_deleted_at = NOW()
		WHERE issuerepository_id = :id
	`

	args := map[string]interface{}{
		"id": id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
