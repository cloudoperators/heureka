// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/openfga"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

type supportGroupHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewSupportGroupHandler(database database.Database, eventRegistry event.EventRegistry, authz openfga.Authorization) SupportGroupHandler {
	return &supportGroupHandler{
		database:      database,
		eventRegistry: eventRegistry,
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

func (sg *supportGroupHandler) GetSupportGroup(supportGroupId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": GetSupportGroupEventName,
		"id":    supportGroupId,
	})
	lo := entity.NewListOptions()
	supportGroupFilter := entity.SupportGroupFilter{Id: []*int64{&supportGroupId}}
	supportGroups, err := sg.ListSupportGroups(&supportGroupFilter, lo)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while retrieving supportGroup.")
	}

	if len(supportGroups.Elements) != 1 {
		return nil, NewSupportGroupHandlerError(fmt.Sprintf("SupportGroup %d not found.", supportGroupId))
	}

	sg.eventRegistry.PushEvent(&GetSupportGroupEvent{
		SupportGroupID: supportGroupId,
		SupportGroup:   supportGroups.Elements[0].SupportGroup,
	})

	return supportGroups.Elements[0].SupportGroup, nil
}

func (sg *supportGroupHandler) ListSupportGroups(filter *entity.SupportGroupFilter, options *entity.ListOptions) (*entity.List[entity.SupportGroupResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginatedX(&filter.PaginatedX)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListSupportGroupsEventName,
		"filter": filter,
	})

	res, err := sg.database.GetSupportGroups(filter, options.Order)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Error while filtering for SupportGroups")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := sg.database.GetAllSupportGroupCursors(filter, options.Order)
			if err != nil {
				l.Error(err)
				return nil, NewSupportGroupHandlerError("Error while getting all cursors")
			}
			pageInfo = common.GetPageInfoX(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = sg.database.CountSupportGroups(filter)
		if err != nil {
			l.Error(err)
			return nil, NewSupportGroupHandlerError("Error while total count of SupportGroups")
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

func (sg *supportGroupHandler) CreateSupportGroup(supportGroup *entity.SupportGroup) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateSupportGroupEventName,
		"object": supportGroup,
	})

	f := &entity.SupportGroupFilter{
		CCRN: []*string{&supportGroup.CCRN},
	}

	var err error
	supportGroup.CreatedBy, err = common.GetCurrentUserId(sg.database)
	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while creating supportGroup (GetUserId).")
	}
	supportGroup.UpdatedBy = supportGroup.CreatedBy

	lo := entity.NewListOptions()
	supportGroups, err := sg.ListSupportGroups(f, lo)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while creating supportGroup.")
	}

	if len(supportGroups.Elements) > 0 {
		return nil, NewSupportGroupHandlerError(fmt.Sprintf("Duplicated entry %s for ccrn.", supportGroup.CCRN))
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

func (sg *supportGroupHandler) UpdateSupportGroup(supportGroup *entity.SupportGroup) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateSupportGroupEventName,
		"object": supportGroup,
	})

	var err error
	supportGroup.UpdatedBy, err = common.GetCurrentUserId(sg.database)
	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while updating supportGroup (GetUserId).")
	}

	err = sg.database.UpdateSupportGroup(supportGroup)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while updating supportGroup.")
	}

	sg.eventRegistry.PushEvent(&UpdateSupportGroupEvent{SupportGroup: supportGroup})

	return sg.GetSupportGroup(supportGroup.Id)
}

func (sg *supportGroupHandler) DeleteSupportGroup(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteSupportGroupEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(sg.database)
	if err != nil {
		l.Error(err)
		return NewSupportGroupHandlerError("Internal error while deleting supportGroup (GetUserId).")
	}

	err = sg.database.DeleteSupportGroup(id, userId)

	if err != nil {
		l.Error(err)
		return NewSupportGroupHandlerError("Internal error while deleting supportGroup.")
	}

	sg.eventRegistry.PushEvent(&DeleteSupportGroupEvent{SupportGroupID: id})

	return nil
}

func (sg *supportGroupHandler) AddServiceToSupportGroup(supportGroupId int64, serviceId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          AddServiceToSupportGroupEventName,
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.AddServiceToSupportGroup(supportGroupId, serviceId)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while adding service to supportGroup.")
	}

	sg.eventRegistry.PushEvent(&AddServiceToSupportGroupEvent{SupportGroupID: supportGroupId, ServiceID: serviceId})

	return sg.GetSupportGroup(supportGroupId)
}

func (sg *supportGroupHandler) RemoveServiceFromSupportGroup(supportGroupId int64, serviceId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          RemoveServiceFromSupportGroupEventName,
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.RemoveServiceFromSupportGroup(supportGroupId, serviceId)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while removing service from supportGroup.")
	}

	sg.eventRegistry.PushEvent(&RemoveServiceFromSupportGroupEvent{SupportGroupID: supportGroupId, ServiceID: serviceId})

	return sg.GetSupportGroup(supportGroupId)
}

func (sg *supportGroupHandler) AddUserToSupportGroup(supportGroupId int64, userId int64) (*entity.SupportGroup, error) {
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

	sg.eventRegistry.PushEvent(&AddUserToSupportGroupEvent{SupportGroupID: supportGroupId, UserID: userId})

	return sg.GetSupportGroup(supportGroupId)
}

func (sg *supportGroupHandler) RemoveUserFromSupportGroup(supportGroupId int64, userId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          RemoveUserFromSupportGroupEventName,
		"userId":         userId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.RemoveUserFromSupportGroup(supportGroupId, userId)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while removing user from supportGroup.")
	}

	sg.eventRegistry.PushEvent(&RemoveUserFromSupportGroupEvent{SupportGroupID: supportGroupId, UserID: userId})

	return sg.GetSupportGroup(supportGroupId)
}

func (sg *supportGroupHandler) ListSupportGroupCcrns(filter *entity.SupportGroupFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListSupportGroupCcrnsEventName,
		"filter": filter,
	})

	supportGroupCcrns, err := sg.database.GetSupportGroupCcrns(filter)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupHandlerError("Internal error while retrieving supportGroupCcrns.")
	}

	sg.eventRegistry.PushEvent(&ListSupportGroupCcrnsEvent{
		Filter:  filter,
		Options: options,
		Ccrns:   supportGroupCcrns,
	})

	return supportGroupCcrns, nil
}
