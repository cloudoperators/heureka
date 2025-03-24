// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package activity

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
)

const (
	ActivityCreateEventName            event.EventName = "ActivityCreate"
	ActivityUpdateEventName            event.EventName = "ActivityUpdate"
	ActivityDeleteEventName            event.EventName = "ActivityDelete"
	AddServiceToActivityEventName      event.EventName = "AddServiceToActivity"
	RemoveServiceFromActivityEventName event.EventName = "RemoveServiceFromActivity"
	AddIssueToActivityEventName        event.EventName = "AddIssueToActivity"
	RemoveIssueFromActivityEventName   event.EventName = "RemoveIssueFromActivity"
	ListActivitiesEventName            event.EventName = "ListActivities"
	GetActivityEventName               event.EventName = "GetActivity"
)

type ActivityCreateEvent struct {
	Activity *entity.Activity
}

func (e ActivityCreateEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ActivityCreateEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (a *ActivityCreateEvent) Name() event.EventName {
	return ActivityCreateEventName
}

type ActivityUpdateEvent struct {
	Activity *entity.Activity
}

func (e ActivityUpdateEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ActivityUpdateEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (a *ActivityUpdateEvent) Name() event.EventName {
	return ActivityUpdateEventName
}

type ActivityDeleteEvent struct {
	ActivityID int64
}

func (e ActivityDeleteEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ActivityDeleteEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (a *ActivityDeleteEvent) Name() event.EventName {
	return ActivityDeleteEventName
}

type AddServiceToActivityEvent struct {
	ActivityID int64
	ServiceID  int64
}

func (e AddServiceToActivityEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &AddServiceToActivityEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (a *AddServiceToActivityEvent) Name() event.EventName {
	return AddServiceToActivityEventName
}

type RemoveServiceFromActivityEvent struct {
	ActivityID int64
	ServiceID  int64
}

func (e RemoveServiceFromActivityEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &RemoveServiceFromActivityEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (a *RemoveServiceFromActivityEvent) Name() event.EventName {
	return RemoveServiceFromActivityEventName
}

type AddIssueToActivityEvent struct {
	ActivityID int64
	IssueID    int64
}

func (e AddIssueToActivityEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &AddIssueToActivityEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (a *AddIssueToActivityEvent) Name() event.EventName {
	return AddIssueToActivityEventName
}

type RemoveIssueFromActivityEvent struct {
	ActivityID int64
	IssueID    int64
}

func (e RemoveIssueFromActivityEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &RemoveIssueFromActivityEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (a *RemoveIssueFromActivityEvent) Name() event.EventName {
	return RemoveIssueFromActivityEventName
}

type ListActivitiesEvent struct {
	Filter     *entity.ActivityFilter
	Options    *entity.ListOptions
	Activities *entity.List[entity.ActivityResult]
}

func (e ListActivitiesEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListActivitiesEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (l *ListActivitiesEvent) Name() event.EventName {
	return ListActivitiesEventName
}

type GetActivityEvent struct {
	ActivityID int64
	Activity   *entity.Activity
}

func (e GetActivityEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &GetActivityEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (g *GetActivityEvent) Name() event.EventName {
	return GetActivityEventName
}
