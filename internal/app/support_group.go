// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getSupportGroupResults(filter *entity.SupportGroupFilter) ([]entity.SupportGroupResult, error) {
	var supportGroupResults []entity.SupportGroupResult
	supportGroups, err := h.database.GetSupportGroups(filter)
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

func (h *HeurekaApp) GetSupportGroup(supportGroupId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.GetSupportGroup",
		"id":    supportGroupId,
	})
	supportGroupFilter := entity.SupportGroupFilter{Id: []*int64{&supportGroupId}}
	supportGroups, err := h.ListSupportGroups(&supportGroupFilter, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving supportGroup.")
	}

	if len(supportGroups.Elements) != 1 {
		return nil, heurekaError(fmt.Sprintf("SupportGroup %d not found.", supportGroupId))
	}

	return supportGroups.Elements[0].SupportGroup, nil
}

func (h *HeurekaApp) ListSupportGroups(filter *entity.SupportGroupFilter, options *entity.ListOptions) (*entity.List[entity.SupportGroupResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListSupportGroups",
		"filter": filter,
	})

	res, err := h.getSupportGroupResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for SupportGroups")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllSupportGroupIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountSupportGroups(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of SupportGroups")
		}
	}

	return &entity.List[entity.SupportGroupResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) CreateSupportGroup(supportGroup *entity.SupportGroup) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateSupportGroup",
		"object": supportGroup,
	})

	f := &entity.SupportGroupFilter{
		Name: []*string{&supportGroup.Name},
	}

	supportGroups, err := h.ListSupportGroups(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating supportGroup.")
	}

	if len(supportGroups.Elements) > 0 {
		return nil, heurekaError(fmt.Sprintf("Duplicated entry %s for name.", supportGroup.Name))
	}

	newSupportGroup, err := h.database.CreateSupportGroup(supportGroup)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating supportGroup.")
	}

	return newSupportGroup, nil
}

func (h *HeurekaApp) UpdateSupportGroup(supportGroup *entity.SupportGroup) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateSupportGroup",
		"object": supportGroup,
	})

	err := h.database.UpdateSupportGroup(supportGroup)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating supportGroup.")
	}
	return h.GetSupportGroup(supportGroup.Id)
}

func (h *HeurekaApp) DeleteSupportGroup(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteSupportGroup",
		"id":    id,
	})

	err := h.database.DeleteSupportGroup(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while deleting supportGroup.")
	}

	return nil
}

func (h *HeurekaApp) AddServiceToSupportGroup(supportGroupId int64, serviceId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          "app.AddServiceToSupportGroup",
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
	})

	err := h.database.AddServiceToSupportGroup(supportGroupId, serviceId)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while adding service to supportGroup.")
	}

	return h.GetSupportGroup(supportGroupId)
}

func (h *HeurekaApp) RemoveServiceFromSupportGroup(supportGroupId int64, serviceId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          "app.RemoveServiceFromSupportGroup",
		"serviceId":      serviceId,
		"supportGroupId": supportGroupId,
	})

	err := h.database.RemoveServiceFromSupportGroup(supportGroupId, serviceId)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while removing service from supportGroup.")
	}

	return h.GetSupportGroup(supportGroupId)
}

func (h *HeurekaApp) AddUserToSupportGroup(supportGroupId int64, userId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          "app.AddUserToSupportGroup",
		"userId":         userId,
		"supportGroupId": supportGroupId,
	})

	err := h.database.AddUserToSupportGroup(supportGroupId, userId)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while adding user to supportGroup.")
	}

	return h.GetSupportGroup(supportGroupId)
}

func (h *HeurekaApp) RemoveUserFromSupportGroup(supportGroupId int64, userId int64) (*entity.SupportGroup, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":          "app.RemoveUserFromSupportGroup",
		"userId":         userId,
		"supportGroupId": supportGroupId,
	})

	err := h.database.RemoveUserFromSupportGroup(supportGroupId, userId)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while removing user from supportGroup.")
	}

	return h.GetSupportGroup(supportGroupId)
}

func (h *HeurekaApp) ListSupportGroupNames(filter *entity.SupportGroupFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListSupportGroupNames",
		"filter": filter,
	})

	supportGroupNames, err := h.database.GetSupportGroupNames(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving supportGroupNames.")
	}

	return supportGroupNames, nil
}
