// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

var userObject = DbObject[*entity.User, *entity.UserFilter, entity.UserResult, *any]{
	Prefix:       "user",
	TableName:    "User",
	TableKey:     "U",
	DefaultOrder: entity.Order{By: entity.UserID, Direction: entity.OrderDirectionAsc},
	Properties: []*Property[*entity.User]{
		NewProperty("user_name", func(u *entity.User) (any, bool) { return u.Name, u.Name != "" }),
		NewProperty("user_unique_user_id", func(u *entity.User) (any, bool) { return u.UniqueUserID, u.UniqueUserID != "" }),
		NewProperty("user_type", func(u *entity.User) (any, bool) { return u.Type, u.Type != entity.InvalidUserType }),
		NewProperty("user_email", func(u *entity.User) (any, bool) { return u.Email, u.Email != "" }),
		NewProperty("user_created_by", func(u *entity.User) (any, bool) { return u.CreatedBy, NoUpdate }),
		NewProperty("user_updated_by", func(u *entity.User) (any, bool) { return u.UpdatedBy, u.UpdatedBy != 0 }),
	},
	FilterProperties: []*FilterProperty[*entity.UserFilter]{
		NewFilterProperty("U.user_id = ?", func(filter *entity.UserFilter) any { return filter.Id }),
		NewFilterProperty("U.user_name = ?", func(filter *entity.UserFilter) any { return filter.Name }),
		NewFilterProperty("U.user_unique_user_id = ?", func(filter *entity.UserFilter) any { return filter.UniqueUserID }),
		NewFilterProperty("U.user_type = ?", func(filter *entity.UserFilter) any { return filter.Type }),
		NewFilterProperty("U.user_email = ?", func(filter *entity.UserFilter) any { return filter.Email }),
		NewFilterProperty("SGU.supportgroupuser_support_group_id = ?", func(filter *entity.UserFilter) any { return filter.SupportGroupId }),
		NewFilterProperty("O.owner_service_id = ?", func(filter *entity.UserFilter) any { return filter.ServiceId }),
		NewStateFilterProperty("U.user", func(filter *entity.UserFilter) any { return filter.State }),
	},
	JoinDefs: []*JoinDef[*entity.UserFilter]{
		{
			Name:      "SGU",
			Type:      LeftJoin,
			Table:     "SupportGroupUser SGU",
			On:        "U.user_id = SGU.supportgroupuser_user_id",
			Condition: func(f *entity.UserFilter, _ *Order) bool { return len(f.SupportGroupId) > 0 },
		},
		{
			Name:      "O",
			Type:      LeftJoin,
			Table:     "Owner O",
			On:        "U.user_id = O.owner_user_id",
			Condition: func(f *entity.UserFilter, _ *Order) bool { return len(f.ServiceId) > 0 },
		},
	},
	Attributes: []Attr{
		{Name: "unique_user_id", Order: entity.Order{By: entity.UserUniqueUserID, Direction: entity.OrderDirectionAsc}},
		{Name: "name", Order: entity.Order{By: entity.UserName, Direction: entity.OrderDirectionAsc}},
	},
	RowToData: func(e RowComposite, order []entity.Order) (*entity.User, string) {
		u := e.AsUser()

		cursor, _ := EncodeCursor(WithUser(order, u))

		return &u, cursor
	},
	NewResult: func(u *entity.User, _ *any, cursor string) entity.UserResult {
		return entity.UserResult{
			WithCursor: entity.WithCursor{Value: cursor},
			User:       u,
		}
	},
}

func (s *SqlDatabase) GetAllUserIds(ctx context.Context, filter *entity.UserFilter) ([]int64, error) {
	return userObject.GetAllIds(ctx, s.db, filter)
}

func (s *SqlDatabase) GetAllUserCursors(ctx context.Context, filter *entity.UserFilter, order []entity.Order) ([]string, error) {
	return userObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetUsers(ctx context.Context, filter *entity.UserFilter, order []entity.Order) ([]entity.UserResult, error) {
	return userObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountUsers(ctx context.Context, filter *entity.UserFilter) (int64, error) {
	return userObject.Count(ctx, s.db, filter)
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
	return userObject.GetAttr(ctx, s.db, "name", filter)
}

func (s *SqlDatabase) GetUniqueUserIDs(ctx context.Context, filter *entity.UserFilter) ([]string, error) {
	return userObject.GetAttr(ctx, s.db, "unique_user_id", filter)
}
