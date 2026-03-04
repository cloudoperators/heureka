// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version

import (
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

const (
	ListComponentVersionsEventName  event.EventName = "ListComponentVersions"
	CreateComponentVersionEventName event.EventName = "CreateComponentVersion"
	UpdateComponentVersionEventName event.EventName = "UpdateComponentVersion"
	DeleteComponentVersionEventName event.EventName = "DeleteComponentVersion"
)

type ListComponentVersionsEvent struct {
	Filter            *entity.ComponentVersionFilter
	Options           *entity.ListOptions
	ComponentVersions *entity.List[entity.ComponentVersionResult]
}

func (e *ListComponentVersionsEvent) Name() event.EventName {
	return ListComponentVersionsEventName
}

type CreateComponentVersionEvent struct {
	ComponentVersion *entity.ComponentVersion
}

func (e *CreateComponentVersionEvent) Name() event.EventName {
	return CreateComponentVersionEventName
}

type UpdateComponentVersionEvent struct {
	ComponentVersion *entity.ComponentVersion
}

func (e *UpdateComponentVersionEvent) Name() event.EventName {
	return UpdateComponentVersionEventName
}

type DeleteComponentVersionEvent struct {
	ComponentVersionID int64
}

func (e *DeleteComponentVersionEvent) Name() event.EventName {
	return DeleteComponentVersionEventName
}

// OnComponentVersionCreateAuthz is a handler for the CreateComponentVersionEvent
// It creates an OpenFGA relation tuple for the component version and the current user
func OnComponentVersionCreateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnComponentVersionCreateAuthz")

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnComponentVersionCreateAuthz",
		"payload": e,
	})

	if createEvent, ok := e.(*CreateComponentVersionEvent); ok {
		userId := openfga.UserIdFromInt(createEvent.ComponentVersion.CreatedBy)

		relations := []openfga.RelationInput{
			{
				UserType:   openfga.TypeRole,
				UserId:     userId,
				Relation:   openfga.RelRole,
				ObjectType: openfga.TypeComponentVersion,
				ObjectId:   openfga.ObjectIdFromInt(createEvent.ComponentVersion.Id),
			},
		}

		err := authz.AddRelationBulk(relations)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "ComponentVersion", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewComponentVersionHandlerError("OnComponentVersionCreateAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "ComponentVersion", "", err)
		l.Error(wrappedErr)
	}
}

// OnComponentVersionUpdateAuthz is a handler for the UpdateComponentVersionEvent
// Fields that can be updated in Component Version which affect tuple relations include:
// componentversion_component_id
func OnComponentVersionUpdateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnComponentVersionUpdateAuthz")

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnComponentVersionUpdateAuthz",
		"payload": e,
	})

	if updateEvent, ok := e.(*UpdateComponentVersionEvent); ok {
		newComponentId := strconv.FormatInt(updateEvent.ComponentVersion.ComponentId, 10)

		if newComponentId != "" {
			// Remove any existing relation where this component_version is connected to any component
			removeInput := openfga.RelationInput{
				UserType:   openfga.TypeComponentVersion,
				UserId:     openfga.UserIdFromInt(updateEvent.ComponentVersion.Id),
				Relation:   openfga.RelComponentVersion,
				ObjectType: openfga.TypeComponent,
				// ObjectId left empty to match any component
			}
			newRelation := openfga.RelationInput{
				UserType:   openfga.TypeComponentVersion,
				UserId:     openfga.UserIdFromInt(updateEvent.ComponentVersion.Id),
				Relation:   openfga.RelComponentVersion,
				ObjectType: openfga.TypeComponent,
				ObjectId:   openfga.ObjectIdFromInt(updateEvent.ComponentVersion.ComponentId),
			}
			err := authz.UpdateRelation(newRelation, removeInput)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), "ComponentVersion", "", err)
				l.Error(wrappedErr)
			}
		}
	} else {
		err := NewComponentVersionHandlerError("OnComponentVersionUpdateAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "ComponentVersion", "", err)
		l.Error(wrappedErr)
	}
}

// OnComponentVersionDeleteAuthz is a handler for the DeleteComponentVersionEvent
func OnComponentVersionDeleteAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnComponentVersionDeleteAuthz")

	deleteInput := []openfga.RelationInput{}

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnComponentVersionDeleteAuthz",
		"payload": e,
	})

	if deleteEvent, ok := e.(*DeleteComponentVersionEvent); ok {
		// Delete all tuples where object is the component_version
		deleteInput = append(deleteInput, openfga.RelationInput{
			ObjectType: openfga.TypeComponentVersion,
			ObjectId:   openfga.ObjectIdFromInt(deleteEvent.ComponentVersionID),
		})

		// Delete all tuples where user is the component_version
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType:   openfga.TypeComponentVersion,
			UserId:     openfga.UserIdFromInt(deleteEvent.ComponentVersionID),
			ObjectType: openfga.TypeComponent,
		})

		err := authz.RemoveRelationBulk(deleteInput)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "ComponentVersion", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewComponentVersionHandlerError("OnComponentVersionDeleteAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "ComponentVersion", "", err)
		l.Error(wrappedErr)
	}
}
