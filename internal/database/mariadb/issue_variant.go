// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var issueVariantObject = DbObject[*entity.IssueVariant]{
	Prefix:    "issuevariant",
	TableName: "IssueVariant",
	Properties: []*Property{
		NewProperty(
			"issuevariant_issue_id",
			WrapAccess(
				func(iv *entity.IssueVariant) (int64, bool) { return iv.IssueId, iv.IssueId != 0 },
			),
		),
		NewProperty(
			"issuevariant_repository_id",
			WrapAccess(
				func(iv *entity.IssueVariant) (int64, bool) { return iv.IssueRepositoryId, iv.IssueRepositoryId != 0 },
			),
		),
		// if rating but not vector is passed, we need to include the vector in the update in order
		// to overwrite any existing vector
		NewProperty("issuevariant_vector", WrapAccess(func(iv *entity.IssueVariant) (string, bool) {
			return iv.Severity.Cvss.Vector, iv.Severity.Cvss.Vector != "" ||
				(iv.Severity.Value != "" && iv.Severity.Cvss.Vector == "")
		})),
		NewProperty(
			"issuevariant_rating",
			WrapAccess(
				func(iv *entity.IssueVariant) (string, bool) { return iv.Severity.Value, iv.Severity.Value != "" },
			),
		),
		NewProperty(
			"issuevariant_secondary_name",
			WrapAccess(
				func(iv *entity.IssueVariant) (string, bool) { return iv.SecondaryName, iv.SecondaryName != "" },
			),
		),
		NewProperty(
			"issuevariant_description",
			WrapAccess(
				func(iv *entity.IssueVariant) (string, bool) { return iv.Description, iv.Description != "" },
			),
		),
		NewProperty(
			"issuevariant_external_url",
			WrapAccess(
				func(iv *entity.IssueVariant) (string, bool) { return iv.ExternalUrl, iv.ExternalUrl != "" },
			),
		),
		NewProperty(
			"issuevariant_created_by",
			WrapAccess(
				func(iv *entity.IssueVariant) (int64, bool) { return iv.CreatedBy, NoUpdate },
			),
		),
		NewProperty(
			"issuevariant_updated_by",
			WrapAccess(
				func(iv *entity.IssueVariant) (int64, bool) { return iv.UpdatedBy, iv.UpdatedBy != 0 },
			),
		),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty(
			"IV.issuevariant_id = ?",
			WrapRetSlice(func(filter *entity.IssueVariantFilter) []*int64 { return filter.Id }),
		),
		NewFilterProperty(
			"IV.issuevariant_secondary_name = ?",
			WrapRetSlice(
				func(filter *entity.IssueVariantFilter) []*string { return filter.SecondaryName },
			),
		),
		NewFilterProperty(
			"IV.issuevariant_issue_id = ?",
			WrapRetSlice(
				func(filter *entity.IssueVariantFilter) []*int64 { return filter.IssueId },
			),
		),
		NewFilterProperty(
			"IV.issuevariant_repository_id = ?",
			WrapRetSlice(
				func(filter *entity.IssueVariantFilter) []*int64 { return filter.IssueRepositoryId },
			),
		),
		NewFilterProperty(
			"IRS.issuerepositoryservice_service_id = ?",
			WrapRetSlice(
				func(filter *entity.IssueVariantFilter) []*int64 { return filter.ServiceId },
			),
		),
		NewFilterProperty(
			"IM.issuematch_id = ?",
			WrapRetSlice(
				func(filter *entity.IssueVariantFilter) []*int64 { return filter.IssueMatchId },
			),
		),
		NewStateFilterProperty(
			"IV.issuevariant",
			WrapRetState(
				func(filter *entity.IssueVariantFilter) []entity.StateFilterType { return filter.State },
			),
		),
	},
	JoinDefs: []*JoinDef{
		{
			Name:      "IR",
			Type:      InnerJoin,
			Table:     "IssueRepository IR",
			On:        "IV.issuevariant_repository_id = IR.issuerepository_id",
			Condition: DependentJoin,
		},
		{
			Name:      "IRS",
			Type:      InnerJoin,
			Table:     "IssueRepositoryService IRS",
			On:        "IR.issuerepository_id = IRS.issuerepositoryservice_issue_repository_id",
			DependsOn: []string{"IR"},
			Condition: WrapJoinCondition(func(f *entity.IssueVariantFilter, _ *Order) bool {
				return len(f.ServiceId) > 0
			}),
		},
		{
			Name:  "I",
			Type:  InnerJoin,
			Table: "Issue I",
			On:    "IV.issuevariant_issue_id = I.issue_id",
			Condition: WrapJoinCondition(func(f *entity.IssueVariantFilter, _ *Order) bool {
				return len(f.IssueId) > 0
			}),
		},
		{
			Name:      "IM",
			Type:      InnerJoin,
			Table:     "IssueMatch IM",
			On:        "I.issue_id = IM.issuematch_issue_id",
			DependsOn: []string{"I"},
			Condition: WrapJoinCondition(func(f *entity.IssueVariantFilter, _ *Order) bool {
				return len(f.IssueMatchId) > 0
			}),
		},
	},
}

func ensureIssueVariantFilter(filter *entity.IssueVariantFilter) *entity.IssueVariantFilter {
	if filter == nil {
		filter = &entity.IssueVariantFilter{}
	}

	return EnsurePagination(filter)
}

func (s *SqlDatabase) buildIssueVariantStatement(
	baseQuery string,
	filter *entity.IssueVariantFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	filter = ensureIssueVariantFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	cursorFields, err := DecodeCursor(filter.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode IssueVariant cursor: %w", err)
	}

	ord := NewOrder(order, entity.Order{By: entity.IssueVariantID, Direction: entity.OrderDirectionAsc})
	joins := issueVariantObject.GetJoins(filter, ord)
	whereClause, hasFilter := issueVariantObject.GetFilterWhereClause(filter, withCursor)
	cursorQuery := issueVariantObject.GetCursorQuery(&hasFilter, cursorFields, &withCursor, false)

	var query string
	if withCursor {
		query = fmt.Sprintf(baseQuery, joins, whereClause, cursorQuery, ord)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause, ord)
	}

	stmt, err := s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			}).Error(msg)

		return nil, nil, fmt.Errorf("failed to prepare IssueVariant statement: %w", err)
	}

	filterParameters := issueVariantObject.GetFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllIssueVariantCursors(
	filter *entity.IssueVariantFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllIssueVariantCursors",
	})

	baseQuery := `
		SELECT IV.* FROM  IssueVariant IV 
		%s
		%s GROUP BY IV.issuevariant_id ORDER BY %s
	`

	filter = ensureIssueVariantFilter(filter)

	stmt, filterParameters, err := s.buildIssueVariantStatement(baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build IssueVariant cursor query: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error while close statement: %s", err.Error())
		}
	}()

	rows, err := performListScan(
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
	filter *entity.IssueVariantFilter,
	order []entity.Order,
) ([]entity.IssueVariantResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueVariants",
	})

	baseQuery := `
		SELECT IV.* FROM  IssueVariant IV 
		%s
		%s
		%s ORDER BY %s LIMIT ?
    `

	stmt, filterParameters, err := s.buildIssueVariantStatement(baseQuery, filter, true, order, l)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error while close statement: %s", err.Error())
		}
	}()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueVariantResult, e IssueVariantRow) []entity.IssueVariantResult {
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
	)
}

func (s *SqlDatabase) CountIssueVariants(filter *entity.IssueVariantFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssueVariants",
	})

	baseQuery := `
		SELECT count(distinct IV.issuevariant_id) FROM  IssueVariant IV 
		%s
		%s
		ORDER BY %s
    `

	stmt, filterParameters, err := s.buildIssueVariantStatement(
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

	return performCountScan(
		stmt,
		filterParameters,
		l,
	)
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
