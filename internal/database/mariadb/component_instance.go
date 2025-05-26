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

func (s *SqlDatabase) ensureComponentInstanceFilter(f *entity.ComponentInstanceFilter) *entity.ComponentInstanceFilter {
	var first int = 1000
	var after string = ""
	if f == nil {
		return &entity.ComponentInstanceFilter{
			PaginatedX: entity.PaginatedX{
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
	fl = append(fl, buildFilterQuery(filter.Pod, "CI.componentinstance_pod = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Container, "CI.componentinstance_container = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Type, "CI.componentinstance_type = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ParentId, "CI.componentinstance_parent_id = ?", OP_OR))
	fl = append(fl, buildJsonFilterQuery(filter.Context, "CI.componentinstance_context", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueMatchId, "IM.issuematch_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "CI.componentinstance_service_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceCcrn, "S.service_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentVersionId, "CI.componentinstance_component_version_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentVersionVersion, "CV.componentversion_version = ?", OP_OR))
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
	if len(filter.ComponentVersionVersion) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "INNER JOIN ComponentVersion CV on CI.componentinstance_component_version_id = CV.componentversion_id")
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
	if componentInstance.Pod != "" {
		fl = append(fl, "componentinstance_pod = :componentinstance_pod")
	}
	if componentInstance.Container != "" {
		fl = append(fl, "componentinstance_container = :componentinstance_container")
	}
	if componentInstance.Type != "" {
		fl = append(fl, "componentinstance_type = :componentinstance_type")
	}
	if componentInstance.ParentId != 0 {
		fl = append(fl, "componentinstance_parent_id = :componentinstance_parent_id")
	}
	if componentInstance.Context != nil {
		fl = append(fl, "componentinstance_context = :componentinstance_context")
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
	if componentInstance.ParentId != 0 {
		fl = append(fl, "componentinstance_parent_id = :componentinstance_parent_id")
	}
	if componentInstance.UpdatedBy != 0 {
		fl = append(fl, "componentinstance_updated_by = :componentinstance_updated_by")
	}

	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildComponentInstanceStatement(baseQuery string, filter *entity.ComponentInstanceFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureComponentInstanceFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getComponentInstanceFilterString(filter)
	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, nil, err
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.ComponentInstanceId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)
	joins := s.getComponentInstanceJoins(filter)

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
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.CCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.Region)
	filterParameters = buildQueryParameters(filterParameters, filter.Cluster)
	filterParameters = buildQueryParameters(filterParameters, filter.Namespace)
	filterParameters = buildQueryParameters(filterParameters, filter.Domain)
	filterParameters = buildQueryParameters(filterParameters, filter.Project)
	filterParameters = buildQueryParameters(filterParameters, filter.Pod)
	filterParameters = buildQueryParameters(filterParameters, filter.Container)
	filterParameters = buildQueryParameters(filterParameters, filter.Type)
	filterParameters = buildQueryParameters(filterParameters, filter.ParentId)
	filterParameters = buildJsonQueryParameters(filterParameters, filter.Context)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchId)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceCcrn)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionVersion)
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

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllComponentInstanceIds(filter *entity.ComponentInstanceFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponentInstanceIds",
	})

	baseQuery := `
		SELECT CI.componentinstance_id FROM ComponentInstance CI 
		%s
	 	%s GROUP BY CI.componentinstance_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildComponentInstanceStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetComponentInstances(filter *entity.ComponentInstanceFilter, order []entity.Order) ([]entity.ComponentInstanceResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponentInstances",
	})
	baseQuery := `
			SELECT CI.* FROM ComponentInstance CI
			%s
			%s
			%s GROUP BY CI.componentinstance_id ORDER BY %s LIMIT ? 
		`

	stmt, filterParameters, err := s.buildComponentInstanceStatement(baseQuery, filter, true, order, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.ComponentInstanceResult, e RowComposite) []entity.ComponentInstanceResult {
			ci := e.AsComponentInstance()

			cursor, _ := EncodeCursor(WithComponentInstance(order, ci))

			cir := entity.ComponentInstanceResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				ComponentInstance: &ci,
			}

			return append(l, cir)
		},
	)
}

func (s *SqlDatabase) GetAllComponentInstanceCursors(filter *entity.ComponentInstanceFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllComponentInstanceCursors",
	})

	baseQuery := `
		SELECT CI.* FROM ComponentInstance CI 
		%s
	    %s GROUP BY CI.componentinstance_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildComponentInstanceStatement(baseQuery, filter, false, order, l)

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
		ci := row.AsComponentInstance()

		cursor, _ := EncodeCursor(WithComponentInstance(order, ci))

		return cursor
	}), nil
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
		ORDER BY %s
	`

	stmt, filterParameters, err := s.buildComponentInstanceStatement(baseQuery, filter, false, []entity.Order{}, l)

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
			componentinstance_pod,
			componentinstance_container,
			componentinstance_type,
			componentinstance_parent_id,
			componentinstance_context,
			componentinstance_count,
			componentinstance_component_version_id,
			componentinstance_service_id,
			componentinstance_parent_id,
			componentinstance_created_by,
			componentinstance_updated_by
		) VALUES (
			:componentinstance_ccrn,
			:componentinstance_region,
			:componentinstance_cluster,
			:componentinstance_namespace,
			:componentinstance_domain,
			:componentinstance_project,
			:componentinstance_pod,
			:componentinstance_container,
			:componentinstance_type,
			:componentinstance_parent_id,
			:componentinstance_context,
			:componentinstance_count,
			:componentinstance_component_version_id,
			:componentinstance_service_id,
			:componentinstance_parent_id,
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

func (s *SqlDatabase) getComponentInstanceAttr(attrName string, filter *entity.ComponentInstanceFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.getComponentInstanceAttr",
	})

	baseQuery := `
    SELECT CI.componentinstance_%s FROM ComponentInstance CI
    %s
    %s
    ORDER BY %s
    `

	baseQuery = fmt.Sprintf(baseQuery, attrName, "%s", "%s", "%s")

	// Ensure the filter is initialized
	filter = s.ensureComponentInstanceFilter(filter)

	order := []entity.Order{
		{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc},
	}

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildComponentInstanceStatement(baseQuery, filter, false, order, l)
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
	attrVal := []string{}
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}
		attrVal = append(attrVal, name)
	}
	// was wenn nil is
	if err = rows.Err(); err != nil {
		l.Error("Row iteration error: ", err)
		return nil, err
	}

	return attrVal, nil
}

func (s *SqlDatabase) GetCcrn(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("ccrn", filter)
}

func (s *SqlDatabase) GetRegion(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("region", filter)
}

func (s *SqlDatabase) GetCluster(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("cluster", filter)
}

func (s *SqlDatabase) GetNamespace(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("namespace", filter)
}

func (s *SqlDatabase) GetDomain(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("domain", filter)
}

func (s *SqlDatabase) GetProject(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("project", filter)
}

func (s *SqlDatabase) GetPod(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("pod", filter)
}

func (s *SqlDatabase) GetContainer(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("container", filter)
}

func (s *SqlDatabase) GetType(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("type", filter)
}

func (s *SqlDatabase) GetComponentInstanceParent(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("parent_id", filter)
}

func (s *SqlDatabase) GetContext(filter *entity.ComponentInstanceFilter) ([]string, error) {
	return s.getComponentInstanceAttr("context", filter)
}

func (s *SqlDatabase) CreateScannerRunComponentInstanceTracker(componentInstanceId int64, scannerRunUUID string) error {
	query := `
        INSERT INTO ScannerRunComponentInstanceTracker (
			scannerruncomponentinstance_component_instance_id, 
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
