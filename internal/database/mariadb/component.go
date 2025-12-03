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

func (s *SqlDatabase) getComponentFilterString(filter *entity.ComponentFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.CCRN, "C.component_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Id, "C.component_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentVersionId, "CV.componentversion_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceCCRN, "S.service_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentVersionRepository, "CV.componentversion_repository = ?", OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "C.component"))

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) ensureComponentFilter(f *entity.ComponentFilter) *entity.ComponentFilter {
	var first = 1000
	after := ""
	if f == nil {
		return &entity.ComponentFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			CCRN:               nil,
			Id:                 nil,
			ComponentVersionId: nil,
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

func (s *SqlDatabase) getComponentJoins(filter *entity.ComponentFilter, order []entity.Order) string {
	joins := ""
	if s.needComponentVersion(filter, order) {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentVersion CV on C.component_id = CV.componentversion_component_id
		`)
	}
	if s.needComponentInstance(filter) {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentInstance CI on CV.componentversion_id = CI.componentinstance_component_version_id
		`)
	}
	if s.needService(filter) {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN Service S on S.service_id = CI.componentinstance_service_id
		`)
	}
	if s.needSingleComponentByServiceVulnerabilityCounts(filter, order) {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN mvSingleComponentByServiceVulnerabilityCounts CVR on C.component_id = CVR.component_id AND CVR.service_id = S.service_id
		`)
	}
	if s.needAllComponentByServiceVulnerabilityCounts(filter, order) {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN mvAllComponentsByServiceVulnerabilityCounts CVR on CVR.service_id = S.service_id
		`)
	}
	return joins
}

func (s *SqlDatabase) needComponentVersion(filter *entity.ComponentFilter, order []entity.Order) bool {
	orderByRepository := lo.ContainsBy(order, func(o entity.Order) bool {
		return o.By == entity.ComponentVersionRepository
	})
	return len(filter.ComponentVersionId) > 0 ||
		len(filter.ServiceCCRN) > 0 ||
		len(filter.ComponentVersionRepository) > 0 ||
		orderByRepository
}

func (s *SqlDatabase) needComponentInstance(filter *entity.ComponentFilter) bool {
	return len(filter.ServiceCCRN) > 0 || len(filter.ComponentVersionRepository) > 0
}

func (s *SqlDatabase) needService(filter *entity.ComponentFilter) bool {
	return len(filter.ServiceCCRN) > 0 || len(filter.ComponentVersionRepository) > 0
}

func (s *SqlDatabase) needSingleComponentByServiceVulnerabilityCounts(filter *entity.ComponentFilter, order []entity.Order) bool {
	orderByCount := lo.ContainsBy(order, func(o entity.Order) bool {
		return o.By == entity.CriticalCount || o.By == entity.HighCount || o.By == entity.MediumCount || o.By == entity.LowCount || o.By == entity.NoneCount
	})
	return orderByCount && (len(filter.Id) > 0 && (len(filter.ServiceCCRN) > 0 || len(filter.ComponentVersionRepository) > 0))
}

func (s *SqlDatabase) needAllComponentByServiceVulnerabilityCounts(filter *entity.ComponentFilter, order []entity.Order) bool {
	orderByCount := lo.ContainsBy(order, func(o entity.Order) bool {
		return o.By == entity.CriticalCount || o.By == entity.HighCount || o.By == entity.MediumCount || o.By == entity.LowCount || o.By == entity.NoneCount
	})
	return orderByCount && (len(filter.Id) == 0 && (len(filter.ServiceCCRN) > 0 || len(filter.ComponentVersionRepository) > 0))
}

func (s *SqlDatabase) getComponentColumns(order []entity.Order) string {
	columns := "C.*"
	for _, o := range order {
		switch o.By {
		case entity.CriticalCount:
			columns = fmt.Sprintf("%s, CVR.critical_count", columns)
		case entity.HighCount:
			columns = fmt.Sprintf("%s, CVR.high_count", columns)
		case entity.MediumCount:
			columns = fmt.Sprintf("%s, CVR.medium_count", columns)
		case entity.LowCount:
			columns = fmt.Sprintf("%s, CVR.low_count", columns)
		case entity.NoneCount:
			columns = fmt.Sprintf("%s, CVR.none_count", columns)
		case entity.ComponentVersionRepository:
			columns = fmt.Sprintf("%s, CV.componentversion_repository", columns)
		}
	}
	return columns
}

func (s *SqlDatabase) getComponentUpdateFields(component *entity.Component) string {
	fl := []string{}
	if component.CCRN != "" {
		fl = append(fl, "component_ccrn = :component_ccrn")
	}
	if component.Type != "" {
		fl = append(fl, "component_type = :component_type")
	}
	if component.UpdatedBy != 0 {
		fl = append(fl, "component_updated_by = :component_updated_by")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildComponentStatement(baseQuery string, filter *entity.ComponentFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	var query string
	filter = s.ensureComponentFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getComponentFilterString(filter)
	joins := s.getComponentJoins(filter, order)
	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, nil, err
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.ComponentId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

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

	//adding parameters
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.CCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionId)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceCCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionRepository)
	if withCursor {
		p := CreateCursorParameters([]any{}, cursorFields)
		filterParameters = append(filterParameters, p...)
		if filter.PaginatedX.First == nil {
			filterParameters = append(filterParameters, 1000)
		} else {
			filterParameters = append(filterParameters, filter.PaginatedX.First)
		}
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllComponentIds(filter *entity.ComponentFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponentIds",
	})

	baseQuery := `
		SELECT C.component_id FROM Component C 
		%s
	 	%s GROUP BY C.component_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildComponentStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetAllComponentCursors(filter *entity.ComponentFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllComponentCursors",
	})

	baseQuery := `
		SELECT %s FROM Component C 
		%s
	    %s GROUP BY C.component_id ORDER BY %s
    `

	filter = s.ensureComponentFilter(filter)
	columns := s.getComponentColumns(order)
	baseQuery = fmt.Sprintf(baseQuery, columns, "%s", "%s", "%s")
	stmt, filterParameters, err := s.buildComponentStatement(baseQuery, filter, false, order, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

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
		c := row.AsComponent()

		var isc entity.IssueSeverityCounts
		if row.RatingCount != nil {
			isc = row.RatingCount.AsIssueSeverityCounts()
		}

		var cv entity.ComponentVersion
		if row.ComponentVersionRow != nil {
			cv = row.ComponentVersionRow.AsComponentVersion()
		}

		cursor, _ := EncodeCursor(WithComponent(order, c, cv, isc))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetComponents(filter *entity.ComponentFilter, order []entity.Order) ([]entity.ComponentResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponents",
	})

	baseQuery := `
		SELECT %s FROM Component C
		%s
		%s
		%s GROUP BY C.component_id ORDER BY %s LIMIT ?
    `

	filter = s.ensureComponentFilter(filter)
	columns := s.getComponentColumns(order)
	baseQuery = fmt.Sprintf(baseQuery, columns, "%s", "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildComponentStatement(baseQuery, filter, true, order, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ComponentResult, e RowComposite) []entity.ComponentResult {
			c := e.AsComponent()

			var isc entity.IssueSeverityCounts
			if e.RatingCount != nil {
				isc = e.RatingCount.AsIssueSeverityCounts()
			}

			var cv entity.ComponentVersion
			if e.ComponentVersionRow != nil {
				cv = e.ComponentVersionRow.AsComponentVersion()
			}

			cursor, _ := EncodeCursor(WithComponent(order, c, cv, isc))

			cr := entity.ComponentResult{
				WithCursor: entity.WithCursor{Value: cursor},
				Component:  &c,
			}

			return append(l, cr)
		},
	)
}

func (s *SqlDatabase) CountComponents(filter *entity.ComponentFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountComponents",
	})

	baseQuery := `
		SELECT count(distinct C.component_id) FROM Component C
		%s
		%s
		ORDER BY %s
	`
	stmt, filterParameters, err := s.buildComponentStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CountComponentVulnerabilities(filter *entity.ComponentFilter) ([]entity.IssueSeverityCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountComponentVulnerabilities",
	})
	var fl []string
	var filterParameters []interface{}

	filter = s.ensureComponentFilter(filter)

	query := `
		SELECT CVR.critical_count, CVR.high_count, CVR.medium_count, CVR.low_count, CVR.none_count FROM %s AS CVR
	`

	joins := ""
	groupBy := ""

	if len(filter.Id) == 0 && len(filter.ComponentVersionRepository) == 0 {
		query = fmt.Sprintf(query, "mvAllComponentsByServiceVulnerabilityCounts")
	} else {
		query = fmt.Sprintf(query, "mvSingleComponentByServiceVulnerabilityCounts")
		groupBy = "GROUP BY CVR.component_id"
	}

	if len(filter.ServiceCCRN) > 0 {
		joins = fmt.Sprintf("%s INNER JOIN Service S ON S.service_id = CVR.service_id", joins)
		fl = append(fl, buildFilterQuery(filter.ServiceCCRN, "S.service_ccrn = ?", OP_OR))
		filterParameters = buildQueryParameters(filterParameters, filter.ServiceCCRN)
	}

	if len(filter.ComponentVersionRepository) > 0 {
		joins = fmt.Sprintf("%s INNER JOIN ComponentVersion CV ON CV.componentversion_component_id = CVR.component_id", joins)
		fl = append(fl, buildFilterQuery(filter.ComponentVersionRepository, "CV.componentversion_repository = ?", OP_OR))
		filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionRepository)
	}

	if len(filter.Id) > 0 {
		filterParameters = buildQueryParameters(filterParameters, filter.Id)
		fl = append(fl, buildFilterQuery(filter.Id, "CVR.component_id = ?", OP_OR))
	}

	filterStr := combineFilterQueries(fl, OP_AND)
	query = fmt.Sprintf("%s %s", query, joins)
	if filterStr != "" {
		query = fmt.Sprintf("%s WHERE %s", query, filterStr)
	}

	query = fmt.Sprintf("%s %s", query, groupBy)

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

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueSeverityCounts, e RatingCount) []entity.IssueSeverityCounts {
			return append(l, e.AsIssueSeverityCounts())
		},
	)
}

func (s *SqlDatabase) CreateComponent(component *entity.Component) (*entity.Component, error) {
	l := logrus.WithFields(logrus.Fields{
		"component": component,
		"event":     "database.CreateComponent",
	})

	query := `
		INSERT INTO Component (
			component_ccrn,
			component_type,
			component_created_by,
			component_updated_by
		) VALUES (
			:component_ccrn,
			:component_type,
			:component_created_by,
			:component_updated_by
		)
	`

	componentRow := ComponentRow{}
	componentRow.FromComponent(component)

	id, err := performInsert(s, query, componentRow, l)

	if err != nil {
		return nil, err
	}

	component.Id = id

	return component, nil
}

func (s *SqlDatabase) UpdateComponent(component *entity.Component) error {
	l := logrus.WithFields(logrus.Fields{
		"component": component,
		"event":     "database.UpdateComponent",
	})

	baseQuery := `
		UPDATE Component SET
		%s
		WHERE component_id = :component_id
	`

	updateFields := s.getComponentUpdateFields(component)

	query := fmt.Sprintf(baseQuery, updateFields)

	componentRow := ComponentRow{}
	componentRow.FromComponent(component)

	_, err := performExec(s, query, componentRow, l)

	return err
}

func (s *SqlDatabase) DeleteComponent(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteComponent",
	})

	query := `
		UPDATE Component SET
		component_deleted_at = NOW(),
		component_updated_by = :userId
		WHERE component_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) GetComponentCcrns(filter *entity.ComponentFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetComponentCcrns",
	})

	baseQuery := `
    SELECT C.component_ccrn FROM Component C
    %s
    %s
    ORDER BY %s
    `

	// Ensure the filter is initialized
	filter = s.ensureComponentFilter(filter)
	order := []entity.Order{
		{
			By:        entity.ComponentCcrn,
			Direction: entity.OrderDirectionAsc,
		},
	}

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildComponentStatement(baseQuery, filter, false, order, l)
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
	componentCcrns := []string{}
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}
		componentCcrns = append(componentCcrns, name)
	}
	if err = rows.Err(); err != nil {
		l.Error("Row iteration error: ", err)
		return nil, err
	}

	return componentCcrns, nil
}
