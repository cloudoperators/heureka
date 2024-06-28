// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (s *SqlDatabase) getComponentFilterString(filter *entity.ComponentFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Name, "C.component_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Id, "C.component_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentVersionId, "CV.componentversion_id = ?", OP_OR))
	fl = append(fl, "C.component_deleted_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) ensureComponentFilter(f *entity.ComponentFilter) *entity.ComponentFilter {
	var first = 1000
	var after int64 = 0
	if f == nil {
		return &entity.ComponentFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Name:               nil,
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

func (s *SqlDatabase) getComponentJoins(filter *entity.ComponentFilter) string {
	joins := ""
	if len(filter.ComponentVersionId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentVersion CV on C.component_id = CV.componentversion_component_id
		`)
	}
	return joins
}

func (s *SqlDatabase) getComponentUpdateFields(component *entity.Component) string {
	fl := []string{}
	if component.Name != "" {
		fl = append(fl, "component_name = :component_name")
	}
	if component.Type != "" {
		fl = append(fl, "component_type = :component_type")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildComponentStatement(baseQuery string, filter *entity.ComponentFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureComponentFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getComponentFilterString(filter)
	joins := s.getComponentJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "C.component_id > ?")

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
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionId)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
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
	 	%s GROUP BY C.component_id ORDER BY C.component_id
    `

	stmt, filterParameters, err := s.buildComponentStatement(baseQuery, filter, false, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetComponents(filter *entity.ComponentFilter) ([]entity.Component, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponents",
	})

	baseQuery := `
		SELECT C.* FROM Component C
		%s
		%s
		%s GROUP BY C.component_id ORDER BY C.component_id LIMIT ?
    `

	filter = s.ensureComponentFilter(filter)
	baseQuery = fmt.Sprintf(baseQuery, "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildComponentStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.Component, e ComponentRow) []entity.Component {
			return append(l, e.AsComponent())
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
	`
	stmt, filterParameters, err := s.buildComponentStatement(baseQuery, filter, false, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateComponent(component *entity.Component) (*entity.Component, error) {
	l := logrus.WithFields(logrus.Fields{
		"component": component,
		"event":     "database.CreateComponent",
	})

	query := `
		INSERT INTO Component (
			component_name,
			component_type
		) VALUES (
			:component_name,
			:component_type
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

func (s *SqlDatabase) DeleteComponent(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteComponent",
	})

	query := `
		UPDATE Component SET
		component_deleted_at = NOW()
		WHERE component_id = :id
	`

	args := map[string]interface{}{
		"id": id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
