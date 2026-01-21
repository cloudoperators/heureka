// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

const (
	ListSupportGroupsEventName             event.EventName = "ListSupportGroups"
	GetSupportGroupEventName               event.EventName = "GetSupportGroup"
	CreateSupportGroupEventName            event.EventName = "CreateSupportGroup"
	UpdateSupportGroupEventName            event.EventName = "UpdateSupportGroup"
	DeleteSupportGroupEventName            event.EventName = "DeleteSupportGroup"
	AddServiceToSupportGroupEventName      event.EventName = "AddServiceToSupportGroup"
	RemoveServiceFromSupportGroupEventName event.EventName = "RemoveServiceFromSupportGroup"
	AddUserToSupportGroupEventName         event.EventName = "AddUserToSupportGroup"
	RemoveUserFromSupportGroupEventName    event.EventName = "RemoveUserFromSupportGroup"
	ListSupportGroupCcrnsEventName         event.EventName = "ListSupportGroupCcrns"
)

type ListSupportGroupsEvent struct {
	Filter        *entity.SupportGroupFilter
	Options       *entity.ListOptions
	SupportGroups *entity.List[entity.SupportGroupResult]
}

func (e *ListSupportGroupsEvent) Name() event.EventName {
	return ListSupportGroupsEventName
}

type GetSupportGroupEvent struct {
	SupportGroupID int64
	SupportGroup   *entity.SupportGroup
}

func (e *GetSupportGroupEvent) Name() event.EventName {
	return GetSupportGroupEventName
}

type CreateSupportGroupEvent struct {
	SupportGroup *entity.SupportGroup
}

func (e *CreateSupportGroupEvent) Name() event.EventName {
	return CreateSupportGroupEventName
}

type UpdateSupportGroupEvent struct {
	SupportGroup *entity.SupportGroup
}

func (e *UpdateSupportGroupEvent) Name() event.EventName {
	return UpdateSupportGroupEventName
}

type DeleteSupportGroupEvent struct {
	SupportGroupID int64
}

func (e *DeleteSupportGroupEvent) Name() event.EventName {
	return DeleteSupportGroupEventName
}

type AddServiceToSupportGroupEvent struct {
	SupportGroupID int64
	ServiceID      int64
}

func (e *AddServiceToSupportGroupEvent) Name() event.EventName {
	return AddServiceToSupportGroupEventName
}

type RemoveServiceFromSupportGroupEvent struct {
	SupportGroupID int64
	ServiceID      int64
}

func (e *RemoveServiceFromSupportGroupEvent) Name() event.EventName {
	return RemoveServiceFromSupportGroupEventName
}

type AddUserToSupportGroupEvent struct {
	SupportGroupID int64
	UserID         int64
}

func (e *AddUserToSupportGroupEvent) Name() event.EventName {
	return AddUserToSupportGroupEventName
}

type RemoveUserFromSupportGroupEvent struct {
	SupportGroupID int64
	UserID         int64
}

func (e *RemoveUserFromSupportGroupEvent) Name() event.EventName {
	return RemoveUserFromSupportGroupEventName
}

type ListSupportGroupCcrnsEvent struct {
	Filter  *entity.SupportGroupFilter
	Options *entity.ListOptions
	Ccrns   []string
}

func (e *ListSupportGroupCcrnsEvent) Name() event.EventName {
	return ListSupportGroupCcrnsEventName
}

// OnSupportGroupCreateAuthz is a handler for the CreateSupportGroupEvent
// It creates an OpenFGA relation tuple for the support group and the current user
func OnSupportGroupCreateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnSupportGroupCreateAuthz")

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnSupportGroupCreateAuthz",
		"payload": e,
	})

	if createEvent, ok := e.(*CreateSupportGroupEvent); ok {
		userId := openfga.UserIdFromInt(createEvent.SupportGroup.CreatedBy)

		relations := []openfga.RelationInput{
			{
				UserType:   openfga.TypeRole,
				UserId:     userId,
				Relation:   openfga.RelRole,
				ObjectType: openfga.TypeSupportGroup,
				ObjectId:   openfga.ObjectIdFromInt(createEvent.SupportGroup.Id),
			},
		}

		err := authz.AddRelationBulk(relations)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewSupportGroupHandlerError("OnSupportGroupCreateAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
		l.Error(wrappedErr)
	}
}

// OnServiceDeleteAuthz is a handler for the DeleteServiceEvent
func OnSupportGroupDeleteAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnSupportGroupDeleteAuthz")

	deleteInput := []openfga.RelationInput{}

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnSupportGroupDeleteAuthz",
		"payload": e,
	})

	if deleteEvent, ok := e.(*DeleteSupportGroupEvent); ok {

		// Delete all tuples where object is the support_group
		deleteInput = append(deleteInput, openfga.RelationInput{
			ObjectType: openfga.TypeSupportGroup,
			ObjectId:   openfga.ObjectIdFromInt(deleteEvent.SupportGroupID),
		})

		// Delete all tuples where user is the support_group
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType: openfga.TypeSupportGroup,
			UserId:   openfga.UserIdFromInt(deleteEvent.SupportGroupID),
		})

		err := authz.RemoveRelationBulk(deleteInput)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewSupportGroupHandlerError("OnSupportGroupDeleteAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
		l.Error(wrappedErr)
	}
}

// OnAddServiceToSupportGroup is a handler for the AddServiceToSupportGroupEvent
// It creates an OpenFGA relation tuple between the support group and the service
func OnAddServiceToSupportGroup(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnAddServiceToSupportGroup")

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnAddServiceToSupportGroup",
		"payload": e,
	})

	if addEvent, ok := e.(*AddServiceToSupportGroupEvent); ok {
		relations := []openfga.RelationInput{
			{
				UserType:   openfga.TypeSupportGroup,
				UserId:     openfga.UserIdFromInt(addEvent.SupportGroupID),
				ObjectType: openfga.TypeService,
				ObjectId:   openfga.ObjectIdFromInt(addEvent.ServiceID),
				Relation:   openfga.RelSupportGroup,
			},
		}
		err := authz.AddRelationBulk(relations)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewSupportGroupHandlerError("OnAddServiceToSupportGroup: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
		l.Error(wrappedErr)
	}
}

// OnRemoveServiceFromSupportGroup is a handler for the RemoveServiceFromSupportGroupEvent
// It removes the OpenFGA relation tuple between the support group and the service
func OnRemoveServiceFromSupportGroup(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnRemoveServiceFromSupportGroup")

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnRemoveServiceFromSupportGroup",
		"payload": e,
	})

	if removeEvent, ok := e.(*RemoveServiceFromSupportGroupEvent); ok {
		rel := openfga.RelationInput{
			UserType:   openfga.TypeSupportGroup,
			UserId:     openfga.UserIdFromInt(removeEvent.SupportGroupID),
			ObjectType: openfga.TypeService,
			ObjectId:   openfga.ObjectIdFromInt(removeEvent.ServiceID),
			Relation:   openfga.RelSupportGroup,
		}
		err := authz.RemoveRelation(rel)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewSupportGroupHandlerError("OnRemoveServiceFromSupportGroup: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
		l.Error(wrappedErr)
	}
}

// OnAddUserToSupportGroup is a handler for the AddUserToSupportGroupEvent
// It creates an OpenFGA relation tuple between the user and the support group
func OnAddUserToSupportGroup(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnAddUserToSupportGroup")

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnAddUserToSupportGroup",
		"payload": e,
	})

	if addEvent, ok := e.(*AddUserToSupportGroupEvent); ok {
		relations := []openfga.RelationInput{
			{
				UserType:   openfga.TypeUser,
				UserId:     openfga.UserIdFromInt(addEvent.UserID),
				ObjectType: openfga.TypeSupportGroup,
				ObjectId:   openfga.ObjectIdFromInt(addEvent.SupportGroupID),
				Relation:   openfga.RelMember,
			},
		}
		err := authz.AddRelationBulk(relations)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewSupportGroupHandlerError("OnAddUserToSupportGroup: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
		l.Error(wrappedErr)
	}
}

// OnRemoveUserFromSupportGroup is a handler for the RemoveUserFromSupportGroupEvent
// It removes the OpenFGA relation tuple between the user and the support group
func OnRemoveUserFromSupportGroup(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnRemoveUserFromSupportGroup")

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnRemoveUserFromSupportGroup",
		"payload": e,
	})

	if removeEvent, ok := e.(*RemoveUserFromSupportGroupEvent); ok {
		rel := openfga.RelationInput{
			UserType:   openfga.TypeUser,
			UserId:     openfga.UserIdFromInt(removeEvent.UserID),
			ObjectType: openfga.TypeSupportGroup,
			ObjectId:   openfga.ObjectIdFromInt(removeEvent.SupportGroupID),
			Relation:   openfga.RelMember,
		}
		err := authz.RemoveRelation(rel)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewSupportGroupHandlerError("OnRemoveUserFromSupportGroup: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
		l.Error(wrappedErr)
	}
}
