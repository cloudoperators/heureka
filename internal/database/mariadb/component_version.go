// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

var componentVersionObject = DbObject[*entity.ComponentVersion, *entity.ComponentVersionFilter, entity.ComponentVersionResult]{
	Prefix:       "componentversion",
	TableName:    "ComponentVersion",
	TableKey:     "CV",
	DefaultOrder: entity.Order{By: entity.ComponentVersionId, Direction: entity.OrderDirectionAsc},
	Aggregated:   true,
	Properties: []*Property[*entity.ComponentVersion]{
		NewProperty("componentversion_component_id", func(cv *entity.ComponentVersion) (any, bool) { return cv.ComponentId, cv.ComponentId != 0 }),
		NewProperty("componentversion_version", func(cv *entity.ComponentVersion) (any, bool) { return cv.Version, cv.Version != "" }),
		NewProperty("componentversion_tag", func(cv *entity.ComponentVersion) (any, bool) { return cv.Tag, cv.Tag != "" }),
		NewProperty("componentversion_repository", func(cv *entity.ComponentVersion) (any, bool) { return cv.Repository, cv.Repository != "" }),
		NewProperty("componentversion_organization", func(cv *entity.ComponentVersion) (any, bool) { return cv.Organization, cv.Organization != "" }),
		NewProperty("componentversion_created_by", func(cv *entity.ComponentVersion) (any, bool) { return cv.CreatedBy, NoUpdate }),
		NewProperty("componentversion_updated_by", func(cv *entity.ComponentVersion) (any, bool) { return cv.UpdatedBy, cv.UpdatedBy != 0 }),
		NewProperty("componentversion_end_of_life", func(cv *entity.ComponentVersion) (any, bool) {
			return ValueOrDefault(cv.EndOfLife, false), cv.EndOfLife != nil
		}),
	},
	FilterProperties: []*FilterProperty[*entity.ComponentVersionFilter]{
		NewFilterProperty("CV.componentversion_id = ?", func(filter *entity.ComponentVersionFilter) any { return filter.Id }),
		NewFilterProperty("CVI.componentversionissue_issue_id = ?", func(filter *entity.ComponentVersionFilter) any { return filter.IssueId }),
		NewFilterProperty("CV.componentversion_component_id = ?", func(filter *entity.ComponentVersionFilter) any { return filter.ComponentId }),
		NewFilterProperty("CV.componentversion_version = ?", func(filter *entity.ComponentVersionFilter) any { return filter.Version }),
		NewFilterProperty("CV.componentversion_tag = ?", func(filter *entity.ComponentVersionFilter) any { return filter.Tag }),
		NewFilterProperty("CV.componentversion_repository = ?", func(filter *entity.ComponentVersionFilter) any { return filter.Repository }),
		NewFilterProperty("CV.componentversion_organization = ?", func(filter *entity.ComponentVersionFilter) any { return filter.Organization }),
		NewFilterProperty("C.component_ccrn = ?", func(filter *entity.ComponentVersionFilter) any { return filter.ComponentCCRN }),
		NewFilterProperty("S.service_ccrn = ?", func(filter *entity.ComponentVersionFilter) any { return filter.ServiceCCRN }),
		NewFilterProperty("CI.componentinstance_service_id = ?", func(filter *entity.ComponentVersionFilter) any { return filter.ServiceId }),
		NewFilterProperty("IV.issuevariant_repository_id = ?", func(filter *entity.ComponentVersionFilter) any { return filter.IssueRepositoryId }),
		NewFilterProperty("CV.componentversion_end_of_life = ?", func(filter *entity.ComponentVersionFilter) any { return filter.EndOfLife }),
		NewStateFilterProperty("CV.componentversion", func(filter *entity.ComponentVersionFilter) any { return filter.State }),
	},
	JoinDefs: []*JoinDef[*entity.ComponentVersionFilter]{
		{
			Name:      "CVI",
			Type:      LeftJoin,
			Table:     "ComponentVersionIssue CVI",
			On:        "CV.componentversion_id = CVI.componentversionissue_component_version_id",
			Condition: func(f *entity.ComponentVersionFilter, _ *Order) bool { return len(f.IssueId) > 0 },
		},
		{
			Name:      "IV",
			Type:      LeftJoin,
			Table:     "IssueVariant IV",
			On:        "CVI.componentversionissue_issue_id = IV.issuevariant_issue_id",
			DependsOn: []string{"CVI"},
			Condition: func(f *entity.ComponentVersionFilter, order *Order) bool {
				return order.ByCount() || len(f.IssueRepositoryId) > 0
			},
		},
		{
			Name:      "C",
			Type:      LeftJoin,
			Table:     "Component C",
			On:        "CV.componentversion_component_id = C.component_id",
			Condition: func(f *entity.ComponentVersionFilter, _ *Order) bool { return len(f.ComponentCCRN) > 0 },
		},
		{
			Name:      "CI",
			Type:      LeftJoin,
			Table:     "ComponentInstance CI",
			On:        "CV.componentversion_id = CI.componentinstance_component_version_id",
			Condition: func(f *entity.ComponentVersionFilter, _ *Order) bool { return len(f.ServiceId) > 0 },
		},
		{
			Name:      "S",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			DependsOn: []string{"CI"},
			Condition: func(f *entity.ComponentVersionFilter, _ *Order) bool { return len(f.ServiceCCRN) > 0 },
		},
	},
	ExtraColumnsSelector: func(_ *entity.ComponentVersionFilter, order *Order) []string {
		s := []string{}
		for _, o := range order.Sequence() {
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
	},
	RowToData: func(e RowComposite, order []entity.Order) (*entity.ComponentVersion, string) {
		cv := e.AsComponentVersion()

		var isc entity.IssueSeverityCounts
		if e.RatingCount != nil {
			isc = e.AsIssueSeverityCounts()
		}

		cursor, _ := EncodeCursor(WithComponentVersion(order, cv, isc))

		return &cv, cursor
	},
	NewResult: func(cv *entity.ComponentVersion, cursor string) entity.ComponentVersionResult {
		return entity.ComponentVersionResult{
			WithCursor:       entity.WithCursor{Value: cursor},
			ComponentVersion: cv,
		}
	},
}

func (s *SqlDatabase) GetAllComponentVersionCursors(
	ctx context.Context,
	filter *entity.ComponentVersionFilter,
	order []entity.Order,
) ([]string, error) {
	return componentVersionObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetComponentVersions(
	ctx context.Context,
	filter *entity.ComponentVersionFilter,
	order []entity.Order,
) ([]entity.ComponentVersionResult, error) {
	return componentVersionObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountComponentVersions(ctx context.Context, filter *entity.ComponentVersionFilter) (int64, error) {
	return componentVersionObject.Count(ctx, s.db, filter)
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
