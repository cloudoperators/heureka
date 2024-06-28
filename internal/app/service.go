// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

func (h *HeurekaApp) getServiceResults(filter *entity.ServiceFilter) ([]entity.ServiceResult, error) {
	var serviceResults []entity.ServiceResult
	services, err := h.database.GetServices(filter)
	if err != nil {
		return nil, err
	}
	for _, s := range services {
		service := s
		cursor := fmt.Sprintf("%d", service.Id)
		serviceResults = append(serviceResults, entity.ServiceResult{
			WithCursor:          entity.WithCursor{Value: cursor},
			ServiceAggregations: nil,
			Service:             &service,
		})
	}
	return serviceResults, nil
}

func (h *HeurekaApp) ListServices(filter *entity.ServiceFilter, options *entity.ListOptions) (*entity.List[entity.ServiceResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	ensurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.ListServices",
		"filter": filter,
	})

	res, err := h.getServiceResults(filter)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Error while filtering for Services")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := h.database.GetAllServiceIds(filter)
			if err != nil {
				l.Error(err)
				return nil, heurekaError("Error while getting all Ids")
			}
			pageInfo = getPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = h.database.CountServices(filter)
		if err != nil {
			l.Error(err)
			return nil, heurekaError("Error while total count of Services")
		}
	}

	return &entity.List[entity.ServiceResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (h *HeurekaApp) CreateService(service *entity.Service) (*entity.Service, error) {
	f := &entity.ServiceFilter{
		Name: []*string{&service.Name},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  "app.CreateService",
		"object": service,
		"filter": f,
	})

	services, err := h.ListServices(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating service.")
	}

	if len(services.Elements) > 0 {
		return nil, heurekaError(fmt.Sprintf("Duplicated entry %s for name.", service.Name))
	}

	newService, err := h.database.CreateService(service)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while creating service.")
	}

	return newService, nil
}

func (h *HeurekaApp) UpdateService(service *entity.Service) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  "app.UpdateService",
		"object": service,
	})

	err := h.database.UpdateService(service)

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while updating service.")
	}

	serviceResult, err := h.ListServices(&entity.ServiceFilter{Id: []*int64{&service.Id}}, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, heurekaError("Internal error while retrieving updated service.")
	}

	if len(serviceResult.Elements) != 1 {
		l.Error(err)
		return nil, heurekaError("Multiple services found.")
	}

	return serviceResult.Elements[0].Service, nil
}

func (h *HeurekaApp) DeleteService(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": "app.DeleteService",
		"id":    id,
	})

	err := h.database.DeleteService(id)

	if err != nil {
		l.Error(err)
		return heurekaError("Internal error while deleting service.")
	}

	return nil
}
