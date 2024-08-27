package component_version

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
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

func (e *ListComponentVersionsEvent) Name() event.EventName {
	return ListComponentVersionsEventName
}

type CreateComponentVersionEvent struct {
	ComponentVersion *entity.ComponentVersion
}

func (e *CreateComponentVersionEvent) Name() event.EventName {
	return CreateComponentVersionEventName
}

type UpdateComponentVersionEvent struct {
	ComponentVersion *entity.ComponentVersion
}

func (e *UpdateComponentVersionEvent) Name() event.EventName {
	return UpdateComponentVersionEventName
}

type DeleteComponentVersionEvent struct {
	ComponentVersionID int64
}

func (e *DeleteComponentVersionEvent) Name() event.EventName {
	return DeleteComponentVersionEventName
}
