// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

var issueRepositoryObject = DbObject[*entity.IssueRepository, *entity.IssueRepositoryFilter, entity.IssueRepositoryResult, *any]{
	Prefix:       "issuerepository",
	TableName:    "IssueRepository",
	TableKey:     "IR",
	DefaultOrder: entity.Order{By: entity.IssueRepositoryID, Direction: entity.OrderDirectionAsc},
	Properties: []*Property[*entity.IssueRepository]{
		NewProperty("issuerepository_name", func(ir *entity.IssueRepository) (any, bool) { return ir.Name, ir.Name != "" }),
		NewProperty("issuerepository_url", func(ir *entity.IssueRepository) (any, bool) { return ir.Url, ir.Url != "" }),
		NewProperty("issuerepository_created_by", func(ir *entity.IssueRepository) (any, bool) { return ir.BaseIssueRepository.CreatedBy, NoUpdate }),
		NewProperty("issuerepository_updated_by", func(ir *entity.IssueRepository) (any, bool) {
			return ir.BaseIssueRepository.UpdatedBy, ir.BaseIssueRepository.UpdatedBy != 0
		}),
	},
	FilterProperties: []*FilterProperty[*entity.IssueRepositoryFilter]{
		NewFilterProperty("IR.issuerepository_name = ?", func(filter *entity.IssueRepositoryFilter) any { return filter.Name }),
		NewFilterProperty("IR.issuerepository_id = ?", func(filter *entity.IssueRepositoryFilter) any { return filter.Id }),
		NewFilterProperty("S.service_ccrn = ?", func(filter *entity.IssueRepositoryFilter) any { return filter.ServiceCCRN }),
		NewFilterProperty("IRS.issuerepositoryservice_service_id = ?", func(filter *entity.IssueRepositoryFilter) any { return filter.ServiceId }),
		NewStateFilterProperty("IR.issuerepository", func(filter *entity.IssueRepositoryFilter) any { return filter.State }),
	},
	JoinDefs: []*JoinDef[*entity.IssueRepositoryFilter]{
		{
			Name:      "IRS",
			Type:      LeftJoin,
			Table:     "IssueRepositoryService IRS",
			On:        "IR.issuerepository_id = IRS.issuerepositoryservice_issue_repository_id",
			Condition: func(f *entity.IssueRepositoryFilter, _ *Order) bool { return len(f.ServiceId) > 0 },
		},
		{
			Name:      "S",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "IRS.issuerepositoryservice_service_id = S.service_id",
			DependsOn: []string{"IRS"},
			Condition: func(f *entity.IssueRepositoryFilter, _ *Order) bool { return len(f.ServiceCCRN) > 0 },
		},
	},
	RowToData: func(e RowComposite, order []entity.Order) (*entity.IssueRepository, string) {
		ir := e.BaseIssueRepositoryRow.AsIssueRepository()

		cursor, _ := EncodeCursor(WithIssueRepository(order, ir))

		return &ir, cursor
	},
	NewResult: func(ir *entity.IssueRepository, _ *any, cursor string) entity.IssueRepositoryResult {
		return entity.IssueRepositoryResult{
			WithCursor:      entity.WithCursor{Value: cursor},
			IssueRepository: ir,
		}
	},
}

func (s *SqlDatabase) CountIssueRepositories(ctx context.Context, filter *entity.IssueRepositoryFilter) (int64, error) {
	return issueRepositoryObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) GetAllIssueRepositoryCursors(
	ctx context.Context,
	filter *entity.IssueRepositoryFilter,
	order []entity.Order,
) ([]string, error) {
	return issueRepositoryObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetIssueRepositories(
	ctx context.Context,
	filter *entity.IssueRepositoryFilter,
	order []entity.Order,
) ([]entity.IssueRepositoryResult, error) {
	return issueRepositoryObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CreateIssueRepository(
	issueRepository *entity.IssueRepository,
) (*entity.IssueRepository, error) {
	return issueRepositoryObject.Create(s.db, issueRepository)
}

func (s *SqlDatabase) UpdateIssueRepository(issueRepository *entity.IssueRepository) error {
	return issueRepositoryObject.Update(s.db, issueRepository)
}

func (s *SqlDatabase) DeleteIssueRepository(id int64, userId int64) error {
	return issueRepositoryObject.Delete(s.db, id, userId)
}
