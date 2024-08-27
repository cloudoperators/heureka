// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package activity

import (
	"fmt"
	"github.wdf.sap.corp/cc/heureka/internal/app/common"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/database"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

type activityService struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewActivityService(db database.Database, er event.EventRegistry) ActivityService {
	return &activityService{
		database:      db,
		eventRegistry: er,
	}
}

type ActivityServiceError struct {
	msg string
}

func (e *ActivityServiceError) Error() string {
	return fmt.Sprintf("ActivityServiceError: %s", e.msg)
}

func NewActivityServiceError(msg string) *ActivityServiceError {
	return &ActivityServiceError{msg: msg}
}

func (a *activityService) getActivityResults(filter *entity.ActivityFilter) ([]entity.ActivityResult, error) {
	var activityResults []entity.ActivityResult
	activities, err := a.database.GetActivities(filter)
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

func (a *activityService) GetActivity(activityId int64) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": GetActivityEventName,
		"id":    activityId,
	})
	activityFilter := entity.ActivityFilter{Id: []*int64{&activityId}}
	activities, err := a.ListActivities(&activityFilter, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewActivityServiceError("Internal error while retrieving activities.")
	}

	if len(activities.Elements) != 1 {
		return nil, NewActivityServiceError(fmt.Sprintf("Activity %d not found.", activityId))
	}

	a.eventRegistry.PushEvent(&GetActivityEvent{
		ActivityID: activityId,
		Activity:   activities.Elements[0].Activity,
	})

	return activities.Elements[0].Activity, nil
}

func (a *activityService) ListActivities(filter *entity.ActivityFilter, options *entity.ListOptions) (*entity.List[entity.ActivityResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)
	l := logrus.WithFields(logrus.Fields{
		"event":  ListActivitiesEventName,
		"filter": filter,
	})

	res, err := a.getActivityResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewActivityServiceError("Error while filtering for Activities")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := a.database.GetAllActivityIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewActivityServiceError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = a.database.CountActivities(filter)
		if err != nil {
			l.Error(err)
			return nil, NewActivityServiceError("Error while total count of Activities")
		}
	}

	ret := &entity.List[entity.ActivityResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	a.eventRegistry.PushEvent(&ListActivitiesEvent{
		Filter:     filter,
		Options:    options,
		Activities: ret,
	})

	return ret, nil
}

func (a *activityService) CreateActivity(activity *entity.Activity) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ActivityCreateEventName,
		"object": activity,
	})

	newActivity, err := a.database.CreateActivity(activity)

	if err != nil {
		l.Error(err)
		return nil, NewActivityServiceError("Internal error while creating activity.")
	}

	a.eventRegistry.PushEvent(&ActivityCreateEvent{
		Activity: newActivity,
	})

	return newActivity, nil
}

func (a *activityService) UpdateActivity(activity *entity.Activity) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ActivityUpdateEventName,
		"object": activity,
	})

	err := a.database.UpdateActivity(activity)

	if err != nil {
		l.Error(err)
		return nil, NewActivityServiceError("Internal error while updating activity.")
	}

	a.eventRegistry.PushEvent(&ActivityUpdateEvent{
		Activity: activity,
	})

	return a.GetActivity(activity.Id)
}

func (a *activityService) DeleteActivity(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": ActivityDeleteEventName,
		"id":    id,
	})

	err := a.database.DeleteActivity(id)

	if err != nil {
		l.Error(err)
		return NewActivityServiceError("Internal error while deleting activity.")
	}

	a.eventRegistry.PushEvent(&ActivityDeleteEvent{
		ActivityID: id,
	})

	return nil
}

func (a *activityService) AddServiceToActivity(activityId, serviceId int64) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":      AddServiceToActivityEventName,
		"activityId": activityId,
		"serviceId":  serviceId,
	})

	err := a.database.AddServiceToActivity(activityId, serviceId)

	if err != nil {
		l.Error(err)
		return nil, NewActivityServiceError("Internal error while adding service to activity.")
	}

	a.eventRegistry.PushEvent(&AddServiceToActivityEvent{
		ActivityID: activityId,
		ServiceID:  serviceId,
	})

	return a.GetActivity(activityId)
}

func (a *activityService) RemoveServiceFromActivity(activityId, serviceId int64) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":      RemoveServiceFromActivityEventName,
		"activityId": activityId,
		"serviceId":  serviceId,
	})

	err := a.database.RemoveServiceFromActivity(activityId, serviceId)

	if err != nil {
		l.Error(err)
		return nil, NewActivityServiceError("Internal error while removing service from activity.")
	}

	a.eventRegistry.PushEvent(&RemoveServiceFromActivityEvent{
		ActivityID: activityId,
		ServiceID:  serviceId,
	})

	return a.GetActivity(activityId)
}

func (a *activityService) AddIssueToActivity(activityId, issueId int64) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":      AddIssueToActivityEventName,
		"activityId": activityId,
		"issueId":    issueId,
	})

	err := a.database.AddIssueToActivity(activityId, issueId)

	if err != nil {
		l.Error(err)
		return nil, NewActivityServiceError("Internal error while adding issue to activity.")
	}

	a.eventRegistry.PushEvent(&AddIssueToActivityEvent{
		ActivityID: activityId,
		IssueID:    issueId,
	})

	return a.GetActivity(activityId)
}

func (a *activityService) RemoveIssueFromActivity(activityId, issueId int64) (*entity.Activity, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":      RemoveIssueFromActivityEventName,
		"activityId": activityId,
		"issueId":    issueId,
	})

	err := a.database.RemoveIssueFromActivity(activityId, issueId)

	if err != nil {
		l.Error(err)
		return nil, NewActivityServiceError("Internal error while removing issue from activity.")
	}

	a.eventRegistry.PushEvent(&RemoveIssueFromActivityEvent{
		ActivityID: activityId,
		IssueID:    issueId,
	})

	return a.GetActivity(activityId)
}
