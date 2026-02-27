// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var componentInstanceObject = DbObject{
	Properties: []PropertySpec{
		Property{Name: "componentinstance_ccrn", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.CCRN != "" })},
		Property{Name: "componentinstance_region", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.Region != "" })},
		Property{Name: "componentinstance_cluster", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.Cluster != "" })},
		Property{Name: "componentinstance_namespace", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.Namespace != "" })},
		Property{Name: "componentinstance_domain", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.Domain != "" })},
		Property{Name: "componentinstance_project", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.Project != "" })},
		Property{Name: "componentinstance_pod", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.Pod != "" })},
		Property{Name: "componentinstance_container", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.Container != "" })},
		Property{Name: "componentinstance_type", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.Type != "" })},
		Property{Name: "componentinstance_parent_id", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.ParentId != 0 })},
		Property{Name: "componentinstance_context", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.Context != nil })},
		Property{Name: "componentinstance_count", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.Count != 0 })},
		Property{Name: "componentinstance_component_version_id", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.ComponentVersionId != 0 })},
		Property{Name: "componentinstance_service_id", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.ServiceId != 0 })},
		Property{Name: "componentinstance_created_by"},
		Property{Name: "componentinstance_updated_by", IsUpdatePresent: WrapChecker(func(ci *entity.ComponentInstance) bool { return ci.UpdatedBy != 0 })},
	},
	FilterProperties: []FilterPropertySpec{
		FilterProperty{Query: "CI.componentinstance_id = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*int64 { return filter.Id })},
		FilterProperty{Query: "CI.componentinstance_ccrn = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.CCRN })},
		FilterProperty{Query: "CI.componentinstance_region = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Region })},
		FilterProperty{Query: "CI.componentinstance_cluster = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Cluster })},
		FilterProperty{Query: "CI.componentinstance_namespace = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Namespace })},
		FilterProperty{Query: "CI.componentinstance_domain = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Domain })},
		FilterProperty{Query: "CI.componentinstance_project = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Project })},
		FilterProperty{Query: "CI.componentinstance_pod = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Pod })},
		FilterProperty{Query: "CI.componentinstance_container = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Container })},
		FilterProperty{Query: "CI.componentinstance_type = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Type })},
		FilterProperty{Query: "CI.componentinstance_parent_id = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*int64 { return filter.ParentId })},
		JsonFilterProperty{Query: "CI.componentinstance_context", Param: WrapRetJson(func(filter *entity.ComponentInstanceFilter) []*entity.Json { return filter.Context })},
		FilterProperty{Query: "IM.issuematch_id = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*int64 { return filter.IssueMatchId })},
		FilterProperty{Query: "CI.componentinstance_service_id = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*int64 { return filter.ServiceId })},
		FilterProperty{Query: "S.service_ccrn = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.ServiceCcrn })},
		FilterProperty{Query: "CI.componentinstance_component_version_id = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*int64 { return filter.ComponentVersionId })},
		FilterProperty{Query: "CV.componentversion_version = ?", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.ComponentVersionVersion })},
		FilterProperty{Query: "CI.componentinstance_ccrn LIKE Concat('%',?,'%')", Param: WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Search })},
		StateFilterProperty{Prefix: "CI.componentinstance", Param: WrapRetState(func(filter *entity.ComponentInstanceFilter) []entity.StateFilterType { return filter.State })},
	},
}

func ensureComponentInstanceFilter(filter *entity.ComponentInstanceFilter) *entity.ComponentInstanceFilter {
	if filter == nil {
		filter = &entity.ComponentInstanceFilter{}
	}
	return EnsurePagination(filter)
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

func (s *SqlDatabase) buildComponentInstanceStatement(baseQuery string, filter *entity.ComponentInstanceFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	filter = ensureComponentInstanceFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode cursor: %w", err)
	}
	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.ComponentInstanceId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)
	joins := s.getComponentInstanceJoins(filter)

	filterStr := componentInstanceObject.GetFilterQuery(filter)
	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && withCursor && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	// construct final query
	var query string
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
		return nil, nil, fmt.Errorf("failed to prepare ComponentInstance statement: %w", err)
	}

	filterParameters := componentInstanceObject.GetFilterParameters(filter, withCursor, cursorFields)

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
		return nil, fmt.Errorf("failed to build ComponentInstance IDs query: %w", err)
	}
	defer stmt.Close()

	ids, err := performIdScan(stmt, filterParameters, l)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance IDs: %w", err)
	}

	return ids, nil
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
		return nil, fmt.Errorf("failed to build ComponentInstances query: %w", err)
	}
	defer stmt.Close()

	results, err := performListScan(
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
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstances: %w", err)
	}

	return results, nil
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
		return nil, fmt.Errorf("failed to build ComponentInstance cursors query: %w", err)
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
		return nil, fmt.Errorf("failed to get ComponentInstance cursors: %w", err)
	}

	cursors := lo.Map(rows, func(row RowComposite, _ int) string {
		ci := row.AsComponentInstance()
		cursor, _ := EncodeCursor(WithComponentInstance(order, ci))
		return cursor
	})

	return cursors, nil
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
		return -1, fmt.Errorf("failed to build ComponentInstance count query: %w", err)
	}
	defer stmt.Close()

	count, err := performCountScan(stmt, filterParameters, l)
	if err != nil {
		return -1, fmt.Errorf("failed to count ComponentInstances: %w", err)
	}

	return count, nil
}

func (s *SqlDatabase) CreateComponentInstance(componentInstance *entity.ComponentInstance) (*entity.ComponentInstance, error) {
	l := logrus.WithFields(logrus.Fields{
		"componentInstance": componentInstance,
		"event":             "database.CreateComponentInstance",
	})

	componentInstanceRow := ComponentInstanceRow{}
	componentInstanceRow.FromComponentInstance(componentInstance)

	query := componentInstanceObject.InsertQuery("ComponentInstance")
	id, err := performInsert(s, query, componentInstanceRow, l)
	if err != nil {
		return nil, fmt.Errorf("failed to create ComponentInstance with CCRN '%s': %w",
			componentInstance.CCRN, err)
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

	updateFields := componentInstanceObject.GetUpdateFields(componentInstance)
	query := fmt.Sprintf(baseQuery, updateFields)

	componentInstanceRow := ComponentInstanceRow{}
	componentInstanceRow.FromComponentInstance(componentInstance)

	_, err := performExec(s, query, componentInstanceRow, l)
	if err != nil {
		return fmt.Errorf("failed to update ComponentInstance with ID %d (CCRN: '%s'): %w",
			componentInstance.Id, componentInstance.CCRN, err)
	}

	return nil
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
	if err != nil {
		return fmt.Errorf("failed to delete ComponentInstance with ID %d: %w", id, err)
	}

	return nil
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
	filter = ensureComponentInstanceFilter(filter)

	order := []entity.Order{
		{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc},
	}

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildComponentInstanceStatement(baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build ComponentInstance attribute query for %s: %w", attrName, err)
	}
	defer stmt.Close()

	// Execute the query
	rows, err := stmt.Queryx(filterParameters...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute ComponentInstance attribute query for %s: %w", attrName, err)
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
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ComponentInstance attribute rows for %s: %w", attrName, err)
	}

	return attrVal, nil
}

func (s *SqlDatabase) GetCcrn(filter *entity.ComponentInstanceFilter) ([]string, error) {
	ccrns, err := s.getComponentInstanceAttr("ccrn", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance CCRNs: %w", err)
	}
	return ccrns, nil
}

func (s *SqlDatabase) GetRegion(filter *entity.ComponentInstanceFilter) ([]string, error) {
	regions, err := s.getComponentInstanceAttr("region", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance regions: %w", err)
	}
	return regions, nil
}

func (s *SqlDatabase) GetCluster(filter *entity.ComponentInstanceFilter) ([]string, error) {
	clusters, err := s.getComponentInstanceAttr("cluster", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance clusters: %w", err)
	}
	return clusters, nil
}

func (s *SqlDatabase) GetNamespace(filter *entity.ComponentInstanceFilter) ([]string, error) {
	namespaces, err := s.getComponentInstanceAttr("namespace", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance namespaces: %w", err)
	}
	return namespaces, nil
}

func (s *SqlDatabase) GetDomain(filter *entity.ComponentInstanceFilter) ([]string, error) {
	domains, err := s.getComponentInstanceAttr("domain", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance domains: %w", err)
	}
	return domains, nil
}

func (s *SqlDatabase) GetProject(filter *entity.ComponentInstanceFilter) ([]string, error) {
	projects, err := s.getComponentInstanceAttr("project", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance projects: %w", err)
	}
	return projects, nil
}

func (s *SqlDatabase) GetPod(filter *entity.ComponentInstanceFilter) ([]string, error) {
	pods, err := s.getComponentInstanceAttr("pod", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance pods: %w", err)
	}
	return pods, nil
}

func (s *SqlDatabase) GetContainer(filter *entity.ComponentInstanceFilter) ([]string, error) {
	containers, err := s.getComponentInstanceAttr("container", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance containers: %w", err)
	}
	return containers, nil
}

func (s *SqlDatabase) GetType(filter *entity.ComponentInstanceFilter) ([]string, error) {
	types, err := s.getComponentInstanceAttr("type", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance types: %w", err)
	}
	return types, nil
}

func (s *SqlDatabase) GetComponentInstanceParent(filter *entity.ComponentInstanceFilter) ([]string, error) {
	parents, err := s.getComponentInstanceAttr("parent_id", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance parents: %w", err)
	}
	return parents, nil
}

func (s *SqlDatabase) GetContext(filter *entity.ComponentInstanceFilter) ([]string, error) {
	contexts, err := s.getComponentInstanceAttr("context", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance contexts: %w", err)
	}
	return contexts, nil
}

func (s *SqlDatabase) CreateScannerRunComponentInstanceTracker(componentInstanceId int64, scannerRunUUID string) error {
	query := `
        INSERT INTO ScannerRunComponentInstanceTracker (
			scannerruncomponentinstancetracker_component_instance_id, 
			scannerruncomponentinstancetracker_scannerrun_run_id
		)
        VALUES (?, ?)
    `

	sr, err := s.ScannerRunByUUID(scannerRunUUID)
	if err != nil {
		return fmt.Errorf("failed to get scanner run by UUID '%s': %w", scannerRunUUID, err)
	}

	_, err = s.db.Exec(query, componentInstanceId, sr.RunID)
	if err != nil {
		return fmt.Errorf("failed to create scanner run component instance tracker for ComponentInstance %d and ScannerRun '%s': %w",
			componentInstanceId, scannerRunUUID, err)
	}

	return nil
}
