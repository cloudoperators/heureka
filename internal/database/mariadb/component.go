// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var componentObject = DbObject[*entity.Component]{
	Prefix:    "component",
	TableName: "Component",
	Properties: []*Property{
		NewProperty(
			"component_ccrn",
			WrapAccess(func(c *entity.Component) (string, bool) { return c.CCRN, c.CCRN != "" }),
		),
		NewProperty(
			"component_repository",
			WrapAccess(
				func(c *entity.Component) (string, bool) { return c.Repository, c.Repository != "" },
			),
		),
		NewProperty(
			"component_organization",
			WrapAccess(
				func(c *entity.Component) (string, bool) { return c.Organization, c.Organization != "" },
			),
		),
		NewProperty(
			"component_url",
			WrapAccess(func(c *entity.Component) (string, bool) { return c.Url, c.Url != "" }),
		),
		NewProperty(
			"component_type",
			WrapAccess(func(c *entity.Component) (string, bool) { return c.Type, c.Type != "" }),
		),
		NewProperty(
			"component_created_by",
			WrapAccess(func(c *entity.Component) (int64, bool) { return c.CreatedBy, NoUpdate }),
		),
		NewProperty(
			"component_updated_by",
			WrapAccess(
				func(c *entity.Component) (int64, bool) { return c.UpdatedBy, c.UpdatedBy != 0 },
			),
		),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty(
			"C.component_ccrn = ?",
			WrapRetSlice(func(filter *entity.ComponentFilter) []*string { return filter.CCRN }),
		),
		NewFilterProperty(
			"C.component_repository = ?",
			WrapRetSlice(
				func(filter *entity.ComponentFilter) []*string { return filter.Repository },
			),
		),
		NewFilterProperty(
			"C.component_organization = ?",
			WrapRetSlice(
				func(filter *entity.ComponentFilter) []*string { return filter.Organization },
			),
		),
		NewFilterProperty(
			"C.component_id = ?",
			WrapRetSlice(func(filter *entity.ComponentFilter) []*int64 { return filter.Id }),
		),
		NewFilterProperty(
			"CV.componentversion_id = ?",
			WrapRetSlice(
				func(filter *entity.ComponentFilter) []*int64 { return filter.ComponentVersionId },
			),
		),
		NewFilterProperty(
			"S.service_ccrn = ?",
			WrapRetSlice(
				func(filter *entity.ComponentFilter) []*string { return filter.ServiceCCRN },
			),
		),
		NewStateFilterProperty(
			"C.component",
			WrapRetState(
				func(filter *entity.ComponentFilter) []entity.StateFilterType { return filter.State },
			),
		),
	},
	JoinDefs: []*JoinDef{
		{
			Name:  "CV",
			Type:  LeftJoin,
			Table: "ComponentVersion CV",
			On:    "C.component_id = CV.componentversion_component_id",
			Condition: WrapJoinCondition(func(f *entity.ComponentFilter, _ *Order) bool {
				return len(f.ComponentVersionId) > 0
			}),
		},
		{
			Name:      "CI",
			Type:      LeftJoin,
			Table:     "ComponentInstance CI",
			On:        "CV.componentversion_id = CI.componentinstance_component_version_id",
			DependsOn: []string{"CV"},
			Condition: DependentJoin,
		},
		{
			Name:      "S",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			DependsOn: []string{"CI"},
			Condition: WrapJoinCondition(func(f *entity.ComponentFilter, _ *Order) bool {
				return len(f.ServiceCCRN) > 0
			}),
		},
		{
			Name:      "mvSCBSVC",
			Type:      LeftJoin,
			Table:     "mvSingleComponentByServiceVulnerabilityCounts CVR",
			On:        "C.component_id = CVR.component_id AND CVR.service_id = S.service_id",
			DependsOn: []string{"S"},
			Condition: WrapJoinCondition(func(f *entity.ComponentFilter, order *Order) bool {
				return needSingleComponentByServiceVulnerabilityCounts(f, order)
			}),
		},
		{
			Name:      "mvACBSVC",
			Type:      LeftJoin,
			Table:     "mvAllComponentsByServiceVulnerabilityCounts CVR",
			On:        "S.service_id = CVR.service_id",
			DependsOn: []string{"S"},
			Condition: WrapJoinCondition(func(f *entity.ComponentFilter, order *Order) bool {
				return needAllComponentByServiceVulnerabilityCounts(f, order)
			}),
		},
	},
}

func ensureComponentFilter(filter *entity.ComponentFilter) *entity.ComponentFilter {
	if filter == nil {
		filter = &entity.ComponentFilter{}
	}

	return EnsurePagination(filter)
}

func needSingleComponentByServiceVulnerabilityCounts(filter *entity.ComponentFilter, order *Order) bool {
	return order.ByCount() && (len(filter.Id) > 0 && (len(filter.ServiceCCRN) > 0))
}

func needAllComponentByServiceVulnerabilityCounts(filter *entity.ComponentFilter, order *Order) bool {
	return order.ByCount() && (len(filter.Id) == 0 && (len(filter.ServiceCCRN) > 0))
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
		}
	}

	return columns
}

func (s *SqlDatabase) buildComponentStatement(
	baseQuery string,
	filter *entity.ComponentFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	filter = ensureComponentFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	cursorFields, err := DecodeCursor(filter.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode Remediation cursor: %w", err)
	}

	cursorQuery := CreateCursorQuery("", cursorFields)

	ord := NewOrder(order, entity.Order{entity.ComponentId, entity.OrderDirectionAsc})
	filterStr := componentObject.GetFilterQuery(filter)
	joins := componentObject.GetJoins(filter, ord)

	whereClause := ""

	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && withCursor && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	var query string
	if withCursor {
		query = fmt.Sprintf(baseQuery, joins, whereClause, cursorQuery, ord)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause, ord)
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

	filterParameters := componentObject.GetFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllComponentCursors(
	filter *entity.ComponentFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllComponentCursors",
	})

	baseQuery := `
		SELECT %s FROM Component C 
		%s
	    %s GROUP BY C.component_id ORDER BY %s
    `

	filter = ensureComponentFilter(filter)
	columns := s.getComponentColumns(order)
	baseQuery = fmt.Sprintf(baseQuery, columns, "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildComponentStatement(baseQuery, filter, false, order, l)
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
		c := row.AsComponent()

		var isc entity.IssueSeverityCounts
		if row.RatingCount != nil {
			isc = row.AsIssueSeverityCounts()
		}

		cursor, _ := EncodeCursor(WithComponent(order, c, isc))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetComponents(
	filter *entity.ComponentFilter,
	order []entity.Order,
) ([]entity.ComponentResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponents",
	})

	baseQuery := `
		SELECT %s FROM Component C
		%s
		%s
		%s GROUP BY C.component_id ORDER BY %s LIMIT ?
    `

	filter = ensureComponentFilter(filter)
	columns := s.getComponentColumns(order)
	baseQuery = fmt.Sprintf(baseQuery, columns, "%s", "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildComponentStatement(baseQuery, filter, true, order, l)
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
		func(l []entity.ComponentResult, e RowComposite) []entity.ComponentResult {
			c := e.AsComponent()

			var isc entity.IssueSeverityCounts
			if e.RatingCount != nil {
				isc = e.AsIssueSeverityCounts()
			}

			cursor, _ := EncodeCursor(WithComponent(order, c, isc))

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

	stmt, filterParameters, err := s.buildComponentStatement(
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

func (s *SqlDatabase) CountComponentVulnerabilities(
	filter *entity.ComponentFilter,
) ([]entity.IssueSeverityCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountComponentVulnerabilities",
	})

	var (
		fl               []string
		filterParameters []any
	)

	filter = ensureComponentFilter(filter)

	query := `
		SELECT CVR.critical_count, CVR.high_count, CVR.medium_count, CVR.low_count, CVR.none_count FROM %s AS CVR
	`

	joins := ""
	groupBy := ""

	if len(filter.Id) == 0 && len(filter.Repository) == 0 {
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

	if len(filter.Repository) > 0 {
		joins = fmt.Sprintf("%s INNER JOIN Component C ON C.component_id = CVR.component_id", joins)

		fl = append(fl, buildFilterQuery(filter.Repository, "C.component_repository = ?", OP_OR))

		filterParameters = buildQueryParameters(filterParameters, filter.Repository)
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

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

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
	return componentObject.Create(s.db, component)
}

func (s *SqlDatabase) UpdateComponent(component *entity.Component) error {
	return componentObject.Update(s.db, component)
}

func (s *SqlDatabase) DeleteComponent(id int64, userId int64) error {
	return componentObject.Delete(s.db, id, userId)
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
	filter = ensureComponentFilter(filter)
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
