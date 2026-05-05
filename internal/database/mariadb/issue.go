// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var issueObject = DbObject[*entity.Issue]{
	Prefix:    "issue",
	TableName: "Issue",
	Properties: []*Property{
		NewProperty(
			"issue_primary_name",
			WrapAccess(
				func(i *entity.Issue) (string, bool) { return i.PrimaryName, i.PrimaryName != "" },
			),
		),
		NewProperty(
			"issue_type",
			WrapAccess(
				func(i *entity.Issue) (entity.IssueType, bool) { return i.Type, i.Type != "" },
			),
		),
		NewProperty(
			"issue_description",
			WrapAccess(
				func(i *entity.Issue) (string, bool) { return i.Description, i.Description != "" },
			),
		),
		NewProperty(
			"issue_created_by",
			WrapAccess(func(i *entity.Issue) (int64, bool) { return i.CreatedBy, NoUpdate }),
		),
		NewProperty(
			"issue_updated_by",
			WrapAccess(
				func(i *entity.Issue) (int64, bool) { return i.UpdatedBy, i.UpdatedBy != 0 },
			),
		),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty(
			"S.service_ccrn = ?",
			WrapRetSlice(func(filter *entity.IssueFilter) []*string { return filter.ServiceCCRN }),
		),
		NewFilterProperty(
			"CI.componentinstance_service_id = ?",
			WrapRetSlice(func(filter *entity.IssueFilter) []*int64 { return filter.ServiceId }),
		),
		NewFilterProperty(
			"I.issue_id = ?",
			WrapRetSlice(func(filter *entity.IssueFilter) []*int64 { return filter.Id }),
		),
		NewFilterProperty(
			"IM.issuematch_status = ?",
			WrapRetSlice(
				func(filter *entity.IssueFilter) []*string { return filter.IssueMatchStatus },
			),
		),
		NewFilterProperty(
			"IM.issuematch_rating = ?",
			WrapRetSlice(
				func(filter *entity.IssueFilter) []*string { return filter.IssueMatchSeverity },
			),
		),
		NewFilterProperty(
			"IM.issuematch_id = ?",
			WrapRetSlice(func(filter *entity.IssueFilter) []*int64 { return filter.IssueMatchId }),
		),
		NewFilterProperty(
			"CVI.componentversionissue_component_version_id = ?",
			WrapRetSlice(
				func(filter *entity.IssueFilter) []*int64 { return filter.ComponentVersionId },
			),
		),
		NewFilterProperty(
			"IV.issuevariant_id = ?",
			WrapRetSlice(
				func(filter *entity.IssueFilter) []*int64 { return filter.IssueVariantId },
			),
		),
		NewFilterProperty(
			"I.issue_type = ?",
			WrapRetSlice(func(filter *entity.IssueFilter) []*string { return filter.Type }),
		),
		NewFilterProperty(
			"I.issue_primary_name = ?",
			WrapRetSlice(func(filter *entity.IssueFilter) []*string { return filter.PrimaryName }),
		),
		NewFilterProperty(
			"IV.issuevariant_repository_id = ?",
			WrapRetSlice(
				func(filter *entity.IssueFilter) []*int64 { return filter.IssueRepositoryId },
			),
		),
		NewFilterProperty(
			"SG.supportgroup_ccrn = ?",
			WrapRetSlice(
				func(filter *entity.IssueFilter) []*string { return filter.SupportGroupCCRN },
			),
		),
		NewFilterProperty(
			"CV.componentversion_component_id = ?",
			WrapRetSlice(func(filter *entity.IssueFilter) []*int64 { return filter.ComponentId }),
		),
		NewNFilterProperty(
			"IV.issuevariant_secondary_name LIKE Concat('%',?,'%') OR I.issue_primary_name LIKE Concat('%',?,'%')",
			WrapRetSlice(func(filter *entity.IssueFilter) []*string { return filter.Search }),
			2,
		),
		NewStateFilterProperty(
			"I.issue",
			WrapRetState(
				func(filter *entity.IssueFilter) []entity.StateFilterType { return filter.State },
			),
		),
		NewCustomFilterProperty(
			WrapBuilder(func(is []entity.IssueStatus) string {
				if len(is) != 1 {
					panic(fmt.Sprintf("Unexpected number of elements for IssueStatus: %d", len(is)))
				}
				switch is[0] {
				case entity.IssueStatusOpen:
					return "( R.remediation_id IS NULL OR R.remediation_expiration_date < CURDATE() )"
				case entity.IssueStatusRemediated:
					return "( R.remediation_id IS NOT NULL AND R.remediation_expiration_date > CURDATE() )"
				}
				return ""
			}),
			WrapRetSlice(
				func(filter *entity.IssueFilter) []entity.IssueStatus { return []entity.IssueStatus{filter.Status} },
			),
		),
	},
	JoinDefs: []*JoinDef{
		{
			Name:  "IM_RJ",
			Type:  RightJoin,
			Table: "IssueMatch IM",
			On:    "I.issue_id = IM.issuematch_issue_id",
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				return f.HasIssueMatches
			}),
		},
		{
			Name:  "IM_LJ",
			Type:  LeftJoin,
			Table: "IssueMatch IM",
			On:    "I.issue_id = IM.issuematch_issue_id",
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.IssueMatchStatus) > 0 || len(f.IssueMatchId) > 0 || len(f.IssueMatchSeverity) > 0
			}),
		},
		{
			Name:      "CI with IM_RJ",
			Type:      LeftJoin,
			Table:     "ComponentInstance CI",
			On:        "IM.issuematch_component_instance_id = CI.componentinstance_id",
			DependsOn: []string{"IM_RJ"},
			Condition: DependentJoin,
		},
		{
			Name:      "CI with IM_LJ",
			Type:      LeftJoin,
			Table:     "ComponentInstance CI",
			On:        "IM.issuematch_component_instance_id = CI.componentinstance_id",
			DependsOn: []string{"IM_LJ"},
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.ServiceId) > 0
			}),
		},
		{
			Name:      "CV with IM_RJ",
			Type:      LeftJoin,
			Table:     "ComponentVersion CV",
			On:        "CI.componentinstance_component_version_id = CV.componentversion_id",
			DependsOn: []string{"CI with IM_RJ"},
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				return f.AllServices
			}),
		}, // Looks like this case is not used because of mv_vulnerabilities
		{
			Name:      "CV with IM_LJ",
			Type:      LeftJoin,
			Table:     "ComponentVersion CV",
			On:        "CI.componentinstance_component_version_id = CV.componentversion_id",
			DependsOn: []string{"CI with IM_LJ"},
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.ServiceCCRN) > 0
			}),
		},
		{
			Name:      "S with IM_RJ",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			DependsOn: []string{"CI with IM_RJ"},
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				return f.AllServices
			}),
		}, // Looks like this case is not used because of mv_vulnerabilities
		{
			Name:      "S with IM_LJ",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			DependsOn: []string{"CI with IM_LJ"},
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.ServiceCCRN) > 0
			}),
		},
		{
			Name:      "SGS",
			Type:      LeftJoin,
			Table:     "SupportGroupService SGS",
			On:        "CI.componentinstance_service_id = SGS.supportgroupservice_service_id",
			DependsOn: []string{"CI with IM_LJ"},
			Condition: DependentJoin,
		},
		{
			Name:      "SG",
			Type:      LeftJoin,
			Table:     "SupportGroup SG",
			On:        "SGS.supportgroupservice_support_group_id = SG.supportgroup_id",
			DependsOn: []string{"SGS"},
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.SupportGroupCCRN) > 0
			}),
		},
		{
			Name:  "IV",
			Type:  LeftJoin,
			Table: "IssueVariant IV",
			On:    "I.issue_id = IV.issuevariant_issue_id",
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, order *Order) bool {
				return len(f.IssueVariantId) > 0 || len(f.IssueRepositoryId) > 0 || len(f.Search) > 0 || order.ByRating()
			}),
		},
		{
			Name:  "CVI",
			Type:  LeftJoin,
			Table: "ComponentVersionIssue CVI",
			On:    "I.issue_id = CVI.componentversionissue_issue_id",
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.ComponentVersionId) > 0
			}),
		},
		{
			Name:      "CV using CVI",
			Type:      LeftJoin,
			Table:     "ComponentVersion CV",
			On:        "CVI.componentversionissue_component_version_id = CV.componentversion_id",
			DependsOn: []string{"CVI"},
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.ComponentId) > 0 && (len(f.ServiceId) == 0 && len(f.ServiceCCRN) == 0 && len(f.SupportGroupCCRN) == 0 && !f.AllServices)
			}),
		},
		{
			Name:      "R has S and C",
			Type:      LeftJoin,
			Table:     "Remediation R",
			On:        "I.issue_id = R.remediation_issue_id AND R.remediation_deleted_at IS NULL AND CI.componentinstance_service_id = R.remediation_service_id AND CV.componentversion_component_id = R.remediation_component_id",
			DependsOn: []string{"CI with IM_LJ", "CV using CVI"},
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				hasService := len(f.ServiceCCRN) > 0 || len(f.ServiceId) > 0
				hasComponent := len(f.ComponentId) > 0
				return (f.Status == entity.IssueStatusOpen || f.Status == entity.IssueStatusRemediated) && hasService && hasComponent
			}),
		}, // Missing test
		{
			Name:      "R has S",
			Type:      LeftJoin,
			Table:     "Remediation R",
			On:        "I.issue_id = R.remediation_issue_id AND R.remediation_deleted_at IS NULL AND CI.componentinstance_service_id = R.remediation_service_id",
			DependsOn: []string{"CI with IM_LJ"},
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				hasService := len(f.ServiceCCRN) > 0 || len(f.ServiceId) > 0
				hasComponent := len(f.ComponentId) > 0
				return (f.Status == entity.IssueStatusOpen || f.Status == entity.IssueStatusRemediated) && hasService && !hasComponent
			}),
		}, // Missing test
		{
			Name:      "R has C",
			Type:      LeftJoin,
			Table:     "Remediation R",
			On:        "I.issue_id = R.remediation_issue_id AND R.remediation_deleted_at IS NULL AND CV.componentversion_component_id = R.remediation_component_id",
			DependsOn: []string{"CV using CVI"},
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				hasService := len(f.ServiceCCRN) > 0 || len(f.ServiceId) > 0
				hasComponent := len(f.ComponentId) > 0
				return (f.Status == entity.IssueStatusOpen || f.Status == entity.IssueStatusRemediated) && !hasService && hasComponent
			}),
		}, // Missing test
		{
			Name:  "R",
			Type:  LeftJoin,
			Table: "Remediation R",
			On:    "I.issue_id = R.remediation_issue_id AND R.remediation_deleted_at IS NULL",
			Condition: WrapJoinCondition(func(f *entity.IssueFilter, _ *Order) bool {
				hasService := len(f.ServiceCCRN) > 0 || len(f.ServiceId) > 0
				hasComponent := len(f.ComponentId) > 0
				return (f.Status == entity.IssueStatusOpen || f.Status == entity.IssueStatusRemediated) && !hasService && !hasComponent
			}),
		},
	},
}

func ensureIssueFilter(filter *entity.IssueFilter) *entity.IssueFilter {
	if filter == nil {
		filter = &entity.IssueFilter{}
	}

	return EnsurePagination(filter)
}

func getIssueColumns(order []entity.Order) string {
	columns := ""

	for _, o := range order {
		switch o.By {
		case entity.IssueVariantRating:
			columns = fmt.Sprintf(
				"%s, MAX(CAST(IV.issuevariant_rating AS UNSIGNED)) AS issuevariant_rating_num",
				columns,
			)
		}
	}

	return columns
}

func getIssueQueryWithCursor(
	baseQuery string,
	order []entity.Order,
	filter *entity.IssueFilter,
	cursorFields []Field,
) string {
	issueColumns := getIssueColumns(order)
	ord := NewOrder(order, entity.Order{By: entity.IssueId, Direction: entity.OrderDirectionAsc})
	joins := issueObject.GetJoins(filter, ord)
	whereClause, hasFilter := issueObject.GetFilterWhereClause(filter, false)
	issueCursor := issueObject.GetCursorQuery(&hasFilter, cursorFields, nil, true)

	return fmt.Sprintf(baseQuery, issueColumns, joins, whereClause, issueCursor, ord)
}

func getIssueQuery(baseQuery string, order []entity.Order, filter *entity.IssueFilter) string {
	issueColumns := getIssueColumns(order)
	ord := NewOrder(order, entity.Order{By: entity.IssueId, Direction: entity.OrderDirectionAsc})
	joins := issueObject.GetJoins(filter, ord)
	whereClause, _ := issueObject.GetFilterWhereClause(filter, false)

	return fmt.Sprintf(baseQuery, issueColumns, joins, whereClause, ord)
}

func (s *SqlDatabase) buildIssueStatementWithCursor(
	ctx context.Context,
	baseQuery string,
	filter *entity.IssueFilter,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	ifilter := ensureIssueFilter(filter)
	l.WithFields(logrus.Fields{"filter": ifilter})

	cursorFields, err := DecodeCursor(ifilter.After)
	if err != nil {
		return nil, nil, err
	}

	query := getIssueQueryWithCursor(baseQuery, order, ifilter, cursorFields)

	// construct prepared statement and if where clause does exist add parameters
	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			},
		).Error(msg)

		return nil, nil, fmt.Errorf("%s", msg)
	}

	// adding parameters
	filterParameters := issueObject.GetFilterParameters(ifilter, true, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) buildIssueStatement(
	ctx context.Context,
	baseQuery string,
	filter *entity.IssueFilter,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	ifilter := ensureIssueFilter(filter)
	l.WithFields(logrus.Fields{"filter": ifilter})

	cursorFields, err := DecodeCursor(ifilter.After)
	if err != nil {
		return nil, nil, err
	}

	query := getIssueQuery(baseQuery, order, ifilter)

	// construct prepared statement and if where clause does exist add parameters
	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			},
		).Error(msg)

		return nil, nil, fmt.Errorf("%s", msg)
	}

	// adding parameters
	filterParameters := issueObject.GetFilterParameters(ifilter, false, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetIssuesWithAggregations(
	ctx context.Context,
	filter *entity.IssueFilter,
	order []entity.Order,
) ([]entity.IssueResult, error) {
	filter = ensureIssueFilter(filter)
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssuesWithAggregations",
	})

	baseCiQuery := `
        SELECT I.*, SUM(CI.componentinstance_count) AS agg_affected_component_instances %s FROM Issue I
        LEFT JOIN IssueMatch IM on I.issue_id = IM.issuematch_issue_id
        LEFT JOIN ComponentInstance CI on IM.issuematch_component_instance_id = CI.componentinstance_id
        %s
        %s
        GROUP BY I.issue_id %s ORDER BY %s LIMIT ?
    `

	// count(distinct activity_id) as agg_activities,
	// LEFT JOIN ActivityHasIssue AHI on I.issue_id = AHI.activityhasissue_issue_id
	// LEFT JOIN Activity A on AHI.activityhasissue_activity_id = A.activity_id~

	baseAggQuery := `
		SELECT I.*,
		count(distinct issuematch_id) as agg_issue_matches,
		count(distinct service_ccrn) as agg_affected_services,
		count(distinct componentversionissue_component_version_id) as agg_component_versions,
		min(issuematch_target_remediation_date) as agg_earliest_target_remediation_date,
		min(issuematch_created_at) agg_earliest_discovery_date
		%s
        FROM Issue I
        LEFT JOIN IssueMatch IM on I.issue_id = IM.issuematch_issue_id
        LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
        LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id
        LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
        LEFT JOIN ComponentVersionIssue CVI ON I.issue_id = CVI.componentversionissue_issue_id
		%s
		%s
		GROUP BY I.issue_id %s ORDER BY %s LIMIT ?
    `

	baseQuery := `
        With ComponentInstanceCounts AS (
            %s
        ),
        Aggs AS (
            %s
        )
        SELECT A.*, CIC.*
        FROM ComponentInstanceCounts CIC
        JOIN Aggs A ON CIC.issue_id = A.issue_id;
    `

	filter = ensureIssueFilter(filter)
	joins := issueObject.GetJoins(filter, NewOrder(order, entity.Order{})) // It seems that this join is redundant for baseAppQuery
	// We should improve testing and remove redundant joins from query

	cursorFields, err := DecodeCursor(filter.After)
	if err != nil {
		return nil, err
	}

	columns := getIssueColumns(order)
	ord := NewOrder(order, entity.Order{By: entity.IssueId, Direction: entity.OrderDirectionAsc})

	whereClause, _ := issueObject.GetFilterWhereClause(filter, false)

	cursorQuery := CreateCursorQuery("", cursorFields)

	filterStr := issueObject.GetFilterQuery(filter)
	if filterStr != "" && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	ciQuery := fmt.Sprintf(baseCiQuery, columns, joins, whereClause, cursorQuery, ord)
	aggQuery := fmt.Sprintf(baseAggQuery, columns, joins, whereClause, cursorQuery, ord)
	query := fmt.Sprintf(baseQuery, ciQuery, aggQuery)

	stmt, err := s.db.PreparexContext(ctx, query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			},
		).Error(msg)

		return nil, fmt.Errorf("%s", msg)
	}

	// parameters for component instance query
	filterParameters := issueObject.GetFilterParameters(filter, true, cursorFields)
	// parameters for agg query
	filterParameters = append(
		filterParameters,
		issueObject.GetFilterParameters(filter, true, cursorFields)...,
	)

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
		func(l []entity.IssueResult, e RowComposite) []entity.IssueResult {
			gibr := GetIssuesByRow{
				IssueAggregationsRow: *e.IssueAggregationsRow,
				IssueRow:             *e.IssueRow,
			}
			issue := gibr.AsIssueWithAggregations()

			var ivRating int64
			if e.IssueVariantRow != nil {
				ivRating = e.RatingNumerical.Int64
			}

			cursor, _ := EncodeCursor(WithIssue(ord.Sequence(), issue.Issue, ivRating))

			sr := entity.IssueResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				Issue:             &issue.Issue,
				IssueAggregations: &issue.IssueAggregations,
			}

			return append(l, sr)
		},
	)
}

func (s *SqlDatabase) CountIssues(ctx context.Context, filter *entity.IssueFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssues",
	})

	baseQuery := `
		SELECT COUNT(distinct I.issue_id) %s FROM Issue I
		%s
		%s
		ORDER BY %s
	`

	stmt, filterParameters, err := s.buildIssueStatement(ctx, baseQuery, filter, []entity.Order{}, l)
	if err != nil {
		return -1, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	return performCountScan(ctx, stmt, filterParameters, l)
}

func (s *SqlDatabase) CountIssueTypes(ctx context.Context, filter *entity.IssueFilter) (*entity.IssueTypeCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssueTypes",
	})

	baseQuery := `
		SELECT I.issue_type AS issue_value, COUNT(distinct I.issue_id) as issue_count %s FROM Issue I
		%s
		%s
		GROUP BY I.issue_type ORDER BY %s
	`

	stmt, filterParameters, err := s.buildIssueStatement(ctx, baseQuery, filter, []entity.Order{}, l)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	counts, err := performListScan(
		ctx,
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueCount, e IssueCountRow) []entity.IssueCount {
			return append(l, e.AsIssueCount())
		},
	)
	if err != nil {
		return nil, err
	}

	var issueTypeCounts entity.IssueTypeCounts

	for _, count := range counts {
		switch count.Value {
		case entity.IssueTypeVulnerability.String():
			issueTypeCounts.VulnerabilityCount = count.Count
		case entity.IssueTypePolicyViolation.String():
			issueTypeCounts.PolicyViolationCount = count.Count
		case entity.IssueTypeSecurityEvent.String():
			issueTypeCounts.SecurityEventCount = count.Count
		}
	}

	return &issueTypeCounts, nil
}

func (s *SqlDatabase) GetAllIssueCursors(
	ctx context.Context,
	filter *entity.IssueFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllIssueCursors",
	})

	baseQuery := `
		SELECT I.* %s FROM Issue I 
		%s
	    %s GROUP BY I.issue_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildIssueStatement(ctx, baseQuery, filter, order, l)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	rows, err := performListScan(
		context.Background(),
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
		issue := row.IssueRow.AsIssue()

		var ivRating int64
		if row.IssueVariantRow != nil {
			ivRating = row.RatingNumerical.Int64
		}

		cursor, _ := EncodeCursor(WithIssue(order, issue, ivRating))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetIssues(
	ctx context.Context,
	filter *entity.IssueFilter,
	order []entity.Order,
) ([]entity.IssueResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssues",
	})

	baseQuery := `
		SELECT I.* %s FROM Issue I
		%s
		%s
		GROUP BY I.issue_id %s ORDER BY %s LIMIT ?
    `

	filter = ensureIssueFilter(filter)

	stmt, filterParameters, err := s.buildIssueStatementWithCursor(ctx, baseQuery, filter, order, l)
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
		func(l []entity.IssueResult, e RowComposite) []entity.IssueResult {
			issue := e.IssueRow.AsIssue()

			var ivRating int64
			if e.IssueVariantRow != nil {
				ivRating = e.RatingNumerical.Int64
			}

			cursor, _ := EncodeCursor(WithIssue(order, issue, ivRating))

			sr := entity.IssueResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				Issue: &issue,
			}

			return append(l, sr)
		},
	)
}

func (s *SqlDatabase) CreateIssue(issue *entity.Issue) (*entity.Issue, error) {
	return issueObject.Create(s.db, issue)
}

func (s *SqlDatabase) UpdateIssue(issue *entity.Issue) error {
	return issueObject.Update(s.db, issue)
}

func (s *SqlDatabase) DeleteIssue(id int64, userId int64) error {
	return issueObject.Delete(s.db, id, userId)
}

func (s *SqlDatabase) AddComponentVersionToIssue(issueId int64, componentVersionId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"issueId":            issueId,
		"componentVersionId": componentVersionId,
		"event":              "database.AddComponentVersionToIssue",
	})

	query := `
		INSERT INTO ComponentVersionIssue (
			componentversionissue_issue_id,
			componentversionissue_component_version_id
		) VALUES (
			:issue_id,
			:component_version_id
		)
	`

	args := map[string]any{
		"issue_id":             issueId,
		"component_version_id": componentVersionId,
	}

	var mysqlErr *mysql.MySQLError

	_, err := performExec(s, query, args, l)
	if err != nil {
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == database.ErrCodeDuplicateEntry {
				return nil
			}
		}

		return err
	}

	return nil
}

func (s *SqlDatabase) RemoveComponentVersionFromIssue(
	issueId int64,
	componentVersionId int64,
) error {
	l := logrus.WithFields(logrus.Fields{
		"issueId":            issueId,
		"componentVersionId": componentVersionId,
		"event":              "database.RemoveComponentVersionFromIssue",
	})

	query := `
		DELETE FROM ComponentVersionIssue
		WHERE
			componentversionissue_issue_id = :issue_id
			AND componentversionissue_component_version_id = :component_version_id
	`

	args := map[string]any{
		"issue_id":             issueId,
		"component_version_id": componentVersionId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) RemoveAllIssuesFromComponentVersion(componentVersionId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"componentVersionId": componentVersionId,
		"event":              "database.RemoveAllIssuesFromComponentVersion",
	})

	query := `
		DELETE FROM ComponentVersionIssue
		WHERE
			componentversionissue_component_version_id = :component_version_id
	`

	args := map[string]any{
		"component_version_id": componentVersionId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) GetIssueNames(ctx context.Context, filter *entity.IssueFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssueNames",
	})

	baseQuery := `
    SELECT I.issue_primary_name FROM Issue I
    %s
    %s
    ORDER BY %s
    `

	order := []entity.Order{
		{By: entity.IssuePrimaryName, Direction: entity.OrderDirectionAsc},
	}

	// Ensure the filter is initialized
	filter = ensureIssueFilter(filter)

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildIssueStatement(ctx, baseQuery, filter, order, l)
	if err != nil {
		l.Error("Error preparing statement: ", err)
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	// Execute the query
	rows, err := stmt.QueryxContext(ctx, filterParameters...)
	if err != nil {
		l.Error("Error executing query: ", err)
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Warnf("error during close rows: %s", err)
		}
	}()

	// Collect the results
	issueNames := []string{}

	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}

		issueNames = append(issueNames, name)
	}

	if err = rows.Err(); err != nil {
		l.Error("Row iteration error: ", err)
		return nil, err
	}

	return issueNames, nil
}
