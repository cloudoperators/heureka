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
	ListComponentsEventName     event.EventName = "ListComponents"
	CreateComponentEventName    event.EventName = "CreateComponent"
	UpdateComponentEventName    event.EventName = "UpdateComponent"
	DeleteComponentEventName    event.EventName = "DeleteComponent"
	ListComponentCcrnsEventName event.EventName = "ListComponentCcrns"
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

// OnComponentCreateAuthz is a handler for the CreateComponentEvent
// It creates an OpenFGA relation tuple for the component and the current user
func OnComponentCreateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()
	ResourceType := "component"
	ResourceRelation := "role"

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentCreateAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if createEvent, ok := e.(*CreateComponentEvent); ok {
		resourceId := strconv.FormatInt(createEvent.Component.Id, 10)
		user := authz.GetCurrentUser()
		userFieldName := "role"

		authz.HandleCreateAuthzRelation(
			userFieldName,
			user,
			resourceId,
			ResourceType,
			ResourceRelation,
		)
	} else {
		l.Error("Wrong event")
	}
}
