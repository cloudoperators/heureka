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

func (s *SqlDatabase) ensureEvidenceFilter(f *entity.EvidenceFilter) *entity.EvidenceFilter {
	var first int = 1000
	var after int64 = 0
	if f == nil {
		return &entity.EvidenceFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
		}
	}
	if f.First == nil {
		f.First = &first
	}
	if f.After == nil {
		f.After = &after
	}
	return f
}

func (s *SqlDatabase) getEvidenceJoins(filter *entity.EvidenceFilter) string {
	joins := ""
	if len(filter.IssueMatchId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN IssueMatchEvidence IME on IME.issuematchevidence_evidence_id = E.evidence_id
		`)
	}
	return joins
}

func (s *SqlDatabase) getEvidenceFilterString(filter *entity.EvidenceFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "E.evidence_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ActivityId, "E.evidence_activity_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.UserId, "E.author_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueMatchId, "IME.issuematchevidence_issue_match_id = ?", OP_OR))
	fl = append(fl, "E.evidence_deleted_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) getEvidenceUpdateFields(evidence *entity.Evidence) string {
	fl := []string{}
	if evidence.UserId != 0 {
		fl = append(fl, "evidence_author_id = :evidence_author_id")
	}
	if evidence.ActivityId != 0 {
		fl = append(fl, "evidence_activity_id = :evidence_activity_id")
	}
	if evidence.Type != "" {
		fl = append(fl, "evidence_type = :evidence_type")
	}
	if evidence.Description != "" {
		fl = append(fl, "evidence_description = :evidence_description")
	}
	if evidence.Severity.Cvss.Vector != "" {
		fl = append(fl, "evidence_vector = :evidence_vector")
	}
	if evidence.Severity.Value != "" {
		fl = append(fl, "evidence_rating = :evidence_rating")
	}
	if !evidence.RaaEnd.IsZero() {
		fl = append(fl, "evidence_raa_end = :evidence_raa_end")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildEvidenceStatement(baseQuery string, filter *entity.EvidenceFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureEvidenceFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getEvidenceFilterString(filter)
	joins := s.getEvidenceJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "E.evidence_id > ?")

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
	filterParameters = buildQueryParameters(filterParameters, filter.ActivityId)
	filterParameters = buildQueryParameters(filterParameters, filter.UserId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueMatchId)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllEvidenceIds(filter *entity.EvidenceFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetEvidenceIds",
	})

	baseQuery := `
		SELECT E.evidence_id FROM Evidence E 
		%s
	 	%s GROUP BY E.evidence_id ORDER BY E.evidence_id
    `

	stmt, filterParameters, err := s.buildEvidenceStatement(baseQuery, filter, false, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetEvidences(filter *entity.EvidenceFilter) ([]entity.Evidence, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetEvidences",
	})

	baseQuery := `
		SELECT E.* FROM Evidence E
		%s
		%s
		%s GROUP BY E.evidence_id ORDER BY E.evidence_id LIMIT ?
    `

	filter = s.ensureEvidenceFilter(filter)
	baseQuery = fmt.Sprintf(baseQuery, "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildEvidenceStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.Evidence, e EvidenceRow) []entity.Evidence {
			return append(l, e.AsEvidence())
		},
	)
}

func (s *SqlDatabase) CountEvidences(filter *entity.EvidenceFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountEvidences",
	})

	baseQuery := `
		SELECT count(distinct E.evidence_id) FROM Evidence E
		%s
		%s
	`
	stmt, filterParameters, err := s.buildEvidenceStatement(baseQuery, filter, false, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateEvidence(evidence *entity.Evidence) (*entity.Evidence, error) {
	l := logrus.WithFields(logrus.Fields{
		"evidence": evidence,
		"event":    "database.CreateEvidence",
	})

	query := `
		INSERT INTO Evidence (
			evidence_author_id,
			evidence_activity_id,
			evidence_type,
			evidence_description,
			evidence_vector,
			evidence_rating,
			evidence_raa_end
		) VALUES (
			:evidence_author_id,
			:evidence_activity_id,
			:evidence_type,
			:evidence_description,
			:evidence_vector,
			:evidence_rating,
			:evidence_raa_end
		)
	`

	evidenceRow := EvidenceRow{}
	evidenceRow.FromEvidence(evidence)

	id, err := performInsert(s, query, evidenceRow, l)

	if err != nil {
		return nil, err
	}

	evidence.Id = id

	return evidence, nil
}

func (s *SqlDatabase) UpdateEvidence(evidence *entity.Evidence) error {
	l := logrus.WithFields(logrus.Fields{
		"evidence": evidence,
		"event":    "database.UpdateEvidence",
	})

	baseQuery := `
		UPDATE Evidence SET
		%s
		WHERE evidence_id = :evidence_id
	`

	updateFields := s.getEvidenceUpdateFields(evidence)

	query := fmt.Sprintf(baseQuery, updateFields)

	evidenceRow := EvidenceRow{}
	evidenceRow.FromEvidence(evidence)

	_, err := performExec(s, query, evidenceRow, l)

	return err
}

func (s *SqlDatabase) DeleteEvidence(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteEvidence",
	})

	query := `
		UPDATE Evidence SET
		evidence_deleted_at = NOW()
		WHERE evidence_id = :id
	`

	args := map[string]interface{}{
		"id": id,
	}

	_, err := performExec(s, query, args, l)

	return err
}
