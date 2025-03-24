// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
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
	AddIssueRepositoryToServiceEventName      event.EventName = "AddIssueRepositoryToService"
	RemoveIssueRepositoryFromServiceEventName event.EventName = "RemoveIssueRepositoryFromService"
)

type CreateServiceEvent struct {
	Service *entity.Service
}

func (e CreateServiceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateServiceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *CreateServiceEvent) Name() event.EventName {
	return CreateServiceEventName
}

type UpdateServiceEvent struct {
	Service *entity.Service
}

func (e UpdateServiceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateServiceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *UpdateServiceEvent) Name() event.EventName {
	return UpdateServiceEventName
}

type DeleteServiceEvent struct {
	ServiceID int64
}

func (e DeleteServiceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &DeleteServiceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *DeleteServiceEvent) Name() event.EventName {
	return DeleteServiceEventName
}

type AddOwnerToServiceEvent struct {
	ServiceID int64
	OwnerID   int64
}

func (e AddOwnerToServiceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &AddOwnerToServiceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *AddOwnerToServiceEvent) Name() event.EventName {
	return AddOwnerToServiceEventName
}

type RemoveOwnerFromServiceEvent struct {
	ServiceID int64
	OwnerID   int64
}

func (e RemoveOwnerFromServiceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &RemoveOwnerFromServiceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *RemoveOwnerFromServiceEvent) Name() event.EventName {
	return RemoveOwnerFromServiceEventName
}

type ListServicesEvent struct {
	Filter   *entity.ServiceFilter
	Options  *entity.ListOptions
	Services *entity.List[entity.ServiceResult]
}

func (e ListServicesEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListServicesEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListServicesEvent) Name() event.EventName {
	return ListServicesEventName
}

type GetServiceEvent struct {
	ServiceID int64
	Service   *entity.Service
}

func (e GetServiceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &GetServiceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *GetServiceEvent) Name() event.EventName {
	return GetServiceEventName
}

type ListServiceCcrnsEvent struct {
	Filter  *entity.ServiceFilter
	Options *entity.ListOptions
	Ccrns   []string
}

func (e ListServiceCcrnsEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListServiceCcrnsEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListServiceCcrnsEvent) Name() event.EventName {
	return ListServiceCcrnsEventName
}

type AddIssueRepositoryToServiceEvent struct {
	ServiceID    int64
	RepositoryID int64
}

func (e AddIssueRepositoryToServiceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &AddIssueRepositoryToServiceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *AddIssueRepositoryToServiceEvent) Name() event.EventName {
	return AddIssueRepositoryToServiceEventName
}

type RemoveIssueRepositoryFromServiceEvent struct {
	ServiceID    int64
	RepositoryID int64
}

func (e RemoveIssueRepositoryFromServiceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &RemoveIssueRepositoryFromServiceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *RemoveIssueRepositoryFromServiceEvent) Name() event.EventName {
	return RemoveIssueRepositoryFromServiceEventName
}

// OnServiceCreate is a handler for the CreateServiceEvent
// Is creating a single default priority for the default issue repository
func OnServiceCreate(db database.Database, e event.Event) {
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
