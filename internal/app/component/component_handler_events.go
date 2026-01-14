// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

const (
	ListComponentsEventName                  event.EventName = "ListComponents"
	CreateComponentEventName                 event.EventName = "CreateComponent"
	UpdateComponentEventName                 event.EventName = "UpdateComponent"
	DeleteComponentEventName                 event.EventName = "DeleteComponent"
	ListComponentCcrnsEventName              event.EventName = "ListComponentCcrns"
	GetComponentIssueSeverityCountsEventName event.EventName = "GetComponentIssueSeverityCounts"
)

type ListComponentsEvent struct {
	Filter     *entity.ComponentFilter
	Options    *entity.ListOptions
	Components *entity.List[entity.ComponentResult]
}

func (e *ListComponentsEvent) Name() event.EventName {
	return ListComponentsEventName
}

type CreateComponentEvent struct {
	Component *entity.Component
}

func (e *CreateComponentEvent) Name() event.EventName {
	return CreateComponentEventName
}

type UpdateComponentEvent struct {
	Component *entity.Component
}

func (e *UpdateComponentEvent) Name() event.EventName {
	return UpdateComponentEventName
}

type DeleteComponentEvent struct {
	ComponentID int64
}

func (e *DeleteComponentEvent) Name() event.EventName {
	return DeleteComponentEventName
}

type ListComponentCcrnsEvent struct {
	Filter  *entity.ComponentFilter
	Options *entity.ListOptions
	CCRNs   []string
}

func (e *ListComponentCcrnsEvent) Name() event.EventName {
	return ListComponentCcrnsEventName
}

type GetComponentIssueSeverityCountsEvent struct {
	Filter *entity.ComponentFilter
	Counts []entity.IssueSeverityCounts
}

func (e *GetComponentIssueSeverityCountsEvent) Name() event.EventName {
	return GetComponentIssueSeverityCountsEventName
}

// OnComponentCreateAuthz is a handler for the CreateComponentEvent
// It creates an OpenFGA relation tuple for the component and the current user
func OnComponentCreateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnComponentCreateAuthz")

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnComponentCreateAuthz",
		"payload": e,
	})

	if createEvent, ok := e.(*CreateComponentEvent); ok {
		userId := openfga.UserIdFromInt(createEvent.Component.CreatedBy)

		relations := []openfga.RelationInput{
			{
				UserType:   openfga.TypeRole,
				UserId:     userId,
				Relation:   openfga.RelRole,
				ObjectType: openfga.TypeComponent,
				ObjectId:   openfga.ObjectIdFromInt(createEvent.Component.Id),
			},
		}

		openfga.AddRelations(authz, relations)
	} else {
		err := NewComponentHandlerError("OnComponentCreateAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "Component", "", err)
		l.Error(wrappedErr)
	}
}

// OnComponentDeleteAuthz is a handler for the DeleteComponentEvent
func OnComponentDeleteAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnComponentDeleteAuthz")

	deleteInput := []openfga.RelationInput{}

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnComponentDeleteAuthz",
		"payload": e,
	})

	if deleteEvent, ok := e.(*DeleteComponentEvent); ok {
		deleteElement := openfga.RelationInput{
			ObjectType: openfga.TypeComponent,
			ObjectId:   openfga.ObjectIdFromInt(deleteEvent.ComponentID),
		}
		deleteInput = append(deleteInput, deleteElement)

		authz.RemoveRelationBulk(deleteInput)
	} else {
		err := NewComponentHandlerError("OnComponentDeleteAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "Component", "", err)
		l.Error(wrappedErr)
	}
}
