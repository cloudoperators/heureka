// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getComponentVersionResults(filter *entity.ComponentVersionFilter) ([]entity.ComponentVersionResult, error) {
	var componentVersionResults []entity.ComponentVersionResult
	componentVersions, err := h.database.GetComponentVersions(filter)
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

func (h *HeurekaApp) ListComponentVersions(filter *entity.ComponentVersionFilter, options *entity.ListOptions) (*entity.List[entity.ComponentVersionResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListComponentVersions",
		"filter": filter,
	})

	res, err := h.getComponentVersionResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for ComponentVersions")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllComponentVersionIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountComponentVersions(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of ComponentVersions")
		}
	}

	return &entity.List[entity.ComponentVersionResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) CreateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateComponentVersion",
		"object": componentVersion,
	})

	newComponent, err := h.database.CreateComponentVersion(componentVersion)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating componentVersion.")
	}

	return newComponent, nil
}

func (h *HeurekaApp) UpdateComponentVersion(componentVersion *entity.ComponentVersion) (*entity.ComponentVersion, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateComponentVersion",
		"object": componentVersion,
	})

	err := h.database.UpdateComponentVersion(componentVersion)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating componentVersion.")
	}

	componentVersionResult, err := h.ListComponentVersions(&entity.ComponentVersionFilter{Id: []*int64{&componentVersion.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving updated componentVersion.")
	}

	if len(componentVersionResult.Elements) != 1 {
		l.Error(err)
		return nil, heurekaError("Multiple componentVersions found.")
	}

	return componentVersionResult.Elements[0].ComponentVersion, nil
}

func (h *HeurekaApp) DeleteComponentVersion(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteComponentVersion",
		"id":    id,
	})

	err := h.database.DeleteComponentVersion(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while deleting componentVersion.")
	}

	return nil
}
