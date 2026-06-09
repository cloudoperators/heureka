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
}

func (s *SqlDatabase) buildIssueVariantStatement(
	ctx context.Context,
	baseQuery sq.SelectBuilder,
	filter *entity.IssueVariantFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	statement := Statement[*entity.IssueVariantFilter]{
		Db:         s.db,
		L:          l,
		Obj:        &issueVariantObject,
		BaseQuery:  baseQuery,
		Order:      NewOrder(order, issueVariantObject.DefaultOrder),
		WithCursor: withCursor,
	}

	return BuildStatement(ctx, statement, filter)
}

func (s *SqlDatabase) GetAllIssueVariantCursors(
	ctx context.Context,
	filter *entity.IssueVariantFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllIssueVariantCursors",
	})

	baseQuery := sq.Select("IV.*").From("IssueVariant IV").GroupBy("IV.issuevariant_id")

	stmt, filterParameters, err := s.buildIssueVariantStatement(ctx, baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build IssueVariant cursor query: %w", err)
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
		func(l []IssueVariantRow, e IssueVariantRow) []IssueVariantRow {
			return append(l, e)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get IssueVariant cursors: %w", err)
	}

	return lo.Map(rows, func(row IssueVariantRow, _ int) string {
		iv := row.AsIssueVariant()

		cursor, _ := EncodeCursor(WithIssueVariant(order, iv))

		return cursor
	}), nil
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
