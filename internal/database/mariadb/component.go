// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var componentObject = DbObject[*entity.Component, *entity.ComponentFilter, entity.ComponentResult, *any]{
	Prefix:       "component",
	TableName:    "Component",
	TableKey:     "C",
	DefaultOrder: entity.Order{By: entity.ComponentId, Direction: entity.OrderDirectionAsc},
	OrderPrefix:  "CVR",
	Properties: []*Property[*entity.Component]{
		NewProperty("component_ccrn", func(c *entity.Component) (any, bool) { return c.CCRN, c.CCRN != "" }),
		NewProperty("component_repository", func(c *entity.Component) (any, bool) { return c.Repository, c.Repository != "" }),
		NewProperty("component_organization", func(c *entity.Component) (any, bool) { return c.Organization, c.Organization != "" }),
		NewProperty("component_url", func(c *entity.Component) (any, bool) { return c.Url, c.Url != "" }),
		NewProperty("component_type", func(c *entity.Component) (any, bool) { return c.Type, c.Type != "" }),
		NewProperty("component_created_by", func(c *entity.Component) (any, bool) { return c.CreatedBy, NoUpdate }),
		NewProperty("component_updated_by", func(c *entity.Component) (any, bool) { return c.UpdatedBy, c.UpdatedBy != 0 }),
	},
	FilterProperties: []*FilterProperty[*entity.ComponentFilter]{
		NewFilterProperty("C.component_ccrn = ?", func(filter *entity.ComponentFilter) any { return filter.CCRN }),
		NewFilterProperty("C.component_repository = ?", func(filter *entity.ComponentFilter) any { return filter.Repository }),
		NewFilterProperty("C.component_organization = ?", func(filter *entity.ComponentFilter) any { return filter.Organization }),
		NewFilterProperty("C.component_id = ?", func(filter *entity.ComponentFilter) any { return filter.Id }),
		NewFilterProperty("CV.componentversion_id = ?", func(filter *entity.ComponentFilter) any { return filter.ComponentVersionId }),
		NewFilterProperty("S.service_ccrn = ?", func(filter *entity.ComponentFilter) any { return filter.ServiceCCRN }),
		NewStateFilterProperty("C.component", func(filter *entity.ComponentFilter) any { return filter.State }),
	},
	JoinDefs: []*JoinDef[*entity.ComponentFilter]{
		// --- Legacy path: CV→CI→S join chain (used when UseMvComponentService is false) ---
		{
			Name:      "CV",
			Type:      LeftJoin,
			Table:     "ComponentVersion CV",
			On:        "C.component_id = CV.componentversion_component_id",
			Condition: func(f *entity.ComponentFilter, _ *Order) bool { return len(f.ComponentVersionId) > 0 },
		},
		{
			Name:      "CI",
			Type:      LeftJoin,
			Table:     "ComponentInstance CI",
			On:        "CV.componentversion_id = CI.componentinstance_component_version_id",
			DependsOn: []string{"CV"},
			Condition: DependentJoin[*entity.ComponentFilter],
		},
		{
			// Legacy service join via ComponentInstance — expensive on large tables.
			// Guarded: only activates when the MV path is not in use.
			Name:      "S",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			DependsOn: []string{"CI"},
			Condition: func(f *entity.ComponentFilter, _ *Order) bool {
				return !f.UseMvComponentService && len(f.ServiceCCRN) > 0
			},
		},
		{
			Name:      "mvSCBSVC",
			Type:      LeftJoin,
			Table:     "mvSingleComponentByServiceVulnerabilityCounts CVR",
			On:        "C.component_id = CVR.component_id AND CVR.service_id = S.service_id",
			DependsOn: []string{"S"},
			Condition: func(f *entity.ComponentFilter, order *Order) bool {
				return !f.UseMvComponentService && needSingleComponentByServiceVulnerabilityCounts(f, order)
			},
		},
		{
			Name:      "mvACBSVC",
			Type:      LeftJoin,
			Table:     "mvAllComponentsByServiceVulnerabilityCounts CVR",
			On:        "S.service_id = CVR.service_id",
			DependsOn: []string{"S"},
			Condition: func(f *entity.ComponentFilter, order *Order) bool {
				return false
			},
		},

		// --- Optimized MV path: mvComponentService replaces the CV→CI→S chain ---
		// Uses a pre-computed junction table of (service_id, component_id) to avoid
		// scanning millions of ComponentInstance rows at query time.
		{
			Name:  "MCS",
			Type:  InnerJoin,
			Table: "mvComponentService MCS",
			On:    "C.component_id = MCS.component_id",
			Condition: func(f *entity.ComponentFilter, _ *Order) bool {
				return f.UseMvComponentService && len(f.ServiceCCRN) > 0
			},
		},
		{
			// Service join via mvComponentService — activates when MV path is used
			// with a service filter. This replaces the expensive CV→CI→S chain.
			Name:      "S via MCS",
			Type:      InnerJoin,
			Table:     "Service S",
			On:        "MCS.service_id = S.service_id",
			DependsOn: []string{"MCS"},
			Condition: func(f *entity.ComponentFilter, _ *Order) bool {
				return f.UseMvComponentService && len(f.ServiceCCRN) > 0
			},
		},
		{
			Name:      "mvSCBSVC via MCS",
			Type:      LeftJoin,
			Table:     "mvSingleComponentByServiceVulnerabilityCounts CVR",
			On:        "C.component_id = CVR.component_id AND MCS.service_id = CVR.service_id",
			DependsOn: []string{"MCS"},
			Condition: func(f *entity.ComponentFilter, order *Order) bool {
				return f.UseMvComponentService && needSingleComponentByServiceVulnerabilityCounts(f, order)
			},
		},
	},
	Attributes: []Attr{{Name: "ccrn", Order: entity.Order{By: entity.ComponentCcrn, Direction: entity.OrderDirectionAsc}}},
	ExtraColumnsSelector: func(_ *entity.ComponentFilter, order *Order) []string {
		s := []string{}
		for _, o := range order.Sequence() {
			switch o.By {
			case entity.CriticalCount:
				s = append(s, "CVR.critical_count")
			case entity.HighCount:
				s = append(s, "CVR.high_count")
			case entity.MediumCount:
				s = append(s, "CVR.medium_count")
			case entity.LowCount:
				s = append(s, "CVR.low_count")
			case entity.NoneCount:
				s = append(s, "CVR.none_count")
			}
		}
		return s
	},
	RowToData: func(e RowComposite, order []entity.Order) (*entity.Component, string) {
		c := e.AsComponent()

		var isc entity.IssueSeverityCounts
		if e.RatingCount != nil {
			isc = e.AsIssueSeverityCounts()
		}

		cursor, _ := EncodeCursor(WithComponent(order, c, isc))

		return &c, cursor
	},
	NewResult: func(c *entity.Component, _ *any, cursor string) entity.ComponentResult {
		return entity.ComponentResult{
			WithCursor: entity.WithCursor{Value: cursor},
			Component:  c,
		}
	},
}

func needSingleComponentByServiceVulnerabilityCounts(filter *entity.ComponentFilter, order *Order) bool {
	return order.ByCount() && len(filter.ServiceCCRN) > 0
}

func (s *SqlDatabase) GetAllComponentCursors(ctx context.Context, filter *entity.ComponentFilter, order []entity.Order) ([]string, error) {
	return componentObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetComponents(ctx context.Context, filter *entity.ComponentFilter, order []entity.Order) ([]entity.ComponentResult, error) {
	return componentObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountComponents(ctx context.Context, filter *entity.ComponentFilter) (int64, error) {
	return componentObject.Count(ctx, s.db, filter)
}

// TODO use DbObject
func (s *SqlDatabase) CountComponentVulnerabilities(
	ctx context.Context,
	filter *entity.ComponentFilter,
) ([]entity.IssueSeverityCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountComponentVulnerabilities",
	})

	var (
		fl               []string
		filterParameters []any
	)

	filter = EnsureFilter(filter)

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

	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			},
		).Error(msg)

		return nil, fmt.Errorf("%s", msg)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	return performListScan(
		ctx,
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

func (s *SqlDatabase) GetComponentCcrns(ctx context.Context, filter *entity.ComponentFilter) ([]string, error) {
	return componentObject.GetAttr(ctx, s.db, "ccrn", filter)
}
