// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

var serviceIssueVariantObject = DbObject[*entity.ServiceIssueVariant, *entity.ServiceIssueVariantFilter, entity.ServiceIssueVariantResult]{
	DefaultOrder: entity.Order{By: entity.ServiceIssueVariantID, Direction: entity.OrderDirectionAsc},
	Properties:   []*Property[*entity.ServiceIssueVariant]{},
	FilterProperties: []*FilterProperty[*entity.ServiceIssueVariantFilter]{
		NewFilterProperty("CI.componentinstance_id = ?", func(filter *entity.ServiceIssueVariantFilter) any { return filter.ComponentInstanceId }),
		NewFilterProperty("I.issue_id = ?", func(filter *entity.ServiceIssueVariantFilter) any { return filter.IssueId }),
		NewStateFilterProperty("IV.issuevariant", func(filter *entity.ServiceIssueVariantFilter) any { return filter.State }),
	},
	JoinDefs: []*JoinDef[*entity.ServiceIssueVariantFilter]{
		{
			Name:  "CV",
			Type:  InnerJoin,
			Table: "ComponentVersion CV",
			On:    "CI.componentinstance_component_version_id = CV.componentversion_id",
		},
		{
			Name:      "CVI",
			Type:      InnerJoin,
			Table:     "ComponentVersionIssue CVI",
			On:        "CV.componentversion_id = CVI.componentversionissue_component_version_id",
			DependsOn: []string{"CV"},
		},
		{
			Name:      "I",
			Type:      InnerJoin,
			Table:     "Issue I",
			On:        "CVI.componentversionissue_issue_id = I.issue_id",
			DependsOn: []string{"CVI"},
		},
		{
			Name:  "S",
			Type:  InnerJoin,
			Table: "Service S",
			On:    "CI.componentinstance_service_id = S.service_id",
		},
		{
			Name:      "IRS",
			Type:      InnerJoin,
			Table:     "IssueRepositoryService IRS",
			On:        "S.service_id = IRS.issuerepositoryservice_service_id",
			DependsOn: []string{"S"},
		},
		{
			Name:      "IR",
			Type:      InnerJoin,
			Table:     "IssueRepository IR",
			On:        "IRS.issuerepositoryservice_issue_repository_id = IR.issuerepository_id",
			DependsOn: []string{"IRS"},
		}, // S, IRS, IR - Join path to Repository
		{
			Name:      "IV",
			Type:      InnerJoin,
			Table:     "IssueVariant IV",
			On:        "I.issue_id = IV.issuevariant_issue_id and IR.issuerepository_id = IV.issuevariant_repository_id",
			DependsOn: []string{"I", "IR"},
		}, // Join to from repo and issue to IssueVariant
	},
	ExtraColumnsSelector: func(_ *entity.ServiceIssueVariantFilter, _ *Order) []string {
		return []string{"IRS.issuerepositoryservice_priority", "IV.*"}
	},
	ForceFrom: "ComponentInstance CI",
	GetItemAppender: func(l []entity.ServiceIssueVariantResult, e RowComposite, order []entity.Order) []entity.ServiceIssueVariantResult {
		r := ServiceIssueVariantRow{
			IssueVariantRow: *e.IssueVariantRow,
			IssueRepositoryRow: IssueRepositoryRow{
				IssueRepositoryServiceRow: *e.IssueRepositoryServiceRow,
			},
		}.AsServiceIssueVariantEntry()

		cursor, _ := EncodeCursor(WithServiceIssueVariant(order, r))

		rr := entity.ServiceIssueVariantResult{
			WithCursor: entity.WithCursor{
				Value: cursor,
			},
			ServiceIssueVariant: &r,
		}

		return append(l, rr)
	},
}

func (s *SqlDatabase) GetServiceIssueVariants(
	ctx context.Context,
	filter *entity.ServiceIssueVariantFilter,
	order []entity.Order,
) ([]entity.ServiceIssueVariantResult, error) {
	return serviceIssueVariantObject.Get(ctx, s.db, filter, order)
}
