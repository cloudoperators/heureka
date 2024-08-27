package component_version

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/app/common"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/database"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

type componentVersionService struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewComponentVersionService(database database.Database, eventRegistry event.EventRegistry) ComponentVersionService {
	return &componentVersionService{
		database:      database,
		eventRegistry: eventRegistry,
	}
}

type ComponentVersionServiceError struct {
	message string
}

func NewComponentVersionServiceError(message string) *ComponentVersionServiceError {
	return &ComponentVersionServiceError{message: message}
}

func (e *ComponentVersionServiceError) Error() string {
	return e.message
}

func (cv *componentVersionService) getComponentVersionResults(filter *entity.ComponentVersionFilter) ([]entity.ComponentVersionResult, error) {
	var componentVersionResults []entity.ComponentVersionResult
	componentVersions, err := cv.database.GetComponentVersions(filter)
	if err != nil {
		return nil, err
	}
	for _, cv := range componentVersions {
		componentVersion := cv
		cursor := fmt.Sprintf("%d", componentVersion.Id)
		componentVersionResults = append(componentVersionResults, entity.ComponentVersionResult{
			WithCursor:                   entity.WithCursor{Value: cursor},
			ComponentVersionAggregations: nil,
			ComponentVersion:             &componentVersion,
		})
	}
	return componentVersionResults, nil
}

func (cv *componentVersionService) ListComponentVersions(filter *entity.ComponentVersionFilter, options *entity.ListOptions) (*entity.List[entity.ComponentVersionResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListComponentVersionsEventName,
		"filter": filter,
	})

	res, err := cv.getComponentVersionResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionServiceError("Error while filtering for ComponentVersions")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := cv.database.GetAllComponentVersionIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewComponentVersionServiceError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = cv.database.CountComponentVersions(filter)
		if err != nil {
			l.Error(err)
			return nil, NewComponentVersionServiceError("Error while total count of ComponentVersions")
		}
	}

	cv.eventRegistry.PushEvent(&ListComponentVersionsEvent{
		Filter:  filter,
		Options: options,
		ComponentVersions: &entity.List[entity.ComponentVersionResult]{
			TotalCount: &count,
			PageInfo:   pageInfo,
			Elements:   res,
		},
	})

	return &entity.List[entity.ComponentVersionResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (cv *componentVersionService) CreateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateComponentVersionEventName,
		"object": componentVersion,
	})

	newComponent, err := cv.database.CreateComponentVersion(componentVersion)

	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionServiceError("Internal error while creating componentVersion.")
	}

	cv.eventRegistry.PushEvent(&CreateComponentVersionEvent{
		ComponentVersion: newComponent,
	})

	return newComponent, nil
}

func (cv *componentVersionService) UpdateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateComponentVersionEventName,
		"object": componentVersion,
	})

	err := cv.database.UpdateComponentVersion(componentVersion)

	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionServiceError("Internal error while updating componentVersion.")
	}

	componentVersionResult, err := cv.ListComponentVersions(&entity.ComponentVersionFilter{Id: []*int64{&componentVersion.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewComponentVersionServiceError("Internal error while retrieving updated componentVersion.")
	}

	if len(componentVersionResult.Elements) != 1 {
		l.Error(err)
		return nil, NewComponentVersionServiceError("Multiple componentVersions found.")
	}

	cv.eventRegistry.PushEvent(&UpdateComponentVersionEvent{
		ComponentVersion: componentVersion,
	})

	return componentVersionResult.Elements[0].ComponentVersion, nil
}

func (cv *componentVersionService) DeleteComponentVersion(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteComponentVersionEventName,
		"id":    id,
	})

	err := cv.database.DeleteComponentVersion(id)

	if err != nil {
		l.Error(err)
		return NewComponentVersionServiceError("Internal error while deleting componentVersion.")
	}

	cv.eventRegistry.PushEvent(&DeleteComponentVersionEvent{
		ComponentVersionID: id,
	})

	return nil
}
