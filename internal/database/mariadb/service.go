// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"errors"
	"fmt"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/go-sql-driver/mysql"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var serviceObject = DbObject[*entity.Service]{
	Prefix:    "service",
	TableName: "Service",
	Properties: []*Property{
		NewProperty(
			"service_ccrn",
			WrapAccess(func(s *entity.Service) (string, bool) { return s.CCRN, s.CCRN != "" }),
		),
		NewProperty(
			"service_domain",
			WrapAccess(func(s *entity.Service) (string, bool) { return s.Domain, s.Domain != "" }),
		),
		NewProperty(
			"service_region",
			WrapAccess(func(s *entity.Service) (string, bool) { return s.Region, s.Region != "" }),
		),
		NewProperty(
			"service_created_by",
			WrapAccess(
				func(s *entity.Service) (int64, bool) { return s.BaseService.CreatedBy, NoUpdate },
			),
		),
		NewProperty(
			"service_updated_by",
			WrapAccess(
				func(s *entity.Service) (int64, bool) { return s.BaseService.UpdatedBy, s.BaseService.UpdatedBy != 0 },
			),
		),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty(
			"S.service_ccrn = ?",
			WrapRetSlice(func(filter *entity.ServiceFilter) []*string { return filter.CCRN }),
		),
		NewFilterProperty(
			"S.service_domain = ?",
			WrapRetSlice(func(filter *entity.ServiceFilter) []*string { return filter.Domain }),
		),
		NewFilterProperty(
			"S.service_region = ?",
			WrapRetSlice(func(filter *entity.ServiceFilter) []*string { return filter.Region }),
		),
		NewFilterProperty(
			"S.service_id = ?",
			WrapRetSlice(func(filter *entity.ServiceFilter) []*int64 { return filter.Id }),
		),
		NewFilterProperty(
			"SG.supportgroup_ccrn = ?",
			WrapRetSlice(
				func(filter *entity.ServiceFilter) []*string { return filter.SupportGroupCCRN },
			),
		),
		NewFilterProperty(
			"U.user_name = ?",
			WrapRetSlice(func(filter *entity.ServiceFilter) []*string { return filter.OwnerName }),
		),
		NewFilterProperty(
			"IM.issuematch_issue_id = ?",
			WrapRetSlice(func(filter *entity.ServiceFilter) []*int64 { return filter.IssueId }),
		),
		NewFilterProperty(
			"CI.componentinstance_id = ?",
			WrapRetSlice(
				func(filter *entity.ServiceFilter) []*int64 { return filter.ComponentInstanceId },
			),
		),
		NewFilterProperty(
			"IRS.issuerepositoryservice_issue_repository_id = ?",
			WrapRetSlice(
				func(filter *entity.ServiceFilter) []*int64 { return filter.IssueRepositoryId },
			),
		),
		NewFilterProperty(
			"SGS.supportgroupservice_support_group_id = ?",
			WrapRetSlice(
				func(filter *entity.ServiceFilter) []*int64 { return filter.SupportGroupId },
			),
		),
		NewFilterProperty(
			"O.owner_user_id = ?",
			WrapRetSlice(func(filter *entity.ServiceFilter) []*int64 { return filter.OwnerId }),
		),
		NewFilterProperty(
			"S.service_ccrn LIKE Concat('%',?,'%')",
			WrapRetSlice(func(filter *entity.ServiceFilter) []*string { return filter.Search }),
		),
		NewStateFilterProperty(
			"S.service",
			WrapRetState(
				func(filter *entity.ServiceFilter) []entity.StateFilterType { return filter.State },
			),
		),
	},
	JoinDefs: []*JoinDef{
		{
			Name:  "O",
			Type:  LeftJoin,
			Table: "Owner O",
			On:    "S.service_id = O.owner_service_id",
			Condition: WrapJoinCondition(func(f *entity.ServiceFilter, _ []entity.Order) bool {
				return len(f.OwnerId) > 0
			}),
		},
		{
			Name:      "U",
			Type:      LeftJoin,
			Table:     "User U",
			On:        "O.owner_user_id = U.user_id",
			DependsOn: []string{"O"},
			Condition: WrapJoinCondition(func(f *entity.ServiceFilter, _ []entity.Order) bool {
				return len(f.OwnerName) > 0
			}),
		},
		{
			Name:  "SGS",
			Type:  LeftJoin,
			Table: "SupportGroupService SGS",
			On:    "S.service_id = SGS.supportgroupservice_service_id",
			Condition: WrapJoinCondition(func(f *entity.ServiceFilter, _ []entity.Order) bool {
				return len(f.SupportGroupId) > 0
			}),
		},
		{
			Name:      "SG",
			Type:      LeftJoin,
			Table:     "SupportGroup SG",
			On:        "SGS.supportgroupservice_support_group_id = SG.supportgroup_id",
			DependsOn: []string{"SGS"},
			Condition: WrapJoinCondition(func(f *entity.ServiceFilter, _ []entity.Order) bool {
				return len(f.SupportGroupCCRN) > 0
			}),
		},
		{
			Name:  "CI",
			Type:  LeftJoin,
			Table: "ComponentInstance CI",
			On:    "S.service_id = CI.componentinstance_service_id",
			Condition: WrapJoinCondition(func(f *entity.ServiceFilter, _ []entity.Order) bool {
				return len(f.ComponentInstanceId) > 0
			}),
		},
		{
			Name:      "IM",
			Type:      LeftJoin,
			Table:     "IssueMatch IM",
			On:        "CI.componentinstance_id = IM.issuematch_component_instance_id",
			DependsOn: []string{"CI"},
			Condition: WrapJoinCondition(func(f *entity.ServiceFilter, _ []entity.Order) bool {
				return len(f.IssueId) > 0
			}),
		},
		{
			Name:  "IRS",
			Type:  LeftJoin,
			Table: "IssueRepositoryService IRS",
			On:    "S.service_id = IRS.issuerepositoryservice_service_id",
			Condition: WrapJoinCondition(func(f *entity.ServiceFilter, _ []entity.Order) bool {
				return len(f.IssueRepositoryId) > 0
			}),
		},
		{
			Name:  "SIC",
			Type:  LeftJoin,
			Table: "mvServiceIssueCounts SIC",
			On:    "S.service_id = SIC.service_id",
			Condition: WrapJoinCondition(func(f *entity.ServiceFilter, order []entity.Order) bool {
				return OrderByCount(order)
			}),
		},
	},
}

func ensureServiceFilter(filter *entity.ServiceFilter) *entity.ServiceFilter {
	if filter == nil {
		filter = &entity.ServiceFilter{}
	}

	return EnsurePagination(filter)
}

func (s *SqlDatabase) getServiceJoins(filter *entity.ServiceFilter, order []entity.Order) string {
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

	if len(filter.ComponentInstanceId) > 0 || len(filter.IssueId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentInstance CI on S.service_id = CI.componentinstance_service_id
		`)

		if len(filter.IssueId) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, `
				LEFT JOIN IssueMatch IM on IM.issuematch_component_instance_id = CI.componentinstance_id
			`)
		}
	}

	if len(filter.IssueRepositoryId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueRepositoryService IRS on IRS.issuerepositoryservice_service_id = S.service_id
		`)
	}
	if OrderByCount(order) {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN mvServiceIssueCounts SIC ON S.service_id = SIC.service_id
		`)
	}

	return joins
}

func (s *SqlDatabase) getServiceColumns(filter *entity.ServiceFilter, order []entity.Order) string {
	columns := "S.*"
	if len(filter.IssueRepositoryId) > 0 {
		columns = fmt.Sprintf("%s, %s", columns, "IRS.*")
	}

	for _, o := range order {
		switch o.By {
		case entity.CriticalCount:
			columns = fmt.Sprintf("%s, SIC.critical_count", columns)
		case entity.HighCount:
			columns = fmt.Sprintf("%s, SIC.high_count", columns)
		case entity.MediumCount:
			columns = fmt.Sprintf("%s, SIC.medium_count", columns)
		case entity.LowCount:
			columns = fmt.Sprintf("%s, SIC.low_count", columns)
		case entity.NoneCount:
			columns = fmt.Sprintf("%s, SIC.none_count", columns)
		}
	}

	return columns
}

func (s *SqlDatabase) buildServiceStatement(
	baseQuery string,
	filter *entity.ServiceFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	var query string

	filter = ensureServiceFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	cursorFields, err := DecodeCursor(filter.After)
	if err != nil {
		return nil, nil, err
	}

	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.ServiceId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)
	joins := s.getServiceJoins(filter, order)

	filterStr := serviceObject.GetFilterQuery(filter)

	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && withCursor && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" HAVING (%s)", cursorQuery)
	}

	// construct final query
	if withCursor {
		query = fmt.Sprintf(baseQuery, joins, whereClause, cursorQuery, orderStr)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause, orderStr)
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
	filterParameters := serviceObject.GetFilterParameters(filter, withCursor, cursorFields)

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

	stmt, filterParameters, err := s.buildServiceStatement(
		baseQuery,
		filter,
		false,
		[]entity.Order{},
		l,
	)
	if err != nil {
		return -1, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetServices(
	filter *entity.ServiceFilter,
	order []entity.Order,
) ([]entity.ServiceResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetServices",
	})

	baseQuery := `
		SELECT %s FROM Service S
		%s
		%s
		GROUP BY S.service_id %s ORDER BY %s LIMIT ?
    `

	filter = ensureServiceFilter(filter)
	columns := s.getServiceColumns(filter, order)
	baseQuery = fmt.Sprintf(baseQuery, columns, "%s", "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildServiceStatement(baseQuery, filter, true, order, l)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ServiceResult, e RowComposite) []entity.ServiceResult {
			s := entity.Service{
				BaseService: e.AsBaseService(),
			}

			var isc entity.IssueSeverityCounts
			if e.RatingCount != nil {
				isc = e.AsIssueSeverityCounts()
			}

			cursor, _ := EncodeCursor(WithService(order, s, isc))

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

func (s *SqlDatabase) GetServicesWithAggregations(
	filter *entity.ServiceFilter,
	order []entity.Order,
) ([]entity.ServiceResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetServicesWithAggregations",
	})

	baseImQuery := `
        SELECT %s, COUNT(IM.issuematch_id) AS service_agg_issue_matches FROM Service S
        %s
        LEFT JOIN IssueMatch IM on CI.componentinstance_id = IM.issuematch_component_instance_id
        %s
        GROUP BY S.service_id %s ORDER BY %s LIMIT ?
    `

	baseCiQuery := `
        SELECT %s, SUM(CI.componentinstance_count) AS service_agg_component_instances FROM Service S
        %s
        %s
        GROUP BY S.service_id %s ORDER BY %s LIMIT ?
    `

	orderBySeverity := lo.ContainsBy(order, func(o entity.Order) bool {
		return o.By == entity.CriticalCount || o.By == entity.HighCount ||
			o.By == entity.MediumCount ||
			o.By == entity.LowCount ||
			o.By == entity.NoneCount
	})

	if !orderBySeverity {
		baseImQuery = fmt.Sprintf(
			baseImQuery,
			"%s",
			"%s LEFT JOIN ComponentInstance CI on S.service_id = CI.componentinstance_service_id",
			"%s",
			"%s",
			"%s",
		)
		baseCiQuery = fmt.Sprintf(
			baseCiQuery,
			"%s",
			"%s LEFT JOIN ComponentInstance CI on S.service_id = CI.componentinstance_service_id",
			"%s",
			"%s",
			"%s",
		)
	}

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
	filter = ensureServiceFilter(filter)
	order = GetDefaultOrder(order, entity.ServiceId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)
	joins := s.getServiceJoins(filter, order)
	columns := s.getServiceColumns(filter, order)

	cursorFields, err := DecodeCursor(filter.After)
	if err != nil {
		return nil, err
	}

	cursorQuery := CreateCursorQuery("", cursorFields)

	filterStr := serviceObject.GetFilterQuery(filter)

	whereClause := ""
	if filterStr != "" || cursorQuery != "" {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" HAVING (%s)", cursorQuery)
	}

	imQuery := fmt.Sprintf(baseImQuery, columns, joins, whereClause, cursorQuery, orderStr)
	ciQuery := fmt.Sprintf(baseCiQuery, columns, joins, whereClause, cursorQuery, orderStr)
	query := fmt.Sprintf(baseQuery, imQuery, ciQuery)

	stmt, err := s.db.Preparex(query)
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
	filterParameters := serviceObject.GetFilterParameters(filter, true, cursorFields)
	// parameters for component instance query
	filterParameters = append(
		filterParameters,
		serviceObject.GetFilterParameters(filter, true, cursorFields)...)

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ServiceResult, e RowComposite) []entity.ServiceResult {
			service := entity.Service{
				BaseService: e.AsBaseService(),
			}
			aggregations := e.AsServiceAggregations()

			var isc entity.IssueSeverityCounts
			if e.RatingCount != nil {
				isc = e.AsIssueSeverityCounts()
			}

			cursor, _ := EncodeCursor(WithService(order, service, isc))

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

func (s *SqlDatabase) GetAllServiceCursors(
	filter *entity.ServiceFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllServiceCursors",
	})

	baseQuery := `
		SELECT %s FROM Service S 
		%s
	    %s GROUP BY S.service_id ORDER BY %s
    `

	filter = ensureServiceFilter(filter)
	columns := s.getServiceColumns(filter, order)
	baseQuery = fmt.Sprintf(baseQuery, columns, "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildServiceStatement(baseQuery, filter, false, order, l)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

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

		var isc entity.IssueSeverityCounts
		if row.RatingCount != nil {
			isc = row.AsIssueSeverityCounts()
		}

		cursor, _ := EncodeCursor(WithService(order, s, isc))

		return cursor
	}), nil
}

func (s *SqlDatabase) CreateService(service *entity.Service) (*entity.Service, error) {
	return serviceObject.Create(s.db, service)
}

func (s *SqlDatabase) UpdateService(service *entity.Service) error {
	return serviceObject.Update(s.db, service)
}

func (s *SqlDatabase) DeleteService(id int64, userId int64) error {
	return serviceObject.Delete(s.db, id, userId)
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

	args := map[string]any{
		"service_id": serviceId,
		"user_id":    userId,
	}

	var mysqlErr *mysql.MySQLError

	_, err := performExec(s, query, args, l)
	if err != nil {
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == database.ErrCodeDuplicateEntry {
				return nil
			}
		}

		return err
	}

	return nil
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

	args := map[string]any{
		"service_id": serviceId,
		"user_id":    userId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) AddIssueRepositoryToService(
	serviceId int64,
	issueRepositoryId int64,
	priority int64,
) error {
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

	args := map[string]any{
		"service_id":          serviceId,
		"issue_repository_id": issueRepositoryId,
		"priority":            priority,
	}

	var mysqlErr *mysql.MySQLError

	_, err := performExec(s, query, args, l)
	if err != nil {
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == database.ErrCodeDuplicateEntry {
				return nil
			}
		}

		return err
	}

	return nil
}

func (s *SqlDatabase) RemoveIssueRepositoryFromService(
	serviceId int64,
	issueRepositoryId int64,
) error {
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

	args := map[string]any{
		"service_id":          serviceId,
		"issue_repository_id": issueRepositoryId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) getServiceAttr(
	attrName string,
	filter *entity.ServiceFilter,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.getServiceAttr",
	})

	baseQuery := `
    SELECT service_%s FROM Service S
    %s
    %s
    ORDER BY %s
    `

	baseQuery = fmt.Sprintf(baseQuery, attrName, "%s", "%s", "%s")

	// Ensure the filter is initialized
	filter = ensureServiceFilter(filter)
	order := []entity.Order{
		{By: entity.ServiceCcrn, Direction: entity.OrderDirectionAsc},
	}

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildServiceStatement(baseQuery, filter, false, order, l)
	if err != nil {
		l.Error("Error preparing statement: ", err)
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	// Execute the query
	rows, err := stmt.Queryx(filterParameters...)
	if err != nil {
		l.Error("Error executing query: ", err)
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Warnf("error during close rows: %s", err)
		}
	}()

	// Collect the results
	serviceAttrs := []string{}

	var attrVal string
	for rows.Next() {
		if err := rows.Scan(&attrVal); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}

		serviceAttrs = append(serviceAttrs, attrVal)
	}

	if err = rows.Err(); err != nil {
		l.Error("Row iteration error: ", err)
		return nil, err
	}

	return serviceAttrs, nil
}

func (s *SqlDatabase) GetServiceCcrns(filter *entity.ServiceFilter) ([]string, error) {
	ccrns, err := s.getServiceAttr("ccrn", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get Service ccrns: %w", err)
	}

	return ccrns, nil
}

func (s *SqlDatabase) GetServiceDomains(filter *entity.ServiceFilter) ([]string, error) {
	domains, err := s.getServiceAttr("domain", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get Service domains: %w", err)
	}

	return domains, nil
}

func (s *SqlDatabase) GetServiceRegions(filter *entity.ServiceFilter) ([]string, error) {
	regions, err := s.getServiceAttr("region", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get Service regions: %w", err)
	}

	return regions, nil
}
