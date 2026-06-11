// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

var issueMatchObject = DbObject[*entity.IssueMatch, *entity.IssueMatchFilter, entity.IssueMatchResult]{
	Prefix:       "issuematch",
	TableName:    "IssueMatch",
	TableKey:     "IM",
	DefaultOrder: entity.Order{By: entity.IssueMatchId, Direction: entity.OrderDirectionAsc},
	Properties: []*Property[*entity.IssueMatch]{
		NewProperty("issuematch_status", func(im *entity.IssueMatch) (any, bool) {
			return im.Status, im.Status != "" && im.Status != entity.IssueMatchStatusValuesNone
		}),
		NewProperty("issuematch_remediation_date", func(im *entity.IssueMatch) (any, bool) { return im.RemediationDate, !im.RemediationDate.IsZero() }),
		NewProperty("issuematch_target_remediation_date", func(im *entity.IssueMatch) (any, bool) {
			return im.TargetRemediationDate, !im.TargetRemediationDate.IsZero()
		}),
		NewProperty("issuematch_vector", func(im *entity.IssueMatch) (any, bool) { return im.Severity.Cvss.Vector, im.Severity.Cvss.Vector != "" }),
		NewProperty("issuematch_rating", func(im *entity.IssueMatch) (any, bool) { return im.Severity.Value, im.Severity.Value != "" }),
		NewProperty("issuematch_user_id", func(im *entity.IssueMatch) (any, bool) { return im.UserId, im.UserId != 0 }),
		NewProperty("issuematch_component_instance_id", func(im *entity.IssueMatch) (any, bool) { return im.ComponentInstanceId, im.ComponentInstanceId != 0 }),
		NewProperty("issuematch_issue_id", func(im *entity.IssueMatch) (any, bool) { return im.IssueId, im.IssueId != 0 }),
		NewProperty("issuematch_created_by", func(im *entity.IssueMatch) (any, bool) { return im.CreatedBy, NoUpdate }),
		NewProperty("issuematch_updated_by", func(im *entity.IssueMatch) (any, bool) { return im.UpdatedBy, im.UpdatedBy != 0 }),
	},
	FilterProperties: []*FilterProperty[*entity.IssueMatchFilter]{
		NewFilterProperty("IM.issuematch_id = ?", func(filter *entity.IssueMatchFilter) any { return filter.Id }),
		NewFilterProperty("IM.issuematch_issue_id = ?", func(filter *entity.IssueMatchFilter) any { return filter.IssueId }),
		NewFilterProperty("IM.issuematch_component_instance_id = ?", func(filter *entity.IssueMatchFilter) any { return filter.ComponentInstanceId }),
		NewFilterProperty("S.service_ccrn = ?", func(filter *entity.IssueMatchFilter) any { return filter.ServiceCCRN }),
		NewFilterProperty("CI.componentinstance_service_id = ?", func(filter *entity.IssueMatchFilter) any { return filter.ServiceId }),
		NewFilterProperty("IM.issuematch_rating = ?", func(filter *entity.IssueMatchFilter) any { return filter.SeverityValue }),
		NewFilterProperty("IM.issuematch_status = ?", func(filter *entity.IssueMatchFilter) any { return filter.Status }),
		NewFilterProperty("SG.supportgroup_ccrn = ?", func(filter *entity.IssueMatchFilter) any { return filter.SupportGroupCCRN }),
		NewFilterProperty("I.issue_primary_name = ?", func(filter *entity.IssueMatchFilter) any { return filter.PrimaryName }),
		NewFilterProperty("C.component_ccrn = ?", func(filter *entity.IssueMatchFilter) any { return filter.ComponentCCRN }),
		NewFilterProperty("I.issue_type = ?", func(filter *entity.IssueMatchFilter) any { return filter.IssueType }),
		NewFilterProperty("U.user_name = ?", func(filter *entity.IssueMatchFilter) any { return filter.ServiceOwnerUsername }),
		NewFilterProperty("U.user_unique_user_id = ?", func(filter *entity.IssueMatchFilter) any { return filter.ServiceOwnerUniqueUserId }),
		NewNFilterProperty(
			"IV.issuevariant_secondary_name LIKE Concat('%',?,'%') OR I.issue_primary_name LIKE Concat('%',?,'%')",
			func(filter *entity.IssueMatchFilter) any { return filter.Search },
			2,
		),
		NewStateFilterProperty("IM.issuematch", func(filter *entity.IssueMatchFilter) any { return filter.State }),
	},
	JoinDefs: []*JoinDef[*entity.IssueMatchFilter]{
		{
			Name:  "I",
			Type:  LeftJoin,
			Table: "Issue I",
			On:    "IM.issuematch_issue_id = I.issue_id",
			Condition: func(f *entity.IssueMatchFilter, order *Order) bool {
				return len(f.IssueType) > 0 || len(f.PrimaryName) > 0 || order.ByIssuePrimaryName()
			},
		},
		{
			Name:      "IV",
			Type:      LeftJoin,
			Table:     "IssueVariant IV",
			On:        "I.issue_id = IV.issuevariant_issue_id",
			DependsOn: []string{"I"},
			Condition: func(f *entity.IssueMatchFilter, _ *Order) bool { return len(f.Search) > 0 },
		},
		{
			Name:      "CI",
			Type:      LeftJoin,
			Table:     "ComponentInstance CI",
			On:        "IM.issuematch_component_instance_id = CI.componentinstance_id",
			Condition: func(f *entity.IssueMatchFilter, order *Order) bool { return order.ByCiCcrn() || len(f.ServiceId) > 0 },
		},
		{
			Name:      "CV",
			Type:      LeftJoin,
			Table:     "ComponentVersion CV",
			On:        "CI.componentinstance_component_version_id = CV.componentversion_id",
			DependsOn: []string{"CI"},
			Condition: DependentJoin[*entity.IssueMatchFilter],
		},
		{
			Name:      "C",
			Type:      LeftJoin,
			Table:     "Component C",
			On:        "CV.componentversion_component_id = C.component_id",
			DependsOn: []string{"CV"},
			Condition: func(f *entity.IssueMatchFilter, _ *Order) bool { return len(f.ComponentCCRN) > 0 },
		},
		{
			Name:      "S",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			DependsOn: []string{"CI"},
			Condition: func(f *entity.IssueMatchFilter, _ *Order) bool { return len(f.ServiceCCRN) > 0 },
		},
		{
			Name:      "SGS",
			Type:      LeftJoin,
			Table:     "SupportGroupService SGS",
			On:        "S.service_id = SGS.supportgroupservice_service_id",
			DependsOn: []string{"S"},
			Condition: DependentJoin[*entity.IssueMatchFilter],
		},
		{
			Name:      "SG",
			Type:      LeftJoin,
			Table:     "SupportGroup SG",
			On:        "SGS.supportgroupservice_support_group_id = SG.supportgroup_id",
			DependsOn: []string{"SGS"},
			Condition: func(f *entity.IssueMatchFilter, _ *Order) bool { return len(f.SupportGroupCCRN) > 0 },
		},
		{
			Name:      "O",
			Type:      LeftJoin,
			Table:     "Owner O",
			On:        "CI.componentinstance_service_id = O.owner_service_id",
			DependsOn: []string{"CI"},
			Condition: DependentJoin[*entity.IssueMatchFilter],
		},
		{
			Name:      "U",
			Type:      LeftJoin,
			Table:     "User U",
			On:        "O.owner_user_id = U.user_id",
			DependsOn: []string{"O"},
			Condition: func(f *entity.IssueMatchFilter, _ *Order) bool {
				return len(f.ServiceOwnerUsername) > 0 || len(f.ServiceOwnerUniqueUserId) > 0
			},
		},
	},
	ExtraColumnsSelector: func(_ *entity.IssueMatchFilter, order *Order) []string {
		s := []string{}
		for _, o := range order.Sequence() {
			switch o.By {
			case entity.IssuePrimaryName:
				s = append(s, "I.issue_primary_name")
			case entity.ComponentInstanceCcrn:
				s = append(s, "CI.componentinstance_ccrn")
			}
		}
		return s
	},
	GetItemAppender: func(l []entity.IssueMatchResult, e RowComposite, order []entity.Order) []entity.IssueMatchResult {
		im := e.AsIssueMatch()
		if e.IssueRow != nil {
			im.Issue = new(e.IssueRow.AsIssue())
		}

		if e.ComponentInstanceRow != nil {
			im.ComponentInstance = new(e.AsComponentInstance())
		}

		cursor, _ := EncodeCursor(WithIssueMatch(order, im))

		imr := entity.IssueMatchResult{
			WithCursor: entity.WithCursor{
				Value: cursor,
			},
			IssueMatch: &im,
		}

		return append(l, imr)
	},
	GetAllCursorItemAppender: func(l []string, e RowComposite, order []entity.Order) []string {
		im := e.AsIssueMatch()
		if e.IssueRow != nil {
			im.Issue = new(e.IssueRow.AsIssue())
		}

		if e.ComponentInstanceRow != nil {
			im.ComponentInstance = new(e.AsComponentInstance())
		}

		cursor, _ := EncodeCursor(WithIssueMatch(order, im))

		return append(l, cursor)
	},
}

func (s *SqlDatabase) GetAllIssueMatchIds(ctx context.Context, filter *entity.IssueMatchFilter) ([]int64, error) {
	return issueMatchObject.GetAllIds(ctx, s.db, filter)
}

func (s *SqlDatabase) GetAllIssueMatchCursors(ctx context.Context, filter *entity.IssueMatchFilter, order []entity.Order) ([]string, error) {
	return issueMatchObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetIssueMatches(ctx context.Context, filter *entity.IssueMatchFilter, order []entity.Order) ([]entity.IssueMatchResult, error) {
	return issueMatchObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountIssueMatches(ctx context.Context, filter *entity.IssueMatchFilter) (int64, error) {
	return issueMatchObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) CreateIssueMatch(issueMatch *entity.IssueMatch) (*entity.IssueMatch, error) {
	return issueMatchObject.Create(s.db, issueMatch)
}

func (s *SqlDatabase) UpdateIssueMatch(issueMatch *entity.IssueMatch) error {
	return issueMatchObject.Update(s.db, issueMatch)
}

func (s *SqlDatabase) DeleteIssueMatch(id int64, userId int64) error {
	return issueMatchObject.Delete(s.db, id, userId)
}
