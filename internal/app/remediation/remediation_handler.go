// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package remediation

import (
	"context"
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

var (
	CacheTtlGetRemediations          = 12 * time.Hour
	CacheTtlGetAllRemediationCursors = 12 * time.Hour
	CacheTtlCountRemediations        = 12 * time.Hour
)

type remediationHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	cache         cache.Cache
	logger        *logrus.Logger
}

func NewRemediationHandler(handlerContext common.HandlerContext) RemediationHandler {
	return &remediationHandler{
		database:      handlerContext.DB,
		eventRegistry: handlerContext.EventReg,
		cache:         handlerContext.Cache,
		logger:        logrus.New(),
	}
}

func (rh *remediationHandler) ListRemediations(filter *entity.RemediationFilter, options *entity.ListOptions) (*entity.List[entity.RemediationResult], error) {
	op := appErrors.Op("remediationHandler.ListRemediations")
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	res, err := cache.CallCached[[]entity.RemediationResult](
		rh.cache,
		CacheTtlGetRemediations,
		"GetRemediations",
		rh.database.GetRemediations,
		filter,
		options.Order,
	)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Remediations", "", err)
		applog.LogError(rh.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := cache.CallCached[[]string](
				rh.cache,
				CacheTtlGetAllRemediationCursors,
				"GetAllRemediationCursors",
				rh.database.GetAllRemediationCursors,
				filter,
				options.Order,
			)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), "RemediationCursors", "", err)
				applog.LogError(rh.logger, wrappedErr, logrus.Fields{
					"filter": filter,
				})
				return nil, wrappedErr
			}
			pageInfo = common.GetPageInfo(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = cache.CallCached[int64](
			rh.cache,
			CacheTtlCountRemediations,
			"CountRemediations",
			rh.database.CountRemediations,
			filter,
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "RemediationCount", "", err)
			applog.LogError(rh.logger, wrappedErr, logrus.Fields{
				"filter": filter,
			})
			return nil, wrappedErr
		}
	}

	result := &entity.List[entity.RemediationResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	rh.eventRegistry.PushEvent(&ListRemediationsEvent{
		Filter:       filter,
		Options:      options,
		Remediations: result,
	})

	return result, nil
}

func (rh *remediationHandler) CreateRemediation(ctx context.Context, remediation *entity.Remediation) (*entity.Remediation, error) {
	op := appErrors.Op("remediationHandler.CreateRemediation")

	// Input validation - check for required fields
	if remediation == nil {
		err := appErrors.E(op, "Remediation", appErrors.InvalidArgument, "remediation cannot be nil")
		applog.LogError(rh.logger, err, logrus.Fields{})
		return nil, err
	}

	if remediation.Service == "" {
		err := appErrors.E(op, "Remediation", appErrors.InvalidArgument, "Service is required")
		applog.LogError(rh.logger, err, logrus.Fields{
			"remediation": remediation,
		})
		return nil, err
	}

	// Get current user for audit fields
	var err error
	remediation.CreatedBy, err = common.GetCurrentUserId(ctx, rh.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Remediation", "", err)
		applog.LogError(rh.logger, wrappedErr, logrus.Fields{
			"remediation": remediation,
		})
		return nil, wrappedErr
	}
	remediation.UpdatedBy = remediation.CreatedBy

	newRemediation, err := rh.database.CreateRemediation(remediation)
	if err != nil {
		// Generic database error
		wrappedErr := appErrors.InternalError(string(op), "Remediation", "", err)
		applog.LogError(rh.logger, wrappedErr, logrus.Fields{
			"remediation": remediation,
		})
		return nil, wrappedErr
	}

	rh.eventRegistry.PushEvent(&CreateRemediationEvent{
		Remediation: newRemediation,
	})

	return newRemediation, nil
}

func (rh *remediationHandler) UpdateRemediation(ctx context.Context, remediation *entity.Remediation) (*entity.Remediation, error) {
	op := appErrors.Op("remediationHandler.UpdateRemediation")

	// Input validation
	if remediation == nil {
		err := appErrors.E(op, "Remediation", appErrors.InvalidArgument, "remediation cannot be nil")
		applog.LogError(rh.logger, err, logrus.Fields{})
		return nil, err
	}

	if remediation.Id <= 0 {
		err := appErrors.E(op, "Remediation", appErrors.InvalidArgument, fmt.Sprintf("invalid ID: %d", remediation.Id))
		applog.LogError(rh.logger, err, logrus.Fields{"id": remediation.Id})
		return nil, err
	}

	// Get current user for audit fields
	var err error
	remediation.UpdatedBy, err = common.GetCurrentUserId(ctx, rh.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Remediation", strconv.FormatInt(remediation.Id, 10), err)
		applog.LogError(rh.logger, wrappedErr, logrus.Fields{
			"remediation": remediation,
		})
		return nil, wrappedErr
	}

	// Update the component instance in database
	err = rh.database.UpdateRemediation(remediation)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Remediation", strconv.FormatInt(remediation.Id, 10), err)
		applog.LogError(rh.logger, wrappedErr, logrus.Fields{
			"remediation": remediation,
		})
		return nil, wrappedErr
	}

	// Retrieve updated component instance to return fresh data
	lo := entity.NewListOptions()
	remediationResult, err := rh.ListRemediations(&entity.RemediationFilter{Id: []*int64{&remediation.Id}}, lo)
	if err != nil {
		wrappedErr := appErrors.E(op, "Remediation", strconv.FormatInt(remediation.Id, 10), appErrors.Internal, err)
		applog.LogError(rh.logger, wrappedErr, logrus.Fields{
			"remediation": remediation,
		})
		return nil, wrappedErr
	}

	if len(remediationResult.Elements) != 1 {
		err := appErrors.E(op, "Remediation", strconv.FormatInt(remediation.Id, 10), appErrors.Internal,
			fmt.Sprintf("unexpected number of remediations found after update: expected 1, got %d", len(remediationResult.Elements)))
		applog.LogError(rh.logger, err, logrus.Fields{
			"id":          remediation.Id,
			"found_count": len(remediationResult.Elements),
		})
		return nil, err
	}

	updatedRemediation := remediationResult.Elements[0].Remediation

	rh.eventRegistry.PushEvent(&UpdateRemediationEvent{
		Remediation: updatedRemediation,
	})

	return updatedRemediation, nil
}

func (rh *remediationHandler) DeleteRemediation(ctx context.Context, id int64) error {
	op := appErrors.Op("remediationHandler.DeleteRemediation")

	// Input validation
	if id <= 0 {
		err := appErrors.E(op, "Remediation", appErrors.InvalidArgument, fmt.Sprintf("invalid ID: %d", id))
		applog.LogError(rh.logger, err, logrus.Fields{"id": id})
		return err
	}

	// Get current user for audit fields
	userId, err := common.GetCurrentUserId(ctx, rh.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Remediation", strconv.FormatInt(id, 10), err)
		applog.LogError(rh.logger, wrappedErr, logrus.Fields{
			"id": id,
		})
		return wrappedErr
	}

	err = rh.database.DeleteRemediation(id, userId)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Remediation", strconv.FormatInt(id, 10), err)
		applog.LogError(rh.logger, wrappedErr, logrus.Fields{
			"id":      id,
			"user_id": userId,
		})
		return wrappedErr
	}

	rh.eventRegistry.PushEvent(&DeleteRemediationEvent{
		RemediationID: id,
	})

	return nil
}
