// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var serviceIssueVariantObject = DbObject[*entity.ServiceIssueVariant, *ServiceIssueVariantRow]{
	Properties: []*Property{},
	FilterProperties: []*FilterProperty{
		NewFilterProperty("CI.componentinstance_id = ?", WrapRetSlice(func(filter *entity.ServiceIssueVariantFilter) []*int64 { return filter.ComponentInstanceId })),
		NewFilterProperty("I.issue_id = ?", WrapRetSlice(func(filter *entity.ServiceIssueVariantFilter) []*int64 { return filter.IssueId })),
		NewStateFilterProperty("IV.issuevariant", WrapRetState(func(filter *entity.ServiceIssueVariantFilter) []entity.StateFilterType { return filter.State })),
	},
	NewRow: func() *ServiceIssueVariantRow { return &ServiceIssueVariantRow{} },
}

func ensureServiceIssueVariantFilter(filter *entity.ServiceIssueVariantFilter) *entity.ServiceIssueVariantFilter {
	if filter == nil {
		filter = &entity.ServiceIssueVariantFilter{}
	}
	return EnsurePagination(filter)
}

func (s *SqlDatabase) buildServiceIssueVariantStatement(baseQuery string, filter *entity.ServiceIssueVariantFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensureServiceIssueVariantFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	cursorFields, err := DecodeCursor(filter.Paginated.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode ServiceIssueVariant cursor: %w", err)
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.ServiceIssueVariantID, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

	filterStr := serviceIssueVariantObject.GetFilterQuery(filter)
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

	filterParameters := serviceIssueVariantObject.GetFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetServiceIssueVariants(filter *entity.ServiceIssueVariantFilter, order []entity.Order) ([]entity.ServiceIssueVariantResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueVariants",
	})

	baseQuery := `
		SELECT IRS.issuerepositoryservice_priority, IV.*  FROM  ComponentInstance CI
			# Join path to Issue
			INNER JOIN ComponentVersion CV on CI.componentinstance_component_version_id = CV.componentversion_id
			INNER JOIN ComponentVersionIssue CVI on CV.componentversion_id = CVI.componentversionissue_component_version_id
			INNER JOIN Issue I on CVI.componentversionissue_issue_id = I.issue_id

			# Join path to Repository
			INNER JOIN Service S on CI.componentinstance_service_id = S.service_id
			INNER JOIN IssueRepositoryService IRS on IRS.issuerepositoryservice_service_id = S.service_id
			INNER JOIN IssueRepository IR on IR.issuerepository_id = IRS.issuerepositoryservice_issue_repository_id

			# Join to from repo and issue to IssueVariant
			INNER JOIN IssueVariant IV on I.issue_id = IV.issuevariant_issue_id and IV.issuevariant_repository_id = IR.issuerepository_id
		%s
		%s ORDER BY %s LIMIT ?
    `

	stmt, filterParameters, err := s.buildServiceIssueVariantStatement(baseQuery, filter, true, order, l)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ServiceIssueVariantResult, e ServiceIssueVariantRow) []entity.ServiceIssueVariantResult {
			r := e.AsServiceIssueVariantEntry()
			cursor, _ := EncodeCursor(WithServiceIssueVariant(order, r))

			rr := entity.ServiceIssueVariantResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				ServiceIssueVariant: &r,
			}

			return append(l, rr)
		},
	)
}
