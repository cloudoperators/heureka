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

func (s *SqlDatabase) ensureIssueMatchFilter(f *entity.IssueMatchFilter) *entity.IssueMatchFilter {
	if f != nil {
		return f
	}

	var first = 1000
	var after int64 = 0
	return &entity.IssueMatchFilter{
		Paginated: entity.Paginated{
			First: &first,
			After: &after,
		},
		IssueId: nil,
	}
}

func (s *SqlDatabase) getIssueMatchFilterString(filter *entity.IssueMatchFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "IM.issuematch_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueId, "IM.issuematch_issue_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentInstanceId, "IM.issuematch_component_instance_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.EvidenceId, "IME.issuematchevidence_evidence_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.AffectedServiceName, "S.service_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.SeverityValue, "IM.issuematch_rating = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Status, "IM.issuematch_status = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.SupportGroupName, "SG.supportgroup_name = ?", OP_OR))
	fl = append(fl, "IM.issuematch_deleted_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) getIssueMatchJoins(filter *entity.IssueMatchFilter) string {
	joins := ""
	if len(filter.EvidenceId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueMatchEvidence IME on IME.issuematchevidence_issue_match_id = IM.issuematch_id
		`)
	}
	if len(filter.AffectedServiceName) > 0 || len(filter.SupportGroupName) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentInstance CI on CI.componentinstance_id = IM.issuematch_component_instance_id
			LEFT JOIN Service S on S.service_id = CI.componentinstance_service_id
		`)
	}
	if len(filter.SupportGroupName) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN SupportGroupService SGS on S.service_id = SGS.supportgroupservice_service_id
			LEFT JOIN SupportGroup SG on SG.supportgroup_id = SGS.supportgroupservice_support_group_id
		`)
	}
	return joins
}

func (s *SqlDatabase) getIssueMatchUpdateFields(issueMatch *entity.IssueMatch) string {
	fl := []string{}
	if issueMatch.Status != "" && issueMatch.Status != entity.IssueMatchStatusValuesNone {
		fl = append(fl, "issuematch_status = :issuematch_status")
	}
	if issueMatch.Severity.Cvss.Vector != "" {
		fl = append(fl, "issuematch_vector = :issuematch_vector")
	}
	if issueMatch.Severity.Value != "" {
		fl = append(fl, "issuematch_rating = :issuematch_rating")
	}
	if issueMatch.UserId != 0 {
		fl = append(fl, "issuematch_user_id = :issuematch_user_id")
	}
	if issueMatch.ComponentInstanceId != 0 {
		fl = append(fl, "issuematch_component_instance_id = :issuematch_component_instance_id")
	}
	if issueMatch.IssueId != 0 {
		fl = append(fl, "issuematch_issue_id = :issuematch_issue_id")
	}
	if !issueMatch.RemediationDate.IsZero() {
		fl = append(fl, "issuematch_remediation_date = :issuematch_remediation_date")
	}
	if !issueMatch.TargetRemediationDate.IsZero() {
		fl = append(fl, "issuematch_target_remediation_date = :issuematch_target_remediation_date")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) GetAllIssueMatchIds(filter *entity.IssueMatchFilter) ([]int64, error) {

	filter = s.ensureIssueMatchFilter(filter)
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssueMatches",
	})

	baseQuery := `
		SELECT IM.issuematch_id FROM IssueMatch IM 
		%s
	 	%s GROUP BY IM.issuematch_id ORDER BY IM.issuematch_id
    `

	filterStr := s.getIssueMatchFilterString(filter)
	if filterStr != "" {
		filterStr = fmt.Sprintf("WHERE %s", filterStr)
	}
	joins := s.getIssueMatchJoins(filter)

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
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentInstanceId)
	filterParameters = buildQueryParameters(filterParameters, filter.EvidenceId)
	filterParameters = buildQueryParameters(filterParameters, filter.AffectedServiceName)
	filterParameters = buildQueryParameters(filterParameters, filter.SeverityValue)
	filterParameters = buildQueryParameters(filterParameters, filter.Status)
	filterParameters = buildQueryParameters(filterParameters, filter.SupportGroupName)

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetIssueMatches(filter *entity.IssueMatchFilter) ([]entity.IssueMatch, error) {
	filter = s.ensureIssueMatchFilter(filter)
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssueMatches",
	})

	baseQuery := `
		SELECT IM.* FROM IssueMatch IM 
		%s
		WHERE %s %s GROUP BY IM.issuematch_id ORDER BY IM.issuematch_id LIMIT ?
    `

	filterStr := s.getIssueMatchFilterString(filter)
	joins := s.getIssueMatchJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "IM.issuematch_id > ?")

	query := fmt.Sprintf(baseQuery, joins, filterStr, cursor.Statement)
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
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentInstanceId)
	filterParameters = buildQueryParameters(filterParameters, filter.EvidenceId)
	filterParameters = buildQueryParameters(filterParameters, filter.AffectedServiceName)
	filterParameters = buildQueryParameters(filterParameters, filter.SeverityValue)
	filterParameters = buildQueryParameters(filterParameters, filter.Status)
	filterParameters = buildQueryParameters(filterParameters, filter.SupportGroupName)
	filterParameters = append(filterParameters, cursor.Value)
	filterParameters = append(filterParameters, cursor.Limit)

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueMatch, e IssueMatchRow) []entity.IssueMatch {
			return append(l, e.AsIssueMatch())
		},
	)
}

func (s *SqlDatabase) CountIssueMatches(filter *entity.IssueMatchFilter) (int64, error) {
	filter = s.ensureIssueMatchFilter(filter)
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.CountIssueMatches",
	})

	baseQuery := `
		SELECT count(distinct IM.issuematch_id) FROM IssueMatch IM 
		%s
		%s
    `

	filterStr := s.getIssueMatchFilterString(filter)
	joins := s.getIssueMatchJoins(filter)
	if filterStr != "" {
		filterStr = fmt.Sprintf("WHERE %s", filterStr)
	}

	query := fmt.Sprintf(baseQuery, joins, filterStr)

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
		return -1, fmt.Errorf("%s", msg)
	}
	defer stmt.Close()

	//adding parameters
	var filterParameters []interface{}
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentInstanceId)
	filterParameters = buildQueryParameters(filterParameters, filter.EvidenceId)
	filterParameters = buildQueryParameters(filterParameters, filter.AffectedServiceName)
	filterParameters = buildQueryParameters(filterParameters, filter.SeverityValue)
	filterParameters = buildQueryParameters(filterParameters, filter.Status)
	filterParameters = buildQueryParameters(filterParameters, filter.SupportGroupName)

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateIssueMatch(issueMatch *entity.IssueMatch) (*entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"issueMatch": issueMatch,
		"event":      "database.CreateIssueMatch",
	})

	query := `
		INSERT INTO IssueMatch (
			issuematch_status,
			issuematch_remediation_date,
			issuematch_target_remediation_date,
			issuematch_vector,
			issuematch_rating,
			issuematch_user_id,
			issuematch_component_instance_id,
			issuematch_issue_id
		) VALUES (
			:issuematch_status,
			:issuematch_remediation_date,
			:issuematch_target_remediation_date,
			:issuematch_vector,
			:issuematch_rating,
			:issuematch_user_id,
			:issuematch_component_instance_id,
			:issuematch_issue_id
		)
	`

	issueMatchRow := IssueMatchRow{}
	issueMatchRow.FromIssueMatch(issueMatch)

	id, err := performInsert(s, query, issueMatchRow, l)

	if err != nil {
		return nil, err
	}

	issueMatch.Id = id

	return issueMatch, nil
}

func (s *SqlDatabase) UpdateIssueMatch(issueMatch *entity.IssueMatch) error {
	l := logrus.WithFields(logrus.Fields{
		"issueMatch": issueMatch,
		"event":      "database.UpdateIssueMatch",
	})

	baseQuery := `
		UPDATE IssueMatch SET
		%s
		WHERE issuematch_id = :issuematch_id
	`

	updateFields := s.getIssueMatchUpdateFields(issueMatch)

	query := fmt.Sprintf(baseQuery, updateFields)

	issueMatchRow := IssueMatchRow{}
	issueMatchRow.FromIssueMatch(issueMatch)

	_, err := performExec(s, query, issueMatchRow, l)

	return err
}

func (s *SqlDatabase) DeleteIssueMatch(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssueMatch",
	})

	query := `
		UPDATE IssueMatch SET
		issuematch_deleted_at = NOW()
		WHERE issuematch_id = :id
	`

	args := map[string]interface{}{
		"id": id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
