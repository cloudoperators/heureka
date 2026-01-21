// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
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
// It creates an OpenFGA relation tuple for the component instance and the current user
func OnComponentInstanceCreateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnComponentInstanceCreateAuthz")

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnComponentInstanceCreateAuthz",
		"payload": e,
	})

	if createEvent, ok := e.(*CreateComponentInstanceEvent); ok {
		userId := openfga.UserIdFromInt(createEvent.ComponentInstance.CreatedBy)

		relations := []openfga.RelationInput{
			{
				UserType:   openfga.TypeRole,
				UserId:     userId,
				Relation:   openfga.RelRole,
				ObjectType: openfga.TypeComponentInstance,
				ObjectId:   openfga.ObjectIdFromInt(createEvent.ComponentInstance.Id),
			},
			{
				UserType:   openfga.TypeService,
				UserId:     openfga.UserIdFromInt(createEvent.ComponentInstance.ServiceId),
				Relation:   openfga.RelRelatedService,
				ObjectType: openfga.TypeComponentInstance,
				ObjectId:   openfga.ObjectIdFromInt(createEvent.ComponentInstance.Id),
			},
			{
				UserType:   openfga.TypeComponentInstance,
				UserId:     openfga.UserIdFromInt(createEvent.ComponentInstance.Id),
				Relation:   openfga.RelComponentInstance,
				ObjectType: openfga.TypeComponentVersion,
				ObjectId:   openfga.ObjectIdFromInt(createEvent.ComponentInstance.ComponentVersionId),
			},
		}

		err := authz.AddRelationBulk(relations)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewComponentInstanceHandlerError("OnComponentInstanceCreateAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
		l.Error(wrappedErr)
	}
}

// OnComponentInstanceUpdateAuthz is a handler for the UpdateComponentInstanceEvent
// Fields that can be updated in Component Instance which affect tuple relations include:
// componentinstance_component_version_id
// componentinstance_service_id
func OnComponentInstanceUpdateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnComponentInstanceUpdateAuthz")

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnComponentInstanceUpdateAuthz",
		"payload": e,
	})

	if updateEvent, ok := e.(*UpdateComponentInstanceEvent); ok {
		// check if the service relation needs to be updated by looking up existing relations
		existingServiceRelations, err := authz.ListRelations([]openfga.RelationInput{
			{
				UserType:   openfga.TypeService,
				UserId:     openfga.UserIdFromInt(updateEvent.ComponentInstance.ServiceId),
				Relation:   openfga.RelRelatedService,
				ObjectType: openfga.TypeComponentInstance,
				ObjectId:   openfga.ObjectIdFromInt(updateEvent.ComponentInstance.Id),
			},
		})
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
			l.Error(wrappedErr)
			return
		}
		// If no existing relation found, then the service relation needs to be updated
		if len(existingServiceRelations) == 0 {
			removeServiceInput := openfga.RelationInput{
				Relation:   openfga.RelRelatedService,
				ObjectType: openfga.TypeComponentInstance,
				ObjectId:   openfga.ObjectIdFromInt(updateEvent.ComponentInstance.Id),
				UserType:   openfga.TypeService,
				// UserId left empty to match any service
			}
			newServiceRelation := openfga.RelationInput{
				UserType:   openfga.TypeService,
				UserId:     openfga.UserIdFromInt(updateEvent.ComponentInstance.ServiceId),
				Relation:   openfga.RelRelatedService,
				ObjectType: openfga.TypeComponentInstance,
				ObjectId:   openfga.ObjectIdFromInt(updateEvent.ComponentInstance.Id),
			}
			err := authz.UpdateRelation(removeServiceInput, newServiceRelation)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
				l.Error(wrappedErr)
			}
		}

		// check if the component_version relation needs to be updated by looking up existing relations
		existingComponentVersionRelations, err := authz.ListRelations([]openfga.RelationInput{
			{
				Relation:   openfga.RelComponentInstance,
				ObjectType: openfga.TypeComponentVersion,
				ObjectId:   openfga.ObjectIdFromInt(updateEvent.ComponentInstance.ComponentVersionId),
				UserType:   openfga.TypeComponentInstance,
				UserId:     openfga.UserIdFromInt(updateEvent.ComponentInstance.Id),
			},
		})
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
			l.Error(wrappedErr)
			return
		}
		// If no existing relation found, then the component_version relation needs to be updated
		if len(existingComponentVersionRelations) == 0 {
			removeComponentVersionInput := openfga.RelationInput{
				UserType:   openfga.TypeComponentInstance,
				UserId:     openfga.UserIdFromInt(updateEvent.ComponentInstance.Id),
				Relation:   openfga.RelComponentInstance,
				ObjectType: openfga.TypeComponentVersion,
				// ObjectId left empty to match any component_version
			}
			newComponentVersionRelation := openfga.RelationInput{
				UserType:   openfga.TypeComponentInstance,
				UserId:     openfga.UserIdFromInt(updateEvent.ComponentInstance.Id),
				Relation:   openfga.RelComponentInstance,
				ObjectType: openfga.TypeComponentVersion,
				ObjectId:   openfga.ObjectIdFromInt(updateEvent.ComponentInstance.ComponentVersionId),
			}
			err = authz.UpdateRelation(removeComponentVersionInput, newComponentVersionRelation)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
				l.Error(wrappedErr)
			}
		}
	} else {
		err := NewComponentInstanceHandlerError("OnComponentInstanceUpdateAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
		l.Error(wrappedErr)
	}
}

// OnComponentInstanceDeleteAuthz is a handler for the DeleteComponentInstanceEvent
// It creates an OpenFGA relation tuple for the component instance and the current user
func OnComponentInstanceDeleteAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	op := appErrors.Op("OnComponentInstanceDeleteAuthz")

	deleteInput := []openfga.RelationInput{}

	l := logrus.WithFields(logrus.Fields{
		"event":   "OnComponentInstanceDeleteAuthz",
		"payload": e,
	})

	if deleteEvent, ok := e.(*DeleteComponentInstanceEvent); ok {
		// Delete all tuples where object is the component_instance
		deleteInput = append(deleteInput, openfga.RelationInput{
			ObjectType: openfga.TypeComponentInstance,
			ObjectId:   openfga.ObjectIdFromInt(deleteEvent.ComponentInstanceID),
		})

		// Delete all tuples where user is the component_instance
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType: openfga.TypeComponentInstance,
			UserId:   openfga.UserIdFromInt(deleteEvent.ComponentInstanceID),
		})

		err := authz.RemoveRelationBulk(deleteInput)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
			l.Error(wrappedErr)
		}
	} else {
		err := NewComponentInstanceHandlerError("OnComponentInstanceDeleteAuthz: triggered with wrong event type")
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
		l.Error(wrappedErr)
	}
}
