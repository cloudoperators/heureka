// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import (
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

func (ci *componentInstanceHandler) ListComponentInstances(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) (*entity.List[entity.ComponentInstanceResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginatedX(&filter.PaginatedX)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListComponentInstancesEventName,
		"filter": filter,
	})

	res, err := ci.database.GetComponentInstances(filter, options.Order)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Error while filtering for ComponentInstances")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := ci.database.GetAllComponentInstanceCursors(filter, options.Order)
			if err != nil {
				return nil, NewComponentInstanceHandlerError("Error while getting all cursors")
			}
			pageInfo = common.GetPageInfoX(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
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

func (ci *componentInstanceHandler) CreateComponentInstance(componentInstance *entity.ComponentInstance, scannerRunUUID string) (*entity.ComponentInstance, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateComponentInstanceEventName,
		"object": componentInstance,
	})

	var err error
	componentInstance.CreatedBy, err = common.GetCurrentUserId(ci.database)
	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while creating componentInstance (GetUserId).")
	}
	componentInstance.UpdatedBy = componentInstance.CreatedBy

	newComponentInstance, err := ci.database.CreateComponentInstance(componentInstance)
	ci.database.CreateScannerRunComponentInstanceTracker(newComponentInstance.Id, scannerRunUUID)

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

	var err error
	componentInstance.UpdatedBy, err = common.GetCurrentUserId(ci.database)
	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while updating componentInstance (GetUserId).")
	}

	err = ci.database.UpdateComponentInstance(componentInstance)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while updating componentInstance.")
	}

	lo := entity.NewListOptions()

	componentInstanceResult, err := ci.ListComponentInstances(&entity.ComponentInstanceFilter{Id: []*int64{&componentInstance.Id}}, lo)

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

	userId, err := common.GetCurrentUserId(ci.database)
	if err != nil {
		l.Error(err)
		return NewComponentInstanceHandlerError("Internal error while deleting componentInstance (GetUserId).")
	}

	err = ci.database.DeleteComponentInstance(id, userId)

	if err != nil {
		l.Error(err)
		return NewComponentInstanceHandlerError("Internal error while deleting componentInstance.")
	}

	ci.eventRegistry.PushEvent(&DeleteComponentInstanceEvent{
		ComponentInstanceID: id,
	})

	return nil
}
func (s *componentInstanceHandler) ListCcrns(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListCcrnEventName,
		"filter": filter,
	})

	ccrn, err := s.database.GetCcrn(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Ccrn.")
	}

	s.eventRegistry.PushEvent(&ListCcrnEvent{Filter: filter, Ccrn: ccrn})

	return ccrn, nil
}
