// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package support_group

import (
	"fmt"
	"github.wdf.sap.corp/cc/heureka/internal/app/common"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/database"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

type supportGroupService struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewSupportGroupService(database database.Database, eventRegistry event.EventRegistry) SupportGroupService {
	return &supportGroupService{
		database:      database,
		eventRegistry: eventRegistry,
	}
}

type SupportGroupServiceError struct {
	message string
}

func NewSupportGroupServiceError(message string) *SupportGroupServiceError {
	return &SupportGroupServiceError{message: message}
}

func (e *SupportGroupServiceError) Error() string {
	return e.message
}

func (sg *supportGroupService) getSupportGroupResults(filter *entity.SupportGroupFilter) ([]entity.SupportGroupResult, error) {
	var supportGroupResults []entity.SupportGroupResult
	supportGroups, err := sg.database.GetSupportGroups(filter)
	if err != nil {
		return nil, err
	}
	for _, sg := range supportGroups {
		supportGroup := sg
		cursor := fmt.Sprintf("%d", supportGroup.Id)
		supportGroupResults = append(supportGroupResults, entity.SupportGroupResult{
			WithCursor:               entity.WithCursor{Value: cursor},
			SupportGroupAggregations: nil,
			SupportGroup:             &supportGroup,
		})
	}
	return supportGroupResults, nil
}

func (sg *supportGroupService) GetSupportGroup(supportGroupId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": GetSupportGroupEventName,
		"id":    supportGroupId,
	})
	supportGroupFilter := entity.SupportGroupFilter{Id: []*int64{&supportGroupId}}
	supportGroups, err := sg.ListSupportGroups(&supportGroupFilter, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupServiceError("Internal error while retrieving supportGroup.")
	}

	if len(supportGroups.Elements) != 1 {
		return nil, NewSupportGroupServiceError(fmt.Sprintf("SupportGroup %d not found.", supportGroupId))
	}

	sg.eventRegistry.PushEvent(&GetSupportGroupEvent{
		SupportGroupID: supportGroupId,
		SupportGroup:   supportGroups.Elements[0].SupportGroup,
	})

	return supportGroups.Elements[0].SupportGroup, nil
}

func (sg *supportGroupService) ListSupportGroups(filter *entity.SupportGroupFilter, options *entity.ListOptions) (*entity.List[entity.SupportGroupResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListSupportGroupsEventName,
		"filter": filter,
	})

	res, err := sg.getSupportGroupResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupServiceError("Error while filtering for SupportGroups")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := sg.database.GetAllSupportGroupIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewSupportGroupServiceError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = sg.database.CountSupportGroups(filter)
		if err != nil {
			l.Error(err)
			return nil, NewSupportGroupServiceError("Error while total count of SupportGroups")
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

func (sg *supportGroupService) CreateSupportGroup(supportGroup *entity.SupportGroup) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateSupportGroupEventName,
		"object": supportGroup,
	})

	f := &entity.SupportGroupFilter{
		Name: []*string{&supportGroup.Name},
	}

	supportGroups, err := sg.ListSupportGroups(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupServiceError("Internal error while creating supportGroup.")
	}

	if len(supportGroups.Elements) > 0 {
		return nil, NewSupportGroupServiceError(fmt.Sprintf("Duplicated entry %s for name.", supportGroup.Name))
	}

	newSupportGroup, err := sg.database.CreateSupportGroup(supportGroup)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupServiceError("Internal error while creating supportGroup.")
	}

	sg.eventRegistry.PushEvent(&CreateSupportGroupEvent{
		SupportGroup: newSupportGroup,
	})

	return newSupportGroup, nil
}

func (sg *supportGroupService) UpdateSupportGroup(supportGroup *entity.SupportGroup) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateSupportGroupEventName,
		"object": supportGroup,
	})

	err := sg.database.UpdateSupportGroup(supportGroup)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupServiceError("Internal error while updating supportGroup.")
	}

	sg.eventRegistry.PushEvent(&UpdateSupportGroupEvent{SupportGroup: supportGroup})

	return sg.GetSupportGroup(supportGroup.Id)
}

func (sg *supportGroupService) DeleteSupportGroup(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteSupportGroupEventName,
		"id":    id,
	})

	err := sg.database.DeleteSupportGroup(id)

	if err != nil {
		l.Error(err)
		return NewSupportGroupServiceError("Internal error while deleting supportGroup.")
	}

	sg.eventRegistry.PushEvent(&DeleteSupportGroupEvent{SupportGroupID: id})

	return nil
}

func (sg *supportGroupService) AddServiceToSupportGroup(supportGroupId int64, serviceId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          AddServiceToSupportGroupEventName,
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.AddServiceToSupportGroup(supportGroupId, serviceId)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupServiceError("Internal error while adding service to supportGroup.")
	}

	sg.eventRegistry.PushEvent(&AddServiceToSupportGroupEvent{SupportGroupID: supportGroupId, ServiceID: serviceId})

	return sg.GetSupportGroup(supportGroupId)
}

func (sg *supportGroupService) RemoveServiceFromSupportGroup(supportGroupId int64, serviceId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          RemoveServiceFromSupportGroupEventName,
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.RemoveServiceFromSupportGroup(supportGroupId, serviceId)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupServiceError("Internal error while removing service from supportGroup.")
	}

	sg.eventRegistry.PushEvent(&RemoveServiceFromSupportGroupEvent{SupportGroupID: supportGroupId, ServiceID: serviceId})

	return sg.GetSupportGroup(supportGroupId)
}

func (sg *supportGroupService) AddUserToSupportGroup(supportGroupId int64, userId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          AddUserToSupportGroupEventName,
		"userId":         userId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.AddUserToSupportGroup(supportGroupId, userId)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupServiceError("Internal error while adding user to supportGroup.")
	}

	sg.eventRegistry.PushEvent(&AddUserToSupportGroupEvent{SupportGroupID: supportGroupId, UserID: userId})

	return sg.GetSupportGroup(supportGroupId)
}

func (sg *supportGroupService) RemoveUserFromSupportGroup(supportGroupId int64, userId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          RemoveUserFromSupportGroupEventName,
		"userId":         userId,
		"supportGroupId": supportGroupId,
	})

	err := sg.database.RemoveUserFromSupportGroup(supportGroupId, userId)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupServiceError("Internal error while removing user from supportGroup.")
	}

	sg.eventRegistry.PushEvent(&RemoveUserFromSupportGroupEvent{SupportGroupID: supportGroupId, UserID: userId})

	return sg.GetSupportGroup(supportGroupId)
}

func (sg *supportGroupService) ListSupportGroupNames(filter *entity.SupportGroupFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListSupportGroupNamesEventName,
		"filter": filter,
	})

	supportGroupNames, err := sg.database.GetSupportGroupNames(filter)

	if err != nil {
		l.Error(err)
		return nil, NewSupportGroupServiceError("Internal error while retrieving supportGroupNames.")
	}

	sg.eventRegistry.PushEvent(&ListSupportGroupNamesEvent{
		Filter:  filter,
		Options: options,
		Names:   supportGroupNames,
	})

	return supportGroupNames, nil
}
