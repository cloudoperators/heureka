// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var issueRepositoryObject = DbObject{
	Prefix:    "issuerepository",
	TableName: "IssueRepository",
	Properties: []*Property{
		NewProperty("issuerepository_name", WrapChecker(func(ir *entity.IssueRepository) bool { return ir.Name != "" })),
		NewProperty("issuerepository_url", WrapChecker(func(ir *entity.IssueRepository) bool { return ir.Url != "" })),
		NewImmutableProperty("issuerepository_created_by"),
		NewProperty("issuerepository_updated_by", WrapChecker(func(ir *entity.IssueRepository) bool { return ir.BaseIssueRepository.UpdatedBy != 0 })),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty("IR.issuerepository_name = ?", WrapRetSlice(func(filter *entity.IssueRepositoryFilter) []*string { return filter.Name })),
		NewFilterProperty("IR.issuerepository_id = ?", WrapRetSlice(func(filter *entity.IssueRepositoryFilter) []*int64 { return filter.Id })),
		NewFilterProperty("S.service_ccrn = ?", WrapRetSlice(func(filter *entity.IssueRepositoryFilter) []*string { return filter.ServiceCCRN })),
		NewFilterProperty("IRS.issuerepositoryservice_service_id = ?", WrapRetSlice(func(filter *entity.IssueRepositoryFilter) []*int64 { return filter.ServiceId })),
		NewStateFilterProperty("IR.issuerepository", WrapRetState(func(filter *entity.IssueRepositoryFilter) []entity.StateFilterType { return filter.State })),
	},
}

func ensureIssueRepositoryFilter(filter *entity.IssueRepositoryFilter) *entity.IssueRepositoryFilter {
	if filter == nil {
		filter = &entity.IssueRepositoryFilter{}
	}
	return EnsurePagination(filter)
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

func (s *SqlDatabase) buildIssueRepositoryStatement(baseQuery string, filter *entity.IssueRepositoryFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensureIssueRepositoryFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	joins := s.getIssueRepositoryJoins(filter)

	cursorFields, err := DecodeCursor(filter.Paginated.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode IssueRepository cursor: %w", err)
	}

	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.IssueRepositoryID, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

	filterStr := issueRepositoryObject.GetFilterQuery(filter)
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

	filterParameters := issueRepositoryObject.GetFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
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

	issueRepositoryRow := IssueRepositoryRow{}
	issueRepositoryRow.FromIssueRepository(issueRepository)

	query := issueRepositoryObject.InsertQuery()
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

	updateFields := issueRepositoryObject.GetUpdateFields(issueRepository)
	query := fmt.Sprintf(baseQuery, updateFields)

	issueRepositoryRow := IssueRepositoryRow{}
	issueRepositoryRow.FromIssueRepository(issueRepository)

	_, err := performExec(s, query, issueRepositoryRow, l)

	return err
}

func (s *SqlDatabase) DeleteIssueRepository(id int64, userId int64) error {
	return issueRepositoryObject.Delete(s.db, id, userId)
}
