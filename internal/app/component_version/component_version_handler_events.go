// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version

import (
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
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
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentVersionCreateAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if createEvent, ok := e.(*CreateComponentVersionEvent); ok {
		versionId := strconv.FormatInt(createEvent.ComponentVersion.Id, 10)
		userId := authz.GetCurrentUser()

		rlist := []openfga.RelationInput{
			{
				UserType:   "role",
				UserId:     openfga.UserId(userId),
				Relation:   "role",
				ObjectType: "component_version",
				ObjectId:   openfga.ObjectId(versionId),
			},
		}

		for _, rel := range rlist {
			authz.AddRelation(rel)
		}
	} else {
		l.Error("Wrong event")
	}
}

// OnComponentVersionUpdateAuthz is a handler for the UpdateComponentVersionEvent
// Fields that can be updated in Component Version which affect tuple relations include:
// componentversion_component_id
func OnComponentVersionUpdateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentVersionUpdateAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if updateEvent, ok := e.(*UpdateComponentVersionEvent); ok {
		versionId := strconv.FormatInt(updateEvent.ComponentVersion.Id, 10)
		newComponentId := strconv.FormatInt(updateEvent.ComponentVersion.ComponentId, 10)

		if newComponentId != "" {
			// Remove any existing relation where this component_version is connected to any component
			removeInput := openfga.RelationInput{
				UserType:   "component_version",
				UserId:     openfga.UserId(versionId),
				Relation:   "component_version",
				ObjectType: "component",
				// ObjectId left empty to match any component
			}
			newRelation := openfga.RelationInput{
				UserType:   "component_version",
				UserId:     openfga.UserId(versionId),
				Relation:   "component_version",
				ObjectType: "component",
				ObjectId:   openfga.ObjectId(newComponentId),
			}
			authz.UpdateRelation(removeInput, newRelation)
		}
	} else {
		l.Error("Wrong event")
	}
}

// OnComponentVersionDeleteAuthz is a handler for the DeleteComponentVersionEvent
func OnComponentVersionDeleteAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()
	deleteInput := []openfga.RelationInput{}

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentVersionDeleteAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if deleteEvent, ok := e.(*DeleteComponentVersionEvent); ok {
		objectId := strconv.FormatInt(deleteEvent.ComponentVersionID, 10)

		// Delete all tuples where object is the component_version
		deleteInput = append(deleteInput, openfga.RelationInput{
			ObjectType: "component_version",
			ObjectId:   openfga.ObjectId(objectId),
		})

		// Delete all tuples where user is the component_version
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType: "component_version",
			UserId:   openfga.UserId(objectId),
		})

		authz.RemoveRelationBulk(deleteInput)
	} else {
		l.Error("Wrong event")
	}
}
