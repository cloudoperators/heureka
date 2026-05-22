// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var serviceIssueVariantObject = DbObject[*entity.ServiceIssueVariant]{
	Properties: []*Property{},
	FilterProperties: []*FilterProperty{
		NewFilterProperty(
			"CI.componentinstance_id = ?",
			WrapRetSlice(
				func(filter *entity.ServiceIssueVariantFilter) []*int64 { return filter.ComponentInstanceId },
			),
		),
		NewFilterProperty(
			"I.issue_id = ?",
			WrapRetSlice(
				func(filter *entity.ServiceIssueVariantFilter) []*int64 { return filter.IssueId },
			),
		),
		NewStateFilterProperty(
			"IV.issuevariant",
			WrapRetState(
				func(filter *entity.ServiceIssueVariantFilter) []entity.StateFilterType { return filter.State },
			),
		),
	},
	JoinDefs: []*JoinDef{
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
}

func (s *SqlDatabase) buildServiceIssueVariantStatement(
	ctx context.Context,
	baseQuery sq.SelectBuilder,
	filter *entity.ServiceIssueVariantFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	statement := Statement{
		Db:         s.db,
		L:          l,
		Obj:        &serviceIssueVariantObject,
		BaseQuery:  baseQuery,
		Order:      NewOrder(order, entity.Order{By: entity.ServiceIssueVariantID, Direction: entity.OrderDirectionAsc}),
		WithCursor: withCursor,
		Aggregated: false,
	}

	return BuildStatement(ctx, statement, filter)
}

func (s *SqlDatabase) GetServiceIssueVariants(
	ctx context.Context,
	filter *entity.ServiceIssueVariantFilter,
	order []entity.Order,
) ([]entity.ServiceIssueVariantResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueVariants",
	})

	baseQuery := sq.Select("IRS.issuerepositoryservice_priority", "IV.*").From("ComponentInstance CI")

	stmt, filterParameters, err := s.buildServiceIssueVariantStatement(
		ctx,
		baseQuery,
		filter,
		true,
		order,
		l,
	)
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
		func(l []entity.ServiceIssueVariantResult, e ServiceIssueVariantRow) []entity.ServiceIssueVariantResult {
			r := e.AsServiceIssueVariantEntry()
			cursor, _ := EncodeCursor(WithServiceIssueVariant(order, r))

			rr := entity.ServiceIssueVariantResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				ServiceIssueVariant: &r,
			}

			return append(l, rr)
		},
	)
}
