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

func getUserFilterString(filter *entity.UserFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "U.user_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Name, "U.user_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.UniqueUserID, "U.user_unique_user_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Type, "U.user_type = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Email, "U.user_email = ?", OP_OR)) // Add this line
	fl = append(fl, buildFilterQuery(filter.SupportGroupId, "SGU.supportgroupuser_support_group_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "O.owner_service_id = ?", OP_OR))
	fl = append(fl, buildStateFilterQuery(filter.State, "U.user"))

	return combineFilterQueries(fl, OP_AND)
}

func buildUserFilterParameters(filter *entity.UserFilter, withCursor bool, cursorFields []Field) []any {
	var filterParameters []any
	filterParameters = buildQueryParameters(filterParameters, filter.Id)
	filterParameters = buildQueryParameters(filterParameters, filter.Name)
	filterParameters = buildQueryParameters(filterParameters, filter.UniqueUserID)
	filterParameters = buildQueryParameters(filterParameters, filter.Type)
	filterParameters = buildQueryParameters(filterParameters, filter.SupportGroupId)
	filterParameters = buildQueryParameters(filterParameters, filter.Email)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)

	if withCursor {
		filterParameters = append(filterParameters, GetCursorQueryParameters(filter.Paginated.First, cursorFields)...)
	}

	return filterParameters
}

func ensureUserFilter(f *entity.UserFilter) *entity.UserFilter {
	var first int = 1000
	var after string
	if f == nil {
		return &entity.UserFilter{
			Paginated: entity.Paginated{
				First: &first,
				After: &after,
			},
			Id:             nil,
			Name:           nil,
			UniqueUserID:   nil,
			Type:           nil,
			SupportGroupId: nil,
			ServiceId:      nil,
			Email:          nil, // Initialize Email filter
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

func getUserUpdateFields(user *entity.User) string {
	fl := []string{}
	if user.Name != "" {
		fl = append(fl, "user_name = :user_name")
	}
	if user.UniqueUserID != "" {
		fl = append(fl, "user_unique_user_id = :user_unique_user_id")
	}
	if user.Type != entity.InvalidUserType {
		fl = append(fl, "user_type = :user_type")
	}
	if user.UpdatedBy != 0 {
		fl = append(fl, "user_updated_by = :user_updated_by")
	}
	if user.Email != "" { // Add this condition
		fl = append(fl, "user_email = :user_email")
	}

	return strings.Join(fl, ", ")
}

func (s *SqlDatabase) getUserJoins(filter *entity.UserFilter) string {
	joins := ""
	if len(filter.SupportGroupId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN SupportGroupUser SGU on U.user_id = SGU.supportgroupuser_user_id
		`)
	}
	if len(filter.ServiceId) > 0 {
		joins = fmt.Sprintf("%s\n%s", joins, `
			LEFT JOIN Owner O on O.owner_user_id = U.user_id
		`)
	}
	return joins
}

func (s *SqlDatabase) buildUserStatement(baseQuery string, filter *entity.UserFilter, withCursor bool, order []entity.Order, l *logrus.Entry) (Stmt, []any, error) {
	filter = ensureUserFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := getUserFilterString(filter)
	joins := s.getUserJoins(filter)
	cursorFields, err := DecodeCursor(filter.Paginated.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode User cursor: %w", err)
	}

	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.UserID, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

	whereClause := ""
	if filterStr != "" || withCursor {
		whereClause = fmt.Sprintf("WHERE %s", filterStr)
	}

	if filterStr != "" && withCursor && cursorQuery != "" {
		cursorQuery = fmt.Sprintf(" AND (%s)", cursorQuery)
	}

	var query string
	if withCursor {
		query = fmt.Sprintf(baseQuery, joins, whereClause, cursorQuery, orderStr)
	} else {
		query = fmt.Sprintf(baseQuery, joins, whereClause, orderStr)
	}

	stmt, err := s.db.Preparex(query)
	if err != nil {
		msg := ERROR_MSG_PREPARED_STMT
		l.WithFields(
			logrus.Fields{
				"error": err,
				"query": query,
				"stmt":  stmt,
			}).Error(msg)
		return nil, nil, fmt.Errorf("failed to prepare User statement: %w", err)
	}

	filterParameters := buildUserFilterParameters(filter, withCursor, cursorFields)

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllUserIds(filter *entity.UserFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetUserIds",
	})

	baseQuery := `
		SELECT U.user_id FROM User U 
		%s
		%s
	 	GROUP BY U.user_id ORDER BY %s
    `

	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, false, []entity.Order{}, l)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetAllUserCursors(filter *entity.UserFilter, order []entity.Order) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllUserCursors",
	})

	baseQuery := `
		SELECT U.* FROM User U 
		%s
		%s GROUP BY U.user_id ORDER BY %s
	`

	filter = ensureUserFilter(filter)
	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build User cursor query: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error while close statement: %s", err.Error())
		}
	}()

	rows, err := performListScan(
		stmt,
		filterParameters,
		l,
		func(l []RowComposite, e RowComposite) []RowComposite {
			return append(l, e)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get User cursors: %w", err)
	}

	return lo.Map(rows, func(row RowComposite, _ int) string {
		r := row.AsUser()

		cursor, _ := EncodeCursor(WithUser(order, r))

		return cursor
	}), nil
}

func (s *SqlDatabase) GetUsers(filter *entity.UserFilter) ([]entity.UserResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetUsers",
	})

	baseQuery := `
		SELECT U.* FROM User U
		%s
		%s
		%s GROUP BY U.user_id ORDER BY %s LIMIT ?
    `

	filter = ensureUserFilter(filter)

	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, true, []entity.Order{}, l)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.UserResult, e UserRow) []entity.UserResult {
			u := e.AsUser()
			cursor, _ := EncodeCursor(WithUser([]entity.Order{}, u))

			ur := entity.UserResult{
				WithCursor: entity.WithCursor{
					Value: cursor,
				},
				User: &u,
			}

			return append(l, ur)
		},
	)
}

func (s *SqlDatabase) CountUsers(filter *entity.UserFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.CountUsers",
	})

	baseQuery := `
		SELECT count(distinct U.user_id) FROM User U
		%s
		%s
		ORDER BY %s
	`
	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, false, []entity.Order{}, l)
	if err != nil {
		return -1, err
	}

	defer stmt.Close()

	return performCountScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) CreateUser(user *entity.User) (*entity.User, error) {
	l := logrus.WithFields(logrus.Fields{
		"user":  user,
		"event": "database.CreateUser",
	})

	query := `
		INSERT INTO User (
			user_name,
			user_unique_user_id,
			user_type,
			user_email,
			user_created_by,
			user_updated_by
		) VALUES (
			:user_name,
			:user_unique_user_id,
			:user_type,
			:user_email,
			:user_created_by,
			:user_updated_by
		)
	`

	userRow := UserRow{}
	userRow.FromUser(user)

	id, err := performInsert(s, query, userRow, l)
	if err != nil {
		return nil, err
	}

	user.Id = id

	return user, nil
}

func (s *SqlDatabase) UpdateUser(user *entity.User) error {
	l := logrus.WithFields(logrus.Fields{
		"user":  user,
		"event": "database.UpdateUser",
	})

	baseQuery := `
		UPDATE User SET
		%s
		WHERE user_id = :user_id
	`

	updateFields := getUserUpdateFields(user)

	query := fmt.Sprintf(baseQuery, updateFields)

	userRow := UserRow{}
	userRow.FromUser(user)

	_, err := performExec(s, query, userRow, l)

	return err
}

func (s *SqlDatabase) DeleteUser(id int64, userId int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteUser",
	})

	query := `
		UPDATE User SET
		user_deleted_at = NOW(),
		user_updated_by = :userId
		WHERE user_id = :id
	`

	args := map[string]interface{}{
		"userId": userId,
		"id":     id,
	}

	_, err := performExec(s, query, args, l)

	return err
}

func (s *SqlDatabase) GetUserNames(filter *entity.UserFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetUserNames",
	})

	baseQuery := `
		SELECT U.user_name FROM User U
		%s
		%s
		ORDER BY %s
    `

	// Ensure the filter is initialized
	filter = ensureUserFilter(filter)

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, false, []entity.Order{
		{
			By: entity.UserName,
		},
	}, l)
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
	userNames := []string{}
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}
		userNames = append(userNames, name)
	}
	if err = rows.Err(); err != nil {
		l.Error("Row iteration error: ", err)
		return nil, err
	}

	return userNames, nil
}

func (s *SqlDatabase) GetUniqueUserIDs(filter *entity.UserFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetUniqueUserIDs",
	})

	baseQuery := `
		SELECT U.user_unique_user_id FROM User U
		%s
		%s
		ORDER BY %s
    `

	// Ensure the filter is initialized
	filter = ensureUserFilter(filter)

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, false, []entity.Order{
		{
			By: entity.UserUniqueUserID,
		},
	}, l)
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
	uniqueUserID := []string{}
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			l.Error("Error scanning row: ", err)
			continue
		}
		uniqueUserID = append(uniqueUserID, name)
	}
	if err = rows.Err(); err != nil {
		l.Error("Row iteration error: ", err)
		return nil, err
	}

	return uniqueUserID, nil
}
