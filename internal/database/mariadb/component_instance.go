// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func (s *SqlDatabase) ensureComponentInstanceFilter(f *entity.ComponentInstanceFilter) *entity.ComponentInstanceFilter {
	var first int = 1000
	var after int64 = 0
	if f == nil {
		return &entity.ComponentInstanceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			IssueMatchId: nil,
			Id:           nil,
			ServiceId:    nil,
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

const (
	componentInstanceWildCardFilterQuery = "CI.componentinstance_ccrn LIKE Concat('%',?,'%')"
)

func (s *SqlDatabase) getComponentInstanceFilterString(filter *entity.ComponentInstanceFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "CI.componentinstance_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.CCRN, "CI.componentinstance_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Region, "CI.componentinstance_region = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Cluster, "CI.componentinstance_cluster = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Namespace, "CI.componentinstance_namespace = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Domain, "CI.componentinstance_domain = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Project, "CI.componentinstance_project = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueMatchId, "IM.issuematch_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "CI.componentinstance_service_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceCcrn, "S.service_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentVersionId, "CI.componentinstance_component_version_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Search, componentInstanceWildCardFilterQuery, OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "CI.componentinstance"))

	filterStr := combineFilterQueries(fl, OP_AND)
	return filterStr
}

func (s *SqlDatabase) getComponentInstanceJoins(filter *entity.ComponentInstanceFilter) string {
	joins := ""
	if len(filter.IssueMatchId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "INNER JOIN IssueMatch IM on CI.componentinstance_id = IM.issuematch_component_instance_id")
	}
	if len(filter.ServiceCcrn) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "INNER JOIN Service S on CI.componentinstance_service_id = S.service_id")
	}
	return joins
}

func (s *SqlDatabase) getComponentInstanceUpdateFields(componentInstance *entity.ComponentInstance) string {
	fl := []string{}
	if componentInstance.CCRN != "" {
		fl = append(fl, "componentinstance_ccrn = :componentinstance_ccrn")
	}
	if componentInstance.Region != "" {
		fl = append(fl, "componentinstance_region = :componentinstance_region")
	}
	if componentInstance.Cluster != "" {
		fl = append(fl, "componentinstance_cluster = :componentinstance_cluster")
	}
	if componentInstance.Namespace != "" {
		fl = append(fl, "componentinstance_namespace = :componentinstance_namespace")
	}
	if componentInstance.Domain != "" {
		fl = append(fl, "componentinstance_domain = :componentinstance_domain")
	}
	if componentInstance.Project != "" {
		fl = append(fl, "componentinstance_project = :componentinstance_project")
	}
	if componentInstance.Count != 0 {
		fl = append(fl, "componentinstance_count = :componentinstance_count")
	}
	if componentInstance.ComponentVersionId != 0 {
		fl = append(fl, "componentinstance_component_version_id = :componentinstance_component_version_id")
	}
	if componentInstance.ServiceId != 0 {
		fl = append(fl, "componentinstance_service_id = :componentinstance_service_id")
	}
	if componentInstance.UpdatedBy != 0 {
		fl = append(fl, "componentinstance_updated_by = :componentinstance_updated_by")
	}

	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildComponentInstanceStatement(baseQuery string, filter *entity.ComponentInstanceFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureComponentInstanceFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getComponentInstanceFilterString(filter)
	joins := s.getComponentInstanceJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "CI.componentinstance_id > ?")

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
	filterParameters = buildQueryParameters(filterParameters, filter.Region)
	filterParameters = buildQueryParameters(filterParameters, filter.Cluster)
	filterParameters = buildQueryParameters(filterParameters, filter.Namespace)
	filterParameters = buildQueryParameters(filterParameters, filter.Domain)
	filterParameters = buildQueryParameters(filterParameters, filter.Project)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchId)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceCcrn)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionId)
	filterParameters = buildQueryParameters(filterParameters, filter.Search)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllComponentInstanceIds(filter *entity.ComponentInstanceFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponentInstanceIds",
	})

	baseQuery := `
		SELECT CI.componentinstance_id FROM ComponentInstance CI 
		%s
	 	%s GROUP BY CI.componentinstance_id ORDER BY CI.componentinstance_id
    `

	stmt, filterParameters, err := s.buildComponentInstanceStatement(baseQuery, filter, false, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetComponentInstances(filter *entity.ComponentInstanceFilter) ([]entity.ComponentInstance, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponentInstances",
	})
	baseQuery := `
			SELECT CI.* FROM ComponentInstance CI
			%s
			%s
			%s GROUP BY CI.componentinstance_id ORDER BY CI.componentinstance_id LIMIT ? 
		`

	stmt, filterParameters, err := s.buildComponentInstanceStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ComponentInstance, e ComponentInstanceRow) []entity.ComponentInstance {
			return append(l, e.AsComponentInstance())
		},
	)
}

func (s *SqlDatabase) CountComponentInstances(filter *entity.ComponentInstanceFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountComponentInstances",
	})

	// Building the Base Query
	baseQuery := `
		SELECT count(distinct CI.componentinstance_id) FROM ComponentInstance CI
		%s
		%s 
	`

	stmt, filterParameters, err := s.buildComponentInstanceStatement(baseQuery, filter, false, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateComponentInstance(componentInstance *entity.ComponentInstance) (*entity.ComponentInstance, error) {
	l := logrus.WithFields(logrus.Fields{
		"componentInstance": componentInstance,
		"event":             "database.CreateComponentInstance",
	})

	query := `
		INSERT INTO ComponentInstance (
			componentinstance_ccrn,
			componentinstance_region,
			componentinstance_cluster,
			componentinstance_namespace,
			componentinstance_domain,
			componentinstance_project,
			componentinstance_count,
			componentinstance_component_version_id,
			componentinstance_service_id,
			componentinstance_created_by,
			componentinstance_updated_by
		) VALUES (
			:componentinstance_ccrn,
			:componentinstance_region,
			:componentinstance_cluster,
			:componentinstance_namespace,
			:componentinstance_domain,
			:componentinstance_project,
			:componentinstance_count,
			:componentinstance_component_version_id,
			:componentinstance_service_id,
			:componentinstance_created_by,
			:componentinstance_updated_by
		)
	`

	componentInstanceRow := ComponentInstanceRow{}
	componentInstanceRow.FromComponentInstance(componentInstance)

	id, err := performInsert(s, query, componentInstanceRow, l)

	if err != nil {
		return nil, err
	}

	componentInstance.Id = id

	return componentInstance, nil
}

func (s *SqlDatabase) UpdateComponentInstance(componentInstance *entity.ComponentInstance) error {
	l := logrus.WithFields(logrus.Fields{
		"componentInstance": componentInstance,
		"event":             "database.UpdateComponentInstance",
	})

	baseQuery := `
		UPDATE ComponentInstance SET
		%s
		WHERE componentinstance_id = :componentinstance_id
	`

	updateFields := s.getComponentInstanceUpdateFields(componentInstance)

	query := fmt.Sprintf(baseQuery, updateFields)

	componentInstanceRow := ComponentInstanceRow{}
	componentInstanceRow.FromComponentInstance(componentInstance)

	_, err := performExec(s, query, componentInstanceRow, l)

	return err
}

func (s *SqlDatabase) DeleteComponentInstance(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteComponentInstance",
	})

	query := `
		UPDATE ComponentInstance SET
		componentinstance_deleted_at = NOW(),
		componentinstance_updated_by = :userId
		WHERE componentinstance_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
func (s *SqlDatabase) GetCcrn(filter *entity.ComponentInstanceFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetCcrn",
	})

	baseQuery := `
    SELECT CI.componentinstance_ccrn FROM ComponentInstance CI 
    %s
    %s
    ORDER BY CI.componentinstance_ccrn
    `

	// Ensure the filter is initialized
	filter = s.ensureComponentInstanceFilter(filter)

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildComponentInstanceStatement(baseQuery, filter, false, l)
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
	ccrn := []string{}
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}
		ccrn = append(ccrn, name)
	}
	if err = rows.Err(); err != nil {
		l.Error("Row iteration error: ", err)
		return nil, err
	}

	return ccrn, nil
}

func (s *SqlDatabase) CreateScannerRunComponentInstanceTracker(componentInstanceId int64, scannerRunUUID string) error {
	query := `
        INSERT INTO scanner_run_component_instance_tracker (
			component_instance_id, 
			scannerruncomponentinstance_scannerrun_run_id
		)
        VALUES (?, ?)
    `

	sr, err := s.ScannerRunByUUID(scannerRunUUID)

	if err != nil {
		return fmt.Errorf("failed to create scanner run component instance tracker: %w", err)
	}
	_, err = s.db.Exec(query, componentInstanceId, sr.RunID)
	if err != nil {
		return fmt.Errorf("failed to create scanner run component instance tracker: %w", err)
	}
	return nil
}
