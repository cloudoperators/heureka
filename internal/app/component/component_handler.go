// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"

	"github.com/sirupsen/logrus"

	"github.com/cloudoperators/heureka/internal/entity"
)

type componentHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewComponentHandler(db database.Database, er event.EventRegistry) ComponentHandler {
	return &componentHandler{
		database:      db,
		eventRegistry: er,
	}
}

type ComponentHandlerError struct {
	msg string
}

func (e *ComponentHandlerError) Error() string {
	return fmt.Sprintf("ServiceHandlerError: %s", e.msg)
}

func NewUserHandlerError(msg string) *ComponentHandlerError {
	return &ComponentHandlerError{msg: msg}
}

func (cs *componentHandler) getComponentResults(filter *entity.ComponentFilter) ([]entity.ComponentResult, error) {
	var componentResults []entity.ComponentResult
	components, err := cs.database.GetComponents(filter)
	if err != nil {
		return nil, err
	}
	for _, c := range components {
		component := c
		cursor := fmt.Sprintf("%d", component.Id)
		componentResults = append(componentResults, entity.ComponentResult{
			WithCursor:            entity.WithCursor{Value: cursor},
			ComponentAggregations: nil,
			Component:             &component,
		})
	}
	return componentResults, nil
}

func (cs *componentHandler) ListComponents(filter *entity.ComponentFilter, options *entity.ListOptions) (*entity.List[entity.ComponentResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListComponentsEventName,
		"filter": filter,
	})

	res, err := cs.getComponentResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Error while filtering for Components")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := cs.database.GetAllComponentIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewUserHandlerError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = cs.database.CountComponents(filter)
		if err != nil {
			l.Error(err)
			return nil, NewUserHandlerError("Error while total count of Components")
		}
	}

	ret := &entity.List[entity.ComponentResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	cs.eventRegistry.PushEvent(&ListComponentsEvent{Filter: filter, Options: options, Components: ret})

	return ret, nil
}

func (cs *componentHandler) CreateComponent(component *entity.Component) (*entity.Component, error) {
	f := &entity.ComponentFilter{
		Name: []*string{&component.Name},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateComponentEventName,
		"object": component,
		"filter": f,
	})

	components, err := cs.ListComponents(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while creating component.")
	}

	if len(components.Elements) > 0 {
		return nil, NewUserHandlerError(fmt.Sprintf("Duplicated entry %s for name.", component.Name))
	}

	newComponent, err := cs.database.CreateComponent(component)

	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while creating component.")
	}

	cs.eventRegistry.PushEvent(&CreateComponentEvent{Component: newComponent})

	return newComponent, nil
}

func (cs *componentHandler) UpdateComponent(component *entity.Component) (*entity.Component, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateComponentEventName,
		"object": component,
	})

	err := cs.database.UpdateComponent(component)

	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while updating component.")
	}

	componentResult, err := cs.ListComponents(&entity.ComponentFilter{Id: []*int64{&component.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while retrieving updated component.")
	}

	if len(componentResult.Elements) != 1 {
		l.Error(err)
		return nil, NewUserHandlerError("Multiple components found.")
	}

	cs.eventRegistry.PushEvent(&UpdateComponentEvent{Component: component})

	return componentResult.Elements[0].Component, nil
}

func (cs *componentHandler) DeleteComponent(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteComponentEventName,
		"id":    id,
	})

	err := cs.database.DeleteComponent(id)

	if err != nil {
		l.Error(err)
		return NewUserHandlerError("Internal error while deleting component.")
	}

	cs.eventRegistry.PushEvent(&DeleteComponentEvent{ComponentID: id})

	return nil
}

func (cs *componentHandler) ListComponentNames(filter *entity.ComponentFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListComponentNamesEventName,
		"filter": filter,
	})

	componentNames, err := cs.database.GetComponentNames(filter)

	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while retrieving componentNames.")
	}

	cs.eventRegistry.PushEvent(&ListComponentNamesEvent{Filter: filter, Options: options, Names: componentNames})

	return componentNames, nil
}
