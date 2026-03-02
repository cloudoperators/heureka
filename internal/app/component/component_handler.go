// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/openfga"

	"github.com/sirupsen/logrus"
)

var (
	CacheTtlGetComponentCcrns      = 12 * time.Hour
	CacheTtlGetAllComponentCursors = 12 * time.Hour
	CacheTtlCountComponents        = 12 * time.Hour
)

type componentHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	cache         cache.Cache
	openfga       openfga.Authorization
}

func NewComponentHandler(handlerContext common.HandlerContext) ComponentHandler {
	return &componentHandler{
		database:      handlerContext.DB,
		eventRegistry: handlerContext.EventReg,
		cache:         handlerContext.Cache,
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

func (cs *componentHandler) ListComponents(filter *entity.ComponentFilter, options *entity.ListOptions) (*entity.List[entity.ComponentResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)
	options = common.EnsureListOptions(options)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListComponentsEventName,
		"filter": filter,
	})

	res, err := cs.database.GetComponents(filter, options.Order)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Error while filtering for Components")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := cache.CallCached[[]string](
				cs.cache,
				CacheTtlGetAllComponentCursors,
				"GetAllComponentCursors",
				cs.database.GetAllComponentCursors,
				filter,
				options.Order,
			)
			if err != nil {
				l.Error(err)
				return nil, NewUserHandlerError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = cache.CallCached[int64](
			cs.cache,
			CacheTtlCountComponents,
			"CountComponents",
			cs.database.CountComponents,
			filter,
		)
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

func (cs *componentHandler) CreateComponent(ctx context.Context, component *entity.Component) (*entity.Component, error) {
	f := &entity.ComponentFilter{
		CCRN: []*string{&component.CCRN},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateComponentEventName,
		"object": component,
		"filter": f,
	})

	var err error
	component.CreatedBy, err = common.GetCurrentUserId(ctx, cs.database)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while creating component (GetUserId).")
	}
	component.UpdatedBy = component.CreatedBy

	lo := entity.NewListOptions()
	components, err := cs.ListComponents(f, lo)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while creating component.")
	}

	if len(components.Elements) > 0 {
		return nil, NewUserHandlerError(fmt.Sprintf("Duplicated entry %s for ccrn.", component.CCRN))
	}

	newComponent, err := cs.database.CreateComponent(component)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while creating component.")
	}

	cs.eventRegistry.PushEvent(&CreateComponentEvent{Component: newComponent})

	return newComponent, nil
}

func (cs *componentHandler) UpdateComponent(ctx context.Context, component *entity.Component) (*entity.Component, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateComponentEventName,
		"object": component,
	})

	var err error
	component.UpdatedBy, err = common.GetCurrentUserId(ctx, cs.database)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while updating component (GetUserId).")
	}

	err = cs.database.UpdateComponent(component)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while updating component.")
	}

	lo := entity.NewListOptions()
	componentResult, err := cs.ListComponents(&entity.ComponentFilter{Id: []*int64{&component.Id}}, lo)
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

func (cs *componentHandler) DeleteComponent(ctx context.Context, id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteComponentEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(ctx, cs.database)
	if err != nil {
		l.Error(err)
		return NewUserHandlerError("Internal error while deleting component (GetUserId).")
	}

	err = cs.database.DeleteComponent(id, userId)
	if err != nil {
		l.Error(err)
		return NewUserHandlerError("Internal error while deleting component.")
	}

	cs.eventRegistry.PushEvent(&DeleteComponentEvent{ComponentID: id})

	return nil
}

func (cs *componentHandler) ListComponentCcrns(filter *entity.ComponentFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListComponentCcrnsEventName,
		"filter": filter,
	})

	componentCcrns, err := cache.CallCached[[]string](
		cs.cache,
		CacheTtlGetComponentCcrns,
		"GetComponentCcrns",
		cs.database.GetComponentCcrns,
		filter,
	)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while retrieving componentCcrns.")
	}

	cs.eventRegistry.PushEvent(&ListComponentCcrnsEvent{Filter: filter, Options: options, CCRNs: componentCcrns})

	return componentCcrns, nil
}

func (cs *componentHandler) GetComponentVulnerabilityCounts(filter *entity.ComponentFilter) ([]entity.IssueSeverityCounts, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  GetComponentIssueSeverityCountsEventName,
		"filter": filter,
	})

	counts, err := cs.database.CountComponentVulnerabilities(filter)
	if err != nil {
		l.Error(err)
		return nil, NewUserHandlerError("Internal error while retrieving issue severity counts.")
	}

	cs.eventRegistry.PushEvent(&GetComponentIssueSeverityCountsEvent{Filter: filter, Counts: counts})

	return counts, nil
}
