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

func (s *SqlDatabase) ensureActivityFilter(f *entity.ActivityFilter) *entity.ActivityFilter {
	if f != nil {
		return f
	}

	var first = 1000
	var after int64 = 0
	return &entity.ActivityFilter{
		Paginated: entity.Paginated{
			First: &first,
			After: &after,
		},
		Id:          nil,
		Status:      nil,
		ServiceId:   nil,
		ServiceName: nil,
		IssueId:     nil,
		EvidenceId:  nil,
	}
}

func (s *SqlDatabase) getActivityJoins(filter *entity.ActivityFilter) string {
	joins := ""
	if len(filter.ServiceId) > 0 || len(filter.ServiceName) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN ActivityHasService AHS on A.activity_id = AHS.activityhasservice_activity_id")
	}
	if len(filter.ServiceName) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN Service S on AHS.activityhasservice_service_id = S.service_id")
	}
	if len(filter.EvidenceId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN Evidence E on E.evidence_activity_id = A.activity_id")
	}
	if len(filter.IssueId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, "LEFT JOIN ActivityHasIssue AHI on AHI.activityhasissue_activity_id = A.activity_id")
	}
	return joins
}

func (s *SqlDatabase) getActivityFilterString(filter *entity.ActivityFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "A.activity_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Status, "A.activity_status = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "AHS.activityhasservice_service_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceName, "S.service_name= ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.EvidenceId, "E.evidence_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueId, "AHI.activityhasissue_issue_id = ?", OP_OR))
	fl = append(fl, "A.activity_deleted_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) getActivityUpdateFields(activity *entity.Activity) string {
	fl := []string{}
	if activity.Status != "" {
		fl = append(fl, "activity_status = :activity_status")
	}
	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) buildActivityStatement(baseQuery string, filter *entity.ActivityFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureActivityFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getActivityFilterString(filter)
	joins := s.getActivityJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "A.activity_id > ?")

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
	filterParameters = buildQueryParameters(filterParameters, filter.Status)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceName)
	filterParameters = buildQueryParameters(filterParameters, filter.EvidenceId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllActivityIds(filter *entity.ActivityFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetActivityIds",
	})

	baseQuery := `
		SELECT A.activity_id FROM Activity A 
		%s
	 	%s GROUP BY A.activity_id ORDER BY A.activity_id
    `

	stmt, filterParameters, err := s.buildActivityStatement(baseQuery, filter, false, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetActivities(filter *entity.ActivityFilter) ([]entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetActivities",
	})

	baseQuery := `
		SELECT A.* FROM Activity A
		%s
		%s
		%s GROUP BY A.activity_id ORDER BY A.activity_id LIMIT ?
    `

	filter = s.ensureActivityFilter(filter)
	baseQuery = fmt.Sprintf(baseQuery, "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildActivityStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.Activity, e ActivityRow) []entity.Activity {
			return append(l, e.AsActivity())
		},
	)
}

func (s *SqlDatabase) CountActivities(filter *entity.ActivityFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountActivities",
	})

	baseQuery := `
		SELECT count(distinct A.activity_id) FROM Activity A
		%s
		%s
	`
	stmt, filterParameters, err := s.buildActivityStatement(baseQuery, filter, false, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateActivity(activity *entity.Activity) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"activity": activity,
		"event":    "database.CreateActivity",
	})

	query := `
		INSERT INTO Activity (
			activity_status
		) VALUES (
		 	:activity_status
		)
	`

	activityRow := ActivityRow{}
	activityRow.FromActivity(activity)

	id, err := performInsert(s, query, activityRow, l)

	if err != nil {
		return nil, err
	}

	activity.Id = id

	return activity, nil
}

func (s *SqlDatabase) UpdateActivity(activity *entity.Activity) error {
	l := logrus.WithFields(logrus.Fields{
		"activity": activity,
		"event":    "database.UpdateActivity",
	})

	baseQuery := `
		UPDATE Activity SET
		%s
		WHERE activity_id = :activity_id
	`

	updateFields := s.getActivityUpdateFields(activity)

	query := fmt.Sprintf(baseQuery, updateFields)

	activityRow := ActivityRow{}
	activityRow.FromActivity(activity)

	_, err := performExec(s, query, activityRow, l)

	return err
}

func (s *SqlDatabase) DeleteActivity(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteActivity",
	})

	query := `
		UPDATE Activity SET
		activity_deleted_at = NOW()
		WHERE activity_id = :id
	`

	args := map[string]interface{}{
		"id": id,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) AddServiceToActivity(activityId int64, serviceId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"serviceId":  serviceId,
		"activityId": activityId,
		"event":      "database.AddServiceToActivity",
	})

	query := `
		INSERT INTO ActivityHasService (
			activityhasservice_service_id,
			activityhasservice_activity_id
		) VALUES (
		 :service_id,
		 :activity_id
		)
	`

	args := map[string]interface{}{
		"service_id":  serviceId,
		"activity_id": activityId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) RemoveServiceFromActivity(activityId int64, serviceId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"serviceId":  serviceId,
		"activityId": activityId,
		"event":      "database.RemoveServiceFromActivity",
	})

	query := `
		DELETE FROM ActivityHasService
		WHERE activityhasservice_service_id = :service_id
		AND activityhasservice_activity_id = :activity_id
	`

	args := map[string]interface{}{
		"service_id":  serviceId,
		"activity_id": activityId,
	}

	_, err := performExec(s, query, args, l)

	return err
}
