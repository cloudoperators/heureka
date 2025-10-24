// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

const (
	CreateServiceEventName                    event.EventName = "CreateService"
	UpdateServiceEventName                    event.EventName = "UpdateService"
	DeleteServiceEventName                    event.EventName = "DeleteService"
	AddOwnerToServiceEventName                event.EventName = "AddOwnerToService"
	RemoveOwnerFromServiceEventName           event.EventName = "RemoveOwnerFromService"
	ListServicesEventName                     event.EventName = "ListServices"
	GetServiceEventName                       event.EventName = "GetService"
	ListServiceCcrnsEventName                 event.EventName = "ListServiceCcrns"
	ListServiceDomainsEventName               event.EventName = "ListServiceDomains"
	ListServiceRegionsEventName               event.EventName = "ListServiceRegions"
	AddIssueRepositoryToServiceEventName      event.EventName = "AddIssueRepositoryToService"
	RemoveIssueRepositoryFromServiceEventName event.EventName = "RemoveIssueRepositoryFromService"
)

type CreateServiceEvent struct {
	Service *entity.Service
}

func (e *CreateServiceEvent) Name() event.EventName {
	return CreateServiceEventName
}

type UpdateServiceEvent struct {
	Service *entity.Service
}

func (e *UpdateServiceEvent) Name() event.EventName {
	return UpdateServiceEventName
}

type DeleteServiceEvent struct {
	ServiceID int64
}

func (e *DeleteServiceEvent) Name() event.EventName {
	return DeleteServiceEventName
}

type AddOwnerToServiceEvent struct {
	ServiceID int64
	OwnerID   int64
}

func (e *AddOwnerToServiceEvent) Name() event.EventName {
	return AddOwnerToServiceEventName
}

type RemoveOwnerFromServiceEvent struct {
	ServiceID int64
	OwnerID   int64
}

func (e *RemoveOwnerFromServiceEvent) Name() event.EventName {
	return RemoveOwnerFromServiceEventName
}

type ListServicesEvent struct {
	Filter   *entity.ServiceFilter
	Options  *entity.ListOptions
	Services *entity.List[entity.ServiceResult]
}

func (e *ListServicesEvent) Name() event.EventName {
	return ListServicesEventName
}

type GetServiceEvent struct {
	ServiceID int64
	Service   *entity.Service
}

func (e *GetServiceEvent) Name() event.EventName {
	return GetServiceEventName
}

type ListServiceCcrnsEvent struct {
	Filter  *entity.ServiceFilter
	Options *entity.ListOptions
	Ccrns   []string
}

func (e *ListServiceCcrnsEvent) Name() event.EventName {
	return ListServiceCcrnsEventName
}

type ListServiceDomainsEvent struct {
	Filter  *entity.ServiceFilter
	Options *entity.ListOptions
	Domains []string
}

func (e *ListServiceDomainsEvent) Name() event.EventName {
	return ListServiceDomainsEventName
}

type ListServiceRegionsEvent struct {
	Filter  *entity.ServiceFilter
	Options *entity.ListOptions
	Regions []string
}

func (e *ListServiceRegionsEvent) Name() event.EventName {
	return ListServiceRegionsEventName
}

type AddIssueRepositoryToServiceEvent struct {
	ServiceID    int64
	RepositoryID int64
}

func (e *AddIssueRepositoryToServiceEvent) Name() event.EventName {
	return AddIssueRepositoryToServiceEventName
}

type RemoveIssueRepositoryFromServiceEvent struct {
	ServiceID    int64
	RepositoryID int64
}

func (e *RemoveIssueRepositoryFromServiceEvent) Name() event.EventName {
	return RemoveIssueRepositoryFromServiceEventName
}

// OnServiceCreate is a handler for the CreateServiceEvent
// Is creating a single default priority for the default issue repository
func OnServiceCreate(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnServiceCreate",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if createEvent, ok := e.(*CreateServiceEvent); ok {
		serviceId := createEvent.Service.Id
		l.WithField("event-step", "GetIssueRepository").Debug("Fetching Issue Repository by name")

		// Fetch IssueRepositories
		issueRepositories, err := db.GetIssueRepositories(&entity.IssueRepositoryFilter{
			Name: []*string{&defaultRepoName},
		})

		if err != nil {
			l.WithField("event-step", "GetIssueRepository").WithError(err).Error("Error while fetching issue repository by name")
			return
		}

		if len(issueRepositories) == 0 {
			l.WithField("event-step", "GetIssueRepository").Error("No Issue Repository found by name")
			return
		}

		l.WithField("event-step", "AddIssueRepositoryToService").Debug("Adding Issue Repository to Service")

		err = db.AddIssueRepositoryToService(serviceId, issueRepositories[0].Id, defaultPrio)
		if err != nil {
			l.WithField("event-step", "AddIssueRepositoryToService").WithError(err).Error("Error while adding issue repository to service")
		}
	} else {
		l.Error("Wrong event")
	}
}

// OnServiceCreateAuthz is a handler for the CreateServiceEvent
// It creates an OpenFGA relation tuple for the service and the current user
func OnServiceCreateAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnServiceCreateAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if createEvent, ok := e.(*CreateServiceEvent); ok {
		serviceId := strconv.FormatInt(createEvent.Service.Id, 10)
		userId := authz.GetCurrentUser()

		rlist := []openfga.RelationInput{
			{
				UserType:   "role",
				UserId:     openfga.UserId(userId),
				Relation:   "role",
				ObjectType: "service",
				ObjectId:   openfga.ObjectId(serviceId),
			},
		}

		for _, rel := range rlist {
			authz.AddRelation(rel)
		}
	} else {
		l.Error("Wrong event")
	}
}

// OnServiceDeleteAuthz is a handler for the DeleteServiceEvent
// It deletes all OpenFGA relation tuples containing that service
func OnServiceDeleteAuthz(db database.Database, e event.Event, authz openfga.Authorization) {
	defaultPrio := db.GetDefaultIssuePriority()
	defaultRepoName := db.GetDefaultRepositoryName()
	deleteInput := []openfga.RelationInput{}

	l := logrus.WithFields(logrus.Fields{
		"event":             "OnServiceDeleteAuthz",
		"payload":           e,
		"default_priority":  defaultPrio,
		"default_repo_name": defaultRepoName,
	})

	if deleteEvent, ok := e.(*DeleteServiceEvent); ok {
		objectId := strconv.FormatInt(deleteEvent.ServiceID, 10)

		// Delete all tuples where object is the service
		deleteInput = append(deleteInput, openfga.RelationInput{
			ObjectType: "service",
			ObjectId:   openfga.ObjectId(objectId),
		})

		// Delete all tuples where user is the service
		deleteInput = append(deleteInput, openfga.RelationInput{
			UserType: "service",
			UserId:   openfga.UserId(objectId),
		})

		authz.RemoveRelationBulk(deleteInput)
	} else {
		l.Error("Wrong event")
	}
}

// OnAddOwnerToService is a handler for the AddOwnerToServiceEvent
// It creates an OpenFGA relation tuple between the owner and the service
func OnAddOwnerToService(db database.Database, e event.Event, authz openfga.Authorization) {
	l := logrus.WithFields(logrus.Fields{
		"event":   "OnAddOwnerToService",
		"payload": e,
	})

	if addEvent, ok := e.(*AddOwnerToServiceEvent); ok {
		serviceId := strconv.FormatInt(addEvent.ServiceID, 10)
		ownerId := strconv.FormatInt(addEvent.OwnerID, 10)

		rel := openfga.RelationInput{
			UserType:   "user",
			UserId:     openfga.UserId(ownerId),
			ObjectType: "service",
			ObjectId:   openfga.ObjectId(serviceId),
			Relation:   "owner",
		}
		authz.AddRelation(rel)
	} else {
		l.Error("Wrong event")
	}
}

// OnRemoveOwnerFromService is a handler for the RemoveOwnerFromServiceEvent
// It removes the OpenFGA relation tuple between the owner and the service
func OnRemoveOwnerFromService(db database.Database, e event.Event, authz openfga.Authorization) {
	l := logrus.WithFields(logrus.Fields{
		"event":   "OnRemoveOwnerFromService",
		"payload": e,
	})

	if removeEvent, ok := e.(*RemoveOwnerFromServiceEvent); ok {
		serviceId := strconv.FormatInt(removeEvent.ServiceID, 10)
		ownerId := strconv.FormatInt(removeEvent.OwnerID, 10)

		rel := openfga.RelationInput{
			UserType:   "user",
			UserId:     openfga.UserId(ownerId),
			ObjectType: "service",
			ObjectId:   openfga.ObjectId(serviceId),
			Relation:   "owner",
		}
		authz.RemoveRelation(rel)
	} else {
		l.Error("Wrong event")
	}
}
