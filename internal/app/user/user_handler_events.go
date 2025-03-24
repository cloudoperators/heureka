// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
)

const (
	ListUsersEventName         event.EventName = "ListUsers"
	CreateUserEventName        event.EventName = "CreateUser"
	UpdateUserEventName        event.EventName = "UpdateUser"
	DeleteUserEventName        event.EventName = "DeleteUser"
	ListUserNamesEventName     event.EventName = "ListUserNames"
	ListUniqueUserIDsEventName event.EventName = "ListUniqueUserIDs"
)

type ListUsersEvent struct {
	Filter  *entity.UserFilter
	Options *entity.ListOptions
	Users   *entity.List[entity.UserResult]
}

func (e ListUsersEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListUsersEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListUsersEvent) Name() event.EventName {
	return ListUsersEventName
}

type CreateUserEvent struct {
	User *entity.User
}

func (e CreateUserEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateUserEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *CreateUserEvent) Name() event.EventName {
	return CreateUserEventName
}

type UpdateUserEvent struct {
	User *entity.User
}

func (e UpdateUserEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateUserEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *UpdateUserEvent) Name() event.EventName {
	return UpdateUserEventName
}

type DeleteUserEvent struct {
	UserID int64
}

func (e DeleteUserEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &DeleteUserEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *DeleteUserEvent) Name() event.EventName {
	return DeleteUserEventName
}

type ListUserNamesEvent struct {
	Filter  *entity.UserFilter
	Options *entity.ListOptions
	Names   []string
}

func (e ListUserNamesEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListUserNamesEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListUserNamesEvent) Name() event.EventName {
	return ListUserNamesEventName
}

type ListUniqueUserIDsEvent struct {
	Filter  *entity.UserFilter
	Options *entity.ListOptions
	IDs     []string
}

func (e ListUniqueUserIDsEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListUniqueUserIDsEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListUniqueUserIDsEvent) Name() event.EventName {
	return ListUniqueUserIDsEventName
}
