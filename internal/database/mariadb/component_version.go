// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/database"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) ensureComponentVersionFilter(f *entity.ComponentVersionFilter) *entity.ComponentVersionFilter {
	var first int = 1000
	var after int64 = 0
	if f == nil {
		return &entity.ComponentVersionFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id:            nil,
			CCRN:          nil,
			IssueId:       nil,
			ComponentCCRN: nil,
			ComponentId:   nil,
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

func (s *SqlDatabase) getComponentVersionJoins(filter *entity.ComponentVersionFilter) string {
	joins := ""
	if len(filter.IssueId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN ComponentVersionIssue CVI on CV.componentversion_id = CVI.componentversionissue_component_version_id")
	}
	if len(filter.ComponentCCRN) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN Component C on CV.componentversion_component_id = C.component_id")
	}
	return joins
}

func (s *SqlDatabase) getComponentVersionUpdateFields(componentVersion *entity.ComponentVersion) string {
	fl := []string{}
	if componentVersion.CCRN != "" {
		fl = append(fl, "componentversion_ccrn = :componentversion_ccrn")
	}
	if componentVersion.Version != "" {
		fl = append(fl, "componentversion_version = :componentversion_version")
	}
	if componentVersion.ComponentId != 0 {
		fl = append(fl, "componentversion_component_id = :componentversion_component_id")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) getComponentVersionFilterString(filter *entity.ComponentVersionFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "CV.componentversion_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.CCRN, "CV.componentversion_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueId, "CVI.componentversionissue_issue_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentId, "CV.componentversion_component_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Version, "CV.componentversion_version = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentCCRN, "C.component_ccrn = ?", OP_OR))
	fl = append(fl, "CV.componentversion_deleted_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) buildComponentVersionStatement(baseQuery string, filter *entity.ComponentVersionFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureComponentVersionFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getComponentVersionFilterString(filter)
	joins := s.getComponentVersionJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "CV.componentversion_id > ?")

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
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.CCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentId)
	filterParameters = buildQueryParameters(filterParameters, filter.Version)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentCCRN)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllComponentVersionIds(filter *entity.ComponentVersionFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponentVersionIds",
	})

	baseQuery := `
		SELECT CV.componentversion_id FROM ComponentVersion CV 
		%s
	 	%s GROUP BY CV.componentversion_id ORDER BY CV.componentversion_id
    `

	stmt, filterParameters, err := s.buildComponentVersionStatement(baseQuery, filter, false, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetComponentVersions(filter *entity.ComponentVersionFilter) ([]entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponentVersions",
	})

	baseQuery := `
		SELECT CV.* FROM ComponentVersion CV 
		%s
		%s
		%s GROUP BY CV.componentversion_id ORDER BY CV.componentversion_id LIMIT ?
    `

	filter = s.ensureComponentVersionFilter(filter)
	baseQuery = fmt.Sprintf(baseQuery, "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildComponentVersionStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ComponentVersion, e ComponentVersionRow) []entity.ComponentVersion {
			return append(l, e.AsComponentVersion())
		},
	)
}

func (s *SqlDatabase) CountComponentVersions(filter *entity.ComponentVersionFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountComponentVersions",
	})

	baseQuery := `
		SELECT count(distinct CV.componentversion_id) FROM ComponentVersion CV 
		%s
		%s
	`
	stmt, filterParameters, err := s.buildComponentVersionStatement(baseQuery, filter, false, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"componentVersion": componentVersion,
		"event":            "database.CreateComponentVersion",
	})

	query := `
		INSERT INTO ComponentVersion (
			componentversion_component_id,
			componentversion_ccrn,
			componentversion_version
		) VALUES (
			:componentversion_component_id,
			:componentversion_ccrn,
			:componentversion_version
		)
	`

	componentVersionRow := ComponentVersionRow{}
	componentVersionRow.FromComponentVersion(componentVersion)

	id, err := performInsert(s, query, componentVersionRow, l)

	if err != nil {
		if strings.HasPrefix(err.Error(), "Error 1062") {
			return nil, database.NewDuplicateEntryDatabaseError(fmt.Sprintf("for ComponentVersion: %s ", componentVersion.Version))
		}
		return nil, err
	}

	componentVersion.Id = id

	return componentVersion, nil
}

func (s *SqlDatabase) UpdateComponentVersion(componentVersion *entity.ComponentVersion) error {
	l := logrus.WithFields(logrus.Fields{
		"componentVersion": componentVersion,
		"event":            "database.UpdateComponentVersion",
	})

	baseQuery := `
		UPDATE ComponentVersion SET
		%s
		WHERE componentversion_id = :componentversion_id
	`

	updateFields := s.getComponentVersionUpdateFields(componentVersion)

	query := fmt.Sprintf(baseQuery, updateFields)

	componentVersionRow := ComponentVersionRow{}
	componentVersionRow.FromComponentVersion(componentVersion)

	_, err := performExec(s, query, componentVersionRow, l)

	return err
}

func (s *SqlDatabase) DeleteComponentVersion(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteComponentVersion",
	})

	query := `
		UPDATE ComponentVersion SET
		componentversion_deleted_at = NOW()
		WHERE componentversion_id = :id
	`

	args := map[string]interface{}{
		"id": id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
