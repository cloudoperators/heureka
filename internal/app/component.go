// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getComponentResults(filter *entity.ComponentFilter) ([]entity.ComponentResult, error) {
	var componentResults []entity.ComponentResult
	components, err := h.database.GetComponents(filter)
	if err != nil {
		return nil, err
	}
	for _, c := range components {
		component := c
		cursor := fmt.Sprintf("%d", component.Id)
		componentResults = append(componentResults, entity.ComponentResult{
			WithCursor:            entity.WithCursor{Value: cursor},
			ComponentAggregations: nil,
			Component:             &component,
		})
	}
	return componentResults, nil
}

func (h *HeurekaApp) ListComponents(filter *entity.ComponentFilter, options *entity.ListOptions) (*entity.List[entity.ComponentResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListComponents",
		"filter": filter,
	})

	res, err := h.getComponentResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for Components")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllComponentIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountComponents(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of Components")
		}
	}

	return &entity.List[entity.ComponentResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) CreateComponent(component *entity.Component) (*entity.Component, error) {
	f := &entity.ComponentFilter{
		Name: []*string{&component.Name},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateComponent",
		"object": component,
		"filter": f,
	})

	components, err := h.ListComponents(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating component.")
	}

	if len(components.Elements) > 0 {
		return nil, heurekaError(fmt.Sprintf("Duplicated entry %s for name.", component.Name))
	}

	newComponent, err := h.database.CreateComponent(component)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating component.")
	}

	return newComponent, nil
}

func (h *HeurekaApp) UpdateComponent(component *entity.Component) (*entity.Component, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateComponent",
		"object": component,
	})

	err := h.database.UpdateComponent(component)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating component.")
	}

	componentResult, err := h.ListComponents(&entity.ComponentFilter{Id: []*int64{&component.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving updated component.")
	}

	if len(componentResult.Elements) != 1 {
		l.Error(err)
		return nil, heurekaError("Multiple components found.")
	}

	return componentResult.Elements[0].Component, nil
}

func (h *HeurekaApp) DeleteComponent(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteComponent",
		"id":    id,
	})

	err := h.database.DeleteComponent(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while deleting component.")
	}

	return nil
}
