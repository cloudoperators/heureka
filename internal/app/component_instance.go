// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getComponentInstanceResults(filter *entity.ComponentInstanceFilter) ([]entity.ComponentInstanceResult, error) {
	var componentInstanceResults []entity.ComponentInstanceResult
	entries, err := h.database.GetComponentInstances(filter)
	if err != nil {
		return nil, err
	}

	for _, ci := range entries {
		componentInstance := ci
		cursor := fmt.Sprintf("%d", componentInstance.Id)
		componentInstanceResults = append(componentInstanceResults, entity.ComponentInstanceResult{
			WithCursor:                    entity.WithCursor{Value: cursor},
			ComponentInstanceAggregations: nil,
			ComponentInstance:             &componentInstance,
		})
	}

	return componentInstanceResults, nil
}

func (h *HeurekaApp) ListComponentInstances(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) (*entity.List[entity.ComponentInstanceResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListComponentInstances",
		"filter": filter,
	})

	res, err := h.getComponentInstanceResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for ComponentInstances")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllComponentInstanceIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountComponentInstances(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of ComponentInstances")
		}
	}

	return &entity.List[entity.ComponentInstanceResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) CreateComponentInstance(componentInstance *entity.ComponentInstance) (*entity.ComponentInstance, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateComponentInstance",
		"object": componentInstance,
	})

	newComponentInstance, err := h.database.CreateComponentInstance(componentInstance)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating componentInstance.")
	}

	return newComponentInstance, nil
}

func (h *HeurekaApp) UpdateComponentInstance(componentInstance *entity.ComponentInstance) (*entity.ComponentInstance, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateComponentInstance",
		"object": componentInstance,
	})

	err := h.database.UpdateComponentInstance(componentInstance)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating componentInstance.")
	}

	componentInstanceResult, err := h.ListComponentInstances(&entity.ComponentInstanceFilter{Id: []*int64{&componentInstance.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving updated componentInstance.")
	}

	if len(componentInstanceResult.Elements) != 1 {
		l.Error(err)
		return nil, heurekaError("Multiple componentInstances found.")
	}

	return componentInstanceResult.Elements[0].ComponentInstance, nil
}

func (h *HeurekaApp) DeleteComponentInstance(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteComponentInstance",
		"id":    id,
	})

	err := h.database.DeleteComponentInstance(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while deleting componentInstance.")
	}

	return nil
}
