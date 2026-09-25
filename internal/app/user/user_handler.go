// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
)

type userHandler struct {
	common.BaseHandler[entity.UserResult, *entity.UserFilter]
}

func NewUserHandler(handlerContext common.HandlerContext) UserHandler {
	return &userHandler{
		BaseHandler: common.NewBaseHandler(handlerContext, common.BaseConfig[entity.UserResult, *entity.UserFilter]{
			Op:              appErrors.Op("userHandler"),
			Entity:          "Users",
			CursorEntity:    "UserCursors",
			CountEntity:     "UserCount",
			GetFn:           handlerContext.DB.GetUsers,
			CursorsFn:       handlerContext.DB.GetAllUserCursors,
			CountFn:         handlerContext.DB.CountUsers,
			Authz:           handlerContext.Authz,
			AuthzObjectType: openfga.TypeSupportGroup,
			AuthzApplyFn: func(f *entity.UserFilter, ids []*int64) {
				f.SupportGroupId = common.CombineFilterWithAccessibleIds(f.SupportGroupId, ids)
			},
			ListEventFn: func(f *entity.UserFilter, o *entity.ListOptions, r *entity.List[entity.UserResult]) any {
				return &ListUsersEvent{Filter: f, Options: o, Users: r}
			},
			DeleteFn:      handlerContext.DB.DeleteUser,
			DeleteEventFn: func(id int64) any { return &DeleteUserEvent{UserID: id} },
		}),
	}
}

func (u *userHandler) ListUsers(
	ctx context.Context,
	filter *entity.UserFilter,
	options *entity.ListOptions,
) (*entity.List[entity.UserResult], error) {
	return u.List(ctx, appErrors.CallerOp(), filter, options)
}

func (u *userHandler) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	op := appErrors.CallerOp()

	var err error

	user.CreatedBy, err = common.GetCurrentUserId(ctx, u.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "User", "", err)
	}

	user.UpdatedBy = user.CreatedBy

	existing, err := u.ListUsers(
		ctx,
		&entity.UserFilter{UniqueUserID: []*string{&user.UniqueUserID}},
		entity.NewListOptions(),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "User", "", err)
	}

	if len(existing.Elements) > 0 {
		return nil, appErrors.AlreadyExistsError(string(op), "User", user.UniqueUserID)
	}

	newUser, err := u.DB().CreateUser(user)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "User", "", err)
	}

	u.PushEvent(&CreateUserEvent{User: newUser})

	return newUser, nil
}

func (u *userHandler) UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	op := appErrors.CallerOp()
	id := strconv.FormatInt(user.Id, 10)

	var err error

	user.UpdatedBy, err = common.GetCurrentUserId(ctx, u.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "User", id, err)
	}

	if err = u.DB().UpdateUser(user); err != nil {
		return nil, appErrors.InternalError(string(op), "User", id, err)
	}

	result, err := u.ListUsers(
		ctx,
		&entity.UserFilter{Id: []*int64{&user.Id}},
		entity.NewListOptions(),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "User", id, err)
	}

	if len(result.Elements) != 1 {
		return nil, appErrors.E(op, "User", id, appErrors.Internal,
			fmt.Sprintf("unexpected result count: %d", len(result.Elements)))
	}

	u.PushEvent(&UpdateUserEvent{User: user})

	return result.Elements[0].User, nil
}

func (u *userHandler) DeleteUser(ctx context.Context, id int64) error {
	return u.Delete(ctx, id)
}

func (u *userHandler) ListUserNames(
	ctx context.Context,
	filter *entity.UserFilter,
	options *entity.ListOptions,
) ([]string, error) {
	op := appErrors.CallerOp()

	userNames, err := u.DB().GetUserNames(ctx, filter)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "UserNames", "", err)
	}

	u.PushEvent(&ListUserNamesEvent{Filter: filter, Options: options, Names: userNames})

	return userNames, nil
}

func (u *userHandler) ListUniqueUserIDs(
	ctx context.Context,
	filter *entity.UserFilter,
	options *entity.ListOptions,
) ([]string, error) {
	op := appErrors.CallerOp()

	uniqueUserIDs, err := u.DB().GetUniqueUserIDs(ctx, filter)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "UniqueUserIDs", "", err)
	}

	u.PushEvent(&ListUniqueUserIDsEvent{Filter: filter, Options: options, IDs: uniqueUserIDs})

	return uniqueUserIDs, nil
}

func (u *userHandler) ListUserNamesAndIds(
	ctx context.Context,
	filter *entity.UserFilter,
	options *entity.ListOptions,
) ([]string, []string, error) {
	op := appErrors.CallerOp()

	users, err := u.DB().GetUsers(ctx, filter, options.Order)
	if err != nil {
		return nil, nil, appErrors.InternalError(string(op), "Users", "", err)
	}

	names := []string{}
	ids := []string{}

	for _, user := range users {
		names = append(names, user.Name)
		ids = append(ids, user.UniqueUserID)
	}

	u.PushEvent(&ListUserNamesAndIdsEvent{Filter: filter, Options: options, Names: names, Ids: ids})

	return names, ids, nil
}
