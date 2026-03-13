// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var (
	CacheTtlGetAllUserCursors = 12 * time.Hour
	CacheTtlGetUsers          = 12 * time.Hour
)

type userHandler struct {
	database      database.Database
	cache         cache.Cache
	eventRegistry event.EventRegistry
	authz         openfga.Authorization
	logger        *logrus.Logger
}

func NewUserHandler(handlerContext common.HandlerContext) UserHandler {
	return &userHandler{
		database:      handlerContext.DB,
		cache:         handlerContext.Cache,
		eventRegistry: handlerContext.EventReg,
		authz:         handlerContext.Authz,
		logger:        logrus.New(),
	}
}

type UserHandlerError struct {
	msg string
}

func (e *UserHandlerError) Error() string {
	return fmt.Sprintf("UserHandlerError: %s", e.msg)
}

func NewUserHandlerError(msg string) *UserHandlerError {
	return &UserHandlerError{msg: msg}
}

func (u *userHandler) ListUsers(ctx context.Context, filter *entity.UserFilter, options *entity.ListOptions) (*entity.List[entity.UserResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	op := appErrors.Op("userHandler.ListUsers")

	common.EnsurePaginated(&filter.Paginated)

	// get current user id
	currentUserId, err := common.GetCurrentUserId(ctx, u.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Users", "", err)
		applog.LogError(u.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	// Authorization check
	accessibleSupportGroupIds, err := u.authz.GetListOfAccessibleObjectIds(openfga.UserId(fmt.Sprint(currentUserId)), openfga.TypeSupportGroup)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Users", "", err)
		applog.LogError(u.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	// Update the filter.Id based on accessibleSupportGroupIds
	filter.SupportGroupId = common.CombineFilterWithAccessibleIds(filter.SupportGroupId, accessibleSupportGroupIds)

	res, err := cache.CallCached[[]entity.UserResult](
		u.cache,
		CacheTtlGetUsers,
		"GetUsers",
		u.database.GetUsers,
		filter,
	)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Users", "", err)
		applog.LogError(u.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := cache.CallCached[[]string](
				u.cache,
				CacheTtlGetAllUserCursors,
				"GetAllUserCursors",
				u.database.GetAllUserCursors,
				filter,
				options.Order,
			)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), "Users", "", err)
				applog.LogError(u.logger, wrappedErr, logrus.Fields{
					"filter": filter,
				})
				return nil, wrappedErr
			}

			pageInfo = common.GetPageInfo(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = u.database.CountUsers(filter)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "Users", "", err)
			applog.LogError(u.logger, wrappedErr, logrus.Fields{
				"filter": filter,
			})
			return nil, wrappedErr
		}
	}
	ret := &entity.List[entity.UserResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	u.eventRegistry.PushEvent(&ListUsersEvent{Filter: filter, Options: options, Users: ret})

	return ret, nil
}

func (u *userHandler) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	f := &entity.UserFilter{
		UniqueUserID: []*string{&user.UniqueUserID},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateUserEventName,
		"object": user,
	})

	var err error
	user.CreatedBy, err = common.GetCurrentUserId(ctx, u.database)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while creating user (GetUserId).")
	}
	user.UpdatedBy = user.CreatedBy

	users, err := u.ListUsers(ctx, f, &entity.ListOptions{})
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while creating user.")
	}

	if len(users.Elements) > 0 {
		return nil, NewUserHandlerError(fmt.Sprintf("Duplicated entry %s for UniqueUserID.", user.UniqueUserID))
	}

	newUser, err := u.database.CreateUser(user)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while creating user.")
	}

	u.eventRegistry.PushEvent(&CreateUserEvent{User: newUser})

	return newUser, nil
}

func (u *userHandler) UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateUserEventName,
		"object": user,
	})

	var err error
	user.UpdatedBy, err = common.GetCurrentUserId(ctx, u.database)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while updating user (GetUserId).")
	}

	err = u.database.UpdateUser(user)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while updating user.")
	}

	userResult, err := u.ListUsers(ctx, &entity.UserFilter{Id: []*int64{&user.Id}}, &entity.ListOptions{})
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while retrieving updated user.")
	}

	if len(userResult.Elements) != 1 {
		l.Error(err)
		return nil, NewUserHandlerError("Multiple users found.")
	}

	u.eventRegistry.PushEvent(&UpdateUserEvent{User: user})

	return userResult.Elements[0].User, nil
}

func (u *userHandler) DeleteUser(ctx context.Context, id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteUserEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(ctx, u.database)
	if err != nil {
		l.Error(err)
		return NewUserHandlerError("Internal error while deleting user (GetUserId).")
	}

	err = u.database.DeleteUser(id, userId)
	if err != nil {
		l.Error(err)
		return NewUserHandlerError("Internal error while deleting user.")
	}

	u.eventRegistry.PushEvent(&DeleteUserEvent{UserID: id})

	return nil
}

func (u *userHandler) ListUserNames(filter *entity.UserFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListUserNamesEventName,
		"filter": filter,
	})

	userNames, err := u.database.GetUserNames(filter)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while retrieving userNames.")
	}

	u.eventRegistry.PushEvent(&ListUserNamesEvent{Filter: filter, Options: options, Names: userNames})

	return userNames, nil
}

func (u *userHandler) ListUniqueUserIDs(filter *entity.UserFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListUniqueUserIDsEventName,
		"filter": filter,
	})

	uniqueUserID, err := u.database.GetUniqueUserIDs(filter)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while retrieving uniqueUserID.")
	}

	u.eventRegistry.PushEvent(&ListUniqueUserIDsEvent{Filter: filter, Options: options, IDs: uniqueUserID})

	return uniqueUserID, nil
}

func (u *userHandler) ListUserNamesAndIds(filter *entity.UserFilter, options *entity.ListOptions) ([]string, []string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListUserNamesAndIdsEventName,
		"filter": filter,
	})

	users, err := u.database.GetUsers(filter)
	if err != nil {
		l.Error(err)
		return nil, nil, NewUserHandlerError("Internal error while retrieving user.")
	}
	names := []string{}
	ids := []string{}
	for _, u := range users {
		names = append(names, u.Name)
		ids = append(ids, u.UniqueUserID)
	}
	u.eventRegistry.PushEvent(&ListUserNamesAndIdsEvent{Filter: filter, Options: options, Names: names, Ids: ids})
	return names, ids, nil
}
