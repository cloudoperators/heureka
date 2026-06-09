// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var issueRepositoryObject = DbObject[*entity.IssueRepository, *entity.IssueRepositoryFilter, entity.IssueRepositoryResult]{
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
	GetItemAppender: func(l []entity.IssueRepositoryResult, e RowComposite, order []entity.Order) []entity.IssueRepositoryResult {
		ir := e.BaseIssueRepositoryRow.AsIssueRepository()
		cursor, _ := EncodeCursor(WithIssueRepository(order, ir))

		irr := entity.IssueRepositoryResult{
			WithCursor: entity.WithCursor{
				Value: cursor,
			},
			IssueRepository: &ir,
		}

		return append(l, irr)
	},
}

func (s *SqlDatabase) CountIssueRepositories(ctx context.Context, filter *entity.IssueRepositoryFilter) (int64, error) {
	return issueRepositoryObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) buildIssueRepositoryStatement(
	ctx context.Context,
	baseQuery sq.SelectBuilder,
	filter *entity.IssueRepositoryFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	statement := Statement[*entity.IssueRepositoryFilter]{
		Db:         s.db,
		L:          l,
		Obj:        &issueRepositoryObject,
		BaseQuery:  baseQuery,
		Order:      NewOrder(order, issueRepositoryObject.DefaultOrder),
		WithCursor: withCursor,
	}

	return BuildStatement(ctx, statement, filter)
}

func (s *SqlDatabase) GetAllIssueRepositoryCursors(
	ctx context.Context,
	filter *entity.IssueRepositoryFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllIssueRepositoryCursors",
	})

	baseQuery := sq.Select("IR.*").From("IssueRepository IR").GroupBy("IR.issuerepository_id")

	stmt, filterParameters, err := s.buildIssueRepositoryStatement(ctx, baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build IssueRepository cursor query: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error while close statement: %s", err.Error())
		}
	}()

	rows, err := performListScan(
		ctx,
		stmt,
		filterParameters,
		l,
		func(l []IssueRepositoryRow, e IssueRepositoryRow) []IssueRepositoryRow {
			return append(l, e)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get IssueRepository cursors: %w", err)
	}

	return lo.Map(rows, func(row IssueRepositoryRow, _ int) string {
		ir := row.AsIssueRepository()

		cursor, _ := EncodeCursor(WithIssueRepository(order, ir))

		return cursor
	}), nil
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
