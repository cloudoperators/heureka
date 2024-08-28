// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package activity

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
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

func (a *ActivityCreateEvent) Name() event.EventName {
	return ActivityCreateEventName
}

type ActivityUpdateEvent struct {
	Activity *entity.Activity
}

func (a *ActivityUpdateEvent) Name() event.EventName {
	return ActivityUpdateEventName
}

type ActivityDeleteEvent struct {
	ActivityID int64
}

func (a *ActivityDeleteEvent) Name() event.EventName {
	return ActivityDeleteEventName
}

type AddServiceToActivityEvent struct {
	ActivityID int64
	ServiceID  int64
}

func (a *AddServiceToActivityEvent) Name() event.EventName {
	return AddServiceToActivityEventName
}

type RemoveServiceFromActivityEvent struct {
	ActivityID int64
	ServiceID  int64
}

func (a *RemoveServiceFromActivityEvent) Name() event.EventName {
	return RemoveServiceFromActivityEventName
}

type AddIssueToActivityEvent struct {
	ActivityID int64
	IssueID    int64
}

func (a *AddIssueToActivityEvent) Name() event.EventName {
	return AddIssueToActivityEventName
}

type RemoveIssueFromActivityEvent struct {
	ActivityID int64
	IssueID    int64
}

func (a *RemoveIssueFromActivityEvent) Name() event.EventName {
	return RemoveIssueFromActivityEventName
}

type ListActivitiesEvent struct {
	Filter     *entity.ActivityFilter
	Options    *entity.ListOptions
	Activities *entity.List[entity.ActivityResult]
}

func (l *ListActivitiesEvent) Name() event.EventName {
	return ListActivitiesEventName
}

type GetActivityEvent struct {
	ActivityID int64
	Activity   *entity.Activity
}

func (g *GetActivityEvent) Name() event.EventName {
	return GetActivityEventName
}
