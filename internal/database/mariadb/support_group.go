// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"errors"
	"fmt"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/go-sql-driver/mysql"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var supportGroupObject = DbObject[*entity.SupportGroup]{
	Prefix:    "supportgroup",
	TableName: "SupportGroup",
	Properties: []*Property{
		NewProperty(
			"supportgroup_ccrn",
			WrapAccess(
				func(sg *entity.SupportGroup) (string, bool) { return sg.CCRN, sg.CCRN != "" },
			),
		),
		NewProperty(
			"supportgroup_created_by",
			WrapAccess(
				func(sg *entity.SupportGroup) (int64, bool) { return sg.CreatedBy, NoUpdate },
			),
		),
		NewProperty(
			"supportgroup_updated_by",
			WrapAccess(
				func(sg *entity.SupportGroup) (int64, bool) { return sg.UpdatedBy, sg.UpdatedBy != 0 },
			),
		),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty(
			"SG.supportgroup_id = ?",
			WrapRetSlice(func(filter *entity.SupportGroupFilter) []*int64 { return filter.Id }),
		),
		NewFilterProperty(
			"SGS.supportgroupservice_service_id = ?",
			WrapRetSlice(
				func(filter *entity.SupportGroupFilter) []*int64 { return filter.ServiceId },
			),
		),
		NewFilterProperty(
			"SG.supportgroup_ccrn = ?",
			WrapRetSlice(func(filter *entity.SupportGroupFilter) []*string { return filter.CCRN }),
		),
		NewFilterProperty(
			"SGU.supportgroupuser_user_id = ?",
			WrapRetSlice(func(filter *entity.SupportGroupFilter) []*int64 { return filter.UserId }),
		),
		NewFilterProperty(
			"IM.issuematch_issue_id = ?",
			WrapRetSlice(
				func(filter *entity.SupportGroupFilter) []*int64 { return filter.IssueId },
			),
		),
		NewStateFilterProperty(
			"SG.supportgroup",
			WrapRetState(
				func(filter *entity.SupportGroupFilter) []entity.StateFilterType { return filter.State },
			),
		),
	},
	JoinDefs: []*JoinDef{
		{
			Name:  "SGS",
			Type:  InnerJoin,
			Table: "SupportGroupService SGS",
			On:    "SG.supportgroup_id = SGS.supportgroupservice_support_group_id",
			Condition: WrapJoinCondition(func(f *entity.SupportGroupFilter, _ []entity.Order) bool {
				return len(f.ServiceId) > 0
			}),
		},
		{
			Name:      "CI",
			Type:      InnerJoin,
			Table:     "ComponentInstance CI",
			On:        "SGS.supportgroupservice_service_id = CI.componentinstance_service_id",
			DependsOn: []string{"SGS"},
			Condition: DependentJoin,
		},
		{
			Name:      "IM",
			Type:      InnerJoin,
			Table:     "IssueMatch IM",
			On:        "CI.componentinstance_id = IM.issuematch_component_instance_id",
			DependsOn: []string{"CI"},
			Condition: WrapJoinCondition(func(f *entity.SupportGroupFilter, _ []entity.Order) bool {
				return len(f.IssueId) > 0
			}),
		},
		{
			Name:  "SGU",
			Type:  InnerJoin,
			Table: "SupportGroupUser SGU",
			On:    "SG.supportgroup_id = SGU.supportgroupuser_support_group_id",
			Condition: WrapJoinCondition(func(f *entity.SupportGroupFilter, _ []entity.Order) bool {
				return len(f.UserId) > 0
			}),
		},
	},
}

func ensureSupportGroupFilter(filter *entity.SupportGroupFilter) *entity.SupportGroupFilter {
	if filter == nil {
		filter = &entity.SupportGroupFilter{}
	}

	return EnsurePagination(filter)
}

func (s *SqlDatabase) buildSupportGroupStatement(
	baseQuery string,
	filter *entity.SupportGroupFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	var query string

	filter = ensureSupportGroupFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	cursorFields, err := DecodeCursor(filter.After)
	if err != nil {
		return nil, nil, err
	}

	cursorQuery := CreateCursorQuery("", cursorFields)
	order = GetDefaultOrder(order, entity.SupportGroupId, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)
	joins := supportGroupObject.GetJoins(filter, order)
	filterStr := supportGroupObject.GetFilterQuery(filter)

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

	// construct prepared statement and if where clause does exist add parameters
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

	filterParameters := supportGroupObject.GetFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllSupportGroupCursors(
	filter *entity.SupportGroupFilter,
	order []entity.Order,
) ([]string, error) {
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

func (s *SqlDatabase) GetSupportGroups(
	filter *entity.SupportGroupFilter,
	order []entity.Order,
) ([]entity.SupportGroupResult, error) {
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

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

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

	stmt, filterParameters, err := s.buildSupportGroupStatement(
		baseQuery,
		filter,
		false,
		[]entity.Order{},
		l,
	)
	if err != nil {
		return -1, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateSupportGroup(
	supportGroup *entity.SupportGroup,
) (*entity.SupportGroup, error) {
	return supportGroupObject.Create(s.db, supportGroup)
}

func (s *SqlDatabase) UpdateSupportGroup(supportGroup *entity.SupportGroup) error {
	return supportGroupObject.Update(s.db, supportGroup)
}

func (s *SqlDatabase) DeleteSupportGroup(id int64, userId int64) error {
	return supportGroupObject.Delete(s.db, id, userId)
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

	args := map[string]any{
		"service_id":       serviceId,
		"support_group_id": supportGroupId,
	}

	var mysqlErr *mysql.MySQLError

	_, err := performExec(s, query, args, l)
	if err != nil {
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == database.ErrCodeDuplicateEntry {
				return nil
			}
		}

		return err
	}

	return nil
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

	args := map[string]any{
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

	args := map[string]any{
		"user_id":          userId,
		"support_group_id": supportGroupId,
	}

	var mysqlErr *mysql.MySQLError

	_, err := performExec(s, query, args, l)
	if err != nil {
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == database.ErrCodeDuplicateEntry {
				return nil
			}
		}

		return err
	}

	return nil
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

	args := map[string]any{
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

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	// Execute the query
	rows, err := stmt.Queryx(filterParameters...)
	if err != nil {
		l.Error("Error executing query: ", err)
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			logrus.Warnf("error during close rows: %s", err)
		}
	}()

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
