// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version

import (
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/event"
)

const (
	ListComponentVersionsEventName  event.EventName = "ListComponentVersions"
	CreateComponentVersionEventName event.EventName = "CreateComponentVersion"
	UpdateComponentVersionEventName event.EventName = "UpdateComponentVersion"
	DeleteComponentVersionEventName event.EventName = "DeleteComponentVersion"
)

type ListComponentVersionsEvent struct {
	Filter            *entity.ComponentVersionFilter
	Options           *entity.ListOptions
	ComponentVersions *entity.List[entity.ComponentVersionResult]
}

func (e ListComponentVersionsEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &ListComponentVersionsEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *ListComponentVersionsEvent) Name() event.EventName {
	return ListComponentVersionsEventName
}

type CreateComponentVersionEvent struct {
	ComponentVersion *entity.ComponentVersion
}

func (e CreateComponentVersionEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &CreateComponentVersionEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *CreateComponentVersionEvent) Name() event.EventName {
	return CreateComponentVersionEventName
}

type UpdateComponentVersionEvent struct {
	ComponentVersion *entity.ComponentVersion
}

func (e UpdateComponentVersionEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &UpdateComponentVersionEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *UpdateComponentVersionEvent) Name() event.EventName {
	return UpdateComponentVersionEventName
}

type DeleteComponentVersionEvent struct {
	ComponentVersionID int64
}

func (e DeleteComponentVersionEvent) Unmarshal(data []byte) (event.Event, error) {
	event := &DeleteComponentVersionEvent{}
	err := json.Unmarshal(data, event)
	return event, err
}

func (e *DeleteComponentVersionEvent) Name() event.EventName {
	return DeleteComponentVersionEventName
}
