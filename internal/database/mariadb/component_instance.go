// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var componentInstanceObject = DbObject[*entity.ComponentInstance]{
	Prefix:    "componentinstance",
	TableName: "ComponentInstance",
	Properties: []*Property{
		NewProperty("componentinstance_ccrn", WrapAccess(func(ci *entity.ComponentInstance) (string, bool) { return ci.CCRN, ci.CCRN != "" })),
		NewProperty("componentinstance_region", WrapAccess(func(ci *entity.ComponentInstance) (string, bool) { return ci.Region, ci.Region != "" })),
		NewProperty("componentinstance_cluster", WrapAccess(func(ci *entity.ComponentInstance) (string, bool) { return ci.Cluster, ci.Cluster != "" })),
		NewProperty("componentinstance_namespace", WrapAccess(func(ci *entity.ComponentInstance) (string, bool) { return ci.Namespace, ci.Namespace != "" })),
		NewProperty("componentinstance_domain", WrapAccess(func(ci *entity.ComponentInstance) (string, bool) { return ci.Domain, ci.Domain != "" })),
		NewProperty("componentinstance_project", WrapAccess(func(ci *entity.ComponentInstance) (string, bool) { return ci.Project, ci.Project != "" })),
		NewProperty("componentinstance_pod", WrapAccess(func(ci *entity.ComponentInstance) (string, bool) { return ci.Pod, ci.Pod != "" })),
		NewProperty("componentinstance_container", WrapAccess(func(ci *entity.ComponentInstance) (string, bool) { return ci.Container, ci.Container != "" })),
		NewProperty("componentinstance_type", WrapAccess(func(ci *entity.ComponentInstance) (entity.ComponentInstanceType, bool) { return ci.Type, ci.Type != "" })),
		NewProperty("componentinstance_parent_id", WrapAccess(func(ci *entity.ComponentInstance) (int64, bool) { return ci.ParentId, ci.ParentId != 0 })),
		NewProperty("componentinstance_context", WrapAccess(func(ci *entity.ComponentInstance) (*entity.Json, bool) { return ci.Context, ci.Context != nil })),
		NewProperty("componentinstance_count", WrapAccess(func(ci *entity.ComponentInstance) (int16, bool) { return ci.Count, ci.Count != 0 })),
		NewProperty("componentinstance_component_version_id", WrapAccess(func(ci *entity.ComponentInstance) (int64, bool) {
			return ci.ComponentVersionId, ci.ComponentVersionId != 0
		})),
		NewProperty("componentinstance_service_id", WrapAccess(func(ci *entity.ComponentInstance) (int64, bool) { return ci.ServiceId, ci.ServiceId != 0 })),
		NewProperty("componentinstance_created_by", WrapAccess(func(ci *entity.ComponentInstance) (int64, bool) { return ci.CreatedBy, NoUpdate })),
		NewProperty("componentinstance_updated_by", WrapAccess(func(ci *entity.ComponentInstance) (int64, bool) { return ci.UpdatedBy, ci.UpdatedBy != 0 })),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty("CI.componentinstance_id = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*int64 { return filter.Id })),
		NewFilterProperty("CI.componentinstance_ccrn = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.CCRN })),
		NewFilterProperty("CI.componentinstance_region = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Region })),
		NewFilterProperty("CI.componentinstance_cluster = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Cluster })),
		NewFilterProperty("CI.componentinstance_namespace = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Namespace })),
		NewFilterProperty("CI.componentinstance_domain = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Domain })),
		NewFilterProperty("CI.componentinstance_project = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Project })),
		NewFilterProperty("CI.componentinstance_pod = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Pod })),
		NewFilterProperty("CI.componentinstance_container = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Container })),
		NewFilterProperty("CI.componentinstance_type = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Type })),
		NewFilterProperty("CI.componentinstance_parent_id = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*int64 { return filter.ParentId })),
		NewJsonFilterProperty("CI.componentinstance_context", WrapRetJson(func(filter *entity.ComponentInstanceFilter) []*entity.Json { return filter.Context })),
		NewFilterProperty("IM.issuematch_id = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*int64 { return filter.IssueMatchId })),
		NewFilterProperty("CI.componentinstance_service_id = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*int64 { return filter.ServiceId })),
		NewFilterProperty("S.service_ccrn = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.ServiceCcrn })),
		NewFilterProperty("CI.componentinstance_component_version_id = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*int64 { return filter.ComponentVersionId })),
		NewFilterProperty("CV.componentversion_version = ?", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.ComponentVersionVersion })),
		NewFilterProperty("CI.componentinstance_ccrn LIKE Concat('%',?,'%')", WrapRetSlice(func(filter *entity.ComponentInstanceFilter) []*string { return filter.Search })),
		NewStateFilterProperty("CI.componentinstance", WrapRetState(func(filter *entity.ComponentInstanceFilter) []entity.StateFilterType { return filter.State })),
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

	cursorFields, err := DecodeCursor(filter.Paginated.After)
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
	return componentInstanceObject.Create(s.db, componentInstance)
}

func (s *SqlDatabase) UpdateComponentInstance(componentInstance *entity.ComponentInstance) error {
	return componentInstanceObject.Update(s.db, componentInstance)
}

func (s *SqlDatabase) DeleteComponentInstance(id int64, userId int64) error {
	return componentInstanceObject.Delete(s.db, id, userId)
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
