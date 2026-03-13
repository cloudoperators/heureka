// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var userObject = DbObject[*entity.User]{
	Prefix:    "user",
	TableName: "User",
	Properties: []*Property{
		NewProperty("user_name", WrapAccess(func(u *entity.User) (string, bool) { return u.Name, u.Name != "" })),
		NewProperty("user_unique_user_id", WrapAccess(func(u *entity.User) (string, bool) { return u.UniqueUserID, u.UniqueUserID != "" })),
		NewProperty("user_type", WrapAccess(func(u *entity.User) (entity.UserType, bool) { return u.Type, u.Type != entity.InvalidUserType })),
		NewProperty("user_email", WrapAccess(func(u *entity.User) (string, bool) { return u.Email, u.Email != "" })),
		NewProperty("user_created_by", WrapAccess(func(u *entity.User) (int64, bool) { return u.CreatedBy, NoUpdate })),
		NewProperty("user_updated_by", WrapAccess(func(u *entity.User) (int64, bool) { return u.UpdatedBy, u.UpdatedBy != 0 })),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty("U.user_id = ?", WrapRetSlice(func(filter *entity.UserFilter) []*int64 { return filter.Id })),
		NewFilterProperty("U.user_name = ?", WrapRetSlice(func(filter *entity.UserFilter) []*string { return filter.Name })),
		NewFilterProperty("U.user_unique_user_id = ?", WrapRetSlice(func(filter *entity.UserFilter) []*string { return filter.UniqueUserID })),
		NewFilterProperty("U.user_type = ?", WrapRetSlice(func(filter *entity.UserFilter) []entity.UserType { return filter.Type })),
		NewFilterProperty("U.user_email = ?", WrapRetSlice(func(filter *entity.UserFilter) []*string { return filter.Email })),
		NewFilterProperty("SGU.supportgroupuser_support_group_id = ?", WrapRetSlice(func(filter *entity.UserFilter) []*int64 { return filter.SupportGroupId })),
		NewFilterProperty("O.owner_service_id = ?", WrapRetSlice(func(filter *entity.UserFilter) []*int64 { return filter.ServiceId })),
		NewStateFilterProperty("U.user", WrapRetState(func(filter *entity.UserFilter) []entity.StateFilterType { return filter.State })),
	},
}

func ensureUserFilter(filter *entity.UserFilter) *entity.UserFilter {
	if filter == nil {
		return &entity.UserFilter{}
	}
	return EnsurePagination(filter)
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

	joins := s.getUserJoins(filter)
	cursorFields, err := DecodeCursor(filter.Paginated.After)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode User cursor: %w", err)
	}

	cursorQuery := CreateCursorQuery("", cursorFields)

	order = GetDefaultOrder(order, entity.UserID, entity.OrderDirectionAsc)
	orderStr := CreateOrderString(order)

	filterStr := userObject.GetFilterQuery(filter)
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

	filterParameters := userObject.GetFilterParameters(filter, withCursor, cursorFields)

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
	return userObject.Create(s.db, user)
}

func (s *SqlDatabase) UpdateUser(user *entity.User) error {
	return userObject.Update(s.db, user)
}

func (s *SqlDatabase) DeleteUser(id int64, userId int64) error {
	return userObject.Delete(s.db, id, userId)
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
