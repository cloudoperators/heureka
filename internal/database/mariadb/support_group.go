// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"
	"strings"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func getSupportGroupFilterString(filter *entity.SupportGroupFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "SG.supportgroup_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "SGS.supportgroupservice_service_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.CCRN, "SG.supportgroup_ccrn = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.UserId, "SGU.supportgroupuser_user_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.IssueId, "IM.issuematch_issue_id = ?", OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "SG.supportgroup"))

	return combineFilterQueries(fl, OP_AND)
}

func getSupportGroupUpdateFields(supportGroup *entity.SupportGroup) string {
	fl := []string{}
	if supportGroup.CCRN != "" {
		fl = append(fl, "supportgroup_ccrn = :supportgroup_ccrn")
	}
	if supportGroup.UpdatedBy != 0 {
		fl = append(fl, "supportgroup_updated_by = :supportgroup_updated_by")
	}

	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) getSupportGroupJoins(filter *entity.SupportGroupFilter) string {
	joins := ""
	if len(filter.ServiceId) > 0 || len(filter.IssueId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, ` 
				INNER JOIN SupportGroupService SGS on SG.supportgroup_id = SGS.supportgroupservice_support_group_id
		`)
		if len(filter.IssueId) > 0 {
			joins = fmt.Sprintf("%s\n%s", joins, `
				INNER JOIN ComponentInstance CI on SGS.supportgroupservice_service_id = CI.componentinstance_service_id
				INNER JOIN IssueMatch IM on CI.componentinstance_id = IM.issuematch_component_instance_id
			`)
		}
	}
	if len(filter.UserId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, ` 
				INNER JOIN SupportGroupUser SGU on SG.supportgroup_id = SGU.supportgroupuser_support_group_id
		`)
	}
	return joins
}

func ensureSupportGroupFilter(f *entity.SupportGroupFilter) *entity.SupportGroupFilter {
	var first int = 1000
	var after string = ""
	if f == nil {
		return &entity.SupportGroupFilter{
			PaginatedX: entity.PaginatedX{
				First: &first,
				After: &after,
			},
			Id:        nil,
			ServiceId: nil,
			UserId:    nil,
			IssueId:   nil,
			CCRN:      nil,
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

func (s *SqlDatabase) buildSupportGroupStatement(baseQuery string, filter *entity.SupportGroupFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []interface{}, error) {
	var query string
	filter = ensureSupportGroupFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := getSupportGroupFilterString(filter)
	cursorFields, err := DecodeCursor(filter.PaginatedX.After)
	if err != nil {
		return nil, nil, err
	}
	cursorQuery := CreateCursorQuery("", cursorFields)
	order = GetDefaultOrder(order, entity.SupportGroupId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)
	joins := s.getSupportGroupJoins(filter)

	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && withCursor && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	// construct final query
	if withCursor {
		query = fmt.Sprintf(baseQuery, joins, whereClause, cursorQuery, orderStr)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause, orderStr)
	}

	//construct prepared statement and if where clause does exist add parameters
	stmt, err := s.db.Preparex(query)
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
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	filterParameters = buildQueryParameters(filterParameters, filter.CCRN)
	filterParameters = buildQueryParameters(filterParameters, filter.UserId)
	filterParameters = buildQueryParameters(filterParameters, filter.IssueId)
	if withCursor {
		filterParameters = append(filterParameters, GetCursorQueryParameters(filter.PaginatedX.First, cursorFields)...)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllSupportGroupIds(filter *entity.SupportGroupFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetSupportGroupIds",
	})

	baseQuery := `
		SELECT SG.supportgroup_id FROM SupportGroup SG 
		%s
	 	%s GROUP BY SG.supportgroup_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildSupportGroupStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetAllSupportGroupCursors(filter *entity.SupportGroupFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllSupportGroupCursors",
	})

	baseQuery := `
		SELECT SG.* FROM SupportGroup SG 
		%s
	    %s GROUP BY SG.supportgroup_id ORDER BY %s
    `

	filter = ensureSupportGroupFilter(filter)
	stmt, filterParameters, err := s.buildSupportGroupStatement(baseQuery, filter, false, order, l)

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
		sg := row.AsSupportGroup()

		cursor, _ := EncodeCursor(WithSupportGroup(order, sg))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetSupportGroups(filter *entity.SupportGroupFilter, order []entity.Order) ([]entity.SupportGroupResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetSupportGroups",
	})

	baseQuery := `
		SELECT SG.* FROM SupportGroup SG
		%s
		%s
		%s
		GROUP BY SG.supportgroup_id ORDER BY %s LIMIT ?
    `

	stmt, filterParameters, err := s.buildSupportGroupStatement(baseQuery, filter, true, order, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.SupportGroupResult, e RowComposite) []entity.SupportGroupResult {
			sg := e.AsSupportGroup()
			cursor, _ := EncodeCursor(WithSupportGroup(order, sg))

			sgr := entity.SupportGroupResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				SupportGroup: &sg,
			}
			return append(l, sgr)
		},
	)
}

func (s *SqlDatabase) CountSupportGroups(filter *entity.SupportGroupFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountSupportGroups",
	})

	baseQuery := `
		SELECT count(distinct SG.supportgroup_id) FROM SupportGroup SG
		%s
		%s
		ORDER BY %s
	`
	stmt, filterParameters, err := s.buildSupportGroupStatement(baseQuery, filter, false, []entity.Order{}, l)

	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateSupportGroup(supportGroup *entity.SupportGroup) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"supportGroup": supportGroup,
		"event":        "database.CreateSupportGroup",
	})

	query := `
		INSERT INTO SupportGroup (
			supportgroup_ccrn,
			supportgroup_created_by,
			supportgroup_updated_by
		) VALUES (
			:supportgroup_ccrn,
			:supportgroup_created_by,
			:supportgroup_updated_by
		)
	`

	supportGroupRow := SupportGroupRow{}
	supportGroupRow.FromSupportGroup(supportGroup)

	id, err := performInsert(s, query, supportGroupRow, l)

	if err != nil {
		return nil, err
	}

	supportGroup.Id = id

	return supportGroup, nil
}

func (s *SqlDatabase) UpdateSupportGroup(supportGroup *entity.SupportGroup) error {
	l := logrus.WithFields(logrus.Fields{
		"supportGroup": supportGroup,
		"event":        "database.UpdateSupportGroup",
	})

	baseQuery := `
		UPDATE SupportGroup SET
		%s
		WHERE supportgroup_id = :supportgroup_id
	`

	updateFields := getSupportGroupUpdateFields(supportGroup)

	query := fmt.Sprintf(baseQuery, updateFields)

	supportGroupRow := SupportGroupRow{}
	supportGroupRow.FromSupportGroup(supportGroup)

	_, err := performExec(s, query, supportGroupRow, l)

	return err
}

func (s *SqlDatabase) DeleteSupportGroup(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteSupportGroup",
	})

	query := `
		UPDATE SupportGroup SET
		supportgroup_deleted_at = NOW(),
		supportgroup_updated_by = :userId
		WHERE supportgroup_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) AddServiceToSupportGroup(supportGroupId int64, serviceId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
		"event":          "database.AddServiceToSupportGroup",
	})

	query := `
		INSERT INTO SupportGroupService (
			supportgroupservice_service_id,
			supportgroupservice_support_group_id
		) VALUES (
			:service_id,
			:support_group_id
		)
	`

	args := map[string]interface{}{
		"service_id":       serviceId,
		"support_group_id": supportGroupId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) RemoveServiceFromSupportGroup(supportGroupId int64, serviceId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
		"event":          "database.RemoveServiceFromSupportGroup",
	})

	query := `
		DELETE FROM SupportGroupService
		WHERE supportgroupservice_service_id = :service_id
		AND supportgroupservice_support_group_id = :support_group_id
	`

	args := map[string]interface{}{
		"service_id":       serviceId,
		"support_group_id": supportGroupId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) AddUserToSupportGroup(supportGroupId int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"userId":         userId,
		"supportGroupId": supportGroupId,
		"event":          "database.AddUserToSupportGroup",
	})

	query := `
		INSERT INTO SupportGroupUser (
			supportgroupuser_user_id,
			supportgroupuser_support_group_id
		) VALUES (
			:user_id,
			:support_group_id
		)
	`

	args := map[string]interface{}{
		"user_id":          userId,
		"support_group_id": supportGroupId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) RemoveUserFromSupportGroup(supportGroupId int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"userId":         userId,
		"supportGroupId": supportGroupId,
		"event":          "database.RemoveUserFromSupportGroup",
	})

	query := `
		DELETE FROM SupportGroupUser
		WHERE supportgroupuser_user_id = :user_id
		AND supportgroupuser_support_group_id = :support_group_id
	`

	args := map[string]interface{}{
		"user_id":          userId,
		"support_group_id": supportGroupId,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) GetSupportGroupCcrns(filter *entity.SupportGroupFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetSupportGroupCcrns",
	})

	baseQuery := `
    SELECT SG.supportgroup_ccrn FROM SupportGroup SG
    %s
    %s
    ORDER BY %s
    `

	order := []entity.Order{
		{
			By:        entity.SupportGroupCcrn,
			Direction: entity.OrderDirectionAsc,
		},
	}

	// Ensure the filter is initialized
	filter = ensureSupportGroupFilter(filter)

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildSupportGroupStatement(baseQuery, filter, false, order, l)
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
	supportGroupCcrns := []string{}
	var ccrn string
	for rows.Next() {
		if err := rows.Scan(&ccrn); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}
		supportGroupCcrns = append(supportGroupCcrns, ccrn)
	}
	if err = rows.Err(); err != nil {
		l.Error("Row iteration error: ", err)
		return nil, err
	}

	return supportGroupCcrns, nil
}
