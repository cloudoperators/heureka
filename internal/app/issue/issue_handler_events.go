// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

const (
	CreateIssueEventName                     event.EventName = "CreateIssue"
	UpdateIssueEventName                     event.EventName = "UpdateIssue"
	DeleteIssueEventName                     event.EventName = "DeleteIssue"
	AddComponentVersionToIssueEventName      event.EventName = "AddComponentVersionToIssue"
	RemoveComponentVersionFromIssueEventName event.EventName = "RemoveComponentVersionFromIssue"
	ListIssuesEventName                      event.EventName = "ListIssues"
	GetIssueEventName                        event.EventName = "GetIssue"
	ListIssueNamesEventName                  event.EventName = "ListIssueNames"
)

type CreateIssueEvent struct {
	Issue *entity.Issue
}

func (e *CreateIssueEvent) Name() event.EventName {
	return CreateIssueEventName
}

type UpdateIssueEvent struct {
	Issue *entity.Issue
}

func (e *UpdateIssueEvent) Name() event.EventName {
	return UpdateIssueEventName
}

type DeleteIssueEvent struct {
	IssueID int64
}

func (e *DeleteIssueEvent) Name() event.EventName {
	return DeleteIssueEventName
}

type AddComponentVersionToIssueEvent struct {
	IssueID            int64
	ComponentVersionID int64
}

func (e *AddComponentVersionToIssueEvent) Name() event.EventName {
	return AddComponentVersionToIssueEventName
}

type RemoveComponentVersionFromIssueEvent struct {
	IssueID            int64
	ComponentVersionID int64
}

func (e *RemoveComponentVersionFromIssueEvent) Name() event.EventName {
	return RemoveComponentVersionFromIssueEventName
}

type ListIssuesEvent struct {
	Filter  *entity.IssueFilter
	Options *entity.IssueListOptions
	Issues  *entity.IssueList
}

func (e *ListIssuesEvent) Name() event.EventName {
	return ListIssuesEventName
}

type GetIssueEvent struct {
	IssueID int64
	Issue   *entity.Issue
}

func (e *GetIssueEvent) Name() event.EventName {
	return GetIssueEventName
}

type ListIssueNamesEvent struct {
	Filter  *entity.IssueFilter
	Options *entity.ListOptions
	Names   []string
}

func (e *ListIssueNamesEvent) Name() event.EventName {
	return ListIssueNamesEventName
}
