// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var componentVersionObject = DbObject[*entity.ComponentVersion]{
	Prefix:    "componentversion",
	TableName: "ComponentVersion",
	Properties: []*Property{
		NewProperty(
			"componentversion_component_id",
			WrapAccess(
				func(cv *entity.ComponentVersion) (int64, bool) { return cv.ComponentId, cv.ComponentId != 0 },
			),
		),
		NewProperty(
			"componentversion_version",
			WrapAccess(
				func(cv *entity.ComponentVersion) (string, bool) { return cv.Version, cv.Version != "" },
			),
		),
		NewProperty(
			"componentversion_tag",
			WrapAccess(
				func(cv *entity.ComponentVersion) (string, bool) { return cv.Tag, cv.Tag != "" },
			),
		),
		NewProperty(
			"componentversion_repository",
			WrapAccess(
				func(cv *entity.ComponentVersion) (string, bool) { return cv.Repository, cv.Repository != "" },
			),
		),
		NewProperty(
			"componentversion_organization",
			WrapAccess(
				func(cv *entity.ComponentVersion) (string, bool) { return cv.Organization, cv.Organization != "" },
			),
		),
		NewProperty(
			"componentversion_created_by",
			WrapAccess(
				func(cv *entity.ComponentVersion) (int64, bool) { return cv.CreatedBy, NoUpdate },
			),
		),
		NewProperty(
			"componentversion_updated_by",
			WrapAccess(
				func(cv *entity.ComponentVersion) (int64, bool) { return cv.UpdatedBy, cv.UpdatedBy != 0 },
			),
		),
		NewProperty(
			"componentversion_end_of_life",
			WrapAccess(func(cv *entity.ComponentVersion) (bool, bool) {
				return ValueOrDefault(cv.EndOfLife, false), cv.EndOfLife != nil
			}),
		),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty(
			"CV.componentversion_id = ?",
			WrapRetSlice(func(filter *entity.ComponentVersionFilter) []*int64 { return filter.Id }),
		),
		NewFilterProperty(
			"CVI.componentversionissue_issue_id = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*int64 { return filter.IssueId },
			),
		),
		NewFilterProperty(
			"CV.componentversion_component_id = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*int64 { return filter.ComponentId },
			),
		),
		NewFilterProperty(
			"CV.componentversion_version = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*string { return filter.Version },
			),
		),
		NewFilterProperty(
			"CV.componentversion_tag = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*string { return filter.Tag },
			),
		),
		NewFilterProperty(
			"CV.componentversion_repository = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*string { return filter.Repository },
			),
		),
		NewFilterProperty(
			"CV.componentversion_organization = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*string { return filter.Organization },
			),
		),
		NewFilterProperty(
			"C.component_ccrn = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*string { return filter.ComponentCCRN },
			),
		),
		NewFilterProperty(
			"S.service_ccrn = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*string { return filter.ServiceCCRN },
			),
		),
		NewFilterProperty(
			"CI.componentinstance_service_id = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*int64 { return filter.ServiceId },
			),
		),
		NewFilterProperty(
			"IV.issuevariant_repository_id = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*int64 { return filter.IssueRepositoryId },
			),
		),
		NewFilterProperty(
			"CV.componentversion_end_of_life = ?",
			WrapRetSlice(
				func(filter *entity.ComponentVersionFilter) []*bool { return filter.EndOfLife },
			),
		),
		NewStateFilterProperty(
			"CV.componentversion",
			WrapRetState(
				func(filter *entity.ComponentVersionFilter) []entity.StateFilterType { return filter.State },
			),
		),
	},
	JoinDefs: []*JoinDef{
		{
			Name:  "CVI",
			Type:  LeftJoin,
			Table: "ComponentVersionIssue CVI",
			On:    "CV.componentversion_id = CVI.componentversionissue_component_version_id",
			Condition: WrapJoinCondition(func(f *entity.ComponentVersionFilter, _ *Order) bool {
				return len(f.IssueId) > 0
			}),
		},
		{
			Name:      "IV",
			Type:      LeftJoin,
			Table:     "IssueVariant IV",
			On:        "CVI.componentversionissue_issue_id = IV.issuevariant_issue_id",
			DependsOn: []string{"CVI"},
			Condition: WrapJoinCondition(func(f *entity.ComponentVersionFilter, order *Order) bool {
				return order.ByCount() || len(f.IssueRepositoryId) > 0
			}),
		},
		{
			Name:  "C",
			Type:  LeftJoin,
			Table: "Component C",
			On:    "CV.componentversion_component_id = C.component_id",
			Condition: WrapJoinCondition(func(f *entity.ComponentVersionFilter, _ *Order) bool {
				return len(f.ComponentCCRN) > 0
			}),
		},
		{
			Name:  "CI",
			Type:  LeftJoin,
			Table: "ComponentInstance CI",
			On:    "CV.componentversion_id = CI.componentinstance_component_version_id",
			Condition: WrapJoinCondition(func(f *entity.ComponentVersionFilter, _ *Order) bool {
				return len(f.ServiceId) > 0
			}),
		},
		{
			Name:      "S",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			DependsOn: []string{"CI"},
			Condition: WrapJoinCondition(func(f *entity.ComponentVersionFilter, _ *Order) bool {
				return len(f.ServiceCCRN) > 0
			}),
		},
	},
}

func appendComponentVersionColumns(s []string, order []entity.Order) []string {
	for _, o := range order {
		switch o.By {
		case entity.CriticalCount:
			s = append(s, "COUNT(distinct CASE WHEN IV.issuevariant_rating = 'Critical' THEN IV.issuevariant_issue_id END) as critical_count")
		case entity.HighCount:
			s = append(s, "COUNT(distinct CASE WHEN IV.issuevariant_rating = 'High' THEN IV.issuevariant_issue_id END) as high_count")
		case entity.MediumCount:
			s = append(s, "COUNT(distinct CASE WHEN IV.issuevariant_rating = 'Medium' THEN IV.issuevariant_issue_id END) as medium_count")
		case entity.LowCount:
			s = append(s, "COUNT(distinct CASE WHEN IV.issuevariant_rating = 'Low' THEN IV.issuevariant_issue_id END) as low_count")
		case entity.NoneCount:
			s = append(s, "COUNT(distinct CASE WHEN IV.issuevariant_rating = 'None' THEN IV.issuevariant_issue_id END) as none_count")
		}
	}

	return s
}

func (s *SqlDatabase) buildComponentVersionStatement(
	ctx context.Context,
	baseQuery sq.SelectBuilder,
	filter *entity.ComponentVersionFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	statement := Statement{
		Db:         s.db,
		L:          l,
		Obj:        &componentVersionObject,
		BaseQuery:  baseQuery,
		Order:      NewOrder(order, entity.Order{By: entity.ComponentVersionId, Direction: entity.OrderDirectionAsc}),
		WithCursor: withCursor,
		//CheckCursorInWhere: false,
		//CheckCursor:        true,
		//CheckFilter:        false,
		Aggregated: true,
	}

	return BuildStatement(ctx, statement, filter)
}

func (s *SqlDatabase) GetAllComponentVersionCursors(
	ctx context.Context,
	filter *entity.ComponentVersionFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllComponentVersionCursors",
	})

	baseQuery := sq.Select(appendComponentVersionColumns([]string{"CV.*"}, order)...).From("ComponentVersion CV").GroupBy("CV.componentversion_id")
	stmt, filterParameters, err := s.buildComponentVersionStatement(ctx, baseQuery, filter, false, order, l)
	if err != nil {
		return nil, err
	}

	rows, err := performListScan(
		ctx,
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
		cv := row.AsComponentVersion()

		var isc entity.IssueSeverityCounts
		if row.RatingCount != nil {
			isc = row.AsIssueSeverityCounts()
		}

		cursor, _ := EncodeCursor(WithComponentVersion(order, cv, isc))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetComponentVersions(
	ctx context.Context,
	filter *entity.ComponentVersionFilter,
	order []entity.Order,
) ([]entity.ComponentVersionResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetComponentVersions",
	})

	baseQuery := sq.Select(appendComponentVersionColumns([]string{"CV.*"}, order)...).From("ComponentVersion CV").GroupBy("CV.componentversion_id")
	stmt, filterParameters, err := s.buildComponentVersionStatement(ctx, baseQuery, filter, true, order, l)
	if err != nil {
		return nil, err
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
		func(l []entity.ComponentVersionResult, e RowComposite) []entity.ComponentVersionResult {
			cv := e.AsComponentVersion()

			var isc entity.IssueSeverityCounts
			if e.RatingCount != nil {
				isc = e.AsIssueSeverityCounts()
			}

			cursor, _ := EncodeCursor(WithComponentVersion(order, cv, isc))

			cvr := entity.ComponentVersionResult{
				WithCursor:       entity.WithCursor{Value: cursor},
				ComponentVersion: &cv,
			}

			return append(l, cvr)
		},
	)
}

func (s *SqlDatabase) CountComponentVersions(ctx context.Context, filter *entity.ComponentVersionFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountComponentVersions",
	})

	baseQuery := sq.Select("count(distinct CV.componentversion_id)").From("ComponentVersion CV")
	stmt, filterParameters, err := s.buildComponentVersionStatement(
		ctx,
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

	return performCountScan(ctx, stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateComponentVersion(
	componentVersion *entity.ComponentVersion,
) (*entity.ComponentVersion, error) {
	return componentVersionObject.Create(s.db, componentVersion)
}

func (s *SqlDatabase) UpdateComponentVersion(componentVersion *entity.ComponentVersion) error {
	return componentVersionObject.Update(s.db, componentVersion)
}

func (s *SqlDatabase) DeleteComponentVersion(id int64, userId int64) error {
	return componentVersionObject.Delete(s.db, id, userId)
}
