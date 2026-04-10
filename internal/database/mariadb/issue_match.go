// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var issueMatchObject = DbObject[*entity.IssueMatch]{
	Prefix:    "issuematch",
	TableName: "IssueMatch",
	Properties: []*Property{
		NewProperty(
			"issuematch_status",
			WrapAccess(func(im *entity.IssueMatch) (entity.IssueMatchStatusValue, bool) {
				return im.Status, im.Status != "" && im.Status != entity.IssueMatchStatusValuesNone
			}),
		),
		NewProperty(
			"issuematch_remediation_date",
			WrapAccess(func(im *entity.IssueMatch) (time.Time, bool) {
				return im.RemediationDate, !im.RemediationDate.IsZero()
			}),
		),
		NewProperty(
			"issuematch_target_remediation_date",
			WrapAccess(func(im *entity.IssueMatch) (time.Time, bool) {
				return im.TargetRemediationDate, !im.TargetRemediationDate.IsZero()
			}),
		),
		NewProperty("issuematch_vector", WrapAccess(func(im *entity.IssueMatch) (string, bool) {
			return im.Severity.Cvss.Vector, im.Severity.Cvss.Vector != ""
		})),
		NewProperty(
			"issuematch_rating",
			WrapAccess(
				func(im *entity.IssueMatch) (string, bool) { return im.Severity.Value, im.Severity.Value != "" },
			),
		),
		NewProperty(
			"issuematch_user_id",
			WrapAccess(
				func(im *entity.IssueMatch) (int64, bool) { return im.UserId, im.UserId != 0 },
			),
		),
		NewProperty(
			"issuematch_component_instance_id",
			WrapAccess(
				func(im *entity.IssueMatch) (int64, bool) { return im.ComponentInstanceId, im.ComponentInstanceId != 0 },
			),
		),
		NewProperty(
			"issuematch_issue_id",
			WrapAccess(
				func(im *entity.IssueMatch) (int64, bool) { return im.IssueId, im.IssueId != 0 },
			),
		),
		NewProperty(
			"issuematch_created_by",
			WrapAccess(func(im *entity.IssueMatch) (int64, bool) { return im.CreatedBy, NoUpdate }),
		),
		NewProperty(
			"issuematch_updated_by",
			WrapAccess(
				func(im *entity.IssueMatch) (int64, bool) { return im.UpdatedBy, im.UpdatedBy != 0 },
			),
		),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty(
			"IM.issuematch_id = ?",
			WrapRetSlice(func(filter *entity.IssueMatchFilter) []*int64 { return filter.Id }),
		),
		NewFilterProperty(
			"IM.issuematch_issue_id = ?",
			WrapRetSlice(func(filter *entity.IssueMatchFilter) []*int64 { return filter.IssueId }),
		),
		NewFilterProperty(
			"IM.issuematch_component_instance_id = ?",
			WrapRetSlice(
				func(filter *entity.IssueMatchFilter) []*int64 { return filter.ComponentInstanceId },
			),
		),
		NewFilterProperty(
			"S.service_ccrn = ?",
			WrapRetSlice(
				func(filter *entity.IssueMatchFilter) []*string { return filter.ServiceCCRN },
			),
		),
		NewFilterProperty(
			"CI.componentinstance_service_id = ?",
			WrapRetSlice(
				func(filter *entity.IssueMatchFilter) []*int64 { return filter.ServiceId },
			),
		),
		NewFilterProperty(
			"IM.issuematch_rating = ?",
			WrapRetSlice(
				func(filter *entity.IssueMatchFilter) []*string { return filter.SeverityValue },
			),
		),
		NewFilterProperty(
			"IM.issuematch_status = ?",
			WrapRetSlice(func(filter *entity.IssueMatchFilter) []*string { return filter.Status }),
		),
		NewFilterProperty(
			"SG.supportgroup_ccrn = ?",
			WrapRetSlice(
				func(filter *entity.IssueMatchFilter) []*string { return filter.SupportGroupCCRN },
			),
		),
		NewFilterProperty(
			"I.issue_primary_name = ?",
			WrapRetSlice(
				func(filter *entity.IssueMatchFilter) []*string { return filter.PrimaryName },
			),
		),
		NewFilterProperty(
			"C.component_ccrn = ?",
			WrapRetSlice(
				func(filter *entity.IssueMatchFilter) []*string { return filter.ComponentCCRN },
			),
		),
		NewFilterProperty(
			"I.issue_type = ?",
			WrapRetSlice(
				func(filter *entity.IssueMatchFilter) []*string { return filter.IssueType },
			),
		),
		NewFilterProperty(
			"U.user_name = ?",
			WrapRetSlice(
				func(filter *entity.IssueMatchFilter) []*string { return filter.ServiceOwnerUsername },
			),
		),
		NewFilterProperty(
			"U.user_unique_user_id = ?",
			WrapRetSlice(
				func(filter *entity.IssueMatchFilter) []*string { return filter.ServiceOwnerUniqueUserId },
			),
		),
		NewNFilterProperty(
			"IV.issuevariant_secondary_name LIKE Concat('%',?,'%') OR I.issue_primary_name LIKE Concat('%',?,'%')",
			WrapRetSlice(func(filter *entity.IssueMatchFilter) []*string { return filter.Search }),
			2,
		),
		NewStateFilterProperty(
			"IM.issuematch",
			WrapRetState(
				func(filter *entity.IssueMatchFilter) []entity.StateFilterType { return filter.State },
			),
		),
	},
	JoinDefs: []*JoinDef{
		{
			Name:  "I",
			Type:  LeftJoin,
			Table: "Issue I",
			On:    "IM.issuematch_issue_id = I.issue_id",
			Condition: WrapJoinCondition(func(f *entity.IssueMatchFilter, order []entity.Order) bool {
				orderByIssuePrimaryName := lo.ContainsBy(order, func(o entity.Order) bool {
					return o.By == entity.IssuePrimaryName
				})
				return len(f.IssueType) > 0 || len(f.PrimaryName) > 0 || orderByIssuePrimaryName
			}),
		},
		{
			Name:      "IV",
			Type:      LeftJoin,
			Table:     "IssueVariant IV",
			On:        "I.issue_id = IV.issuevariant_issue_id",
			DependsOn: []string{"I"},
			Condition: WrapJoinCondition(func(f *entity.IssueMatchFilter, _ []entity.Order) bool {
				return len(f.Search) > 0
			}),
		},
		{
			Name:  "CI",
			Type:  LeftJoin,
			Table: "ComponentInstance CI",
			On:    "IM.issuematch_component_instance_id = CI.componentinstance_id",
			Condition: WrapJoinCondition(func(f *entity.IssueMatchFilter, order []entity.Order) bool {
				orderByCiCcrn := lo.ContainsBy(order, func(o entity.Order) bool {
					return o.By == entity.ComponentInstanceCcrn
				})
				return orderByCiCcrn || len(f.ServiceId) > 0
			}),
		},
		{
			Name:      "CV",
			Type:      LeftJoin,
			Table:     "ComponentVersion CV",
			On:        "CI.componentinstance_component_version_id = CV.componentversion_id",
			DependsOn: []string{"CI"},
			Condition: DependentJoin,
		},
		{
			Name:      "C",
			Type:      LeftJoin,
			Table:     "Componen C",
			On:        "CV.componentversion_component_id = C.component_id",
			DependsOn: []string{"CV"},
			Condition: WrapJoinCondition(func(f *entity.IssueMatchFilter, _ []entity.Order) bool {
				return len(f.ComponentCCRN) > 0
			}),
		},
		{
			Name:      "S",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			DependsOn: []string{"CI"},
			Condition: WrapJoinCondition(func(f *entity.IssueMatchFilter, _ []entity.Order) bool {
				return len(f.ServiceCCRN) > 0
			}),
		},
		{
			Name:      "SGS",
			Type:      LeftJoin,
			Table:     "SupportGroupService SGS",
			On:        "S.service_id = SS.supportgroupservice_service_id",
			DependsOn: []string{"S"},
			Condition: DependentJoin,
		},
		{
			Name:      "SG",
			Type:      LeftJoin,
			Table:     "SupportGroup SG",
			On:        "SGS.supportgroupservice_support_group_id = SG.supportgroup_id",
			DependsOn: []string{"SGS"},
			Condition: WrapJoinCondition(func(f *entity.IssueMatchFilter, _ []entity.Order) bool {
				return len(f.SupportGroupCCRN) > 0
			}),
		},
		{
			Name:      "O",
			Type:      LeftJoin,
			Table:     "Owner O",
			On:        "CI.componentinstance_service_id = O.owner_service_id",
			DependsOn: []string{"CI"},
			Condition: DependentJoin,
		},
		{
			Name:      "U",
			Type:      LeftJoin,
			Table:     "User U",
			On:        "O.owner_user_id = U.user_id",
			DependsOn: []string{"O"},
			Condition: WrapJoinCondition(func(f *entity.IssueMatchFilter, _ []entity.Order) bool {
				return len(f.ServiceOwnerUsername) > 0 || len(f.ServiceOwnerUniqueUserId) > 0
			}),
		},
	},
}

func ensureIssueMatchFilter(filter *entity.IssueMatchFilter) *entity.IssueMatchFilter {
	if filter == nil {
		filter = &entity.IssueMatchFilter{}
	}

	return EnsurePagination(filter)
}

func (s *SqlDatabase) getIssueMatchJoins(
	filter *entity.IssueMatchFilter,
	order []entity.Order,
) string {
	joins := ""
	orderByIssuePrimaryName := lo.ContainsBy(order, func(o entity.Order) bool {
		return o.By == entity.IssuePrimaryName
	})
	orderByCiCcrn := lo.ContainsBy(order, func(o entity.Order) bool {
		return o.By == entity.ComponentInstanceCcrn
	})

	if len(filter.Search) > 0 || len(filter.IssueType) > 0 || len(filter.PrimaryName) > 0 ||
		orderByIssuePrimaryName {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN Issue I on I.issue_id = IM.issuematch_issue_id
		`)
		if len(filter.Search) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueVariant IV on IV.issuevariant_issue_id = I.issue_id
			`)
		}
	}

	if orderByCiCcrn || len(filter.ServiceId) > 0 || len(filter.ServiceCCRN) > 0 ||
		len(filter.SupportGroupCCRN) > 0 ||
		len(filter.ComponentCCRN) > 0 ||
		len(filter.ServiceOwnerUsername) > 0 ||
		len(filter.ServiceOwnerUniqueUserId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentInstance CI on CI.componentinstance_id = IM.issuematch_component_instance_id
			
		`)

		if len(filter.ComponentCCRN) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, `
                LEFT JOIN ComponentVersion CV on CV.componentversion_id = CI.componentinstance_component_version_id
				LEFT JOIN Component C on C.component_id = CV.componentversion_component_id
			`)
		}

		if len(filter.ServiceCCRN) > 0 || len(filter.SupportGroupCCRN) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, `
				LEFT JOIN Service S on S.service_id = CI.componentinstance_service_id
			`)
		}

		if len(filter.SupportGroupCCRN) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, `
				LEFT JOIN SupportGroupService SGS on S.service_id = SGS.supportgroupservice_service_id
				LEFT JOIN SupportGroup SG on SG.supportgroup_id = SGS.supportgroupservice_support_group_id
			`)
		}

		if len(filter.ServiceOwnerUsername) > 0 || len(filter.ServiceOwnerUniqueUserId) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, `
				LEFT JOIN Owner O on O.owner_service_id = CI.componentinstance_service_id
				LEFT JOIN User U on U.user_id = O.owner_user_id
			`)
		}
	}

	return joins
}

func (s *SqlDatabase) getIssueMatchColumns(order []entity.Order) string {
	columns := ""

	for _, o := range order {
		switch o.By {
		case entity.IssuePrimaryName:
			columns = fmt.Sprintf("%s, I.issue_primary_name", columns)
		case entity.ComponentInstanceCcrn:
			columns = fmt.Sprintf("%s, CI.componentinstance_ccrn", columns)
		}
	}

	return columns
}

func (s *SqlDatabase) buildIssueMatchStatement(
	baseQuery string,
	filter *entity.IssueMatchFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	filter = ensureIssueMatchFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	cursorFields, err := DecodeCursor(filter.After)
	if err != nil {
		return nil, nil, err
	}

	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.IssueMatchId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)
	columns := s.getIssueMatchColumns(order)
	joins := s.getIssueMatchJoins(filter, order)

	filterStr := issueMatchObject.GetFilterQuery(filter)

	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && withCursor && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	var query string
	if withCursor {
		query = fmt.Sprintf(baseQuery, columns, joins, whereClause, cursorQuery, orderStr)
	} else {
		query = fmt.Sprintf(baseQuery, columns, joins, whereClause, orderStr)
	}

	// construct prepared statement and if where clause does exist add parameters
	stmt, err := s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			}).Error(msg)

		return nil, nil, fmt.Errorf("%s", msg)
	}

	filterParameters := issueMatchObject.GetFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllIssueMatchIds(filter *entity.IssueMatchFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllIssueMatchIds",
	})

	baseQuery := `
		SELECT IM.issuematch_id %s FROM IssueMatch IM 
		%s
	 	%s GROUP BY IM.issuematch_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildIssueMatchStatement(
		baseQuery,
		filter,
		false,
		[]entity.Order{},
		l,
	)
	if err != nil {
		return nil, err
	}

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetAllIssueMatchCursors(
	filter *entity.IssueMatchFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssueAllIssueMatchCursors",
	})

	baseQuery := `
		SELECT IM.* %s FROM IssueMatch IM 
		%s
	    %s GROUP BY IM.issuematch_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildIssueMatchStatement(baseQuery, filter, false, order, l)
	if err != nil {
		return nil, err
	}

	rows, err := performListScan(
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
		im := row.AsIssueMatch()
		if row.IssueRow != nil {
			im.Issue = new(row.IssueRow.AsIssue())
		}

		if row.ComponentInstanceRow != nil {
			im.ComponentInstance = new(row.AsComponentInstance())
		}

		cursor, _ := EncodeCursor(WithIssueMatch(order, im))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetIssueMatches(
	filter *entity.IssueMatchFilter,
	order []entity.Order,
) ([]entity.IssueMatchResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssueMatches",
	})

	baseQuery := `
		SELECT IM.* %s FROM IssueMatch IM 
		%s
	    %s %s GROUP BY IM.issuematch_id ORDER BY %s LIMIT ?
    `

	stmt, filterParameters, err := s.buildIssueMatchStatement(baseQuery, filter, true, order, l)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueMatchResult, e RowComposite) []entity.IssueMatchResult {
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
	)
}

func (s *SqlDatabase) CountIssueMatches(filter *entity.IssueMatchFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.CountIssueMatches",
	})

	baseQuery := `
		SELECT count(distinct IM.issuematch_id) %s FROM IssueMatch IM 
		%s
		%s
		ORDER BY %s
    `

	stmt, filterParameters, err := s.buildIssueMatchStatement(
		baseQuery,
		filter,
		false,
		[]entity.Order{},
		l,
	)
	if err != nil {
		return -1, err
	}

	return performCountScan(stmt, filterParameters, l)
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
