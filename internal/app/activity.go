// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getActivityResults(filter *entity.ActivityFilter) ([]entity.ActivityResult, error) {
	var activityResults []entity.ActivityResult
	activities, err := h.database.GetActivities(filter)
	if err != nil {
		return nil, err
	}
	for _, a := range activities {
		activity := a
		cursor := fmt.Sprintf("%d", activity.Id)
		activityResults = append(activityResults, entity.ActivityResult{
			WithCursor:           entity.WithCursor{Value: cursor},
			ActivityAggregations: nil,
			Activity:             &activity,
		})
	}
	return activityResults, nil
}

func (h *HeurekaApp) GetActivity(activityId int64) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.GetActivity",
		"id":    activityId,
	})
	activityFilter := entity.ActivityFilter{Id: []*int64{&activityId}}
	activities, err := h.ListActivities(&activityFilter, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving activities.")
	}

	if len(activities.Elements) != 1 {
		return nil, heurekaError(fmt.Sprintf("Activity %d not found.", activityId))
	}

	return activities.Elements[0].Activity, nil
}

func (h *HeurekaApp) ListActivities(filter *entity.ActivityFilter, options *entity.ListOptions) (*entity.List[entity.ActivityResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListActivities",
		"filter": filter,
	})

	res, err := h.getActivityResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for Activities")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllActivityIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountActivities(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of Activities")
		}
	}

	return &entity.List[entity.ActivityResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) CreateActivity(activity *entity.Activity) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateActivity",
		"object": activity,
	})

	newActivity, err := h.database.CreateActivity(activity)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating activity.")
	}

	return newActivity, nil
}

func (h *HeurekaApp) UpdateActivity(activity *entity.Activity) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateActivity",
		"object": activity,
	})

	err := h.database.UpdateActivity(activity)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating activity.")
	}

	return h.GetActivity(activity.Id)
}

func (h *HeurekaApp) DeleteActivity(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteActivity",
		"id":    id,
	})

	err := h.database.DeleteActivity(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while deleting activity.")
	}

	return nil
}

func (h *HeurekaApp) AddServiceToActivity(activityId, serviceId int64) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":      "app.AddServiceToActivity",
		"activityId": activityId,
		"serviceId":  serviceId,
	})

	err := h.database.AddServiceToActivity(activityId, serviceId)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while adding service to activity.")
	}

	return h.GetActivity(activityId)
}

func (h *HeurekaApp) RemoveServiceFromActivity(activityId, serviceId int64) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":      "app.RemoveServiceFromActivity",
		"activityId": activityId,
		"serviceId":  serviceId,
	})

	err := h.database.RemoveServiceFromActivity(activityId, serviceId)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while removing service from activity.")
	}

	return h.GetActivity(activityId)
}
