// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var componentInstanceObject = DbObject[*entity.ComponentInstance, *entity.ComponentInstanceFilter, entity.ComponentInstanceResult]{
	Prefix:       "componentinstance",
	TableName:    "ComponentInstance",
	TableKey:     "CI",
	DefaultOrder: entity.Order{By: entity.ComponentInstanceId, Direction: entity.OrderDirectionAsc},
	Properties: []*Property[*entity.ComponentInstance]{
		NewProperty("componentinstance_ccrn", func(ci *entity.ComponentInstance) (any, bool) { return ci.CCRN, ci.CCRN != "" }),
		NewProperty("componentinstance_region", func(ci *entity.ComponentInstance) (any, bool) { return ci.Region, ci.Region != "" }),
		NewProperty("componentinstance_cluster", func(ci *entity.ComponentInstance) (any, bool) { return ci.Cluster, ci.Cluster != "" }),
		NewProperty("componentinstance_namespace", func(ci *entity.ComponentInstance) (any, bool) { return ci.Namespace, ci.Namespace != "" }),
		NewProperty("componentinstance_domain", func(ci *entity.ComponentInstance) (any, bool) { return ci.Domain, ci.Domain != "" }),
		NewProperty("componentinstance_project", func(ci *entity.ComponentInstance) (any, bool) { return ci.Project, ci.Project != "" }),
		NewProperty("componentinstance_pod", func(ci *entity.ComponentInstance) (any, bool) { return ci.Pod, ci.Pod != "" }),
		NewProperty("componentinstance_container", func(ci *entity.ComponentInstance) (any, bool) { return ci.Container, ci.Container != "" }),
		NewProperty("componentinstance_type", func(ci *entity.ComponentInstance) (any, bool) { return ci.Type, ci.Type != "" }),
		NewProperty("componentinstance_parent_id",
			func(ci *entity.ComponentInstance) (any, bool) {
				return sql.NullInt64{Int64: ci.ParentId, Valid: IsValidId(ci.ParentId)}, ci.ParentId != 0
			}),
		NewProperty("componentinstance_context", func(ci *entity.ComponentInstance) (any, bool) { return ci.Context, ci.Context != nil }),
		NewProperty("componentinstance_count", func(ci *entity.ComponentInstance) (any, bool) { return ci.Count, ci.Count != 0 }),
		NewProperty("componentinstance_component_version_id",
			func(ci *entity.ComponentInstance) (any, bool) {
				return sql.NullInt64{Int64: ci.ComponentVersionId, Valid: IsValidId(ci.ComponentVersionId)}, ci.ComponentVersionId != 0
			}),
		NewProperty("componentinstance_service_id", func(ci *entity.ComponentInstance) (any, bool) { return ci.ServiceId, ci.ServiceId != 0 }),
		NewProperty("componentinstance_created_by", func(ci *entity.ComponentInstance) (any, bool) { return ci.CreatedBy, NoUpdate }),
		NewProperty("componentinstance_updated_by", func(ci *entity.ComponentInstance) (any, bool) { return ci.UpdatedBy, ci.UpdatedBy != 0 }),
	},
	FilterProperties: []*FilterProperty[*entity.ComponentInstanceFilter]{
		NewFilterProperty("CI.componentinstance_id = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.Id }),
		NewFilterProperty("CI.componentinstance_ccrn = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.CCRN }),
		NewFilterProperty("CI.componentinstance_region = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.Region }),
		NewFilterProperty("CI.componentinstance_cluster = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.Cluster }),
		NewFilterProperty("CI.componentinstance_namespace = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.Namespace }),
		NewFilterProperty("CI.componentinstance_domain = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.Domain }),
		NewFilterProperty("CI.componentinstance_project = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.Project }),
		NewFilterProperty("CI.componentinstance_pod = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.Pod }),
		NewFilterProperty("CI.componentinstance_container = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.Container }),
		NewFilterProperty("CI.componentinstance_type = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.Type }),
		NewFilterProperty("CI.componentinstance_parent_id = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.ParentId }),
		NewJsonFilterProperty("CI.componentinstance_context", func(filter *entity.ComponentInstanceFilter) any { return filter.Context }),
		NewFilterProperty("IM.issuematch_id = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.IssueMatchId }),
		NewFilterProperty("CI.componentinstance_service_id = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.ServiceId }),
		NewFilterProperty("S.service_ccrn = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.ServiceCcrn }),
		NewFilterProperty("CI.componentinstance_component_version_id = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.ComponentVersionId }),
		NewFilterProperty("CV.componentversion_version = ?", func(filter *entity.ComponentInstanceFilter) any { return filter.ComponentVersionVersion }),
		NewFilterProperty("CI.componentinstance_ccrn LIKE Concat('%',?,'%')", func(filter *entity.ComponentInstanceFilter) any { return filter.Search }),
		NewStateFilterProperty("CI.componentinstance", func(filter *entity.ComponentInstanceFilter) any { return filter.State }),
	},
	JoinDefs: []*JoinDef[*entity.ComponentInstanceFilter]{
		{
			Name:      "IM",
			Type:      InnerJoin,
			Table:     "IssueMatch IM",
			On:        "CI.componentinstance_id = IM.issuematch_component_instance_id",
			Condition: func(f *entity.ComponentInstanceFilter, _ *Order) bool { return len(f.IssueMatchId) > 0 },
		},
		{
			Name:      "S",
			Type:      InnerJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			Condition: func(f *entity.ComponentInstanceFilter, _ *Order) bool { return len(f.ServiceCcrn) > 0 },
		},
		{
			Name:      "CV",
			Type:      InnerJoin,
			Table:     "ComponentVersion CV",
			On:        "CI.componentinstance_component_version_id = CV.componentversion_id",
			Condition: func(f *entity.ComponentInstanceFilter, _ *Order) bool { return len(f.ComponentVersionVersion) > 0 },
		},
	},
	GetItemAppender: func(l []entity.ComponentInstanceResult, e RowComposite, order []entity.Order) []entity.ComponentInstanceResult {
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
	GetAllCursorItemAppender: func(l []string, e RowComposite, order []entity.Order) []string {
		ci := e.AsComponentInstance()

		cursor, _ := EncodeCursor(WithComponentInstance(order, ci))

		return append(l, cursor)
	},
}

func (s *SqlDatabase) buildComponentInstanceStatement(
	ctx context.Context,
	baseQuery sq.SelectBuilder,
	filter *entity.ComponentInstanceFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	statement := Statement[*entity.ComponentInstanceFilter]{
		Db:         s.db,
		L:          l,
		Obj:        &componentInstanceObject,
		BaseQuery:  baseQuery,
		Order:      NewOrder(order, componentInstanceObject.DefaultOrder),
		WithCursor: withCursor,
	}

	return BuildStatement(ctx, statement, filter)
}

func (s *SqlDatabase) GetComponentInstances(
	ctx context.Context,
	filter *entity.ComponentInstanceFilter,
	order []entity.Order,
) ([]entity.ComponentInstanceResult, error) {
	return componentInstanceObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetAllComponentInstanceCursors(
	ctx context.Context,
	filter *entity.ComponentInstanceFilter,
	order []entity.Order,
) ([]string, error) {
	return componentInstanceObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountComponentInstances(ctx context.Context, filter *entity.ComponentInstanceFilter) (int64, error) {
	return componentInstanceObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) CreateComponentInstance(
	componentInstance *entity.ComponentInstance,
) (*entity.ComponentInstance, error) {
	return componentInstanceObject.Create(s.db, componentInstance)
}

func (s *SqlDatabase) UpdateComponentInstance(componentInstance *entity.ComponentInstance) error {
	return componentInstanceObject.Update(s.db, componentInstance)
}

func (s *SqlDatabase) DeleteComponentInstance(id int64, userId int64) error {
	return componentInstanceObject.Delete(s.db, id, userId)
}

func (s *SqlDatabase) getComponentInstanceAttr(
	ctx context.Context,
	attrName string,
	filter *entity.ComponentInstanceFilter,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.getComponentInstanceAttr",
	})

	baseQuery := sq.Select(fmt.Sprintf("CI.componentinstance_%s", attrName)).From("ComponentInstance CI")
	order := []entity.Order{
		{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc},
	}

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildComponentInstanceStatement(
		ctx,
		baseQuery,
		filter,
		false,
		order,
		l,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to build ComponentInstance attribute query for %s: %w",
			attrName,
			err,
		)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	// Execute the query
	rows, err := stmt.QueryxContext(ctx, filterParameters...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to execute ComponentInstance attribute query for %s: %w",
			attrName,
			err,
		)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Warnf("error during close rows: %s", err)
		}
	}()

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
		return nil, fmt.Errorf(
			"error iterating ComponentInstance attribute rows for %s: %w",
			attrName,
			err,
		)
	}

	return attrVal, nil
}

func (s *SqlDatabase) GetCcrn(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	ccrns, err := s.getComponentInstanceAttr(ctx, "ccrn", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance CCRNs: %w", err)
	}

	return ccrns, nil
}

func (s *SqlDatabase) GetRegion(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	regions, err := s.getComponentInstanceAttr(ctx, "region", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance regions: %w", err)
	}

	return regions, nil
}

func (s *SqlDatabase) GetCluster(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	clusters, err := s.getComponentInstanceAttr(ctx, "cluster", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance clusters: %w", err)
	}

	return clusters, nil
}

func (s *SqlDatabase) GetNamespace(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	namespaces, err := s.getComponentInstanceAttr(ctx, "namespace", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance namespaces: %w", err)
	}

	return namespaces, nil
}

func (s *SqlDatabase) GetDomain(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	domains, err := s.getComponentInstanceAttr(ctx, "domain", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance domains: %w", err)
	}

	return domains, nil
}

func (s *SqlDatabase) GetProject(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	projects, err := s.getComponentInstanceAttr(ctx, "project", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance projects: %w", err)
	}

	return projects, nil
}

func (s *SqlDatabase) GetPod(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	pods, err := s.getComponentInstanceAttr(ctx, "pod", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance pods: %w", err)
	}

	return pods, nil
}

func (s *SqlDatabase) GetContainer(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	containers, err := s.getComponentInstanceAttr(ctx, "container", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance containers: %w", err)
	}

	return containers, nil
}

func (s *SqlDatabase) GetType(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	types, err := s.getComponentInstanceAttr(ctx, "type", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance types: %w", err)
	}

	return types, nil
}

func (s *SqlDatabase) GetComponentInstanceParent(
	ctx context.Context,
	filter *entity.ComponentInstanceFilter,
) ([]string, error) {
	parents, err := s.getComponentInstanceAttr(ctx, "parent_id", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance parents: %w", err)
	}

	return parents, nil
}

func (s *SqlDatabase) GetContext(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	contexts, err := s.getComponentInstanceAttr(ctx, "context", filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get ComponentInstance contexts: %w", err)
	}

	return contexts, nil
}

func (s *SqlDatabase) CreateScannerRunComponentInstanceTracker(
	componentInstanceId int64,
	scannerRunUUID string,
) error {
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
		return fmt.Errorf(
			"failed to create scanner run component instance tracker for ComponentInstance %d and ScannerRun '%s': %w",
			componentInstanceId,
			scannerRunUUID,
			err,
		)
	}

	return nil
}
