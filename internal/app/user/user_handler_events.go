// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
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

// OnServiceDeleteAuthz is a handler for the DeleteServiceEvent
func OnUserDeleteAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnUserDeleteAuthz")

	deleteInput := []openfga.RelationInput{}

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnUserDeleteAuthz",
		"payload": e,
	})

	if deleteEvent, ok := e.(*DeleteUserEvent); ok {
		// Delete all tuples where object is the user
		deleteInput = append(deleteInput, openfga.RelationInput{
			ObjectType: openfga.TypeUser,
			ObjectId:   openfga.ObjectIdFromInt(deleteEvent.UserID),
		})

		// Delete all tuples where user is the user
		// includes: service, component, component verison, component instance, issue match, support group, role
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType:   openfga.TypeUser,
			UserId:     openfga.UserIdFromInt(deleteEvent.UserID),
			ObjectType: openfga.TypeService,
		})
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType:   openfga.TypeUser,
			UserId:     openfga.UserIdFromInt(deleteEvent.UserID),
			ObjectType: openfga.TypeComponent,
		})
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType:   openfga.TypeUser,
			UserId:     openfga.UserIdFromInt(deleteEvent.UserID),
			ObjectType: openfga.TypeComponentVersion,
		})
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType:   openfga.TypeUser,
			UserId:     openfga.UserIdFromInt(deleteEvent.UserID),
			ObjectType: openfga.TypeComponentInstance,
		})
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType:   openfga.TypeUser,
			UserId:     openfga.UserIdFromInt(deleteEvent.UserID),
			ObjectType: openfga.TypeIssueMatch,
		})
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType:   openfga.TypeUser,
			UserId:     openfga.UserIdFromInt(deleteEvent.UserID),
			ObjectType: openfga.TypeSupportGroup,
		})
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType:   openfga.TypeUser,
			UserId:     openfga.UserIdFromInt(deleteEvent.UserID),
			ObjectType: openfga.TypeRole,
		})

		err := authz.RemoveRelationBulk(deleteInput)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "User", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewUserHandlerError("OnUserDeleteAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "User", "", err)
		l.Error(wrappedErr)
	}
}
