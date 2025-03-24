// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
)

const (
	ListComponentsEventName     event.EventName = "ListComponents"
	CreateComponentEventName    event.EventName = "CreateComponent"
	UpdateComponentEventName    event.EventName = "UpdateComponent"
	DeleteComponentEventName    event.EventName = "DeleteComponent"
	ListComponentCcrnsEventName event.EventName = "ListComponentCcrns"
)

type ListComponentsEvent struct {
	Filter     *entity.ComponentFilter
	Options    *entity.ListOptions
	Components *entity.List[entity.ComponentResult]
}

func (e ListComponentsEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListComponentsEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListComponentsEvent) Name() event.EventName {
	return ListComponentsEventName
}

type CreateComponentEvent struct {
	Component *entity.Component
}

func (e CreateComponentEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateComponentEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *CreateComponentEvent) Name() event.EventName {
	return CreateComponentEventName
}

type UpdateComponentEvent struct {
	Component *entity.Component
}

func (e UpdateComponentEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateComponentEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *UpdateComponentEvent) Name() event.EventName {
	return UpdateComponentEventName
}

type DeleteComponentEvent struct {
	ComponentID int64
}

func (e DeleteComponentEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &DeleteComponentEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *DeleteComponentEvent) Name() event.EventName {
	return DeleteComponentEventName
}

type ListComponentCcrnsEvent struct {
	Filter  *entity.ComponentFilter
	Options *entity.ListOptions
	CCRNs   []string
}

func (e ListComponentCcrnsEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListComponentCcrnsEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListComponentCcrnsEvent) Name() event.EventName {
	return ListComponentCcrnsEventName
}
