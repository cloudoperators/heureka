// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/entity"
)

var componentInstanceObject = DbObject[*entity.ComponentInstance, *entity.ComponentInstanceFilter, entity.ComponentInstanceResult, *any]{
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
	Attributes: []Attr{
		{Name: "ccrn", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "region", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "cluster", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "namespace", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "domain", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "project", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "pod", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "container", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "type", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "parent_id", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
		{Name: "context", Order: entity.Order{By: entity.ComponentInstanceCcrn, Direction: entity.OrderDirectionAsc}},
	},
	RowToData: func(e RowComposite, order []entity.Order) (*entity.ComponentInstance, string) {
		ci := e.AsComponentInstance()

		cursor, _ := EncodeCursor(WithComponentInstance(order, ci))

		return &ci, cursor
	},
	NewResult: func(ci *entity.ComponentInstance, _ *any, cursor string) entity.ComponentInstanceResult {
		return entity.ComponentInstanceResult{
			WithCursor:        entity.WithCursor{Value: cursor},
			ComponentInstance: ci,
		}
	},
}

func (s *SqlDatabase) GetComponentInstances(ctx context.Context, filter *entity.ComponentInstanceFilter, order []entity.Order) ([]entity.ComponentInstanceResult, error) {
	return componentInstanceObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetAllComponentInstanceCursors(ctx context.Context, filter *entity.ComponentInstanceFilter, order []entity.Order) ([]string, error) {
	return componentInstanceObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountComponentInstances(ctx context.Context, filter *entity.ComponentInstanceFilter) (int64, error) {
	return componentInstanceObject.Count(ctx, s.db, filter)
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

func (s *SqlDatabase) GetCcrn(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "ccrn", filter)
}

func (s *SqlDatabase) GetRegion(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "region", filter)
}

func (s *SqlDatabase) GetCluster(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "cluster", filter)
}

func (s *SqlDatabase) GetNamespace(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "namespace", filter)
}

func (s *SqlDatabase) GetDomain(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "domain", filter)
}

func (s *SqlDatabase) GetProject(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "project", filter)
}

func (s *SqlDatabase) GetPod(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "pod", filter)
}

func (s *SqlDatabase) GetContainer(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "container", filter)
}

func (s *SqlDatabase) GetType(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "type", filter)
}

func (s *SqlDatabase) GetComponentInstanceParent(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "parent_id", filter)
}

func (s *SqlDatabase) GetContext(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error) {
	return componentInstanceObject.GetAttr(ctx, s.db, "context", filter)
}

func (s *SqlDatabase) CreateScannerRunComponentInstanceTracker(componentInstanceId int64, scannerRunUUID string) error {
	sr, err := s.ScannerRunByUUID(scannerRunUUID)
	if err != nil {
		return fmt.Errorf("failed to get scanner run by UUID '%s': %w", scannerRunUUID, err)
	}

	qb := sq.Insert("ScannerRunComponentInstanceTracker").
		Columns("scannerruncomponentinstancetracker_component_instance_id", "scannerruncomponentinstancetracker_scannerrun_run_id").
		Values(componentInstanceId, sr.RunID)

	query, params, err := qb.ToSql()
	if err != nil {
		return fmt.Errorf(
			"failed to create scanner run component instance tracker for ComponentInstance %d and ScannerRun '%s': %w",
			componentInstanceId,
			scannerRunUUID,
			err,
		)
	}

	_, err = s.db.Exec(query, params...)
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
