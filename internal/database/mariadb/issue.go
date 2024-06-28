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

// buildGetIssuesStatement is building the prepared statement and its parameters from the provided filter
//
// The where clause is build as follows:
//   - Filter entries of the same type (array values) are combined with an "OR"
//   - Filter entries of different types are combined with "AND"
func (s *SqlDatabase) buildGetIssuesStatement(filter *entity.IssueFilter, aggregations []string) (*sqlx.Stmt, []interface{}, error) {
	l := logrus.WithFields(logrus.Fields{"filter": filter})

	baseQuery := `
		SELECT I.* %s FROM Issue I
		%s
		WHERE %s %s GROUP BY I.issue_id ORDER BY I.issue_id LIMIT ? 
	`

	filterStr := s.getIssueFilterString(filter)
	withAggreations := len(aggregations) > 0
	joins := s.getIssueJoins(filter, withAggreations)

	cursor := getCursor(filter.Paginated, filterStr, "I.issue_id > ?")

	ags := ""
	if len(aggregations) > 0 {
		ags = fmt.Sprintf(", %s", strings.Join(aggregations, ", "))
	}

	// construct final query
	query := fmt.Sprintf(baseQuery, ags, joins, filterStr, cursor.Statement)

	//construct a prepared statement and if where clause does exist add parameters
	var stmt *sqlx.Stmt
	var err error
	var filterParameters []interface{}

	stmt, err = s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
			}).Error(msg)
		return nil, nil, fmt.Errorf("%s", msg)
	}

	//adding parameters
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceName)
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchStatus)
	filterParameters = buildQueryParameters(filterParameters, filter.ActivityId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueVariantId)
	filterParameters = buildQueryParameters(filterParameters, filter.Type)
	filterParameters = append(filterParameters, cursor.Value)
	filterParameters = append(filterParameters, cursor.Limit)

	logrus.WithFields(logrus.Fields{
		"event":  "internal/database/mariadb/SqlDatabase/buildGetIssuesStatement/Success",
		"filter": filter,
		"query":  query,
	}).Debugf("Constructed Prepared Statment for GetIssues")

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) buildCountIssuesStatement(filter *entity.IssueFilter) (*sqlx.Stmt, []interface{}, error) {
	filter = s.ensureIssueFilter(filter)
	l := logrus.WithFields(logrus.Fields{"filter": filter})
	// Building the Base Query
	baseQuery := `
		SELECT count(distinct I.issue_id) FROM Issue I
		%s
		%s 
	`
	joins := s.getIssueJoins(filter, false)
	filterStr := s.getIssueFilterString(filter)

	if filterStr != "" {
		filterStr = fmt.Sprintf("WHERE %s", filterStr)
	}
	// construct final query
	query := fmt.Sprintf(baseQuery, joins, filterStr)

	//construct a prepared statement and if where clause does exist add parameters
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

	l.WithFields(logrus.Fields{
		"query": query,
	}).Debugf("Constructed prepared Statment for CountIssues")

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetIssuesWithAggregations(filter *entity.IssueFilter) ([]entity.IssueWithAggregations, error) {
	filter = s.ensureIssueFilter(filter)
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssuesWithAggregations",
	})
	aggregations := []string{
		"count(distinct issuematch_id) as agg_issue_matches",
		"count(distinct activity_id) as agg_activities",
		"count(distinct service_name) as agg_affected_services",
		"count(distinct componentversionissue_component_version_id) as agg_component_versions",
		"sum(componentinstance_count) as agg_affected_component_instances",
		"min(issuematch_target_remediation_date) as agg_earliest_target_remediation_date",
		"min(issuematch_created_at) agg_earliest_discovery_date",
	}

	stmt, filterParameters, err := s.buildGetIssuesStatement(filter, aggregations)

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

func (s *SqlDatabase) GetAllIssueIds(filter *entity.IssueFilter) ([]int64, error) {

	filter = s.ensureIssueFilter(filter)
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllIssueIds",
	})

	baseQuery := `
		SELECT I.issue_id FROM Issue I 
		%s
	 	%s GROUP BY I.issue_id ORDER BY I.issue_id
    `

	filterStr := s.getIssueFilterString(filter)
	if filterStr != "" {
		filterStr = fmt.Sprintf("WHERE %s", filterStr)
	}
	joins := s.getIssueJoins(filter, false)

	query := fmt.Sprintf(baseQuery, joins, filterStr)
	stmt, err := s.db.Preparex(query)
	if err != nil {
		msg := "Error while preparing Statement"
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
			}).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}
	defer stmt.Close()

	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceName)
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchStatus)
	filterParameters = buildQueryParameters(filterParameters, filter.ActivityId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentVersionId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueVariantId)
	filterParameters = buildQueryParameters(filterParameters, filter.Type)

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetIssues(filter *entity.IssueFilter) ([]entity.Issue, error) {
	filter = s.ensureIssueFilter(filter)
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "datanase.GetIssues",
	})

	//Statement preparation
	stmt, filterParameters, err := s.buildGetIssuesStatement(filter, nil)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
			}).Error(msg)
		return nil, fmt.Errorf("%s", msg)
	}
	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.Issue, e GetIssuesByRow) []entity.Issue {
			return append(l, e.AsIssue())
		},
	)
}

func (s *SqlDatabase) CountIssues(filter *entity.IssueFilter) (int64, error) {
	filter = s.ensureIssueFilter(filter)
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.CountIssues",
	})

	stmt, filterParameters, err := s.buildCountIssuesStatement(filter)

	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
			}).Error(msg)
		return -1, fmt.Errorf("%s", msg)
	}
	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}
