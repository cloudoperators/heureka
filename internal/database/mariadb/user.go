// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var userObject = DbObject[*entity.User]{
	Prefix:    "user",
	TableName: "User",
	Properties: []*Property{
		NewProperty(
			"user_name",
			WrapAccess(func(u *entity.User) (string, bool) { return u.Name, u.Name != "" }),
		),
		NewProperty(
			"user_unique_user_id",
			WrapAccess(
				func(u *entity.User) (string, bool) { return u.UniqueUserID, u.UniqueUserID != "" },
			),
		),
		NewProperty(
			"user_type",
			WrapAccess(
				func(u *entity.User) (entity.UserType, bool) { return u.Type, u.Type != entity.InvalidUserType },
			),
		),
		NewProperty(
			"user_email",
			WrapAccess(func(u *entity.User) (string, bool) { return u.Email, u.Email != "" }),
		),
		NewProperty(
			"user_created_by",
			WrapAccess(func(u *entity.User) (int64, bool) { return u.CreatedBy, NoUpdate }),
		),
		NewProperty(
			"user_updated_by",
			WrapAccess(func(u *entity.User) (int64, bool) { return u.UpdatedBy, u.UpdatedBy != 0 }),
		),
	},
	FilterProperties: []*FilterProperty{
		NewFilterProperty(
			"U.user_id = ?",
			WrapRetSlice(func(filter *entity.UserFilter) []*int64 { return filter.Id }),
		),
		NewFilterProperty(
			"U.user_name = ?",
			WrapRetSlice(func(filter *entity.UserFilter) []*string { return filter.Name }),
		),
		NewFilterProperty(
			"U.user_unique_user_id = ?",
			WrapRetSlice(func(filter *entity.UserFilter) []*string { return filter.UniqueUserID }),
		),
		NewFilterProperty(
			"U.user_type = ?",
			WrapRetSlice(func(filter *entity.UserFilter) []entity.UserType { return filter.Type }),
		),
		NewFilterProperty(
			"U.user_email = ?",
			WrapRetSlice(func(filter *entity.UserFilter) []*string { return filter.Email }),
		),
		NewFilterProperty(
			"SGU.supportgroupuser_support_group_id = ?",
			WrapRetSlice(func(filter *entity.UserFilter) []*int64 { return filter.SupportGroupId }),
		),
		NewFilterProperty(
			"O.owner_service_id = ?",
			WrapRetSlice(func(filter *entity.UserFilter) []*int64 { return filter.ServiceId }),
		),
		NewStateFilterProperty(
			"U.user",
			WrapRetState(
				func(filter *entity.UserFilter) []entity.StateFilterType { return filter.State },
			),
		),
	},
	JoinDefs: []*JoinDef{
		{
			Name:      "SGU",
			Type:      LeftJoin,
			Table:     "SupportGroupUser SGU",
			On:        "U.user_id = SGU.supportgroupuser_user_id",
			Condition: WrapJoinCondition(func(f *entity.UserFilter, _ *Order) bool { return len(f.SupportGroupId) > 0 }),
		},
		{
			Name:      "O",
			Type:      LeftJoin,
			Table:     "Owner O",
			On:        "U.user_id = O.owner_user_id",
			Condition: WrapJoinCondition(func(f *entity.UserFilter, _ *Order) bool { return len(f.ServiceId) > 0 }),
		},
	},
}

func (s *SqlDatabase) buildUserStatement(
	ctx context.Context,
	baseQuery sq.SelectBuilder,
	filter *entity.UserFilter,
	withCursor bool,
	order []entity.Order,
	l *logrus.Entry,
) (Stmt, []any, error) {
	statement := Statement{
		Db:         s.db,
		L:          l,
		Obj:        &userObject,
		BaseQuery:  baseQuery,
		Order:      NewOrder(order, entity.Order{By: entity.UserID, Direction: entity.OrderDirectionAsc}),
		WithCursor: withCursor,
		Aggregated: false,
	}

	return BuildStatement(ctx, statement, filter)
}

func (s *SqlDatabase) GetAllUserIds(ctx context.Context, filter *entity.UserFilter) ([]int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetUserIds",
	})

	baseQuery := sq.Select("U.user_id").From("User U").GroupBy("U.user_id")

	stmt, filterParameters, err := s.buildUserStatement(
		ctx,
		baseQuery,
		filter,
		false,
		[]entity.Order{},
		l,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	return performIdScan(ctx, stmt, filterParameters, l)
}

func (s *SqlDatabase) GetAllUserCursors(
	ctx context.Context,
	filter *entity.UserFilter,
	order []entity.Order,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetAllUserCursors",
	})

	baseQuery := sq.Select("U.*").From("User U").GroupBy("U.user_id")

	stmt, filterParameters, err := s.buildUserStatement(ctx, baseQuery, filter, false, order, l)
	if err != nil {
		return nil, fmt.Errorf("failed to build User cursor query: %w", err)
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			l.Warnf("error while close statement: %s", err.Error())
		}
	}()

	rows, err := performListScan(
		ctx,
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

func (s *SqlDatabase) GetUsers(ctx context.Context, filter *entity.UserFilter) ([]entity.UserResult, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetUsers",
	})

	baseQuery := sq.Select("U.*").From("User U").GroupBy("U.user_id")

	stmt, filterParameters, err := s.buildUserStatement(
		ctx,
		baseQuery,
		filter,
		true,
		[]entity.Order{},
		l,
	)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := stmt.Close(); err != nil {
			logrus.Warnf("error during close stmt: %s", err)
		}
	}()

	return performListScan(
		ctx,
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

func (s *SqlDatabase) CountUsers(ctx context.Context, filter *entity.UserFilter) (int64, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.CountUsers",
	})

	baseQuery := sq.Select("count(distinct U.user_id)").From("User U")

	stmt, filterParameters, err := s.buildUserStatement(
		ctx,
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

	return performCountScan(ctx, stmt, filterParameters, l)
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

func (s *SqlDatabase) GetUserNames(ctx context.Context, filter *entity.UserFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetUserNames",
	})

	baseQuery := sq.Select("U.user_name").From("User U")

	stmt, filterParameters, err := s.buildUserStatement(ctx, baseQuery, filter, false, []entity.Order{
		{
			By: entity.UserName,
		},
	}, l)
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
	rows, err := stmt.QueryxContext(ctx, filterParameters...)
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

func (s *SqlDatabase) GetUniqueUserIDs(ctx context.Context, filter *entity.UserFilter) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"event":  "database.GetUniqueUserIDs",
	})

	baseQuery := sq.Select("U.user_unique_user_id").From("User U")

	stmt, filterParameters, err := s.buildUserStatement(ctx, baseQuery, filter, false, []entity.Order{
		{
			By: entity.UserUniqueUserID,
		},
	}, l)
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
	rows, err := stmt.QueryxContext(ctx, filterParameters...)
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
