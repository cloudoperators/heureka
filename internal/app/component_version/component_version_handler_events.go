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
	ResourceType := "component_version"
	ResourceRelation := "role"

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentVersionCreateAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if createEvent, ok := e.(*CreateComponentVersionEvent); ok {
		resourceId := strconv.FormatInt(createEvent.ComponentVersion.Id, 10)
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
