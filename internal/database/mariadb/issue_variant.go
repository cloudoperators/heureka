// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

var issueVariantObject = DbObject[*entity.IssueVariant, *entity.IssueVariantFilter, entity.IssueVariantResult]{
	Prefix:       "issuevariant",
	TableName:    "IssueVariant",
	TableKey:     "IV",
	DefaultOrder: entity.Order{By: entity.IssueVariantID, Direction: entity.OrderDirectionAsc},
	Properties: []*Property[*entity.IssueVariant]{
		NewProperty("issuevariant_issue_id", func(iv *entity.IssueVariant) (any, bool) { return iv.IssueId, iv.IssueId != 0 }),
		NewProperty("issuevariant_repository_id", func(iv *entity.IssueVariant) (any, bool) { return iv.IssueRepositoryId, iv.IssueRepositoryId != 0 }),
		// if rating but not vector is passed, we need to include the vector in the update in order
		// to overwrite any existing vector
		NewProperty("issuevariant_vector", func(iv *entity.IssueVariant) (any, bool) {
			return iv.Severity.Cvss.Vector, iv.Severity.Cvss.Vector != "" ||
				(iv.Severity.Value != "" && iv.Severity.Cvss.Vector == "")
		}),
		NewProperty("issuevariant_rating", func(iv *entity.IssueVariant) (any, bool) { return iv.Severity.Value, iv.Severity.Value != "" }),
		NewProperty("issuevariant_secondary_name", func(iv *entity.IssueVariant) (any, bool) { return iv.SecondaryName, iv.SecondaryName != "" }),
		NewProperty("issuevariant_description", func(iv *entity.IssueVariant) (any, bool) { return iv.Description, iv.Description != "" }),
		NewProperty("issuevariant_external_url", func(iv *entity.IssueVariant) (any, bool) { return iv.ExternalUrl, iv.ExternalUrl != "" }),
		NewProperty("issuevariant_created_by", func(iv *entity.IssueVariant) (any, bool) { return iv.CreatedBy, NoUpdate }),
		NewProperty("issuevariant_updated_by", func(iv *entity.IssueVariant) (any, bool) { return iv.UpdatedBy, iv.UpdatedBy != 0 }),
	},
	FilterProperties: []*FilterProperty[*entity.IssueVariantFilter]{
		NewFilterProperty("IV.issuevariant_id = ?", func(filter *entity.IssueVariantFilter) any { return filter.Id }),
		NewFilterProperty("IV.issuevariant_secondary_name = ?", func(filter *entity.IssueVariantFilter) any { return filter.SecondaryName }),
		NewFilterProperty("IV.issuevariant_issue_id = ?", func(filter *entity.IssueVariantFilter) any { return filter.IssueId }),
		NewFilterProperty("IV.issuevariant_repository_id = ?", func(filter *entity.IssueVariantFilter) any { return filter.IssueRepositoryId }),
		NewFilterProperty("IRS.issuerepositoryservice_service_id = ?", func(filter *entity.IssueVariantFilter) any { return filter.ServiceId }),
		NewFilterProperty("IM.issuematch_id = ?", func(filter *entity.IssueVariantFilter) any { return filter.IssueMatchId }),
		NewStateFilterProperty("IV.issuevariant", func(filter *entity.IssueVariantFilter) any { return filter.State }),
	},
	JoinDefs: []*JoinDef[*entity.IssueVariantFilter]{
		{
			Name:      "IR",
			Type:      InnerJoin,
			Table:     "IssueRepository IR",
			On:        "IV.issuevariant_repository_id = IR.issuerepository_id",
			Condition: DependentJoin[*entity.IssueVariantFilter],
		},
		{
			Name:      "IRS",
			Type:      InnerJoin,
			Table:     "IssueRepositoryService IRS",
			On:        "IR.issuerepository_id = IRS.issuerepositoryservice_issue_repository_id",
			DependsOn: []string{"IR"},
			Condition: func(f *entity.IssueVariantFilter, _ *Order) bool { return len(f.ServiceId) > 0 },
		},
		{
			Name:      "I",
			Type:      InnerJoin,
			Table:     "Issue I",
			On:        "IV.issuevariant_issue_id = I.issue_id",
			Condition: func(f *entity.IssueVariantFilter, _ *Order) bool { return len(f.IssueId) > 0 },
		},
		{
			Name:      "IM",
			Type:      InnerJoin,
			Table:     "IssueMatch IM",
			On:        "I.issue_id = IM.issuematch_issue_id",
			DependsOn: []string{"I"},
			Condition: func(f *entity.IssueVariantFilter, _ *Order) bool { return len(f.IssueMatchId) > 0 },
		},
	},
	GetItemAppender: func(l []entity.IssueVariantResult, e RowComposite, order []entity.Order) []entity.IssueVariantResult {
		iv := e.AsIssueVariant()
		cursor, _ := EncodeCursor(WithIssueVariant(order, iv))

		ivr := entity.IssueVariantResult{
			WithCursor: entity.WithCursor{
				Value: cursor,
			},
			IssueVariant: &iv,
		}

		return append(l, ivr)
	},
	GetAllCursorItemAppender: func(l []string, e RowComposite, order []entity.Order) []string {
		iv := e.AsIssueVariant()

		cursor, _ := EncodeCursor(WithIssueVariant(order, iv))

		return append(l, cursor)
	},
}

func (s *SqlDatabase) GetAllIssueVariantCursors(
	ctx context.Context,
	filter *entity.IssueVariantFilter,
	order []entity.Order,
) ([]string, error) {
	return issueVariantObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetIssueVariants(
	ctx context.Context,
	filter *entity.IssueVariantFilter,
	order []entity.Order,
) ([]entity.IssueVariantResult, error) {
	return issueVariantObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountIssueVariants(ctx context.Context, filter *entity.IssueVariantFilter) (int64, error) {
	return issueVariantObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) CreateIssueVariant(
	issueVariant *entity.IssueVariant,
) (*entity.IssueVariant, error) {
	return issueVariantObject.Create(s.db, issueVariant)
}

func (s *SqlDatabase) UpdateIssueVariant(issueVariant *entity.IssueVariant) error {
	return issueVariantObject.Update(s.db, issueVariant)
}

func (s *SqlDatabase) DeleteIssueVariant(id int64, userId int64) error {
	return issueVariantObject.Delete(s.db, id, userId)
}
