// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
)

const (
	ListSupportGroupsEventName             event.EventName = "ListSupportGroups"
	GetSupportGroupEventName               event.EventName = "GetSupportGroup"
	CreateSupportGroupEventName            event.EventName = "CreateSupportGroup"
	UpdateSupportGroupEventName            event.EventName = "UpdateSupportGroup"
	DeleteSupportGroupEventName            event.EventName = "DeleteSupportGroup"
	AddServiceToSupportGroupEventName      event.EventName = "AddServiceToSupportGroup"
	RemoveServiceFromSupportGroupEventName event.EventName = "RemoveServiceFromSupportGroup"
	AddUserToSupportGroupEventName         event.EventName = "AddUserToSupportGroup"
	RemoveUserFromSupportGroupEventName    event.EventName = "RemoveUserFromSupportGroup"
	ListSupportGroupCcrnsEventName         event.EventName = "ListSupportGroupCcrns"
)

type ListSupportGroupsEvent struct {
	Filter        *entity.SupportGroupFilter
	Options       *entity.ListOptions
	SupportGroups *entity.List[entity.SupportGroupResult]
}

func (e ListSupportGroupsEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListSupportGroupsEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListSupportGroupsEvent) Name() event.EventName {
	return ListSupportGroupsEventName
}

type GetSupportGroupEvent struct {
	SupportGroupID int64
	SupportGroup   *entity.SupportGroup
}

func (e GetSupportGroupEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &GetSupportGroupEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *GetSupportGroupEvent) Name() event.EventName {
	return GetSupportGroupEventName
}

type CreateSupportGroupEvent struct {
	SupportGroup *entity.SupportGroup
}

func (e CreateSupportGroupEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateSupportGroupEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *CreateSupportGroupEvent) Name() event.EventName {
	return CreateSupportGroupEventName
}

type UpdateSupportGroupEvent struct {
	SupportGroup *entity.SupportGroup
}

func (e UpdateSupportGroupEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateSupportGroupEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *UpdateSupportGroupEvent) Name() event.EventName {
	return UpdateSupportGroupEventName
}

type DeleteSupportGroupEvent struct {
	SupportGroupID int64
}

func (e DeleteSupportGroupEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &DeleteSupportGroupEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *DeleteSupportGroupEvent) Name() event.EventName {
	return DeleteSupportGroupEventName
}

type AddServiceToSupportGroupEvent struct {
	SupportGroupID int64
	ServiceID      int64
}

func (e AddServiceToSupportGroupEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &AddServiceToSupportGroupEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *AddServiceToSupportGroupEvent) Name() event.EventName {
	return AddServiceToSupportGroupEventName
}

type RemoveServiceFromSupportGroupEvent struct {
	SupportGroupID int64
	ServiceID      int64
}

func (e RemoveServiceFromSupportGroupEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &RemoveServiceFromSupportGroupEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *RemoveServiceFromSupportGroupEvent) Name() event.EventName {
	return RemoveServiceFromSupportGroupEventName
}

type AddUserToSupportGroupEvent struct {
	SupportGroupID int64
	UserID         int64
}

func (e AddUserToSupportGroupEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &AddUserToSupportGroupEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *AddUserToSupportGroupEvent) Name() event.EventName {
	return AddUserToSupportGroupEventName
}

type RemoveUserFromSupportGroupEvent struct {
	SupportGroupID int64
	UserID         int64
}

func (e RemoveUserFromSupportGroupEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &RemoveUserFromSupportGroupEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *RemoveUserFromSupportGroupEvent) Name() event.EventName {
	return RemoveUserFromSupportGroupEventName
}

type ListSupportGroupCcrnsEvent struct {
	Filter  *entity.SupportGroupFilter
	Options *entity.ListOptions
	Ccrns   []string
}

func (e ListSupportGroupCcrnsEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListSupportGroupCcrnsEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListSupportGroupCcrnsEvent) Name() event.EventName {
	return ListSupportGroupCcrnsEventName
}
