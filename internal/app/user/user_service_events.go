package user

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
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
