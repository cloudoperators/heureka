// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) ensureServiceIssueVariantFilter(f *entity.ServiceIssueVariantFilter) *entity.ServiceIssueVariantFilter {
	var first = 1000
	var after int64 = 0
	if f == nil {
		return &entity.ServiceIssueVariantFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			ComponentInstanceId: nil,
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

func (s *SqlDatabase) getServiceIssueVariantJoins(filter *entity.ServiceIssueVariantFilter) string {
	joins := ""
	if filter.ComponentInstanceId != nil {
		joins = `
		  INNER JOIN Issue I on IV.issuevariant_issue_id = i.issue_id
		  INNER JOIN ComponentVersionIssue CVI on I.issue_id = CVI.componentversionissue_issue_id
		  INNER JOIN ComponentVersion CV on CVI.componentversionissue_component_version_id = CV.componentversion_id
		  INNER JOIN ComponentInstance CI on CV.componentversion_id= CI.componentinstance_component_version_id
		  INNER JOIN Service S on CI.componentinstance_service_id = S.service_id
		  INNER JOIN IssueRepositoryService IRS on IRS.issuerepositoryservice_service_id = S.service_id
		  INNER JOIN IssueRepository IR on IR.issuerepository_id = IRS.issuerepositoryservice_issue_repository_id`
	}

	return joins
}

func (s *SqlDatabase) getServiceIssueVariantFilterString(filter *entity.ServiceIssueVariantFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.ComponentInstanceId, "CI.componentinstance_id = ?", OP_OR))
	fl = append(fl, "IV.issuevariant_deleted_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) buildServiceIssueVariantStatement(baseQuery string, filter *entity.ServiceIssueVariantFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureServiceIssueVariantFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getServiceIssueVariantFilterString(filter)
	joins := s.getServiceIssueVariantJoins(filter)
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
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentInstanceId)

	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetServiceIssueVariants(filter *entity.ServiceIssueVariantFilter) ([]entity.ServiceIssueVariant, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueVariants",
	})

	baseQuery := `
		SELECT IV.* FROM  IssueVariant IV 
		%s
		%s
		%s ORDER BY IV.issuevariant_id LIMIT ?
    `

	stmt, filterParameters, err := s.buildServiceIssueVariantStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ServiceIssueVariant, e ServiceIssueVariantRow) []entity.ServiceIssueVariant {
			return append(l, e.AsServiceIssueVariantEntry())
		},
	)
}
