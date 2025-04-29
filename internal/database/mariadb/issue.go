// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/samber/lo"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

const (
	wildCardFilterQuery      = "IV.issuevariant_secondary_name LIKE Concat('%',?,'%') OR I.issue_primary_name LIKE Concat('%',?,'%')"
	wildCardFilterParamCount = 2
)

func (s *SqlDatabase) buildIssueFilterParameters(filter *entity.IssueFilter, withCursor bool, cursorFields []Field) []interface{} {
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceCCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchStatus)
	filterParameters = buildQueryParameters(filterParameters, filter.ActivityId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueVariantId)
	filterParameters = buildQueryParameters(filterParameters, filter.Type)
	filterParameters = buildQueryParameters(filterParameters, filter.PrimaryName)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueRepositoryId)
	filterParameters = buildQueryParameters(filterParameters, filter.SupportGroupCCRN)
	filterParameters = buildQueryParametersCount(filterParameters, filter.Search, wildCardFilterParamCount)
	if withCursor {
		p := CreateCursorParameters([]any{}, cursorFields)
		filterParameters = append(filterParameters, p...)
		if filter.PaginatedX.First == nil {
			filterParameters = append(filterParameters, 1000)
		} else {
			filterParameters = append(filterParameters, filter.PaginatedX.First)
		}
	}

	return filterParameters
}

func (s *SqlDatabase) getIssueFilterString(filter *entity.IssueFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.ServiceCCRN, "S.service_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "CI.componentinstance_service_id= ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Id, "I.issue_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueMatchStatus, "IM.issuematch_status = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ActivityId, "A.activity_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueMatchId, "IM.issuematch_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentVersionId, "CVI.componentversionissue_component_version_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueVariantId, "IV.issuevariant_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Type, "I.issue_type = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.PrimaryName, "I.issue_primary_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueRepositoryId, "IV.issuevariant_repository_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.SupportGroupCCRN, "SG.supportgroup_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Search, wildCardFilterQuery, OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "I.issue"))

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) getIssueJoins(filter *entity.IssueFilter, order []entity.Order) string {
	joins := ""
	orderByRating := lo.ContainsBy(order, func(o entity.Order) bool {
		return o.By == entity.IssueVariantRating
	})
	if len(filter.ActivityId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ActivityHasIssue AHI on I.issue_id = AHI.activityhasissue_issue_id
         	LEFT JOIN Activity A on AHI.activityhasissue_activity_id = A.activity_id
		`)
	}
	if len(filter.IssueMatchStatus) > 0 || len(filter.ServiceId) > 0 || len(filter.ServiceCCRN) > 0 || len(filter.IssueMatchId) > 0 || len(filter.SupportGroupCCRN) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
		`)
	}
	if filter.AllServices {
		joins = fmt.Sprintf("%s\n%s", joins, `
			RIGHT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
		`)
	}
	if len(filter.ServiceId) > 0 || len(filter.ServiceCCRN) > 0 || len(filter.SupportGroupCCRN) > 0 || filter.AllServices {

		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
		`)
		if len(filter.ServiceCCRN) > 0 || filter.AllServices {
			joins = fmt.Sprintf("%s\n%s", joins, `
				LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id
				LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
			`)
		}
		if len(filter.SupportGroupCCRN) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, `
				LEFT JOIN SupportGroupService SGS ON SGS.supportgroupservice_service_id = CI.componentinstance_service_id
				LEFT JOIN SupportGroup SG ON SGS.supportgroupservice_support_group_id = SG.supportgroup_id
			`)
		}
	}

	if len(filter.ComponentVersionId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentVersionIssue CVI ON I.issue_id = CVI.componentversionissue_issue_id
		`)
	}

	if len(filter.IssueRepositoryId) > 0 || len(filter.IssueVariantId) > 0 || len(filter.Search) > 0 || orderByRating {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueVariant IV ON I.issue_id = IV.issuevariant_issue_id
		`)
	}

	return joins
}

func (s *SqlDatabase) ensureIssueFilter(f *entity.IssueFilter) *entity.IssueFilter {
	var first = 1000
	var after string = ""
	if f == nil {
		return &entity.IssueFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			ServiceCCRN:                     nil,
			Id:                              nil,
			ActivityId:                      nil,
			IssueMatchStatus:                nil,
			IssueMatchDiscoveryDate:         nil,
			IssueMatchTargetRemediationDate: nil,
			IssueMatchId:                    nil,
			ServiceId:                       nil,
			ComponentVersionId:              nil,
			IssueVariantId:                  nil,
			Type:                            nil,
		}
	}

	if f.After == nil {
		f.After = &after
	}
	if f.First == nil {
		f.First = &first
	}
	return f
}

func (s *SqlDatabase) getIssueUpdateFields(issue *entity.Issue) string {
	fl := []string{}
	if issue.PrimaryName != "" {
		fl = append(fl, "issue_primary_name = :issue_primary_name")
	}
	if issue.Type != "" {
		fl = append(fl, "issue_type = :issue_type")
	}
	if issue.Description != "" {
		fl = append(fl, "issue_description = :issue_description")
	}
	if issue.UpdatedBy != 0 {
		fl = append(fl, "issue_updated_by = :issue_updated_by")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) getIssueColumns(order []entity.Order) string {
	columns := ""
	for _, o := range order {
		switch o.By {
		case entity.IssueVariantRating:
			columns = fmt.Sprintf("%s, MAX(CAST(IV.issuevariant_rating AS UNSIGNED)) AS issuevariant_rating_num", columns)
		}
	}
	return columns
}

func (s *SqlDatabase) buildIssueStatement(baseQuery string, filter *entity.IssueFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureIssueFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getIssueFilterString(filter)
	joins := s.getIssueJoins(filter, order)
	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, nil, err
	}

	cursorQuery := CreateCursorQuery("", cursorFields)
	columns := s.getIssueColumns(order)
	order = GetDefaultOrder(order, entity.IssueId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

	whereClause := ""
	if filterStr != "" {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && withCursor && cursorQuery != "" {
		cursorQuery = fmt.Sprintf("HAVING (%s)", cursorQuery)
	}

	// construct final query
	if withCursor {
		query = fmt.Sprintf(baseQuery, columns, joins, whereClause, cursorQuery, orderStr)
	} else {
		query = fmt.Sprintf(baseQuery, columns, joins, whereClause, orderStr)
	}

	//construct prepared statement and if where clause does exist add parameters
	var stmt *sqlx.Stmt
	stmt, err = s.db.Preparex(query)
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

	//adding parameters
	filterParameters := s.buildIssueFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetIssuesWithAggregations(filter *entity.IssueFilter, order []entity.Order) ([]entity.IssueResult, error) {
	filter = s.ensureIssueFilter(filter)
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

	baseAggQuery := `
		SELECT I.*,
		count(distinct issuematch_id) as agg_issue_matches,
		count(distinct activity_id) as agg_activities,
		count(distinct service_ccrn) as agg_affected_services,
		count(distinct componentversionissue_component_version_id) as agg_component_versions,
		min(issuematch_target_remediation_date) as agg_earliest_target_remediation_date,
		min(issuematch_created_at) agg_earliest_discovery_date
		%s
        FROM Issue I
        LEFT JOIN ActivityHasIssue AHI on I.issue_id = AHI.activityhasissue_issue_id
        LEFT JOIN Activity A on AHI.activityhasissue_activity_id = A.activity_id
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

	filter = s.ensureIssueFilter(filter)
	filterStr := s.getIssueFilterString(filter)
	joins := s.getIssueJoins(filter, order)
	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, err
	}

	cursorQuery := CreateCursorQuery("", cursorFields)
	columns := s.getIssueColumns(order)
	order = GetDefaultOrder(order, entity.IssueId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

	whereClause := ""
	if filterStr != "" {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	ciQuery := fmt.Sprintf(baseCiQuery, columns, joins, whereClause, cursorQuery, orderStr)
	aggQuery := fmt.Sprintf(baseAggQuery, columns, joins, whereClause, cursorQuery, orderStr)
	query := fmt.Sprintf(baseQuery, ciQuery, aggQuery)

	var stmt *sqlx.Stmt

	stmt, err = s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			}).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}

	// parameters for component instance query
	filterParameters := s.buildIssueFilterParameters(filter, true, cursorFields)
	// parameters for agg query
	filterParameters = append(filterParameters, s.buildIssueFilterParameters(filter, true, cursorFields)...)

	defer stmt.Close()

	return performListScan(
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
				ivRating = e.IssueVariantRow.RatingNumerical.Int64

			}

			cursor, _ := EncodeCursor(WithIssue(order, issue.Issue, ivRating))

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

func (s *SqlDatabase) CountIssues(filter *entity.IssueFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssues",
	})

	baseQuery := `
		SELECT count(distinct I.issue_id) %s FROM Issue I
		%s
		%s
		ORDER BY %s
	`
	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CountIssueTypes(filter *entity.IssueFilter) (*entity.IssueTypeCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssueTypes",
	})

	baseQuery := `
		SELECT I.issue_type AS issue_value, COUNT(distinct I.issue_id) as issue_count %s FROM Issue I
		%s
		%s
		GROUP BY I.issue_type ORDER BY %s
	`

	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	counts, err := performListScan(
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

func (s *SqlDatabase) CountIssueRatings(filter *entity.IssueFilter) (*entity.IssueSeverityCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssueRatings",
	})

	filter = s.ensureIssueFilter(filter)

	baseQuery := `
		SELECT IV.issuevariant_rating AS issue_value, %s AS issue_count FROM %s Issue I
		%s
		%s
		%s
		GROUP BY IV.issuevariant_rating ORDER BY %s
	`

	var countColumn string
	if filter.AllServices {
		// Count issues that appear in multiple services and in multiple component versions per service
		countColumn = "COUNT(distinct CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', S.service_id))"
	} else if len(filter.SupportGroupCCRN) > 0 {
		// Count issues that appear in multiple support groups
		countColumn = "COUNT(distinct CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id, ',', SGS.supportgroupservice_service_id, ',', SG.supportgroup_id))"
	} else if len(filter.ServiceCCRN) > 0 || len(filter.ServiceId) > 0 {
		// Count issues that appear in multiple component versions
		countColumn = "COUNT(distinct CONCAT(CI.componentinstance_component_version_id, ',', I.issue_id))"
	} else {
		countColumn = "COUNT(distinct IV.issuevariant_issue_id)"
	}

	baseQuery = fmt.Sprintf(baseQuery, countColumn, "%s", "%s", "%s", "%s", "%s")

	if len(filter.IssueRepositoryId) == 0 {
		baseQuery = fmt.Sprintf(baseQuery, "%s", "LEFT JOIN IssueVariant IV ON IV.issuevariant_issue_id = I.issue_id", "%s", "%s", "%s")
	}

	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	counts, err := performListScan(
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

	var issueSeverityCounts entity.IssueSeverityCounts
	for _, count := range counts {
		switch count.Value {
		case entity.SeverityValuesCritical.String():
			issueSeverityCounts.Critical = count.Count
		case entity.SeverityValuesHigh.String():
			issueSeverityCounts.High = count.Count
		case entity.SeverityValuesMedium.String():
			issueSeverityCounts.Medium = count.Count
		case entity.SeverityValuesLow.String():
			issueSeverityCounts.Low = count.Count
		case entity.SeverityValuesNone.String():
			issueSeverityCounts.None = count.Count
		}
		issueSeverityCounts.Total += count.Count
	}

	return &issueSeverityCounts, nil
}

func (s *SqlDatabase) GetAllIssueIds(filter *entity.IssueFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueIds",
	})

	baseQuery := `
		SELECT I.issue_id %s FROM Issue I 
		%s
	 	%s GROUP BY I.issue_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetAllIssueCursors(filter *entity.IssueFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllIssueCursors",
	})

	baseQuery := `
		SELECT I.* %s FROM Issue I 
		%s
	    %s GROUP BY I.issue_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, false, order, l)

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
		issue := row.IssueRow.AsIssue()
		var ivRating int64
		if row.IssueVariantRow != nil {
			ivRating = row.IssueVariantRow.RatingNumerical.Int64

		}

		cursor, _ := EncodeCursor(WithIssue(order, issue, ivRating))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetIssues(filter *entity.IssueFilter, order []entity.Order) ([]entity.IssueResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssues",
	})

	baseQuery := `
		SELECT I.* %s FROM Issue I
		%s
		%s
		GROUP BY I.issue_id %s ORDER BY %s LIMIT ?
    `

	filter = s.ensureIssueFilter(filter)

	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, true, order, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueResult, e RowComposite) []entity.IssueResult {
			issue := e.IssueRow.AsIssue()

			var ivRating int64
			if e.IssueVariantRow != nil {
				ivRating = e.IssueVariantRow.RatingNumerical.Int64
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
	l := logrus.WithFields(logrus.Fields{
		"issue": issue,
		"event": "database.CreateIssue",
	})

	query := `
		INSERT INTO Issue (
			issue_primary_name,
			issue_type,
			issue_description,
			issue_created_by,
			issue_updated_by
		) VALUES (
			:issue_primary_name,
			:issue_type,
			:issue_description,
			:issue_created_by,
			:issue_updated_by
		)
	`

	issueRow := IssueRow{}
	issueRow.FromIssue(issue)

	id, err := performInsert(s, query, issueRow, l)

	if err != nil {
		return nil, err
	}

	issue.Id = id

	return issue, nil
}

func (s *SqlDatabase) UpdateIssue(issue *entity.Issue) error {
	l := logrus.WithFields(logrus.Fields{
		"issue": issue,
		"event": "database.UpdateIssue",
	})

	baseQuery := `
		UPDATE Issue SET
		%s
		WHERE issue_id = :issue_id
	`

	updateFields := s.getIssueUpdateFields(issue)

	query := fmt.Sprintf(baseQuery, updateFields)

	issueRow := IssueRow{}
	issueRow.FromIssue(issue)

	_, err := performExec(s, query, issueRow, l)

	return err
}

func (s *SqlDatabase) DeleteIssue(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssue",
	})

	query := `
		UPDATE Issue SET
		issue_deleted_at = NOW(),
		issue_updated_by = :userId
		WHERE issue_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
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

	args := map[string]interface{}{
		"issue_id":             issueId,
		"component_version_id": componentVersionId,
	}

	_, err := performExec(s, query, args, l)

	if err != nil {
		if strings.HasPrefix(err.Error(), "Error 1062") {
			return database.NewDuplicateEntryDatabaseError(fmt.Sprintf("for adding ComponentVersion %d to Issue %d ", componentVersionId, issueId))
		}
	}

	return err
}

func (s *SqlDatabase) RemoveComponentVersionFromIssue(issueId int64, componentVersionId int64) error {
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

	args := map[string]interface{}{
		"issue_id":             issueId,
		"component_version_id": componentVersionId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) GetIssueNames(filter *entity.IssueFilter) ([]string, error) {
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
	filter = s.ensureIssueFilter(filter)

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, false, order, l)
	if err != nil {
		l.Error("Error preparing statement: ", err)
		return nil, err
	}
	defer stmt.Close()

	// Execute the query
	rows, err := stmt.Queryx(filterParameters...)
	if err != nil {
		l.Error("Error executing query: ", err)
		return nil, err
	}
	defer rows.Close()

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
