// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/common"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

type supportGroupHandler struct {
	common.BaseHandler[entity.SupportGroupResult, *entity.SupportGroupFilter]
}

func NewSupportGroupHandler(handlerContext common.HandlerContext) SupportGroupHandler {
	return &supportGroupHandler{
		BaseHandler: common.NewBaseHandler(handlerContext, common.BaseConfig[entity.SupportGroupResult, *entity.SupportGroupFilter]{
			Op:              appErrors.Op("supportGroupHandler"),
			Entity:          "SupportGroups",
			CursorEntity:    "SupportGroupCursors",
			CountEntity:     "SupportGroupCount",
			GetFn:           handlerContext.DB.GetSupportGroups,
			CursorsFn:       handlerContext.DB.GetAllSupportGroupCursors,
			CountFn:         handlerContext.DB.CountSupportGroups,
			Authz:           handlerContext.Authz,
			AuthzObjectType: openfga.TypeSupportGroup,
			AuthzApplyFn: func(f *entity.SupportGroupFilter, ids []*int64) {
				f.Id = common.CombineFilterWithAccessibleIds(f.Id, ids)
			},
			ListEventFn: func(f *entity.SupportGroupFilter, o *entity.ListOptions, r *entity.List[entity.SupportGroupResult]) any {
				return &ListSupportGroupsEvent{Filter: f, Options: o, SupportGroups: r}
			},
			DeleteFn:             handlerContext.DB.DeleteSupportGroup,
			DeleteEventFn:        func(id int64) any { return &DeleteSupportGroupEvent{SupportGroupID: id} },
			AuthzWriteObjectType: openfga.TypeSupportGroup,
		}),
	}
}

func (sg *supportGroupHandler) GetSupportGroup(
	ctx context.Context,
	supportGroupId int64,
) (*entity.SupportGroup, error) {
	op := appErrors.CallerOp()

	if err := sg.CheckPermission(ctx, op, "SupportGroup", openfga.RelCanView, openfga.TypeSupportGroup, supportGroupId); err != nil {
		return nil, err
	}

	result, err := sg.ListSupportGroups(
		ctx,
		&entity.SupportGroupFilter{Id: []*int64{&supportGroupId}},
		entity.NewListOptions(),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "SupportGroup", fmt.Sprint(supportGroupId), err)
	}

	if len(result.Elements) != 1 {
		return nil, appErrors.NotFoundError(string(op), "SupportGroup", fmt.Sprint(supportGroupId))
	}

	sg.PushEvent(&GetSupportGroupEvent{
		SupportGroupID: supportGroupId,
		SupportGroup:   result.Elements[0].SupportGroup,
	})

	return result.Elements[0].SupportGroup, nil
}

func (sg *supportGroupHandler) ListSupportGroups(
	ctx context.Context,
	filter *entity.SupportGroupFilter,
	options *entity.ListOptions,
) (*entity.List[entity.SupportGroupResult], error) {
	return sg.List(ctx, appErrors.CallerOp(), filter, options)
}

func (sg *supportGroupHandler) CreateSupportGroup(
	ctx context.Context,
	supportGroup *entity.SupportGroup,
) (*entity.SupportGroup, error) {
	op := appErrors.CallerOp()

	if err := sg.CheckSystemAdminPermission(ctx, op); err != nil {
		return nil, err
	}

	var err error

	supportGroup.CreatedBy, err = common.GetCurrentUserId(ctx, sg.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "SupportGroup", "", err)
	}

	supportGroup.UpdatedBy = supportGroup.CreatedBy

	newSupportGroup, err := sg.DB().CreateSupportGroup(supportGroup)
	if err != nil {
		duplicateEntryError := &database.DuplicateEntryDatabaseError{}
		if errors.As(err, &duplicateEntryError) {
			wrappedErr := appErrors.AlreadyExistsError(string(op), "SupportGroup", supportGroup.CCRN)
			applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{
				"ccrn":                    supportGroup.CCRN,
				"duplicate_entry_details": duplicateEntryError.Error(),
			})

			return nil, wrappedErr
		}

		wrappedErr := appErrors.InternalError(string(op), "SupportGroup", "", err)
		applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{"ccrn": supportGroup.CCRN})

		return nil, wrappedErr
	}

	sg.PushEvent(&CreateSupportGroupEvent{SupportGroup: newSupportGroup})

	return newSupportGroup, nil
}

func (sg *supportGroupHandler) UpdateSupportGroup(
	ctx context.Context,
	supportGroup *entity.SupportGroup,
) (*entity.SupportGroup, error) {
	op := appErrors.CallerOp()

	if err := sg.CheckWritePermission(ctx, op, "SupportGroup", openfga.TypeSupportGroup, supportGroup.Id); err != nil {
		return nil, err
	}

	var err error

	supportGroup.UpdatedBy, err = common.GetCurrentUserId(ctx, sg.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "SupportGroup", strconv.FormatInt(supportGroup.Id, 10), err)
	}

	if err = sg.DB().UpdateSupportGroup(supportGroup); err != nil {
		return nil, appErrors.InternalError(string(op), "SupportGroup", strconv.FormatInt(supportGroup.Id, 10), err)
	}

	sg.PushEvent(&UpdateSupportGroupEvent{SupportGroup: supportGroup})

	return sg.GetSupportGroup(ctx, supportGroup.Id)
}

func (sg *supportGroupHandler) DeleteSupportGroup(ctx context.Context, id int64) error {
	return sg.Delete(ctx, id)
}

func (sg *supportGroupHandler) AddServiceToSupportGroup(
	ctx context.Context,
	supportGroupId int64,
	serviceId int64,
) (*entity.SupportGroup, error) {
	op := appErrors.CallerOp()

	if err := sg.CheckWritePermission(ctx, op, "SupportGroup", openfga.TypeSupportGroup, supportGroupId); err != nil {
		return nil, err
	}

	if err := sg.DB().AddServiceToSupportGroup(supportGroupId, serviceId); err != nil {
		return nil, appErrors.InternalError(string(op), "SupportGroup", fmt.Sprint(supportGroupId), err)
	}

	sg.PushEvent(&AddServiceToSupportGroupEvent{SupportGroupID: supportGroupId, ServiceID: serviceId})

	return sg.GetSupportGroup(ctx, supportGroupId)
}

func (sg *supportGroupHandler) RemoveServiceFromSupportGroup(
	ctx context.Context,
	supportGroupId int64,
	serviceId int64,
) (*entity.SupportGroup, error) {
	op := appErrors.CallerOp()

	if err := sg.CheckWritePermission(ctx, op, "SupportGroup", openfga.TypeSupportGroup, supportGroupId); err != nil {
		return nil, err
	}

	if err := sg.DB().RemoveServiceFromSupportGroup(supportGroupId, serviceId); err != nil {
		return nil, appErrors.InternalError(string(op), "SupportGroup", fmt.Sprint(supportGroupId), err)
	}

	sg.PushEvent(&RemoveServiceFromSupportGroupEvent{SupportGroupID: supportGroupId, ServiceID: serviceId})

	return sg.GetSupportGroup(ctx, supportGroupId)
}

func (sg *supportGroupHandler) AddUserToSupportGroup(
	ctx context.Context,
	supportGroupId int64,
	userId int64,
) (*entity.SupportGroup, error) {
	op := appErrors.CallerOp()

	if err := sg.CheckWritePermission(ctx, op, "SupportGroup", openfga.TypeSupportGroup, supportGroupId); err != nil {
		return nil, err
	}

	if err := sg.DB().AddUserToSupportGroup(supportGroupId, userId); err != nil {
		return nil, appErrors.InternalError(string(op), "SupportGroup", fmt.Sprint(supportGroupId), err)
	}

	sg.PushEvent(&AddUserToSupportGroupEvent{SupportGroupID: supportGroupId, UserID: userId})

	return sg.GetSupportGroup(ctx, supportGroupId)
}

func (sg *supportGroupHandler) RemoveUserFromSupportGroup(
	ctx context.Context,
	supportGroupId int64,
	userId int64,
) (*entity.SupportGroup, error) {
	op := appErrors.CallerOp()

	if err := sg.CheckWritePermission(ctx, op, "SupportGroup", openfga.TypeSupportGroup, supportGroupId); err != nil {
		return nil, err
	}

	if err := sg.DB().RemoveUserFromSupportGroup(supportGroupId, userId); err != nil {
		return nil, appErrors.InternalError(string(op), "SupportGroup", fmt.Sprint(supportGroupId), err)
	}

	sg.PushEvent(&RemoveUserFromSupportGroupEvent{SupportGroupID: supportGroupId, UserID: userId})

	return sg.GetSupportGroup(ctx, supportGroupId)
}

func (sg *supportGroupHandler) ListSupportGroupCcrns(
	ctx context.Context,
	filter *entity.SupportGroupFilter,
	options *entity.ListOptions,
) ([]string, error) {
	op := appErrors.CallerOp()

	ccrns, err := sg.DB().GetSupportGroupCcrns(ctx, filter)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "SupportGroupCcrns", "", err)
	}

	sg.PushEvent(&ListSupportGroupCcrnsEvent{Filter: filter, Options: options, Ccrns: ccrns})

	return ccrns, nil
}
