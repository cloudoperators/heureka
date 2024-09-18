// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_repository

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

const (
	ListIssueRepositoriesEventName event.EventName = "ListIssueRepositories"
	CreateIssueRepositoryEventName event.EventName = "CreateIssueRepository"
	UpdateIssueRepositoryEventName event.EventName = "UpdateIssueRepository"
	DeleteIssueRepositoryEventName event.EventName = "DeleteIssueRepository"
)

type ListIssueRepositoriesEvent struct {
	Filter  *entity.IssueRepositoryFilter
	Options *entity.ListOptions
	Results *entity.List[entity.IssueRepositoryResult]
}

func (e *ListIssueRepositoriesEvent) Name() event.EventName {
	return ListIssueRepositoriesEventName
}

type CreateIssueRepositoryEvent struct {
	IssueRepository *entity.IssueRepository
}

func (e *CreateIssueRepositoryEvent) Name() event.EventName {
	return CreateIssueRepositoryEventName
}

type UpdateIssueRepositoryEvent struct {
	IssueRepository *entity.IssueRepository
}

func (e *UpdateIssueRepositoryEvent) Name() event.EventName {
	return UpdateIssueRepositoryEventName
}

type DeleteIssueRepositoryEvent struct {
	IssueRepositoryID int64
}

func (e *DeleteIssueRepositoryEvent) Name() event.EventName {
	return DeleteIssueRepositoryEventName
}

// OnIssueRepositoryCreate is a handler for the CreateIssueRepositoryEvent
// Is adding the default priority  for the default issue repository
func OnIssueRepositoryCreate(db database.Database, e event.Event) {
	// TODO: make this configureable
	var defaultPrio int64 = 100

	l := logrus.WithFields(logrus.Fields{
		"event":            "OnIssueRepositoryCreate",
		"payload":          e,
		"default_priority": defaultPrio,
	})

	if createEvent, ok := e.(*CreateIssueRepositoryEvent); ok {
		issueRepositoryId := createEvent.IssueRepository.Id

		l.WithField("event-step", "GetIssueRepository").Debug("Fetching Issue Repository by name")

		// Fetch services
		services, err := db.GetServices(&entity.ServiceFilter{})

		if err != nil {
			l.WithField("event-step", "GetServices").WithError(err).Error("Error while fetching services")
			return
		}

		if len(services) == 0 {
			l.WithField("event-step", "GetServices").Error("No services found")
			return
		}

		l.WithField("event-step", "AddIssueRepositoryToService").Debug("Adding Issue Repository to Services")

		for _, srv := range services {
			serviceId := srv.Id
			err = db.AddIssueRepositoryToService(serviceId, issueRepositoryId, defaultPrio)
			if err != nil {
				l.WithField("event-step", "AddIssueRepositoryToService").WithError(err).Error("Error while adding issue repository to service")
			}
		}
	} else {
		l.Error("Wrong event")
	}

}
