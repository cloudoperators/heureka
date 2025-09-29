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
	userId := authz.GetCurrentUser()

	r := openfga.RelationInput{
		ObjectType: "component_version",
		Relation:   "role",
		UserType:   "role",
	}

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentVersionCreateAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if createEvent, ok := e.(*CreateComponentVersionEvent); ok {
		objectId := strconv.FormatInt(createEvent.ComponentVersion.Id, 10)
		r.UserId = openfga.UserId(userId)
		r.ObjectId = openfga.ObjectId(objectId)

		authz.HandleCreateAuthzRelation(r)
	} else {
		l.Error("Wrong event")
	}
}

// OnComponentVersionUpdateAuthz is a handler for the UpdateComponentVersionEvent
func OnComponentVersionUpdateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()
	userId := authz.GetCurrentUser()

	r := openfga.RelationInput{
		ObjectType: "component_version",
		Relation:   "role",
		UserType:   "role",
	}

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentVersionUpdateAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if updateEvent, ok := e.(*UpdateComponentVersionEvent); ok {
		objectId := strconv.FormatInt(updateEvent.ComponentVersion.Id, 10)
		r.UserId = openfga.UserId(userId)
		r.ObjectId = openfga.ObjectId(objectId)

		// Handle Update here:
		//recreate cv - user
		//recreate cv - ci
		//recreate cv - role

		//recreate cv - component
	} else {
		l.Error("Wrong event")
	}
}

// OnComponentVersionDeleteAuthz is a handler for the DeleteComponentVersionEvent
func OnComponentVersionDeleteAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()
	userId := authz.GetCurrentUser()

	r := openfga.RelationInput{
		ObjectType: "component_version",
		Relation:   "role",
		UserType:   "role",
	}

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentVersionDeleteAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if deleteEvent, ok := e.(*DeleteComponentVersionEvent); ok {
		objectId := strconv.FormatInt(deleteEvent.ComponentVersionID, 10)
		r.UserId = openfga.UserId(userId)
		r.ObjectId = openfga.ObjectId(objectId)

		// Handle Delete here:
		//recreate cv - user
		//recreate cv - ci
		//recreate cv - role

		//recreate cv - component
	} else {
		l.Error("Wrong event")
	}
}
