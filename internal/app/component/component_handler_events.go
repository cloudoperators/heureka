// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/entity"
)

const (
	ListComponentsEventName                  event.EventName = "ListComponents"
	CreateComponentEventName                 event.EventName = "CreateComponent"
	UpdateComponentEventName                 event.EventName = "UpdateComponent"
	DeleteComponentEventName                 event.EventName = "DeleteComponent"
	ListComponentCcrnsEventName              event.EventName = "ListComponentCcrns"
	GetComponentIssueSeverityCountsEventName event.EventName = "GetComponentIssueSeverityCounts"
)

type ListComponentsEvent struct {
	Filter     *entity.ComponentFilter
	Options    *entity.ListOptions
	Components *entity.List[entity.ComponentResult]
}

func (e *ListComponentsEvent) Name() event.EventName {
	return ListComponentsEventName
}

type CreateComponentEvent struct {
	Component *entity.Component
}

func (e *CreateComponentEvent) Name() event.EventName {
	return CreateComponentEventName
}

type UpdateComponentEvent struct {
	Component *entity.Component
}

func (e *UpdateComponentEvent) Name() event.EventName {
	return UpdateComponentEventName
}

type DeleteComponentEvent struct {
	ComponentID int64
}

func (e *DeleteComponentEvent) Name() event.EventName {
	return DeleteComponentEventName
}

type ListComponentCcrnsEvent struct {
	Filter  *entity.ComponentFilter
	Options *entity.ListOptions
	CCRNs   []string
}

func (e *ListComponentCcrnsEvent) Name() event.EventName {
	return ListComponentCcrnsEventName
}

type GetComponentIssueSeverityCountsEvent struct {
	Filter *entity.ComponentFilter
	Counts []entity.IssueSeverityCounts
}

func (e *GetComponentIssueSeverityCountsEvent) Name() event.EventName {
	return GetComponentIssueSeverityCountsEventName
}
