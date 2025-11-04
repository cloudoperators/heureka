// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version

import (
	"context"
	"errors"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

var CacheTtlGetComponentVersions = 12 * time.Hour
var CacheTtlGetAllComponentVersionCursors = 12 * time.Hour
var CacheTtlCountComponentVersions = 12 * time.Hour

type componentVersionHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	cache         cache.Cache
}

func NewComponentVersionHandler(handlerContext common.HandlerContext) ComponentVersionHandler {
	return &componentVersionHandler{
		database:      handlerContext.DB,
		eventRegistry: handlerContext.EventReg,
		cache:         handlerContext.Cache,
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

func (cv *componentVersionHandler) ListComponentVersions(filter *entity.ComponentVersionFilter, options *entity.ListOptions) (*entity.List[entity.ComponentVersionResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginatedX(&filter.PaginatedX)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListComponentVersionsEventName,
		"filter": filter,
	})

	res, err := cache.CallCached[[]entity.ComponentVersionResult](
		cv.cache,
		CacheTtlGetComponentVersions,
		"GetComponentVersions",
		cv.database.GetComponentVersions,
		filter,
		options.Order,
	)

	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionHandlerError("Error while filtering for ComponentVersions")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := cache.CallCached[[]string](
				cv.cache,
				CacheTtlGetAllComponentVersionCursors,
				"GetAllComponentVersionCursors",
				cv.database.GetAllComponentVersionCursors,
				filter,
				options.Order,
			)
			if err != nil {
				l.Error(err)
				return nil, NewComponentVersionHandlerError("Error while getting all cursors")
			}
			pageInfo = common.GetPageInfoX(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = cache.CallCached[int64](
			cv.cache,
			CacheTtlCountComponentVersions,
			"CountComponentVersions",
			cv.database.CountComponentVersions,
			filter,
		)
		if err != nil {
			l.Error(err)
			return nil, NewComponentVersionHandlerError("Error while total count of ComponentVersions")
		}
	}

	ret := &entity.List[entity.ComponentVersionResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	cv.eventRegistry.PushEvent(&ListComponentVersionsEvent{
		Filter:            filter,
		Options:           options,
		ComponentVersions: ret,
	})

	return ret, nil
}

func (cv *componentVersionHandler) CreateComponentVersion(ctx context.Context, componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateComponentVersionEventName,
		"object": componentVersion,
	})

	var err error
	componentVersion.CreatedBy, err = common.GetCurrentUserId(ctx, cv.database)
	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionHandlerError("Internal error while creating componentVersion (GetUserId).")
	}
	componentVersion.UpdatedBy = componentVersion.CreatedBy

	newComponent, err := cv.database.CreateComponentVersion(componentVersion)

	if err != nil {
		l.Error(err)
		duplicateEntryError := &database.DuplicateEntryDatabaseError{}
		if errors.As(err, &duplicateEntryError) {
			return nil, NewComponentVersionHandlerError("Entry already Exists")
		}
		return nil, NewComponentVersionHandlerError("Internal error while creating componentVersion.")
	}

	cv.eventRegistry.PushEvent(&CreateComponentVersionEvent{
		ComponentVersion: newComponent,
	})

	return newComponent, nil
}

func (cv *componentVersionHandler) UpdateComponentVersion(ctx context.Context, componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateComponentVersionEventName,
		"object": componentVersion,
	})

	var err error
	componentVersion.UpdatedBy, err = common.GetCurrentUserId(ctx, cv.database)
	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionHandlerError("Internal error while updating componentVersion (GetUserId).")
	}

	err = cv.database.UpdateComponentVersion(componentVersion)

	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionHandlerError("Internal error while updating componentVersion.")
	}

	lo := entity.NewListOptions()
	componentVersionResult, err := cv.ListComponentVersions(&entity.ComponentVersionFilter{Id: []*int64{&componentVersion.Id}}, lo)

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

func (cv *componentVersionHandler) DeleteComponentVersion(ctx context.Context, id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteComponentVersionEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(ctx, cv.database)
	if err != nil {
		l.Error(err)
		return NewComponentVersionHandlerError("Internal error while deleting componentVersion (GetUserId).")
	}

	err = cv.database.DeleteComponentVersion(id, userId)

	if err != nil {
		l.Error(err)
		return NewComponentVersionHandlerError("Internal error while deleting componentVersion.")
	}

	cv.eventRegistry.PushEvent(&DeleteComponentVersionEvent{
		ComponentVersionID: id,
	})

	return nil
}
