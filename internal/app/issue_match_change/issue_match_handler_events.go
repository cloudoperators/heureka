// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match_change

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
)

const (
	ListIssueMatchChangesEventName  event.EventName = "ListIssueMatchChanges"
	CreateIssueMatchChangeEventName event.EventName = "CreateIssueMatchChange"
	UpdateIssueMatchChangeEventName event.EventName = "UpdateIssueMatchChange"
	DeleteIssueMatchChangeEventName event.EventName = "DeleteIssueMatchChange"
)

type ListIssueMatchChangesEvent struct {
	Filter  *entity.IssueMatchChangeFilter
	Options *entity.ListOptions
	Results *entity.List[entity.IssueMatchChangeResult]
}

func (e ListIssueMatchChangesEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListIssueMatchChangesEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListIssueMatchChangesEvent) Name() event.EventName {
	return ListIssueMatchChangesEventName
}

type CreateIssueMatchChangeEvent struct {
	IssueMatchChange *entity.IssueMatchChange
}

func (e CreateIssueMatchChangeEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateIssueMatchChangeEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *CreateIssueMatchChangeEvent) Name() event.EventName {
	return CreateIssueMatchChangeEventName
}

type UpdateIssueMatchChangeEvent struct {
	IssueMatchChange *entity.IssueMatchChange
}

func (e UpdateIssueMatchChangeEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateIssueMatchChangeEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *UpdateIssueMatchChangeEvent) Name() event.EventName {
	return UpdateIssueMatchChangeEventName
}

type DeleteIssueMatchChangeEvent struct {
	IssueMatchChangeID int64
}

func (e DeleteIssueMatchChangeEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &DeleteIssueMatchChangeEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *DeleteIssueMatchChangeEvent) Name() event.EventName {
	return DeleteIssueMatchChangeEventName
}
