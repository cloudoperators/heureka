// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

type userHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewUserHandler(db database.Database, er event.EventRegistry) UserHandler {
	return &userHandler{
		database:      db,
		eventRegistry: er,
	}
}

type UserHandlerError struct {
	msg string
}

func (e *UserHandlerError) Error() string {
	return fmt.Sprintf("ServiceHandlerError: %s", e.msg)
}

func NewUserHandlerError(msg string) *UserHandlerError {
	return &UserHandlerError{msg: msg}
}

func (u *userHandler) getUserResults(filter *entity.UserFilter) ([]entity.UserResult, error) {
	var userResults []entity.UserResult
	users, err := u.database.GetUsers(filter)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		user := u
		cursor := fmt.Sprintf("%d", user.Id)
		userResults = append(userResults, entity.UserResult{
			WithCursor:       entity.WithCursor{Value: cursor},
			UserAggregations: nil,
			User:             &user,
		})
	}
	return userResults, nil
}

func (u *userHandler) ListUsers(filter *entity.UserFilter, options *entity.ListOptions) (*entity.List[entity.UserResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListUsersEventName,
		"filter": filter,
	})

	res, err := u.getUserResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Error while filtering for Users")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := u.database.GetAllUserIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewUserHandlerError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = u.database.CountUsers(filter)
		if err != nil {
			l.Error(err)
			return nil, NewUserHandlerError("Error while total count of Users")
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

func (u *userHandler) CreateUser(user *entity.User) (*entity.User, error) {
	f := &entity.UserFilter{
		UniqueUserID: []*string{&user.UniqueUserID},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateUserEventName,
		"object": user,
	})

	users, err := u.ListUsers(f, &entity.ListOptions{})

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

func (u *userHandler) UpdateUser(user *entity.User) (*entity.User, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateUserEventName,
		"object": user,
	})

	err := u.database.UpdateUser(user)

	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while updating user.")
	}

	userResult, err := u.ListUsers(&entity.UserFilter{Id: []*int64{&user.Id}}, &entity.ListOptions{})

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

func (u *userHandler) DeleteUser(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteUserEventName,
		"id":    id,
	})

	err := u.database.DeleteUser(id)

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
