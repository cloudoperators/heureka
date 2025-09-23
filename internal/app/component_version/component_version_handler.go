// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/sirupsen/logrus"
)

var CacheTtlGetComponentVersions = 12 * time.Hour
var CacheTtlGetAllComponentVersionCursors = 12 * time.Hour
var CacheTtlCountComponentVersions = 12 * time.Hour

type componentVersionHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	logger        *logrus.Logger
	cache         cache.Cache
}

func NewComponentVersionHandler(handlerContext common.HandlerContext) ComponentVersionHandler {
	return &componentVersionHandler{
		database:      handlerContext.DB,
		eventRegistry: handlerContext.EventReg,
		cache:         handlerContext.Cache,
	}
}

func (cv *componentVersionHandler) ListComponentVersions(filter *entity.ComponentVersionFilter, options *entity.ListOptions) (*entity.List[entity.ComponentVersionResult], error) {
	op := appErrors.Op("componentVersionHandler.ListComponentVersions")

	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginatedX(&filter.PaginatedX)

	res, err := cache.CallCached[[]entity.ComponentVersionResult](
		cv.cache,
		CacheTtlGetComponentVersions,
		"GetComponentVersions",
		cv.database.GetComponentVersions,
		filter,
		options.Order,
	)

	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentVersions", "", err)
		applog.LogError(cv.logger, wrappedErr, logrus.Fields{
			"filter":  filter,
			"options": options,
		})
		return nil, wrappedErr
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
				wrappedErr := appErrors.InternalError(string(op), "ComponentVersionCursors", "", err)
				applog.LogError(cv.logger, wrappedErr, logrus.Fields{
					"filter": filter,
				})
				return nil, wrappedErr
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
			wrappedErr := appErrors.InternalError(string(op), "ComponentVersionCount", "", err)
			applog.LogError(cv.logger, wrappedErr, logrus.Fields{
				"filter": filter,
			})
			return nil, wrappedErr
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

func (cv *componentVersionHandler) CreateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	op := appErrors.Op("componentVersionHandler.CreateComponentVersion")

	var err error
	componentVersion.CreatedBy, err = common.GetCurrentUserId(cv.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentVersion", "", err)
		applog.LogError(cv.logger, wrappedErr, logrus.Fields{
			"version":      componentVersion.Version,
			"component_id": componentVersion.ComponentId,
		})
		return nil, wrappedErr
	}
	componentVersion.UpdatedBy = componentVersion.CreatedBy

	newComponent, err := cv.database.CreateComponentVersion(componentVersion)
	if err != nil {
		duplicateEntryError := &database.DuplicateEntryDatabaseError{}
		if errors.As(err, &duplicateEntryError) {
			wrappedErr := appErrors.AlreadyExistsError(string(op), "ComponentVersion", componentVersion.Version)
			applog.LogError(cv.logger, wrappedErr, logrus.Fields{
				"version":      componentVersion.Version,
				"component_id": componentVersion.ComponentId,
			})
			return nil, wrappedErr
		}
		wrappedErr := appErrors.InternalError(string(op), "ComponentVersion", "", err)
		applog.LogError(cv.logger, wrappedErr, logrus.Fields{
			"version":      componentVersion.Version,
			"component_id": componentVersion.ComponentId,
		})
		return nil, wrappedErr
	}

	cv.eventRegistry.PushEvent(&CreateComponentVersionEvent{
		ComponentVersion: newComponent,
	})

	return newComponent, nil
}

func (cv *componentVersionHandler) UpdateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	op := appErrors.Op("componentVersionHandler.UpdateComponentVersion")

	var err error
	componentVersion.UpdatedBy, err = common.GetCurrentUserId(cv.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentVersion", strconv.FormatInt(componentVersion.Id, 10), err)
		applog.LogError(cv.logger, wrappedErr, logrus.Fields{
			"id":      componentVersion.Id,
			"version": componentVersion.Version,
		})
		return nil, wrappedErr
	}

	err = cv.database.UpdateComponentVersion(componentVersion)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentVersion", strconv.FormatInt(componentVersion.Id, 10), err)
		applog.LogError(cv.logger, wrappedErr, logrus.Fields{
			"id":      componentVersion.Id,
			"version": componentVersion.Version,
		})
		return nil, wrappedErr
	}

	lo := entity.NewListOptions()
	componentVersionResult, err := cv.ListComponentVersions(&entity.ComponentVersionFilter{Id: []*int64{&componentVersion.Id}}, lo)
	if err != nil {
		wrappedErr := appErrors.E(op, "ComponentVersion", strconv.FormatInt(componentVersion.Id, 10), appErrors.Internal, err)
		applog.LogError(cv.logger, wrappedErr, logrus.Fields{
			"id":      componentVersion.Id,
			"version": componentVersion.Version,
		})
		return nil, wrappedErr
	}

	if len(componentVersionResult.Elements) != 1 {
		err := appErrors.E(op, "ComponentVersion", strconv.FormatInt(componentVersion.Id, 10), appErrors.Internal,
			fmt.Sprintf("found %d component versions with ID %d, expected 1", len(componentVersionResult.Elements), componentVersion.Id))
		applog.LogError(cv.logger, err, logrus.Fields{
			"id":          componentVersion.Id,
			"found_count": len(componentVersionResult.Elements),
			"version":     componentVersion.Version,
		})
		return nil, err
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

	userId, err := common.GetCurrentUserId(cv.database)
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
