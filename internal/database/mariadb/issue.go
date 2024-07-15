// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (s *SqlDatabase) getIssueFilterString(filter *entity.IssueFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.ServiceName, "S.service_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Id, "I.issue_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueMatchStatus, "IM.issuematch_status = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ActivityId, "A.activity_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueMatchId, "IM.issuematch_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentVersionId, "CVI.componentversionissue_component_version_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueVariantId, "IV.issuevariant_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Type, "I.issue_type = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.PrimaryName, "I.issue_primary_name = ?", OP_OR))
	fl = append(fl, "I.issue_deleted_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) getIssueJoins(filter *entity.IssueFilter, withAggregations bool) string {
	joins := ""
	if len(filter.ActivityId) > 0 || withAggregations {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ActivityHasIssue AHI on I.issue_id = AHI.activityhasissue_issue_id
         	LEFT JOIN Activity A on AHI.activityhasissue_activity_id = A.activity_id
		`)
	}
	if len(filter.IssueMatchStatus) > 0 || len(filter.ServiceName) > 0 || len(filter.IssueMatchId) > 0 || withAggregations {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueMatch IM ON I.issue_id = IM.issuematch_issue_id
		`)
	}
	if len(filter.ServiceName) > 0 || withAggregations {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentInstance CI ON CI.componentinstance_id = IM.issuematch_component_instance_id
			LEFT JOIN ComponentVersion CV ON CI.componentinstance_component_version_id = CV.componentversion_id
			LEFT JOIN Service S ON S.service_id = CI.componentinstance_service_id
		`)
	}

	if len(filter.ComponentVersionId) > 0 || withAggregations {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentVersionIssue CVI ON I.issue_id = CVI.componentversionissue_issue_id
		`)
	}

	if len(filter.IssueVariantId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueVariant IV ON I.issue_id = IV.issuevariant_issue_id
		`)
	}

	return joins
}

func (s *SqlDatabase) ensureIssueFilter(f *entity.IssueFilter) *entity.IssueFilter {
	var first = 1000
	var after int64 = 0
	if f == nil {
		return &entity.IssueFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			ServiceName:                     nil,
			Id:                              nil,
			ActivityId:                      nil,
			IssueMatchStatus:                nil,
			IssueMatchDiscoveryDate:         nil,
			IssueMatchTargetRemediationDate: nil,
			IssueMatchId:                    nil,
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
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildIssueStatement(baseQuery string, filter *entity.IssueFilter, aggregations []string, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureIssueFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getIssueFilterString(filter)
	withAggreations := len(aggregations) > 0
	joins := s.getIssueJoins(filter, withAggreations)
	cursor := getCursor(filter.Paginated, filterStr, "I.issue_id > ?")

	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	ags := ""
	if len(aggregations) > 0 {
		ags = fmt.Sprintf(", %s", strings.Join(aggregations, ", "))
		baseQuery = fmt.Sprintf(baseQuery, ags, "%s", "%s", "%s")
	}

	// construct final query
	if withCursor {
		query = fmt.Sprintf(baseQuery, joins, whereClause, cursor.Statement)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause)
	}

	//construct prepared statement and if where clause does exist add parameters
	var stmt *sqlx.Stmt
	var err error

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
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceName)
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchStatus)
	filterParameters = buildQueryParameters(filterParameters, filter.ActivityId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueVariantId)
	filterParameters = buildQueryParameters(filterParameters, filter.Type)
	filterParameters = buildQueryParameters(filterParameters, filter.PrimaryName)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetIssuesWithAggregations(filter *entity.IssueFilter) ([]entity.IssueWithAggregations, error) {
	filter = s.ensureIssueFilter(filter)
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssuesWithAggregations",
	})

	baseQuery := `
		SELECT I.* %s FROM Issue I
		%s
		%s
		%s GROUP BY I.issue_id ORDER BY I.issue_id LIMIT ?
	`

	aggregations := []string{
		"count(distinct issuematch_id) as agg_issue_matches",
		"count(distinct activity_id) as agg_activities",
		"count(distinct service_name) as agg_affected_services",
		"count(distinct componentversionissue_component_version_id) as agg_component_versions",
		"sum(componentinstance_count) as agg_affected_component_instances",
		"min(issuematch_target_remediation_date) as agg_earliest_target_remediation_date",
		"min(issuematch_created_at) agg_earliest_discovery_date",
	}

	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, aggregations, true, l)

	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error":        err,
				"aggregations": aggregations,
			}).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}
	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueWithAggregations, e GetIssuesByRow) []entity.IssueWithAggregations {
			return append(l, e.AsIssueWithAggregations())
		},
	)
}

func (s *SqlDatabase) CountIssues(filter *entity.IssueFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountIssues",
	})

	baseQuery := `
		SELECT count(distinct I.issue_id) FROM Issue I
		%s
		%s
	`
	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, []string{}, false, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetAllIssueIds(filter *entity.IssueFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssueIds",
	})

	baseQuery := `
		SELECT I.issue_id FROM Issue I 
		%s
	 	%s GROUP BY I.issue_id ORDER BY I.issue_id
    `

	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, []string{}, false, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetIssues(filter *entity.IssueFilter) ([]entity.Issue, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetIssues",
	})

	baseQuery := `
		SELECT I.* FROM Issue I
		%s
		%s
		%s GROUP BY I.issue_id ORDER BY I.issue_id LIMIT ?
    `

	filter = s.ensureIssueFilter(filter)

	stmt, filterParameters, err := s.buildIssueStatement(baseQuery, filter, []string{}, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.Issue, e IssueRow) []entity.Issue {
			return append(l, e.AsIssue())
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
			issue_description
		) VALUES (
			:issue_primary_name,
			:issue_type,
			:issue_description
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

func (s *SqlDatabase) DeleteIssue(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssue",
	})

	query := `
		UPDATE Issue SET
		issue_deleted_at = NOW()
		WHERE issue_id = :id
	`

	args := map[string]interface{}{
		"id": id,
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
