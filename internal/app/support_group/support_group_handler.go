// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group

import (
	"context"
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

type supportGroupHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	authz         openfga.Authorization
	logger        *logrus.Logger
}

func NewSupportGroupHandler(handlerContext common.HandlerContext) SupportGroupHandler {
	return &supportGroupHandler{
		database:      handlerContext.DB,
		eventRegistry: handlerContext.EventReg,
		authz:         handlerContext.Authz,
		logger:        logrus.New(),
	}
}

type SupportGroupHandlerError struct {
	message string
}

func NewSupportGroupHandlerError(message string) *SupportGroupHandlerError {
	return &SupportGroupHandlerError{message: message}
}

func (e *SupportGroupHandlerError) Error() string {
	return e.message
}

func (sg *supportGroupHandler) GetSupportGroup(
	ctx context.Context,
	supportGroupId int64,
) (*entity.SupportGroup, error) {
	op := appErrors.Op("supportGroupHandler.GetSupportGroup")

	// get current user id
	currentUserId, err := common.GetCurrentUserId(ctx, sg.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(
			string(op),
			"SupportGroups",
			fmt.Sprint(supportGroupId),
			err,
		)
		applog.LogError(sg.logger, wrappedErr, logrus.Fields{
			"supportGroupId": supportGroupId,
		})

		return nil, wrappedErr
	}

	// Authorization check
	hasPermission, err := sg.authz.CheckPermission(openfga.RelationInput{
		UserType:   openfga.TypeUser,
		UserId:     openfga.UserId(fmt.Sprint(currentUserId)),
		Relation:   openfga.RelCanView,
		ObjectType: openfga.TypeSupportGroup,
		ObjectId:   openfga.ObjectId(fmt.Sprint(supportGroupId)),
	})
	if err != nil {
		wrappedErr := appErrors.InternalError(
			string(op),
			"SupportGroups",
			fmt.Sprint(supportGroupId),
			err,
		)
		applog.LogError(sg.logger, wrappedErr, logrus.Fields{
			"supportGroupId": supportGroupId,
		})

		return nil, wrappedErr
	}

	if !hasPermission {
		wrappedErr := appErrors.PermissionDeniedError(
			string(op),
			"SupportGroups",
			fmt.Sprint(supportGroupId),
		)
		applog.LogError(sg.logger, wrappedErr, logrus.Fields{
			"supportGroupId": supportGroupId,
			"userId":         currentUserId,
		})

		return nil, wrappedErr
	}

	lo := entity.NewListOptions()
	supportGroupFilter := entity.SupportGroupFilter{Id: []*int64{&supportGroupId}}

	supportGroups, err := sg.ListSupportGroups(ctx, &supportGroupFilter, lo)
	if err != nil {
		wrappedErr := appErrors.InternalError(
			string(op),
			"SupportGroups",
			fmt.Sprint(supportGroupId),
			err,
		)
		applog.LogError(sg.logger, wrappedErr, logrus.Fields{
			"supportGroupId": supportGroupId,
		})

		return nil, wrappedErr
	}

	if len(supportGroups.Elements) != 1 {
		wrappedErr := appErrors.InternalError(
			string(op),
			"SupportGroups",
			fmt.Sprint(supportGroupId),
			err,
		)
		applog.LogError(sg.logger, wrappedErr, logrus.Fields{
			"supportGroupId": supportGroupId,
		})

		return nil, wrappedErr
	}

	sg.eventRegistry.PushEvent(&GetSupportGroupEvent{
		SupportGroupID: supportGroupId,
		SupportGroup:   supportGroups.Elements[0].SupportGroup,
	})

	return supportGroups.Elements[0].SupportGroup, nil
}

func (sg *supportGroupHandler) ListSupportGroups(
	ctx context.Context,
	filter *entity.SupportGroupFilter,
	options *entity.ListOptions,
) (*entity.List[entity.SupportGroupResult], error) {
	var (
		count    int64
		pageInfo *entity.PageInfo
	)

	op := appErrors.Op("supportGroupHandler.ListSupportGroups")

	common.EnsurePaginated(&filter.Paginated)

	// get current user id
	currentUserId, err := common.GetCurrentUserId(ctx, sg.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "SupportGroups", "", err)
		applog.LogError(sg.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})

		return nil, wrappedErr
	}

	// Authorization check
	accessibleSupportGroupIds, err := sg.authz.GetListOfAccessibleObjectIds(
		openfga.UserId(fmt.Sprint(currentUserId)),
		openfga.TypeSupportGroup,
	)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "SupportGroups", "", err)
		applog.LogError(sg.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})

		return nil, wrappedErr
	}

	// Update the filter.Id based on accessibleSupportGroupIds
	filter.Id = common.CombineFilterWithAccessibleIds(filter.Id, accessibleSupportGroupIds)

	res, err := sg.database.GetSupportGroups(ctx, filter, options.Order)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "SupportGroups", "", err)
		applog.LogError(sg.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})

		return nil, wrappedErr
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := sg.database.GetAllSupportGroupCursors(ctx, filter, options.Order)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), "SupportGroups", "", err)
				applog.LogError(sg.logger, wrappedErr, logrus.Fields{
					"filter": filter,
				})

				return nil, wrappedErr
			}

			pageInfo = common.GetPageInfo(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = sg.database.CountSupportGroups(ctx, filter)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "SupportGroups", "", err)
			applog.LogError(sg.logger, wrappedErr, logrus.Fields{
				"filter": filter,
			})

			return nil, wrappedErr
		}
	}

	ret := &entity.List[entity.SupportGroupResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	sg.eventRegistry.PushEvent(&ListSupportGroupsEvent{
		Filter:        filter,
		Options:       options,
		SupportGroups: ret,
	})

	return ret, nil
}

func (sg *supportGroupHandler) CreateSupportGroup(
	ctx context.Context,
	supportGroup *entity.SupportGroup,
) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateSupportGroupEventName,
		"object": supportGroup,
	})

	f := &entity.SupportGroupFilter{
		CCRN: []*string{&supportGroup.CCRN},
	}

	var err error

	supportGroup.CreatedBy, err = common.GetCurrentUserId(ctx, sg.database)
	if err != nil {
		l.Error(err)

		return nil, NewSupportGroupHandlerError(
			"Internal error while creating supportGroup (GetUserId).",
		)
	}

	supportGroup.UpdatedBy = supportGroup.CreatedBy

	lo := entity.NewListOptions()

	supportGroups, err := sg.ListSupportGroups(ctx, f, lo)
	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while creating supportGroup.")
	}

	if len(supportGroups.Elements) > 0 {
		return nil, NewSupportGroupHandlerError(
			fmt.Sprintf("Duplicated entry %s for ccrn.", supportGroup.CCRN),
		)
	}

	newSupportGroup, err := sg.database.CreateSupportGroup(supportGroup)
	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while creating supportGroup.")
	}

	sg.eventRegistry.PushEvent(&CreateSupportGroupEvent{
		SupportGroup: newSupportGroup,
	})

	return newSupportGroup, nil
}

func (sg *supportGroupHandler) UpdateSupportGroup(
	ctx context.Context,
	supportGroup *entity.SupportGroup,
) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateSupportGroupEventName,
		"object": supportGroup,
	})

	var err error

	supportGroup.UpdatedBy, err = common.GetCurrentUserId(ctx, sg.database)
	if err != nil {
		l.Error(err)

		return nil, NewSupportGroupHandlerError(
			"Internal error while updating supportGroup (GetUserId).",
		)
	}

	err = sg.database.UpdateSupportGroup(supportGroup)
	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while updating supportGroup.")
	}

	sg.eventRegistry.PushEvent(&UpdateSupportGroupEvent{SupportGroup: supportGroup})

	return sg.GetSupportGroup(ctx, supportGroup.Id)
}

func (sg *supportGroupHandler) DeleteSupportGroup(ctx context.Context, id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteSupportGroupEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(ctx, sg.database)
	if err != nil {
		l.Error(err)

		return NewSupportGroupHandlerError(
			"Internal error while deleting supportGroup (GetUserId).",
		)
	}

	err = sg.database.DeleteSupportGroup(id, userId)
	if err != nil {
		l.Error(err)
		return NewSupportGroupHandlerError("Internal error while deleting supportGroup.")
	}

	sg.eventRegistry.PushEvent(&DeleteSupportGroupEvent{SupportGroupID: id})

	return nil
}

func (sg *supportGroupHandler) AddServiceToSupportGroup(
	ctx context.Context,
	supportGroupId int64,
	serviceId int64,
) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          AddServiceToSupportGroupEventName,
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.AddServiceToSupportGroup(supportGroupId, serviceId)
	if err != nil {
		l.Error(err)

		return nil, NewSupportGroupHandlerError(
			"Internal error while adding service to supportGroup.",
		)
	}

	sg.eventRegistry.PushEvent(
		&AddServiceToSupportGroupEvent{SupportGroupID: supportGroupId, ServiceID: serviceId},
	)

	return sg.GetSupportGroup(ctx, supportGroupId)
}

func (sg *supportGroupHandler) RemoveServiceFromSupportGroup(ctx context.Context,
	supportGroupId int64, serviceId int64,
) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          RemoveServiceFromSupportGroupEventName,
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.RemoveServiceFromSupportGroup(supportGroupId, serviceId)
	if err != nil {
		l.Error(err)

		return nil, NewSupportGroupHandlerError(
			"Internal error while removing service from supportGroup.",
		)
	}

	sg.eventRegistry.PushEvent(
		&RemoveServiceFromSupportGroupEvent{SupportGroupID: supportGroupId, ServiceID: serviceId},
	)

	return sg.GetSupportGroup(ctx, supportGroupId)
}

func (sg *supportGroupHandler) AddUserToSupportGroup(
	ctx context.Context,
	supportGroupId int64,
	userId int64,
) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          AddUserToSupportGroupEventName,
		"userId":         userId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.AddUserToSupportGroup(supportGroupId, userId)
	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while adding user to supportGroup.")
	}

	sg.eventRegistry.PushEvent(
		&AddUserToSupportGroupEvent{SupportGroupID: supportGroupId, UserID: userId},
	)

	return sg.GetSupportGroup(ctx, supportGroupId)
}

func (sg *supportGroupHandler) RemoveUserFromSupportGroup(ctx context.Context,
	supportGroupId int64, userId int64,
) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          RemoveUserFromSupportGroupEventName,
		"userId":         userId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.RemoveUserFromSupportGroup(supportGroupId, userId)
	if err != nil {
		l.Error(err)

		return nil, NewSupportGroupHandlerError(
			"Internal error while removing user from supportGroup.",
		)
	}

	sg.eventRegistry.PushEvent(
		&RemoveUserFromSupportGroupEvent{SupportGroupID: supportGroupId, UserID: userId},
	)

	return sg.GetSupportGroup(ctx, supportGroupId)
}

func (sg *supportGroupHandler) ListSupportGroupCcrns(
	ctx context.Context,
	filter *entity.SupportGroupFilter,
	options *entity.ListOptions,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListSupportGroupCcrnsEventName,
		"filter": filter,
	})

	supportGroupCcrns, err := sg.database.GetSupportGroupCcrns(ctx, filter)
	if err != nil {
		l.Error(err)

		return nil, NewSupportGroupHandlerError(
			"Internal error while retrieving supportGroupCcrns.",
		)
	}

	sg.eventRegistry.PushEvent(&ListSupportGroupCcrnsEvent{
		Filter:  filter,
		Options: options,
		Ccrns:   supportGroupCcrns,
	})

	return supportGroupCcrns, nil
}
