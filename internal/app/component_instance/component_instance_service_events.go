package component_instance

import (
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

const (
	ListComponentInstancesEventName  event.EventName = "ListComponentInstances"
	CreateComponentInstanceEventName event.EventName = "CreateComponentInstance"
	UpdateComponentInstanceEventName event.EventName = "UpdateComponentInstance"
	DeleteComponentInstanceEventName event.EventName = "DeleteComponentInstance"
)

type ListComponentInstancesEvent struct {
	Filter             *entity.ComponentInstanceFilter
	Options            *entity.ListOptions
	ComponentInstances *entity.List[entity.ComponentInstanceResult]
}

func (e *ListComponentInstancesEvent) Name() event.EventName {
	return ListComponentInstancesEventName
}

type CreateComponentInstanceEvent struct {
	ComponentInstance *entity.ComponentInstance
}

func (e *CreateComponentInstanceEvent) Name() event.EventName {
	return CreateComponentInstanceEventName
}

type UpdateComponentInstanceEvent struct {
	ComponentInstance *entity.ComponentInstance
}

func (e *UpdateComponentInstanceEvent) Name() event.EventName {
	return UpdateComponentInstanceEventName
}

type DeleteComponentInstanceEvent struct {
	ComponentInstanceID int64
}

func (e *DeleteComponentInstanceEvent) Name() event.EventName {
	return DeleteComponentInstanceEventName
}
