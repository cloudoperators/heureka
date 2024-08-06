// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getUserResults(filter *entity.UserFilter) ([]entity.UserResult, error) {
	var userResults []entity.UserResult
	users, err := h.database.GetUsers(filter)
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

func (h *HeurekaApp) ListUsers(filter *entity.UserFilter, options *entity.ListOptions) (*entity.List[entity.UserResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListUsers",
		"filter": filter,
	})

	res, err := h.getUserResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for Users")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllUserIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountUsers(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of Users")
		}
	}

	return &entity.List[entity.UserResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) CreateUser(user *entity.User) (*entity.User, error) {
	f := &entity.UserFilter{
		UniqueUserID: []*string{&user.UniqueUserID},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateUser",
		"object": user,
	})

	users, err := h.ListUsers(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating user.")
	}

	if len(users.Elements) > 0 {
		return nil, heurekaError(fmt.Sprintf("Duplicated entry %s for UniqueUserID.", user.UniqueUserID))
	}

	newUser, err := h.database.CreateUser(user)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating user.")
	}

	return newUser, nil
}

func (h *HeurekaApp) UpdateUser(user *entity.User) (*entity.User, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateUser",
		"object": user,
	})

	err := h.database.UpdateUser(user)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating user.")
	}

	userResult, err := h.ListUsers(&entity.UserFilter{Id: []*int64{&user.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving updated user.")
	}

	if len(userResult.Elements) != 1 {
		l.Error(err)
		return nil, heurekaError("Multiple users found.")
	}

	return userResult.Elements[0].User, nil
}

func (h *HeurekaApp) DeleteUser(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteUser",
		"id":    id,
	})

	err := h.database.DeleteUser(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while deleting user.")
	}

	return nil
}

func (h *HeurekaApp) ListUserNames(filter *entity.UserFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListUserNames",
		"filter": filter,
	})

	userNames, err := h.database.GetUserNames(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving userNames.")
	}

	return userNames, nil
}

func (h *HeurekaApp) ListUniqueUserID(filter *entity.UserFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListUniqueUserID",
		"filter": filter,
	})

	uniqueUserID, err := h.database.GetUniqueUserIDs(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving uniqueUserID.")
	}

	return uniqueUserID, nil
}
