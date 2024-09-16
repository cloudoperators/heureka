// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

type componentVersionHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewComponentVersionHandler(database database.Database, eventRegistry event.EventRegistry) ComponentVersionHandler {
	return &componentVersionHandler{
		database:      database,
		eventRegistry: eventRegistry,
	}
}

type ComponentVersionHandlerError struct {
	message string
}

func NewComponentVersionHandlerError(message string) *ComponentVersionHandlerError {
	return &ComponentVersionHandlerError{message: message}
}

func (e *ComponentVersionHandlerError) Error() string {
	return e.message
}

func (cv *componentVersionHandler) getComponentVersionResults(filter *entity.ComponentVersionFilter) ([]entity.ComponentVersionResult, error) {
	var componentVersionResults []entity.ComponentVersionResult
	componentVersions, err := cv.database.GetComponentVersions(filter)
	if err != nil {
		return nil, err
	}
	for _, cv := range componentVersions {
		componentVersion := cv
		cursor := fmt.Sprintf("%d", componentVersion.Id)
		componentVersionResults = append(componentVersionResults, entity.ComponentVersionResult{
			WithCursor:                   entity.WithCursor{Value: cursor},
			ComponentVersionAggregations: nil,
			ComponentVersion:             &componentVersion,
		})
	}
	return componentVersionResults, nil
}

func (cv *componentVersionHandler) ListComponentVersions(filter *entity.ComponentVersionFilter, options *entity.ListOptions) (*entity.List[entity.ComponentVersionResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListComponentVersionsEventName,
		"filter": filter,
	})

	res, err := cv.getComponentVersionResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionHandlerError("Error while filtering for ComponentVersions")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := cv.database.GetAllComponentVersionIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewComponentVersionHandlerError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = cv.database.CountComponentVersions(filter)
		if err != nil {
			l.Error(err)
			return nil, NewComponentVersionHandlerError("Error while total count of ComponentVersions")
		}
	}

	cv.eventRegistry.PushEvent(&ListComponentVersionsEvent{
		Filter:  filter,
		Options: options,
		ComponentVersions: &entity.List[entity.ComponentVersionResult]{
			TotalCount: &count,
			PageInfo:   pageInfo,
			Elements:   res,
		},
	})

	return &entity.List[entity.ComponentVersionResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (cv *componentVersionHandler) CreateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateComponentVersionEventName,
		"object": componentVersion,
	})

	newComponent, err := cv.database.CreateComponentVersion(componentVersion)

	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionHandlerError("Internal error while creating componentVersion.")
	}

	cv.eventRegistry.PushEvent(&CreateComponentVersionEvent{
		ComponentVersion: newComponent,
	})

	return newComponent, nil
}

func (cv *componentVersionHandler) UpdateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateComponentVersionEventName,
		"object": componentVersion,
	})

	err := cv.database.UpdateComponentVersion(componentVersion)

	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionHandlerError("Internal error while updating componentVersion.")
	}

	componentVersionResult, err := cv.ListComponentVersions(&entity.ComponentVersionFilter{Id: []*int64{&componentVersion.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionHandlerError("Internal error while retrieving updated componentVersion.")
	}

	if len(componentVersionResult.Elements) != 1 {
		l.Error(err)
		return nil, NewComponentVersionHandlerError("Multiple componentVersions found.")
	}

	cv.eventRegistry.PushEvent(&UpdateComponentVersionEvent{
		ComponentVersion: componentVersion,
	})

	return componentVersionResult.Elements[0].ComponentVersion, nil
}

func (cv *componentVersionHandler) DeleteComponentVersion(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteComponentVersionEventName,
		"id":    id,
	})

	err := cv.database.DeleteComponentVersion(id)

	if err != nil {
		l.Error(err)
		return NewComponentVersionHandlerError("Internal error while deleting componentVersion.")
	}

	cv.eventRegistry.PushEvent(&DeleteComponentVersionEvent{
		ComponentVersionID: id,
	})

	return nil
}
