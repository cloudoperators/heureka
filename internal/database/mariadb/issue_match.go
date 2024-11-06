// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"database/sql"
	"fmt"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
	"time"
)

// defaultIssueMatchCursor returns a default cursor for the issue match
func DefaultIssueMatchCursor() *string {
	return lo.ToPtr(MarshalCursor(RowComposite{
		IssueMatchRow: &IssueMatchRow{
			Id:                    sql.NullInt64{Valid: true, Int64: int64(0)},
			Rating:                sql.NullString{Valid: true, String: string(entity.SeverityValuesNone)},
			TargetRemediationDate: sql.NullTime{Valid: true, Time: time.Unix(0, 0)},
		},
		ComponentInstanceRow: &ComponentInstanceRow{
			CCRN: sql.NullString{Valid: true, String: "a"},
		},
	}))
}

func (s *SqlDatabase) ensureIssueMatchFilter(f *entity.IssueMatchFilter) *entity.IssueMatchFilter {
	if f != nil {
		if f.Cursor == nil {
			f.Cursor = DefaultIssueMatchCursor()
		}
		return f
	}

	var first = 1000
	return &entity.IssueMatchFilter{
		Paginated: entity.Paginated{
			First:  &first,
			Cursor: DefaultIssueMatchCursor(),
		},
		IssueId: nil,
	}
}

func (s *SqlDatabase) getIssueMatchFilterString(filter *entity.IssueMatchFilter, cursor *DatabaseCursor) string {
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
	fl = append(fl, "IM.issuematch_deleted_at IS NULL")
	if cursor != nil {
		fl = append(fl, cursor.Statement)
	}

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

// ensureIssueMatchOrderBy ensures that the order slice contains the id as last ordering element if the id is not already
// explicitly set to another position in the ordering.
// This is required to make the cursor based pagination works correctly and that at least one indexed key is in the ordering
func (s *SqlDatabase) ensureIssueMatchOrderBy(order []entity.Order) []entity.Order {
	if !lo.ContainsBy(order, func(o entity.Order) bool {
		return o.By == entity.IssueMatchOrderValuesId
	}) {
		order = append(order, entity.Order{
			By:        entity.IssueMatchOrderValuesId,
			Direction: entity.OrderDirectionValueAsc,
		})
	}
	return order
}

// getIssueMatchCursorField returns the database field for the given cursor field
// it does additional logic for special fields where case ordering is applied
func getIssueMatchCursorField(field string) (string, error) {
	structs := []interface{}{
		IssueMatchRow{},
		ComponentInstanceRow{},
	}
	for _, s := range structs {
		imrType := reflect.TypeOf(s)
		for i := 0; i < imrType.NumField(); i++ {
			structField := imrType.Field(i)
			cursorTag := structField.Tag.Get("cursor")
			if cursorTag == field {
				dbTag := structField.Tag.Get("db")
				if dbTag != "" {
					// Additional logic for special fields
					switch field {
					case string(entity.IssueMatchOrderValuesSeverity):
						return fmt.Sprintf(`CASE
							 WHEN IM.%[1]s = 'Critical' THEN 5
							 WHEN IM.%[1]s = 'High' THEN 4
							 WHEN IM.%[1]s = 'Medium' THEN 3
							 WHEN IM.%[1]s = 'Low' THEN 2
							 WHEN IM.%[1]s = 'None' THEN 1
							 ELSE 0
						 	END`, dbTag), nil
					default:
						return fmt.Sprintf("IM.%s", dbTag), nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("field %s not found", field)
}

func buildIssueMatchCursor(filter *entity.IssueMatchFilter, order []entity.Order) (*DatabaseCursor, error) {
	return buildCursor(filter.Cursor, filter.First, order, RowComposite{
		IssueMatchRow:        &IssueMatchRow{},
		ComponentInstanceRow: &ComponentInstanceRow{},
	}, getIssueMatchCursorField)
}

func (s *SqlDatabase) buildIssueMatchStatement(baseQuery string, filter *entity.IssueMatchFilter, withLimit bool, withOrder bool, order []entity.Order, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	var err error
	var cursor *DatabaseCursor
	filter = s.ensureIssueMatchFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	if withOrder {
		order = s.ensureIssueMatchOrderBy(order)
		l.WithFields(logrus.Fields{"order": order})
		cursor, err = buildIssueMatchCursor(filter, order)

		if err != nil {
			return nil, nil, err
		}
	}

	filterStr := s.getIssueMatchFilterString(filter, cursor)
	joins := s.getIssueMatchJoins(filter)

	whereClause := fmt.Sprintf("WHERE %s", filterStr)

	orderByClause := ""
	if withOrder {
		for i, o := range order {
			// add comma if not first element
			if i != 0 {
				orderByClause = fmt.Sprintf("%s, ", orderByClause)
			}
			cursorField, err := getCursorField(string(o.By), IssueMatchRow{}, "")
			if err != nil {
				return nil, nil, err
			}
			orderByClause = fmt.Sprintf("%s %s %s", orderByClause, cursorField, o.Direction)
		}
		query = fmt.Sprintf(baseQuery, joins, whereClause, orderByClause)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause)
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

	if withOrder {
		filterParameters = append(filterParameters, cursor.ParameterValues...)
	}

	if withLimit {
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllIssueMatchCursors(filter *entity.IssueMatchFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssueAllIssuematchCursors",
	})

	baseQuery := `
		SELECT IM.*, CI.componentinstance_ccrn FROM IssueMatch IM
		LEFT JOIN ComponentInstance CI on CI.componentinstance_id = IM.issuematch_component_instance_id
		%s
	 	%s GROUP BY IM.issuematch_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildIssueMatchStatement(baseQuery, filter, false, true, order, l)

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
		return MarshalCursor(row)
	}), nil
}

func (s *SqlDatabase) GetIssueMatches(filter *entity.IssueMatchFilter, order []entity.Order) ([]entity.IssueMatchResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetIssueMatches",
	})

	baseQuery := `
		SELECT IM.*, CI.componentinstance_ccrn FROM IssueMatch IM 
		LEFT JOIN ComponentInstance CI on CI.componentinstance_id = IM.issuematch_component_instance_id
		%s
	    %s GROUP BY IM.issuematch_id ORDER BY %s LIMIT ?
    `

	stmt, filterParameters, err := s.buildIssueMatchStatement(baseQuery, filter, true, true, order, l)

	if err != nil {
		return nil, err
	}

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.IssueMatchResult, e RowComposite) []entity.IssueMatchResult {
			imr := entity.IssueMatchResult{
				WithCursor: entity.WithCursor{Value: MarshalCursor(e)},
				IssueMatch: lo.ToPtr(e.AsIssueMatch()),
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
		SELECT count(distinct IM.issuematch_id) FROM IssueMatch IM 
		LEFT JOIN ComponentInstance CI on CI.componentinstance_id = IM.issuematch_component_instance_id
		%s
		%s
    `

	stmt, filterParameters, err := s.buildIssueMatchStatement(baseQuery, filter, false, false, nil, l)

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
