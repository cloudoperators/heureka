// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
)

const (
	ListComponentInstancesEventName  event.EventName = "ListComponentInstances"
	CreateComponentInstanceEventName event.EventName = "CreateComponentInstance"
	UpdateComponentInstanceEventName event.EventName = "UpdateComponentInstance"
	DeleteComponentInstanceEventName event.EventName = "DeleteComponentInstance"
	ListCcrnEventName                event.EventName = "ListCcrn"
)

type ListComponentInstancesEvent struct {
	Filter             *entity.ComponentInstanceFilter `json:"filter"`
	Options            *entity.ListOptions
	ComponentInstances *entity.List[entity.ComponentInstanceResult]
}

func (e ListComponentInstancesEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListComponentInstancesEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListComponentInstancesEvent) Name() event.EventName {
	return ListComponentInstancesEventName
}

type CreateComponentInstanceEvent struct {
	ComponentInstance *entity.ComponentInstance
}

func (e CreateComponentInstanceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateComponentInstanceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *CreateComponentInstanceEvent) Name() event.EventName {
	return CreateComponentInstanceEventName
}

type UpdateComponentInstanceEvent struct {
	ComponentInstance *entity.ComponentInstance
}

func (e UpdateComponentInstanceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateComponentInstanceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *UpdateComponentInstanceEvent) Name() event.EventName {
	return UpdateComponentInstanceEventName
}

type DeleteComponentInstanceEvent struct {
	ComponentInstanceID int64
}

func (e DeleteComponentInstanceEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &DeleteComponentInstanceEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *DeleteComponentInstanceEvent) Name() event.EventName {
	return DeleteComponentInstanceEventName
}

type ListCcrnEvent struct {
	Filter *entity.ComponentInstanceFilter
	Ccrn   []string
}

func (e ListCcrnEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListCcrnEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListCcrnEvent) Name() event.EventName {
	return ListCcrnEventName
}
