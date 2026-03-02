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

func getIssueRepositoryFilterString(filter *entity.IssueRepositoryFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Name, "IR.issuerepository_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Id, "IR.issuerepository_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceCCRN, "S.service_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "IRS.issuerepositoryservice_service_id = ?", OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "IR.issuerepository"))

	return combineFilterQueries(fl, OP_AND)
}

func buildIssueRepositoryFilterParameters(filter *entity.IssueRepositoryFilter, withCursor bool, cursorFields []Field) []any {
	var filterParameters []any
	filterParameters = buildQueryParameters(filterParameters, filter.Name)
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceCCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	if withCursor {
		filterParameters = append(filterParameters, GetCursorQueryParameters(filter.Paginated.First, cursorFields)...)
	}

	return filterParameters
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

func getIssueRepositoryUpdateFields(issueRepository *entity.IssueRepository) string {
	fl := []string{}
	if issueRepository.Name != "" {
		fl = append(fl, "issuerepository_name = :issuerepository_name")
	}
	if issueRepository.Url != "" {
		fl = append(fl, "issuerepository_url = :issuerepository_url")
	}
	if issueRepository.BaseIssueRepository.UpdatedBy != 0 {
		fl = append(fl, "issuerepository_updated_by = :issuerepository_updated_by")
	}
	return strings.Join(fl, ", ")
}

func ensureIssueRepositoryFilter(f *entity.IssueRepositoryFilter) *entity.IssueRepositoryFilter {
	var first int = 1000
	var after string
	if f == nil {
		return &entity.IssueRepositoryFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
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

func (s *SqlDatabase) buildIssueRepositoryStatement(baseQuery string, filter *entity.IssueRepositoryFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensureIssueRepositoryFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := getIssueRepositoryFilterString(filter)
	joins := s.getIssueRepositoryJoins(filter)

	cursorFields, err := DecodeCursor(filter.Paginated.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode IssueRepository cursor: %w", err)
	}

	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.IssueRepositoryID, entity.OrderDirectionAsc)
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
		return nil, nil, fmt.Errorf("failed to prepare IssueRepository statement: %w", err)
	}

	filterParameters := buildIssueRepositoryFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllIssueRepositoryIds(filter *entity.IssueRepositoryFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueRepositoryIds",
	})

	baseQuery := `
		SELECT IR.issuerepository_id FROM IssueRepository IR 
		%s
	 	%s GROUP BY IR.issuerepository_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildIssueRepositoryStatement(baseQuery, filter, false, []entity.Order{}, l)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetAllIssueRepositoryCursors(filter *entity.IssueRepositoryFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllIssueRepositoryCursors",
	})

	baseQuery := `
		SELECT IR.* FROM IssueRepository IR 
		%s
		%s GROUP BY IR.issuerepository_id ORDER BY %s
	`

	filter = ensureIssueRepositoryFilter(filter)
	stmt, filterParameters, err := s.buildIssueRepositoryStatement(baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build IssueRepository cursor query: %w", err)
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
		func(l []IssueRepositoryRow, e IssueRepositoryRow) []IssueRepositoryRow {
			return append(l, e)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get IssueRepository cursors: %w", err)
	}

	return lo.Map(rows, func(row IssueRepositoryRow, _ int) string {
		ir := row.AsIssueRepository()

		cursor, _ := EncodeCursor(WithIssueRepository(order, ir))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetIssueRepositories(filter *entity.IssueRepositoryFilter, order []entity.Order) ([]entity.IssueRepositoryResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueRepositories",
	})

	baseQuery := `
		SELECT IR.* FROM IssueRepository IR 
		%s
		%s
		%s GROUP BY IR.issuerepository_id ORDER BY %s LIMIT ?
    `

	filter = ensureIssueRepositoryFilter(filter)

	stmt, filterParameters, err := s.buildIssueRepositoryStatement(baseQuery, filter, true, order, l)
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
		func(l []entity.IssueRepositoryResult, e IssueRepositoryRow) []entity.IssueRepositoryResult {
			ir := e.AsIssueRepository()
			cursor, _ := EncodeCursor(WithIssueRepository(order, ir))

			irr := entity.IssueRepositoryResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				IssueRepository: &ir,
			}

			return append(l, irr)
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
		ORDER BY %s
	`
	stmt, filterParameters, err := s.buildIssueRepositoryStatement(baseQuery, filter, false, []entity.Order{}, l)
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
			issuerepository_url,
			issuerepository_created_by,
			issuerepository_updated_by
		) VALUES (
			:issuerepository_name,
			:issuerepository_url,
			:issuerepository_created_by,
			:issuerepository_updated_by
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

	updateFields := getIssueRepositoryUpdateFields(issueRepository)

	query := fmt.Sprintf(baseQuery, updateFields)

	issueRepositoryRow := IssueRepositoryRow{}
	issueRepositoryRow.FromIssueRepository(issueRepository)

	_, err := performExec(s, query, issueRepositoryRow, l)

	return err
}

func (s *SqlDatabase) DeleteIssueRepository(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssueRepository",
	})

	query := `
		UPDATE IssueRepository SET
		issuerepository_deleted_at = NOW(),
		issuerepository_updated_by = :userId
		WHERE issuerepository_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
