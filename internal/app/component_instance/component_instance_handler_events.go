// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import (
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

const (
	ListComponentInstancesEventName  event.EventName = "ListComponentInstances"
	CreateComponentInstanceEventName event.EventName = "CreateComponentInstance"
	UpdateComponentInstanceEventName event.EventName = "UpdateComponentInstance"
	DeleteComponentInstanceEventName event.EventName = "DeleteComponentInstance"
	ListCcrnEventName                event.EventName = "ListCcrn"
	ListRegionsEventName             event.EventName = "ListRegions"
	ListClustersEventName            event.EventName = "ListClusters"
	ListNamespacesEventName          event.EventName = "ListNamespaces"
	ListDomainsEventName             event.EventName = "ListDomains"
	ListProjectsEventName            event.EventName = "ListProjects"
	ListPodsEventName                event.EventName = "ListPods"
	ListContainersEventName          event.EventName = "ListContainers"
	ListTypesEventName               event.EventName = "ListTypes"
	ListParentsEventName             event.EventName = "ListParents"
	ListContextsEventName            event.EventName = "ListContexts"
)

type ListComponentInstancesEvent struct {
	Filter             *entity.ComponentInstanceFilter
	Options            *entity.ListOptions
	ComponentInstances *entity.List[entity.ComponentInstanceResult]
}

func (e *ListComponentInstancesEvent) Name() event.EventName {
	return ListComponentInstancesEventName
}

type CreateComponentInstanceEvent struct {
	ComponentInstance *entity.ComponentInstance
}

func (e *CreateComponentInstanceEvent) Name() event.EventName {
	return CreateComponentInstanceEventName
}

type UpdateComponentInstanceEvent struct {
	ComponentInstance *entity.ComponentInstance
}

func (e *UpdateComponentInstanceEvent) Name() event.EventName {
	return UpdateComponentInstanceEventName
}

type DeleteComponentInstanceEvent struct {
	ComponentInstanceID int64
}

func (e *DeleteComponentInstanceEvent) Name() event.EventName {
	return DeleteComponentInstanceEventName
}

type ListCcrnEvent struct {
	Filter *entity.ComponentInstanceFilter
	Ccrn   []string
}

func (e *ListCcrnEvent) Name() event.EventName {
	return ListCcrnEventName
}

type ListRegionsEvent struct {
	Filter  *entity.ComponentInstanceFilter
	Regions []string
}

func (e *ListRegionsEvent) Name() event.EventName {
	return ListRegionsEventName
}

type ListClustersEvent struct {
	Filter   *entity.ComponentInstanceFilter
	Clusters []string
}

func (e *ListClustersEvent) Name() event.EventName {
	return ListClustersEventName
}

type ListNamespacesEvent struct {
	Filter     *entity.ComponentInstanceFilter
	Namespaces []string
}

func (e *ListNamespacesEvent) Name() event.EventName {
	return ListNamespacesEventName
}

type ListDomainsEvent struct {
	Filter  *entity.ComponentInstanceFilter
	Domains []string
}

func (e *ListDomainsEvent) Name() event.EventName {
	return ListDomainsEventName
}

type ListProjectsEvent struct {
	Filter   *entity.ComponentInstanceFilter
	Projects []string
}

func (e *ListProjectsEvent) Name() event.EventName {
	return ListProjectsEventName
}

type ListPodsEvent struct {
	Filter *entity.ComponentInstanceFilter
	Pods   []string
}

func (e *ListPodsEvent) Name() event.EventName {
	return ListPodsEventName
}

type ListContainersEvent struct {
	Filter     *entity.ComponentInstanceFilter
	Containers []string
}

func (e *ListContainersEvent) Name() event.EventName {
	return ListContainersEventName
}

type ListTypesEvent struct {
	Filter *entity.ComponentInstanceFilter
	Types  []string
}

func (e *ListTypesEvent) Name() event.EventName {
	return ListTypesEventName
}

type ListParentsEvent struct {
	Filter  *entity.ComponentInstanceFilter
	Parents []string
}

func (e *ListParentsEvent) Name() event.EventName {
	return ListParentsEventName
}

type ListContextsEvent struct {
	Filter   *entity.ComponentInstanceFilter
	Contexts []string
}

func (e *ListContextsEvent) Name() event.EventName {
	return ListContextsEventName
}

// OnComponentInstanceCreateAuthz is a handler for the CreateComponentInstanceEvent
// Is creating a openfga relation tuple for the component instance and the current user
func OnComponentInstanceCreateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()
	ResourceType := "component_instance"
	ResourceRelation := "role"

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnComponentInstanceCreateAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if createEvent, ok := e.(*CreateComponentInstanceEvent); ok {
		resourceId := strconv.FormatInt(createEvent.ComponentInstance.Id, 10)
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
