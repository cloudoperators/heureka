// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var issueObject = DbObject[*entity.Issue, *entity.IssueFilter, entity.IssueResult]{
	Prefix:       "issue",
	TableName:    "Issue",
	TableKey:     "I",
	DefaultOrder: entity.Order{By: entity.IssueId, Direction: entity.OrderDirectionAsc},
	Aggregated:   true,
	Properties: []*Property[*entity.Issue]{
		NewProperty("issue_primary_name", func(i *entity.Issue) (any, bool) { return i.PrimaryName, i.PrimaryName != "" }),
		NewProperty("issue_type", func(i *entity.Issue) (any, bool) { return i.Type, i.Type != "" }),
		NewProperty("issue_description", func(i *entity.Issue) (any, bool) { return i.Description, i.Description != "" }),
		NewProperty("issue_created_by", func(i *entity.Issue) (any, bool) { return i.CreatedBy, NoUpdate }),
		NewProperty("issue_updated_by", func(i *entity.Issue) (any, bool) { return i.UpdatedBy, i.UpdatedBy != 0 }),
	},
	FilterProperties: []*FilterProperty[*entity.IssueFilter]{
		NewFilterProperty("S.service_ccrn = ?", func(filter *entity.IssueFilter) any { return filter.ServiceCCRN }),
		NewFilterProperty("CI.componentinstance_service_id = ?", func(filter *entity.IssueFilter) any { return filter.ServiceId }),
		NewFilterProperty("I.issue_id = ?", func(filter *entity.IssueFilter) any { return filter.Id }),
		NewFilterProperty("IM.issuematch_status = ?", func(filter *entity.IssueFilter) any { return filter.IssueMatchStatus }),
		NewFilterProperty("IM.issuematch_rating = ?", func(filter *entity.IssueFilter) any { return filter.IssueMatchSeverity }),
		// Exclude soft-deleted IssueMatches when the IM table is joined.
		// The IM join (IM_RJ or IM_LJ) activates when HasIssueMatches or IssueMatchStatus is set.
		NewCustomFilterProperty(
			WrapBuilder(func(vals []bool) string {
				if len(vals) > 0 && vals[0] {
					return "IM.issuematch_deleted_at IS NULL"
				}
				return ""
			}),
			func(filter *entity.IssueFilter) any {
				if filter.HasIssueMatches || len(filter.IssueMatchStatus) > 0 || len(filter.IssueMatchId) > 0 || len(filter.IssueMatchSeverity) > 0 {
					return []bool{true}
				}
				return []bool{}
			},
		),
		NewFilterProperty("MVL.max_severity = ?", func(filter *entity.IssueFilter) any { return filter.MvSeverity }),
		NewFilterProperty("IM.issuematch_id = ?", func(filter *entity.IssueFilter) any { return filter.IssueMatchId }),
		NewFilterProperty("CVI.componentversionissue_component_version_id = ?", func(filter *entity.IssueFilter) any { return filter.ComponentVersionId }),
		NewFilterProperty("IV.issuevariant_id = ?", func(filter *entity.IssueFilter) any { return filter.IssueVariantId }),
		NewFilterProperty("I.issue_type = ?", func(filter *entity.IssueFilter) any { return filter.Type }),
		NewFilterProperty("I.issue_primary_name = ?", func(filter *entity.IssueFilter) any { return filter.PrimaryName }),
		NewFilterProperty("IV.issuevariant_repository_id = ?", func(filter *entity.IssueFilter) any { return filter.IssueRepositoryId }),
		NewFilterProperty("SG.supportgroup_ccrn = ?", func(filter *entity.IssueFilter) any { return filter.SupportGroupCCRN }),
		NewFilterProperty("CV.componentversion_component_id = ?", func(filter *entity.IssueFilter) any { return filter.ComponentId }),
		// When HasIssueMatches is set with ComponentId, restrict the IM path to only
		// include IssueMatches whose ComponentInstance belongs to a ComponentVersion
		// of the target component. Without this, IssueMatches from OTHER components
		// in the same service would leak through.
		// Self-contained subquery on IM — does not require CI to be joined.
		NewFilterProperty(
			"IM.issuematch_component_instance_id IN (SELECT ci.componentinstance_id FROM ComponentInstance ci JOIN ComponentVersion cv ON ci.componentinstance_component_version_id = cv.componentversion_id WHERE cv.componentversion_component_id = ?)",
			func(filter *entity.IssueFilter) any {
				if filter.HasIssueMatches {
					return filter.ComponentId
				}
				return []int64{}
			},
		),
		NewNFilterProperty(
			"IV.issuevariant_secondary_name LIKE Concat('%',?,'%') OR I.issue_primary_name LIKE Concat('%',?,'%')",
			func(filter *entity.IssueFilter) any { return filter.Search },
			2,
		),
		NewStateFilterProperty("I.issue", func(filter *entity.IssueFilter) any { return filter.State }),
		NewCustomFilterProperty(
			WrapBuilder(func(is []entity.IssueStatus) string {
				if len(is) != 1 {
					panic(fmt.Sprintf("Unexpected number of elements for IssueStatus: %d", len(is)))
				}
				switch is[0] {
				case entity.IssueStatusOpen:
					return "R.remediation_id IS NULL"
				case entity.IssueStatusRemediated:
					return "R.remediation_id IS NOT NULL"
				}
				return ""
			}),
			func(filter *entity.IssueFilter) any { return []entity.IssueStatus{filter.Status} },
		),
	},
	JoinDefs: []*JoinDef[*entity.IssueFilter]{
		{
			Name:      "IM_RJ",
			Type:      RightJoin,
			Table:     "IssueMatch IM",
			On:        "I.issue_id = IM.issuematch_issue_id",
			Condition: func(f *entity.IssueFilter, _ *Order) bool { return f.HasIssueMatches },
		},
		{
			Name:  "IM_LJ",
			Type:  LeftJoin,
			Table: "IssueMatch IM",
			On:    "I.issue_id = IM.issuematch_issue_id",
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.IssueMatchStatus) > 0 || len(f.IssueMatchId) > 0 || len(f.IssueMatchSeverity) > 0
			},
		},
		{
			Name:      "CI with IM_RJ",
			Type:      LeftJoin,
			Table:     "ComponentInstance CI",
			On:        "IM.issuematch_component_instance_id = CI.componentinstance_id",
			DependsOn: []string{"IM_RJ"},
			Condition: DependentJoin[*entity.IssueFilter],
		},
		{
			Name:      "CI with IM_LJ",
			Type:      LeftJoin,
			Table:     "ComponentInstance CI",
			On:        "IM.issuematch_component_instance_id = CI.componentinstance_id",
			DependsOn: []string{"IM_LJ"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.ServiceId) > 0 && !f.UseMvVulnerabilityList
			},
		},
		{
			Name:      "CV with IM_RJ",
			Type:      LeftJoin,
			Table:     "ComponentVersion CV",
			On:        "CI.componentinstance_component_version_id = CV.componentversion_id",
			DependsOn: []string{"CI with IM_RJ"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool { return f.AllServices },
		}, // Looks like this case is not used because of mv_vulnerabilities
		{
			Name:      "CV with IM_LJ",
			Type:      LeftJoin,
			Table:     "ComponentVersion CV",
			On:        "CI.componentinstance_component_version_id = CV.componentversion_id",
			DependsOn: []string{"CI with IM_LJ"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				// Only activate when NOT using IM_RJ path — otherwise CV comes from CVI path
				return len(f.ServiceCCRN) > 0 && !f.UseMvVulnerabilityList && !f.HasIssueMatches
			},
		},
		{
			Name:      "S with IM_RJ",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			DependsOn: []string{"CI with IM_RJ"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				// Activate for service CCRN filtering when IM_RJ is the active path
				return f.HasIssueMatches && len(f.ServiceCCRN) > 0 || f.AllServices
			},
		},
		{
			Name:      "S with IM_LJ",
			Type:      LeftJoin,
			Table:     "Service S",
			On:        "CI.componentinstance_service_id = S.service_id",
			DependsOn: []string{"CI with IM_LJ"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				// Only activate when NOT using IM_RJ path
				return len(f.ServiceCCRN) > 0 && !f.UseMvVulnerabilityList && !f.HasIssueMatches
			},
		},
		{
			Name:      "SGS",
			Type:      LeftJoin,
			Table:     "SupportGroupService SGS",
			On:        "CI.componentinstance_service_id = SGS.supportgroupservice_service_id",
			DependsOn: []string{"CI with IM_LJ"},
			Condition: DependentJoin[*entity.IssueFilter],
		},
		{
			Name:      "SG",
			Type:      LeftJoin,
			Table:     "SupportGroup SG",
			On:        "SGS.supportgroupservice_support_group_id = SG.supportgroup_id",
			DependsOn: []string{"SGS"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.SupportGroupCCRN) > 0 && !f.UseMvVulnerabilityList
			},
		},
		{
			Name:  "IV",
			Type:  LeftJoin,
			Table: "IssueVariant IV",
			On:    "I.issue_id = IV.issuevariant_issue_id",
			Condition: func(f *entity.IssueFilter, order *Order) bool {
				return len(f.IssueVariantId) > 0 || len(f.IssueRepositoryId) > 0 || len(f.Search) > 0 || order.ByRating()
			},
		},
		{
			Name:  "CVI",
			Type:  LeftJoin,
			Table: "ComponentVersionIssue CVI",
			On:    "I.issue_id = CVI.componentversionissue_issue_id",
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				return len(f.ComponentVersionId) > 0 || len(f.ComponentId) > 0
			},
		},
		{
			Name:      "CV using CVI",
			Type:      LeftJoin,
			Table:     "ComponentVersion CV",
			On:        "CVI.componentversionissue_component_version_id = CV.componentversion_id",
			DependsOn: []string{"CVI"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				// Activate when ComponentId is set AND the ServiceCCRN path won't create
				// a conflicting CV alias. When HasIssueMatches is true, service filtering
				// goes through IM_RJ→CI→S (not IM_LJ→CI→CV), so no alias collision.
				noConflictingServicePath := len(f.ServiceCCRN) == 0 || f.HasIssueMatches
				return len(f.ComponentId) > 0 && noConflictingServicePath && len(f.SupportGroupCCRN) == 0 && !f.AllServices
			},
		},
		{
			Name:      "R has S and C",
			Type:      LeftJoin,
			Table:     "Remediation R",
			On:        "I.issue_id = R.remediation_issue_id AND R.remediation_deleted_at IS NULL AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE()) AND CI.componentinstance_service_id = R.remediation_service_id AND CV.componentversion_component_id = R.remediation_component_id",
			DependsOn: []string{"CI with IM_LJ", "CV using CVI"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				hasService := len(f.ServiceCCRN) > 0 || len(f.ServiceId) > 0
				hasComponent := len(f.ComponentId) > 0
				return (f.Status == entity.IssueStatusOpen || f.Status == entity.IssueStatusRemediated) && hasService && hasComponent
			},
		}, // Missing test
		{
			Name:      "R has S",
			Type:      LeftJoin,
			Table:     "Remediation R",
			On:        "I.issue_id = R.remediation_issue_id AND R.remediation_deleted_at IS NULL AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE()) AND CI.componentinstance_service_id = R.remediation_service_id",
			DependsOn: []string{"CI with IM_LJ"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				hasService := len(f.ServiceCCRN) > 0 || len(f.ServiceId) > 0
				hasComponent := len(f.ComponentId) > 0
				return (f.Status == entity.IssueStatusOpen || f.Status == entity.IssueStatusRemediated) && hasService && !hasComponent
			},
		}, // Missing test
		{
			Name:      "R has C",
			Type:      LeftJoin,
			Table:     "Remediation R",
			On:        "I.issue_id = R.remediation_issue_id AND R.remediation_deleted_at IS NULL AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE()) AND CV.componentversion_component_id = R.remediation_component_id",
			DependsOn: []string{"CV using CVI"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				hasService := len(f.ServiceCCRN) > 0 || len(f.ServiceId) > 0
				hasComponent := len(f.ComponentId) > 0
				return (f.Status == entity.IssueStatusOpen || f.Status == entity.IssueStatusRemediated) && !hasService && hasComponent
			},
		}, // Missing test
		{
			Name:  "R",
			Type:  LeftJoin,
			Table: "Remediation R",
			On:    "I.issue_id = R.remediation_issue_id AND R.remediation_deleted_at IS NULL AND (R.remediation_expiration_date IS NULL OR R.remediation_expiration_date >= CURDATE())",
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				hasService := len(f.ServiceCCRN) > 0 || len(f.ServiceId) > 0
				hasComponent := len(f.ComponentId) > 0
				return (f.Status == entity.IssueStatusOpen || f.Status == entity.IssueStatusRemediated) && !hasService && !hasComponent
			},
		},
		{
			Name:  "MVL",
			Type:  InnerJoin,
			Table: "(SELECT issue_id AS mvl_issue_id, max_severity, earliest_remediation_date, source_url FROM mvVulnerabilityList) MVL",
			On:    "I.issue_id = MVL.mvl_issue_id",
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				return f.UseMvVulnerabilityList
			},
		},
		{
			Name:      "MVS",
			Type:      InnerJoin,
			Table:     "(SELECT issue_id AS mvs_issue_id, service_id AS mvs_service_id FROM mvVulnerabilityService) MVS",
			On:        "I.issue_id = MVS.mvs_issue_id",
			DependsOn: []string{"MVL"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				return f.UseMvVulnerabilityList && (len(f.ServiceCCRN) > 0 || len(f.ServiceId) > 0 || len(f.SupportGroupCCRN) > 0)
			},
		},
		{
			Name:      "S with MVS",
			Type:      InnerJoin,
			Table:     "Service S",
			On:        "MVS.mvs_service_id = S.service_id",
			DependsOn: []string{"MVS"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				return f.UseMvVulnerabilityList && len(f.ServiceCCRN) > 0
			},
		},
		{
			Name:      "SGS with MVS",
			Type:      InnerJoin,
			Table:     "SupportGroupService SGS",
			On:        "MVS.mvs_service_id = SGS.supportgroupservice_service_id",
			DependsOn: []string{"MVS"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				return f.UseMvVulnerabilityList && len(f.SupportGroupCCRN) > 0
			},
		},
		{
			Name:      "SG with MVS",
			Type:      InnerJoin,
			Table:     "SupportGroup SG",
			On:        "SGS.supportgroupservice_support_group_id = SG.supportgroup_id",
			DependsOn: []string{"SGS with MVS"},
			Condition: func(f *entity.IssueFilter, _ *Order) bool {
				return f.UseMvVulnerabilityList && len(f.SupportGroupCCRN) > 0
			},
		},
	},
	Attributes: []Attr{{Name: "primary_name", Order: entity.Order{By: entity.IssuePrimaryName, Direction: entity.OrderDirectionAsc}}},
	ExtraColumnsSelector: func(f *entity.IssueFilter, order *Order) []string {
		var cols []string
		for _, o := range order.Sequence() {
			switch o.By {
			case entity.IssueVariantRating:
				cols = append(cols, "MAX(CAST(IV.issuevariant_rating AS UNSIGNED)) AS issuevariant_rating_num")
			case entity.IssueEarliestTargetRemediationDate:
				if f != nil && f.UseMvVulnerabilityList {
					// MV path: pre-aggregated by the materialized view.
					// COALESCE ensures non-NULL so the cursor can always encode this field.
					// Vulnerabilities with no date sort last (far-future sentinel > any real date).
					cols = append(cols, "COALESCE(MVL.earliest_remediation_date, CAST('9999-12-31 23:59:59' AS DATETIME)) AS issue_earliest_target_remediation_date")
				} else {
					// Non-MV path (image/imageVersion child queries): IM is RIGHT-joined and
					// the query groups by issue_id, so we must aggregate here too.
					// MIN gives a single deterministic value per group, matching MV semantics.
					cols = append(cols, "COALESCE(MIN(IM.issuematch_target_remediation_date), CAST('9999-12-31 23:59:59' AS DATETIME)) AS issue_earliest_target_remediation_date")
				}
			}
		}

		return cols
	},
	RowToData: func(e RowComposite, order []entity.Order) (*entity.Issue, string) {
		issue := e.IssueRow.AsIssue()

		var ivRating int64
		var earliestTargetRemediation sql.NullTime
		if e.IssueVariantRow != nil {
			ivRating = e.RatingNumerical.Int64
			earliestTargetRemediation = e.EarliestTargetRemediation
		}

		cursor, _ := EncodeCursor(WithIssue(order, issue, ivRating, earliestTargetRemediation))

		return &issue, cursor
	},
	NewResult: func(i *entity.Issue, cursor string) entity.IssueResult {
		return entity.IssueResult{
			WithCursor: entity.WithCursor{Value: cursor},
			Issue:      i,
		}
	},
}

func appendIssueColumns(s []string, filter *entity.IssueFilter, order []entity.Order) []string {
	for _, o := range order {
		switch o.By {
		case entity.IssueVariantRating:
			s = append(s, "MAX(CAST(IV.issuevariant_rating AS UNSIGNED)) AS issuevariant_rating_num")
		case entity.IssueEarliestTargetRemediationDate:
			if filter != nil && filter.UseMvVulnerabilityList {
				s = append(s, "COALESCE(MVL.earliest_remediation_date, CAST('9999-12-31 23:59:59' AS DATETIME)) AS issue_earliest_target_remediation_date")
			} else {
				s = append(s, "COALESCE(MIN(IM.issuematch_target_remediation_date), CAST('9999-12-31 23:59:59' AS DATETIME)) AS issue_earliest_target_remediation_date")
			}
		}
	}

	return s
}

func (s *SqlDatabase) buildIssueStatement(ctx context.Context, baseQuery sq.SelectBuilder, filter *entity.IssueFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []any, error) {
	statement := Statement[*entity.IssueFilter]{
		Db:         s.db,
		L:          l,
		Obj:        &issueObject,
		BaseQuery:  baseQuery,
		Order:      NewOrder(order, issueObject.DefaultOrder),
		WithCursor: withCursor,
	}

	return BuildStatement(ctx, statement, filter)
}

// TODO: use DbObject
func (s *SqlDatabase) GetIssuesWithAggregations(ctx context.Context, filter *entity.IssueFilter, order []entity.Order) ([]entity.IssueResult, error) {
	filter = EnsureFilter(filter)
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

	joins := issueObject.GetJoins_tmp(filter, NewOrder(order, entity.Order{})) // It seems that this join is redundant for baseAppQuery
	// We should improve testing and remove redundant joins from query

	cursorFields, err := DecodeCursor(filter.After)
	if err != nil {
		return nil, err
	}

	columns := strings.Join(appendIssueColumns([]string{}, filter, order), ",")
	ord := NewOrder(order, entity.Order{By: entity.IssueId, Direction: entity.OrderDirectionAsc})

	whereClause := issueObject.GetFilterWhereClause_tmp(filter, false)

	cursorQuery := CreateCursorQuery("", cursorFields)

	filterStr := issueObject.GetFilterQuery_tmp(filter)
	if filterStr != "" && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	ciQuery := fmt.Sprintf(baseCiQuery, columns, joins, whereClause, cursorQuery, ord.ToSql())
	aggQuery := fmt.Sprintf(baseAggQuery, columns, joins, whereClause, cursorQuery, ord.ToSql())
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
	filterParameters := issueObject.GetFilterParameters_tmp(filter, true, cursorFields)
	// parameters for agg query
	filterParameters = append(filterParameters, issueObject.GetFilterParameters_tmp(filter, true, cursorFields)...)

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

			var earliestTargetRemediation sql.NullTime

			if e.IssueVariantRow != nil {
				ivRating = e.RatingNumerical.Int64
				earliestTargetRemediation = e.EarliestTargetRemediation
			}

			cursor, _ := EncodeCursor(WithIssue(ord.Sequence(), issue.Issue, ivRating, earliestTargetRemediation))

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
	return issueObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) CountIssueTypes(ctx context.Context, filter *entity.IssueFilter) (*entity.IssueTypeCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssueTypes",
	})

	baseQuery := sq.Select("I.issue_type AS issue_value", "COUNT(distinct I.issue_id) as issue_count").From("Issue I").GroupBy("I.issue_type")

	stmt, filterParameters, err := s.buildIssueStatement(ctx, baseQuery, filter, false, []entity.Order{}, l)
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

func (s *SqlDatabase) GetAllIssueCursors(ctx context.Context, filter *entity.IssueFilter, order []entity.Order) ([]string, error) {
	return issueObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetIssues(ctx context.Context, filter *entity.IssueFilter, order []entity.Order) ([]entity.IssueResult, error) {
	return issueObject.Get(ctx, s.db, filter, order)
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
	return AssociateId(s.db, "ComponentVersionIssue", "componentversionissue", "issue", issueId, "component_version", componentVersionId)
}

func (s *SqlDatabase) RemoveComponentVersionFromIssue(issueId int64, componentVersionId int64) error {
	return DissociateId(s.db, "ComponentVersionIssue", "componentversionissue", "issue", issueId, "component_version", componentVersionId)
}

func (s *SqlDatabase) RemoveAllIssuesFromComponentVersion(componentVersionId int64) error {
	return DissociateAllIds(s.db, "ComponentVersionIssue", "componentversionissue", "component_version", componentVersionId)
}

func (s *SqlDatabase) GetIssueNames(ctx context.Context, filter *entity.IssueFilter) ([]string, error) {
	return issueObject.GetAttr(ctx, s.db, "primary_name", filter)
}
