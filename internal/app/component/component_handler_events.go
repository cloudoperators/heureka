// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
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
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()
	userId := authz.GetCurrentUser()

	r := openfga.RelationInput{
		ObjectType: "component",
		Relation:   "role",
		UserType:   "role",
	}

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentCreateAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if createEvent, ok := e.(*CreateComponentEvent); ok {
		objectId := strconv.FormatInt(createEvent.Component.Id, 10)
		r.UserId = openfga.UserId(userId)
		r.ObjectId = openfga.ObjectId(objectId)

		authz.HandleCreateAuthzRelation(r)
	} else {
		l.Error("Wrong event")
	}
}

// OnComponentUpdateAuthz is a handler for the UpdateComponentEvent
func OnComponentUpdateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()
	userId := authz.GetCurrentUser()

	r := openfga.RelationInput{
		ObjectType: "component",
		Relation:   "role",
		UserType:   "role",
	}

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentCreateAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if updateEvent, ok := e.(*UpdateComponentEvent); ok {
		objectId := strconv.FormatInt(updateEvent.Component.Id, 10)
		r.UserId = openfga.UserId(userId)
		r.ObjectId = openfga.ObjectId(objectId)

		// Handle Update here:
		//recreate component - user
		//recreate component - cv
		//recreate component - role
	} else {
		l.Error("Wrong event")
	}
}

// OnComponentDeleteAuthz is a handler for the DeleteComponentEvent
func OnComponentDeleteAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()
	userId := authz.GetCurrentUser()

	r := openfga.RelationInput{
		ObjectType: "component",
		Relation:   "role",
		UserType:   "role",
	}

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentDeleteAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if deleteEvent, ok := e.(*DeleteComponentEvent); ok {
		objectId := strconv.FormatInt(deleteEvent.ComponentID, 10)
		r.UserId = openfga.UserId(userId)
		r.ObjectId = openfga.ObjectId(objectId)

		// Handle Delete here:
		//delete component - user
		//delete component - cv
		r.ObjectType = "component"
		r.UserType = "role"
		authz.HandleDeleteAuthzRelation(r)
		//delete component - role
		r.ObjectType = "component"
		r.UserType = "role"
		authz.HandleDeleteAuthzRelation(r)

	} else {
		l.Error("Wrong event")
	}
}
