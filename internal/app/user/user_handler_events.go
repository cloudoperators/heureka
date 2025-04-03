// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity"
)

const (
	ListUsersEventName           event.EventName = "ListUsers"
	CreateUserEventName          event.EventName = "CreateUser"
	UpdateUserEventName          event.EventName = "UpdateUser"
	DeleteUserEventName          event.EventName = "DeleteUser"
	ListUserNamesEventName       event.EventName = "ListUserNames"
	ListUniqueUserIDsEventName   event.EventName = "ListUniqueUserIDs"
	ListUserNamesAndIdsEventName event.EventName = "ListUserNamesAndIds"
)

type ListUsersEvent struct {
	Filter  *entity.UserFilter
	Options *entity.ListOptions
	Users   *entity.List[entity.UserResult]
}

func (e *ListUsersEvent) Name() event.EventName {
	return ListUsersEventName
}

type CreateUserEvent struct {
	User *entity.User
}

func (e *CreateUserEvent) Name() event.EventName {
	return CreateUserEventName
}

type UpdateUserEvent struct {
	User *entity.User
}

func (e *UpdateUserEvent) Name() event.EventName {
	return UpdateUserEventName
}

type DeleteUserEvent struct {
	UserID int64
}

func (e *DeleteUserEvent) Name() event.EventName {
	return DeleteUserEventName
}

type ListUserNamesEvent struct {
	Filter  *entity.UserFilter
	Options *entity.ListOptions
	Names   []string
}

func (e *ListUserNamesEvent) Name() event.EventName {
	return ListUserNamesEventName
}

type ListUniqueUserIDsEvent struct {
	Filter  *entity.UserFilter
	Options *entity.ListOptions
	IDs     []string
}

func (e *ListUniqueUserIDsEvent) Name() event.EventName {
	return ListUniqueUserIDsEventName
}

type ListUserNamesAndIdsEvent struct {
	Filter  *entity.UserFilter
	Options *entity.ListOptions
	Names   []string
	Ids     []string
}

func (e *ListUserNamesAndIdsEvent) Name() event.EventName {
	return ListUserNamesAndIdsEventName
}
