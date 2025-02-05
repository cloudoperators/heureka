// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
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
	fl = append(fl, buildFilterQuery(filter.AffectedServiceCCRN, "S.service_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.SeverityValue, "IM.issuematch_rating = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Status, "IM.issuematch_status = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.SupportGroupCCRN, "SG.supportgroup_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.PrimaryName, "I.issue_primary_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ComponentCCRN, "C.component_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueType, "I.issue_type = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Search, wildCardFilterQuery, OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "IM.issuematch"))

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) getIssueMatchJoins(filter *entity.IssueMatchFilter) string {
	joins := ""

	if len(filter.Search) > 0 || len(filter.IssueType) > 0 || len(filter.PrimaryName) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN Issue I on I.issue_id = IM.issuematch_issue_id
		`)
		if len(filter.Search) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueVariant IV on IV.issuevariant_issue_id = I.issue_id
			`)
		}
	}

	if len(filter.EvidenceId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueMatchEvidence IME on IME.issuematchevidence_issue_match_id = IM.issuematch_id
		`)
	}

	if len(filter.AffectedServiceCCRN) > 0 || len(filter.SupportGroupCCRN) > 0 || len(filter.ComponentCCRN) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN ComponentInstance CI on CI.componentinstance_id = IM.issuematch_component_instance_id
			
		`)

		if len(filter.ComponentCCRN) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, `
                LEFT JOIN ComponentVersion CV on CV.componentversion_id = CI.componentinstance_component_version_id
				LEFT JOIN Component C on C.component_id = CV.componentversion_component_id
			`)
		}

		if len(filter.AffectedServiceCCRN) > 0 || len(filter.SupportGroupCCRN) > 0 {
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
	if issueMatch.UpdatedBy != 0 {
		fl = append(fl, "issuematch_updated_by = :issuematch_updated_by")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildIssueMatchStatement(baseQuery string, filter *entity.IssueMatchFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureIssueMatchFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getIssueMatchFilterString(filter)
	joins := s.getIssueMatchJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "IM.issuematch_id > ?")

	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
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
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentInstanceId)
	filterParameters = buildQueryParameters(filterParameters, filter.EvidenceId)
	filterParameters = buildQueryParameters(filterParameters, filter.AffectedServiceCCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.SeverityValue)
	filterParameters = buildQueryParameters(filterParameters, filter.Status)
	filterParameters = buildQueryParameters(filterParameters, filter.SupportGroupCCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.PrimaryName)
	filterParameters = buildQueryParameters(filterParameters, filter.ComponentCCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueType)
	filterParameters = buildQueryParametersCount(filterParameters, filter.Search, wildCardFilterParamCount)

	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllIssueMatchIds(filter *entity.IssueMatchFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssueMatches",
	})

	baseQuery := `
		SELECT IM.issuematch_id FROM IssueMatch IM 
		%s
	 	%s GROUP BY IM.issuematch_id ORDER BY IM.issuematch_id
    `

	stmt, filterParameters, err := s.buildIssueMatchStatement(baseQuery, filter, false, l)

	if err != nil {
		return nil, err
	}

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetIssueMatches(filter *entity.IssueMatchFilter) ([]entity.IssueMatch, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssueMatches",
	})

	baseQuery := `
		SELECT IM.* FROM IssueMatch IM 
		%s
	    %s %s GROUP BY IM.issuematch_id ORDER BY IM.issuematch_id LIMIT ?
    `

	stmt, filterParameters, err := s.buildIssueMatchStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

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
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.CountIssueMatches",
	})

	baseQuery := `
		SELECT count(distinct IM.issuematch_id) FROM IssueMatch IM 
		%s
		%s
    `

	stmt, filterParameters, err := s.buildIssueMatchStatement(baseQuery, filter, false, l)

	if err != nil {
		return -1, err
	}

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
			issuematch_issue_id,
			issuematch_created_by,
			issuematch_updated_by
		) VALUES (
			:issuematch_status,
			:issuematch_remediation_date,
			:issuematch_target_remediation_date,
			:issuematch_vector,
			:issuematch_rating,
			:issuematch_user_id,
			:issuematch_component_instance_id,
			:issuematch_issue_id,
			:issuematch_created_by,
			:issuematch_updated_by
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

func (s *SqlDatabase) DeleteIssueMatch(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteIssueMatch",
	})

	query := `
		UPDATE IssueMatch SET
		issuematch_deleted_at = NOW(),
		issuematch_updated_by = :userId
		WHERE issuematch_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) AddEvidenceToIssueMatch(issueMatchId int64, evidenceId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"issueMatchId": issueMatchId,
		"evidenceId":   evidenceId,
		"event":        "database.AddEvidenceToIssueMatch",
	})

	query := `
		INSERT INTO IssueMatchEvidence (
			issuematchevidence_issue_match_id,
			issuematchevidence_evidence_id
		) VALUES (
			:issuematchevidence_issue_match_id,
			:issuematchevidence_evidence_id
		)
	`

	args := map[string]interface{}{
		"issuematchevidence_issue_match_id": issueMatchId,
		"issuematchevidence_evidence_id":    evidenceId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) RemoveEvidenceFromIssueMatch(issueMatchId int64, evidenceId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"issueMatchId": issueMatchId,
		"evidenceId":   evidenceId,
		"event":        "database.RemoveEvidenceFromIssueMatch",
	})

	query := `
		DELETE FROM IssueMatchEvidence
		WHERE
			issuematchevidence_issue_match_id = :issuematchevidence_issue_match_id
			AND issuematchevidence_evidence_id = :issuematchevidence_evidence_id
	`

	args := map[string]interface{}{
		"issuematchevidence_issue_match_id": issueMatchId,
		"issuematchevidence_evidence_id":    evidenceId,
	}

	_, err := performExec(s, query, args, l)

	return err
}
