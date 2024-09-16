// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

type componentInstanceHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewComponentInstanceHandler(database database.Database, eventRegistry event.EventRegistry) ComponentInstanceHandler {
	return &componentInstanceHandler{
		database:      database,
		eventRegistry: eventRegistry,
	}
}

type ComponentInstanceHandlerError struct {
	message string
}

func NewComponentInstanceHandlerError(message string) *ComponentInstanceHandlerError {
	return &ComponentInstanceHandlerError{message: message}
}

func (e *ComponentInstanceHandlerError) Error() string {
	return e.message
}

func (ci *componentInstanceHandler) getComponentInstanceResults(filter *entity.ComponentInstanceFilter) ([]entity.ComponentInstanceResult, error) {
	var componentInstanceResults []entity.ComponentInstanceResult
	entries, err := ci.database.GetComponentInstances(filter)
	if err != nil {
		return nil, err
	}

	for _, ci := range entries {
		componentInstance := ci
		cursor := fmt.Sprintf("%d", componentInstance.Id)
		componentInstanceResults = append(componentInstanceResults, entity.ComponentInstanceResult{
			WithCursor:                    entity.WithCursor{Value: cursor},
			ComponentInstanceAggregations: nil,
			ComponentInstance:             &componentInstance,
		})
	}

	return componentInstanceResults, nil
}

func (ci *componentInstanceHandler) ListComponentInstances(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) (*entity.List[entity.ComponentInstanceResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListComponentInstancesEventName,
		"filter": filter,
	})

	res, err := ci.getComponentInstanceResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Error while filtering for ComponentInstances")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := ci.database.GetAllComponentInstanceIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewComponentInstanceHandlerError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = ci.database.CountComponentInstances(filter)
		if err != nil {
			l.Error(err)
			return nil, NewComponentInstanceHandlerError("Error while total count of ComponentInstances")
		}
	}

	ci.eventRegistry.PushEvent(&ListComponentInstancesEvent{
		Filter:  filter,
		Options: options,
		ComponentInstances: &entity.List[entity.ComponentInstanceResult]{
			TotalCount: &count,
			PageInfo:   pageInfo,
			Elements:   res,
		},
	})

	return &entity.List[entity.ComponentInstanceResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (ci *componentInstanceHandler) CreateComponentInstance(componentInstance *entity.ComponentInstance) (*entity.ComponentInstance, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateComponentInstanceEventName,
		"object": componentInstance,
	})

	newComponentInstance, err := ci.database.CreateComponentInstance(componentInstance)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while creating componentInstance.")
	}

	ci.eventRegistry.PushEvent(&CreateComponentInstanceEvent{
		ComponentInstance: newComponentInstance,
	})

	return newComponentInstance, nil
}

func (ci *componentInstanceHandler) UpdateComponentInstance(componentInstance *entity.ComponentInstance) (*entity.ComponentInstance, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateComponentInstanceEventName,
		"object": componentInstance,
	})

	err := ci.database.UpdateComponentInstance(componentInstance)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while updating componentInstance.")
	}

	componentInstanceResult, err := ci.ListComponentInstances(&entity.ComponentInstanceFilter{Id: []*int64{&componentInstance.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving updated componentInstance.")
	}

	if len(componentInstanceResult.Elements) != 1 {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Multiple componentInstances found.")
	}

	ci.eventRegistry.PushEvent(&UpdateComponentInstanceEvent{
		ComponentInstance: componentInstance,
	})

	return componentInstanceResult.Elements[0].ComponentInstance, nil
}

func (ci *componentInstanceHandler) DeleteComponentInstance(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteComponentInstanceEventName,
		"id":    id,
	})

	err := ci.database.DeleteComponentInstance(id)

	if err != nil {
		l.Error(err)
		return NewComponentInstanceHandlerError("Internal error while deleting componentInstance.")
	}

	ci.eventRegistry.PushEvent(&DeleteComponentInstanceEvent{
		ComponentInstanceID: id,
	})

	return nil
}
