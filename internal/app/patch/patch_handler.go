// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package patch

import (
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

var (
	CacheTtlGetPatches         = 12 * time.Hour
	CacheTtlGetAllPatchCursors = 12 * time.Hour
	CacheTtlCountPatches       = 12 * time.Hour
)

type patchHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	cache         cache.Cache
	logger        *logrus.Logger
}

func NewPatchHandler(handlerContext common.HandlerContext) PatchHandler {
	return &patchHandler{
		database:      handlerContext.DB,
		eventRegistry: handlerContext.EventReg,
		cache:         handlerContext.Cache,
		logger:        logrus.New(),
	}
}

func (ph *patchHandler) ListPatches(filter *entity.PatchFilter, options *entity.ListOptions) (*entity.List[entity.PatchResult], error) {
	op := appErrors.Op("patchHandler.ListPatches")
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginatedX(&filter.PaginatedX)

	res, err := cache.CallCached[[]entity.PatchResult](
		ph.cache,
		CacheTtlGetPatches,
		"GetPatches",
		ph.database.GetPatches,
		filter,
		options.Order,
	)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Patches", "", err)
		applog.LogError(ph.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := cache.CallCached[[]string](
				ph.cache,
				CacheTtlGetAllPatchCursors,
				"GetAllPatchCursors",
				ph.database.GetAllPatchCursors,
				filter,
				options.Order,
			)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), "PatchCursors", "", err)
				applog.LogError(ph.logger, wrappedErr, logrus.Fields{
					"filter": filter,
				})
				return nil, wrappedErr
			}
			pageInfo = common.GetPageInfoX(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = cache.CallCached[int64](
			ph.cache,
			CacheTtlCountPatches,
			"CountPatches",
			ph.database.CountPatches,
			filter,
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "PatchCount", "", err)
			applog.LogError(ph.logger, wrappedErr, logrus.Fields{
				"filter": filter,
			})
			return nil, wrappedErr
		}
	}

	result := &entity.List[entity.PatchResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	ph.eventRegistry.PushEvent(&ListPatchesEvent{
		Filter:  filter,
		Options: options,
		Patches: result,
	})

	return result, nil
}
