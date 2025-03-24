// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_repository

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
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

func (e ListIssueRepositoriesEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListIssueRepositoriesEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListIssueRepositoriesEvent) Name() event.EventName {
	return ListIssueRepositoriesEventName
}

type CreateIssueRepositoryEvent struct {
	IssueRepository *entity.IssueRepository
}

func (e CreateIssueRepositoryEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateIssueRepositoryEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *CreateIssueRepositoryEvent) Name() event.EventName {
	return CreateIssueRepositoryEventName
}

type UpdateIssueRepositoryEvent struct {
	IssueRepository *entity.IssueRepository
}

func (e UpdateIssueRepositoryEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateIssueRepositoryEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *UpdateIssueRepositoryEvent) Name() event.EventName {
	return UpdateIssueRepositoryEventName
}

type DeleteIssueRepositoryEvent struct {
	IssueRepositoryID int64
}

func (e DeleteIssueRepositoryEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &DeleteIssueRepositoryEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *DeleteIssueRepositoryEvent) Name() event.EventName {
	return DeleteIssueRepositoryEventName
}

// OnIssueRepositoryCreate is a handler for the CreateIssueRepositoryEvent
// Is adding the default priority  for the default issue repository
func OnIssueRepositoryCreate(db database.Database, e event.Event) {
	defaultPrio := db.GetDefaultIssuePriority()

	l := logrus.WithFields(logrus.Fields{
		"event":            "OnIssueRepositoryCreate",
		"payload":          e,
		"default_priority": defaultPrio,
	})

	if createEvent, ok := e.(*CreateIssueRepositoryEvent); ok {
		issueRepositoryId := createEvent.IssueRepository.Id

		l.WithField("event-step", "GetIssueRepository").Debug("Fetching Issue Repository by name")

		// Fetch services
		services, err := db.GetServices(&entity.ServiceFilter{}, []entity.Order{})

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
