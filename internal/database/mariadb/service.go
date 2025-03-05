// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

const (
	serviceWildCardFilterQuery = "S.service_ccrn LIKE Concat('%',?,'%')"
)

func (s *SqlDatabase) buildServiceFilterParameters(filter *entity.ServiceFilter, withCursor bool, cursorFields []Field) []interface{} {
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.SupportGroupCCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.CCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.OwnerName)
	filterParameters = buildQueryParameters(filterParameters, filter.ActivityId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentInstanceId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueRepositoryId)
	filterParameters = buildQueryParameters(filterParameters, filter.SupportGroupId)
	filterParameters = buildQueryParameters(filterParameters, filter.OwnerId)
	filterParameters = buildQueryParameters(filterParameters, filter.Search)
	if withCursor {
		p := CreateCursorParameters([]any{}, cursorFields)
		filterParameters = append(filterParameters, p...)
		if filter.PaginatedX.First == nil {
			filterParameters = append(filterParameters, 1000)
		} else {
			filterParameters = append(filterParameters, filter.PaginatedX.First)
		}
	}
	return filterParameters
}

func (s *SqlDatabase) getServiceFilterString(filter *entity.ServiceFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.CCRN, "S.service_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Id, "S.service_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.SupportGroupCCRN, "SG.supportgroup_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.OwnerName, "U.user_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ActivityId, "A.activity_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentInstanceId, "CI.componentinstance_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueRepositoryId, "IRS.issuerepositoryservice_issue_repository_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.SupportGroupId, "SGS.supportgroupservice_support_group_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.OwnerId, "O.owner_user_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Search, serviceWildCardFilterQuery, OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "S.service"))

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) getServiceJoins(filter *entity.ServiceFilter) string {
	joins := ""
	if len(filter.OwnerName) > 0 || len(filter.OwnerId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN Owner O on S.service_id = O.owner_service_id
		`)
	}
	if len(filter.OwnerName) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN User U on U.user_id = O.owner_user_id
		`)
	}
	if len(filter.SupportGroupCCRN) > 0 || len(filter.SupportGroupId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN SupportGroupService SGS on S.service_id = SGS.supportgroupservice_service_id
		`)
	}
	if len(filter.SupportGroupCCRN) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN SupportGroup SG on SG.supportgroup_id = SGS.supportgroupservice_support_group_id
		`)
	}
	if len(filter.ActivityId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ActivityHasService AHS on S.service_id = AHS.activityhasservice_service_id
         	LEFT JOIN Activity A on AHS.activityhasservice_activity_id = A.activity_id
		`)
	}
	if len(filter.ComponentInstanceId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentInstance CI on S.service_id = CI.componentinstance_service_id
		`)
	}
	if len(filter.IssueRepositoryId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueRepositoryService IRS on IRS.issuerepositoryservice_service_id = S.service_id
		`)
	}
	return joins
}

func (s *SqlDatabase) getServiceColumns(filter *entity.ServiceFilter) string {
	columns := "S.*"
	if len(filter.IssueRepositoryId) > 0 {
		columns = fmt.Sprintf("%s, %s", columns, "IRS.*")
	}
	return columns
}

func (s *SqlDatabase) ensureServiceFilter(f *entity.ServiceFilter) *entity.ServiceFilter {
	var first int = 1000
	var after string = ""
	if f == nil {
		return &entity.ServiceFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			SupportGroupCCRN:  nil,
			CCRN:              nil,
			Id:                nil,
			OwnerName:         nil,
			SupportGroupId:    nil,
			ActivityId:        nil,
			IssueRepositoryId: nil,
			OwnerId:           nil,
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

func (s *SqlDatabase) getServiceUpdateFields(service *entity.Service) string {
	fl := []string{}
	if service.CCRN != "" {
		fl = append(fl, "service_ccrn = :service_ccrn")
	}
	if service.BaseService.UpdatedBy != 0 {
		fl = append(fl, "service_updated_by = :service_updated_by")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildServiceStatement(baseQuery string, filter *entity.ServiceFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureServiceFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getServiceFilterString(filter)
	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, nil, err
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.ServiceId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)
	joins := s.getServiceJoins(filter)

	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && withCursor && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	// construct final query
	if withCursor {
		query = fmt.Sprintf(baseQuery, joins, whereClause, cursorQuery, orderStr)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause, orderStr)
	}

	//construct prepared statement and if where clause does exist add parameters
	var stmt *sqlx.Stmt

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
	filterParameters := s.buildServiceFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) CountServices(filter *entity.ServiceFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountServices",
	})

	baseQuery := `
		SELECT count(distinct S.service_id) FROM Service S
		%s
		%s
        ORDER BY %s
	`
	stmt, filterParameters, err := s.buildServiceStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetAllServiceIds(filter *entity.ServiceFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetServiceIds",
	})

	baseQuery := `
		SELECT S.service_id FROM Service S 
		%s
	 	%s GROUP BY S.service_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildServiceStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetServices(filter *entity.ServiceFilter, order []entity.Order) ([]entity.ServiceResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetServices",
	})

	baseQuery := `
		SELECT %s FROM Service S
		%s
		%s
		%s GROUP BY S.service_id ORDER BY %s LIMIT ?
    `

	filter = s.ensureServiceFilter(filter)
	columns := s.getServiceColumns(filter)
	baseQuery = fmt.Sprintf(baseQuery, columns, "%s", "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildServiceStatement(baseQuery, filter, true, order, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ServiceResult, e RowComposite) []entity.ServiceResult {
			s := entity.Service{
				BaseService: e.AsBaseService(),
			}

			cursor, _ := EncodeCursor(WithService(order, s))

			sr := entity.ServiceResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				Service: &s,
			}
			return append(l, sr)
		},
	)
}

func (s *SqlDatabase) GetServicesWithAggregations(filter *entity.ServiceFilter, order []entity.Order) ([]entity.ServiceResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetServicesWithAggregations",
	})

	baseImQuery := `
        SELECT %s, COUNT(IM.issuematch_id) AS service_agg_issue_matches FROM Service S
        %s
        LEFT JOIN ComponentInstance CI on S.service_id = CI.componentinstance_service_id
        LEFT JOIN IssueMatch IM on CI.componentinstance_id = IM.issuematch_component_instance_id
        %s
        %s GROUP BY S.service_id ORDER BY %s LIMIT ?
    `

	baseCiQuery := `
        SELECT %s, SUM(CI.componentinstance_count) AS service_agg_component_instances FROM Service S
        %s
        LEFT JOIN ComponentInstance CI on S.service_id = CI.componentinstance_service_id
        %s
        %s GROUP BY S.service_id ORDER BY %s LIMIT ?
    `

	baseQuery := `
        WITH IssueMatchCounts AS (
            %s
        ),
        ComponentInstanceCounts AS (
            %s
        )
        SELECT IMC.*, CIC.*
        FROM ComponentInstanceCounts CIC
        JOIN IssueMatchCounts IMC ON CIC.service_id = IMC.service_id;
    `
	filter = s.ensureServiceFilter(filter)
	filterStr := s.getServiceFilterString(filter)
	order = GetDefaultOrder(order, entity.ServiceId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)
	joins := s.getServiceJoins(filter)
	columns := s.getServiceColumns(filter)
	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, err
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	whereClause := ""
	if filterStr != "" || cursorQuery != "" {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	imQuery := fmt.Sprintf(baseImQuery, columns, joins, whereClause, cursorQuery, orderStr)
	ciQuery := fmt.Sprintf(baseCiQuery, columns, joins, whereClause, cursorQuery, orderStr)
	query := fmt.Sprintf(baseQuery, imQuery, ciQuery)

	var stmt *sqlx.Stmt

	stmt, err = s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			}).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}

	// parameters for issue match query
	filterParameters := s.buildServiceFilterParameters(filter, true, cursorFields)
	// parameters for component instance query
	filterParameters = append(filterParameters, s.buildServiceFilterParameters(filter, true, cursorFields)...)

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ServiceResult, e RowComposite) []entity.ServiceResult {
			service := entity.Service{
				BaseService: e.AsBaseService(),
			}
			aggregations := e.AsServiceAggregations()

			cursor, _ := EncodeCursor(WithService(order, service))

			sr := entity.ServiceResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				Service:             &service,
				ServiceAggregations: &aggregations,
			}
			return append(l, sr)
		},
	)
}

func (s *SqlDatabase) GetAllServiceCursors(filter *entity.ServiceFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssueAllServiceCursors",
	})

	baseQuery := `
		SELECT S.* FROM Service S 
		%s
	    %s GROUP BY S.service_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildServiceStatement(baseQuery, filter, false, order, l)

	if err != nil {
		return nil, err
	}

	rows, err := performListScan(
		stmt,
		filterParameters,
		l,
		func(l []RowComposite, e RowComposite) []RowComposite {
			return append(l, e)
		},
	)

	if err != nil {
		return nil, err
	}

	return lo.Map(rows, func(row RowComposite, _ int) string {
		s := entity.Service{
			BaseService: row.AsBaseService(),
		}

		cursor, _ := EncodeCursor(WithService(order, s))

		return cursor
	}), nil
}

func (s *SqlDatabase) CreateService(service *entity.Service) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"service": service,
		"event":   "database.CreateService",
	})

	query := `
		INSERT INTO Service (
			service_ccrn,
			service_created_by,
			service_updated_by
		) VALUES (
			:service_ccrn,
			:service_created_by,
			:service_updated_by
		)
	`

	serviceRow := ServiceRow{}
	serviceRow.FromService(service)

	id, err := performInsert(s, query, serviceRow, l)

	if err != nil {
		return nil, err
	}

	service.Id = id

	return service, nil
}

func (s *SqlDatabase) UpdateService(service *entity.Service) error {
	l := logrus.WithFields(logrus.Fields{
		"service": service,
		"event":   "database.UpdateService",
	})

	baseQuery := `
		UPDATE Service SET
		%s
		WHERE service_id = :service_id
	`

	updateFields := s.getServiceUpdateFields(service)

	query := fmt.Sprintf(baseQuery, updateFields)

	serviceRow := ServiceRow{}
	serviceRow.FromService(service)

	_, err := performExec(s, query, serviceRow, l)

	return err
}

func (s *SqlDatabase) DeleteService(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteService",
	})

	query := `
		UPDATE Service SET
		service_deleted_at = NOW(),
		service_updated_by = :userId
		WHERE service_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) AddOwnerToService(serviceId int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"serviceId": serviceId,
		"userId":    userId,
		"event":     "database.AddOwnerToService",
	})

	query := `
		INSERT INTO Owner (
			owner_service_id,
			owner_user_id
		) VALUES (
			:service_id,
			:user_id
		)
	`

	args := map[string]interface{}{
		"service_id": serviceId,
		"user_id":    userId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) RemoveOwnerFromService(serviceId int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"serviceId": serviceId,
		"userId":    userId,
		"event":     "database.RemoveOwnerFromService",
	})

	query := `
		DELETE FROM Owner
		WHERE owner_service_id = :service_id
		AND owner_user_id = :user_id
	`

	args := map[string]interface{}{
		"service_id": serviceId,
		"user_id":    userId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) AddIssueRepositoryToService(serviceId int64, issueRepositoryId int64, priority int64) error {
	l := logrus.WithFields(logrus.Fields{
		"serviceId":         serviceId,
		"issueRepositoryId": issueRepositoryId,
		"event":             "database.AddIssueRepositoryToService",
	})

	query := `
		INSERT INTO IssueRepositoryService (
			issuerepositoryservice_service_id,
			issuerepositoryservice_issue_repository_id,
			issuerepositoryservice_priority
		) VALUES (
		 :service_id,
		 :issue_repository_id,
		 :priority
		)
	`

	args := map[string]interface{}{
		"service_id":          serviceId,
		"issue_repository_id": issueRepositoryId,
		"priority":            priority,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) RemoveIssueRepositoryFromService(serviceId int64, issueRepositoryId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"serviceId":         serviceId,
		"issueRepositoryId": issueRepositoryId,
		"event":             "database.RemoveIssueRepositoryFromService",
	})

	query := `
		DELETE FROM IssueRepositoryService
		WHERE issuerepositoryservice_service_id = :service_id
		AND issuerepositoryservice_issue_repository_id = :issue_repository_id
	`

	args := map[string]interface{}{
		"service_id":          serviceId,
		"issue_repository_id": issueRepositoryId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) GetServiceCcrns(filter *entity.ServiceFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetServiceCcrns",
	})

	baseQuery := `
    SELECT service_ccrn FROM Service S
    %s
    %s
    ORDER BY %s
    `

	// Ensure the filter is initialized
	filter = s.ensureServiceFilter(filter)
	order := []entity.Order{
		{By: entity.ServiceCcrn, Direction: entity.OrderDirectionAsc},
	}

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildServiceStatement(baseQuery, filter, false, order, l)
	if err != nil {
		l.Error("Error preparing statement: ", err)
		return nil, err
	}
	defer stmt.Close()

	// Execute the query
	rows, err := stmt.Queryx(filterParameters...)
	if err != nil {
		l.Error("Error executing query: ", err)
		return nil, err
	}
	defer rows.Close()

	// Collect the results
	serviceCcrns := []string{}
	var ccrn string
	for rows.Next() {
		if err := rows.Scan(&ccrn); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}
		serviceCcrns = append(serviceCcrns, ccrn)
	}
	if err = rows.Err(); err != nil {
		l.Error("Row iteration error: ", err)
		return nil, err
	}

	return serviceCcrns, nil
}
