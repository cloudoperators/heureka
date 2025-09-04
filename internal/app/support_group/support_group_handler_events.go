// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity"
)

const (
	ListSupportGroupsBatchEventName        event.EventName = "ListSupportGroupsBatch"
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

type ListSupportGroupsBatchEvent struct {
	Filter        *entity.SupportGroupFilter
	Options       *entity.ListOptions
	SupportGroups *entity.List[entity.SupportGroupBatchResult]
}

func (e *ListSupportGroupsBatchEvent) Name() event.EventName {
	return ListSupportGroupsBatchEventName
}

type ListSupportGroupsEvent struct {
	Filter        *entity.SupportGroupFilter
	Options       *entity.ListOptions
	SupportGroups *entity.List[entity.SupportGroupResult]
}

func (e *ListSupportGroupsEvent) Name() event.EventName {
	return ListSupportGroupsEventName
}

type GetSupportGroupEvent struct {
	SupportGroupID int64
	SupportGroup   *entity.SupportGroup
}

func (e *GetSupportGroupEvent) Name() event.EventName {
	return GetSupportGroupEventName
}

type CreateSupportGroupEvent struct {
	SupportGroup *entity.SupportGroup
}

func (e *CreateSupportGroupEvent) Name() event.EventName {
	return CreateSupportGroupEventName
}

type UpdateSupportGroupEvent struct {
	SupportGroup *entity.SupportGroup
}

func (e *UpdateSupportGroupEvent) Name() event.EventName {
	return UpdateSupportGroupEventName
}

type DeleteSupportGroupEvent struct {
	SupportGroupID int64
}

func (e *DeleteSupportGroupEvent) Name() event.EventName {
	return DeleteSupportGroupEventName
}

type AddServiceToSupportGroupEvent struct {
	SupportGroupID int64
	ServiceID      int64
}

func (e *AddServiceToSupportGroupEvent) Name() event.EventName {
	return AddServiceToSupportGroupEventName
}

type RemoveServiceFromSupportGroupEvent struct {
	SupportGroupID int64
	ServiceID      int64
}

func (e *RemoveServiceFromSupportGroupEvent) Name() event.EventName {
	return RemoveServiceFromSupportGroupEventName
}

type AddUserToSupportGroupEvent struct {
	SupportGroupID int64
	UserID         int64
}

func (e *AddUserToSupportGroupEvent) Name() event.EventName {
	return AddUserToSupportGroupEventName
}

type RemoveUserFromSupportGroupEvent struct {
	SupportGroupID int64
	UserID         int64
}

func (e *RemoveUserFromSupportGroupEvent) Name() event.EventName {
	return RemoveUserFromSupportGroupEventName
}

type ListSupportGroupCcrnsEvent struct {
	Filter  *entity.SupportGroupFilter
	Options *entity.ListOptions
	Ccrns   []string
}

func (e *ListSupportGroupCcrnsEvent) Name() event.EventName {
	return ListSupportGroupCcrnsEventName
}
