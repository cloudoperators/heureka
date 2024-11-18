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

func (s *SqlDatabase) getUserFilterString(filter *entity.UserFilter) string {
	var fl []string
	fl = append(fl, buildFilterQuery(filter.Id, "U.user_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Name, "U.user_name = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.UniqueUserID, "U.user_unique_user_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.Type, "U.user_type = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.SupportGroupId, "SGU.supportgroupuser_support_group_id = ?", OP_OR))
	fl = append(fl, buildFilterQuery(filter.ServiceId, "O.owner_service_id = ?", OP_OR))
	fl = append(fl, "U.user_deleted_at IS NULL")

	return combineFilterQueries(fl, OP_AND)
}

func (s *SqlDatabase) ensureUserFilter(f *entity.UserFilter) *entity.UserFilter {
	var first int = 1000
	var after int64 = 0
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

func (s *SqlDatabase) getUserUpdateFields(user *entity.User) string {
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

func (s *SqlDatabase) buildUserStatement(baseQuery string, filter *entity.UserFilter, withCursor bool, l *logrus.Entry) (*sqlx.Stmt, []interface{}, error) {
	var query string
	filter = s.ensureUserFilter(filter)
	l.WithFields(logrus.Fields{"filter": filter})

	filterStr := s.getUserFilterString(filter)
	joins := s.getUserJoins(filter)
	cursor := getCursor(filter.Paginated, filterStr, "U.user_id > ?")

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
	filterParameters = buildQueryParameters(filterParameters, filter.Name)
	filterParameters = buildQueryParameters(filterParameters, filter.UniqueUserID)
	filterParameters = buildQueryParameters(filterParameters, filter.Type)
	filterParameters = buildQueryParameters(filterParameters, filter.SupportGroupId)
	filterParameters = buildQueryParameters(filterParameters, filter.ServiceId)
	if withCursor {
		filterParameters = append(filterParameters, cursor.Value)
		filterParameters = append(filterParameters, cursor.Limit)
	}

	return stmt, filterParameters, nil
}

func (s *SqlDatabase) GetAllUserIds(filter *entity.UserFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetUserIds",
	})

	baseQuery := `
		SELECT U.user_id FROM User U 
		%s
	 	%s GROUP BY U.user_id ORDER BY U.user_id
    `

	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, false, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	return performIdScan(stmt, filterParameters, l)
}

func (s *SqlDatabase) GetUsers(filter *entity.UserFilter) ([]entity.User, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "database.GetUsers",
	})

	baseQuery := `
		SELECT U.* FROM User U
		%s
		%s
		%s GROUP BY U.user_id ORDER BY U.user_id LIMIT ?
    `

	filter = s.ensureUserFilter(filter)
	baseQuery = fmt.Sprintf(baseQuery, "%s", "%s", "%s")

	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, true, l)

	if err != nil {
		return nil, err
	}

	defer stmt.Close()
	return performListScan(
		stmt,
		filterParameters,
		l,
		func(l []entity.User, e UserRow) []entity.User {
			return append(l, e.AsUser())
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
	`
	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, false, l)

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
			user_created_by
		) VALUES (
			:user_name,
			:user_unique_user_id,
			:user_type,
			:user_created_by
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

	updateFields := s.getUserUpdateFields(user)

	query := fmt.Sprintf(baseQuery, updateFields)

	userRow := UserRow{}
	userRow.FromUser(user)

	_, err := performExec(s, query, userRow, l)

	return err
}

func (s *SqlDatabase) DeleteUser(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"id":    id,
		"event": "database.DeleteUser",
	})

	query := `
		UPDATE User SET
		user_deleted_at = NOW()
		WHERE user_id = :id
	`

	args := map[string]interface{}{
		"id": id,
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
    ORDER BY U.user_name
    `

	// Ensure the filter is initialized
	filter = s.ensureUserFilter(filter)

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, false, l)
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
    ORDER BY U.user_unique_user_id
    `

	// Ensure the filter is initialized
	filter = s.ensureUserFilter(filter)

	// Builds full statement with possible joins and filters
	stmt, filterParameters, err := s.buildUserStatement(baseQuery, filter, false, l)
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
