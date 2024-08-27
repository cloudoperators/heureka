// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"
	"github.wdf.sap.corp/cc/heureka/internal/app/common"
	"github.wdf.sap.corp/cc/heureka/internal/app/event"
	"github.wdf.sap.corp/cc/heureka/internal/database"

	"github.com/sirupsen/logrus"
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

type serviceService struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewServiceService(db database.Database, er event.EventRegistry) ServiceService {
	return &serviceService{
		database:      db,
		eventRegistry: er,
	}
}

type ServiceServiceError struct {
	msg string
}

func (e *ServiceServiceError) Error() string {
	return fmt.Sprintf("ServiceServiceError: %s", e.msg)
}

func NewServiceServiceError(msg string) *ServiceServiceError {
	return &ServiceServiceError{msg: msg}
}

func (s *serviceService) getServiceResults(filter *entity.ServiceFilter) ([]entity.ServiceResult, error) {
	var serviceResults []entity.ServiceResult
	services, err := s.database.GetServices(filter)
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

func (s *serviceService) GetService(serviceId int64) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": GetServiceEventName,
		"id":    serviceId,
	})
	serviceFilter := entity.ServiceFilter{Id: []*int64{&serviceId}}
	services, err := s.ListServices(&serviceFilter, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewServiceServiceError("Internal error while retrieving services.")
	}

	if len(services.Elements) != 1 {
		return nil, NewServiceServiceError(fmt.Sprintf("Service %d not found.", serviceId))
	}

	s.eventRegistry.PushEvent(&GetServiceEvent{ServiceID: serviceId, Service: services.Elements[0].Service})

	return services.Elements[0].Service, nil
}

func (s *serviceService) ListServices(filter *entity.ServiceFilter, options *entity.ListOptions) (*entity.List[entity.ServiceResult], error) {
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginated(&filter.Paginated)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListServicesEventName,
		"filter": filter,
	})

	res, err := s.getServiceResults(filter)

	if err != nil {
		l.Error(err)
		return nil, NewServiceServiceError("Error while filtering for Services")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := s.database.GetAllServiceIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewServiceServiceError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = s.database.CountServices(filter)
		if err != nil {
			l.Error(err)
			return nil, NewServiceServiceError("Error while total count of Services")
		}
	}
	ret := &entity.List[entity.ServiceResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	s.eventRegistry.PushEvent(&ListServicesEvent{Filter: filter, Options: options, Services: ret})

	return ret, nil
}

func (s *serviceService) CreateService(service *entity.Service) (*entity.Service, error) {
	f := &entity.ServiceFilter{
		Name: []*string{&service.Name},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateServiceEventName,
		"object": service,
		"filter": f,
	})

	services, err := s.ListServices(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewServiceServiceError("Internal error while creating service.")
	}

	if len(services.Elements) > 0 {
		return nil, NewServiceServiceError(fmt.Sprintf("Duplicated entry %s for name.", service.Name))
	}

	newService, err := s.database.CreateService(service)

	if err != nil {
		l.Error(err)
		return nil, NewServiceServiceError("Internal error while creating service.")
	}

	s.eventRegistry.PushEvent(&CreateServiceEvent{Service: newService})

	return newService, nil
}

func (s *serviceService) UpdateService(service *entity.Service) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateServiceEventName,
		"object": service,
	})

	err := s.database.UpdateService(service)

	if err != nil {
		l.Error(err)
		return nil, NewServiceServiceError("Internal error while updating service.")
	}

	s.eventRegistry.PushEvent(&UpdateServiceEvent{Service: service})

	return s.GetService(service.Id)
}

func (s *serviceService) DeleteService(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteServiceEventName,
		"id":    id,
	})

	err := s.database.DeleteService(id)

	if err != nil {
		l.Error(err)
		return NewServiceServiceError("Internal error while deleting service.")
	}

	s.eventRegistry.PushEvent(&DeleteServiceEvent{ServiceID: id})

	return nil
}

func (s *serviceService) AddOwnerToService(serviceId, ownerId int64) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":     AddOwnerToServiceEventName,
		"serviceId": serviceId,
		"ownerId":   ownerId,
	})

	err := s.database.AddOwnerToService(serviceId, ownerId)

	if err != nil {
		l.Error(err)
		return nil, NewServiceServiceError("Internal error while adding owner to service.")
	}

	s.eventRegistry.PushEvent(&AddOwnerToServiceEvent{ServiceID: serviceId, OwnerID: ownerId})

	return s.GetService(serviceId)
}

func (s *serviceService) RemoveOwnerFromService(serviceId, ownerId int64) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":     RemoveOwnerFromServiceEventName,
		"serviceId": serviceId,
		"ownerId":   ownerId,
	})

	err := s.database.RemoveOwnerFromService(serviceId, ownerId)

	if err != nil {
		l.Error(err)
		return nil, NewServiceServiceError("Internal error while removing owner from service.")
	}

	s.eventRegistry.PushEvent(&RemoveOwnerFromServiceEvent{ServiceID: serviceId, OwnerID: ownerId})

	return s.GetService(serviceId)
}

func (s *serviceService) AddIssueRepositoryToService(serviceId, issueRepositoryId int64, priority int64) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":             AddIssueRepositoryToServiceEventName,
		"serviceId":         serviceId,
		"issueRepositoryId": issueRepositoryId,
	})

	err := s.database.AddIssueRepositoryToService(serviceId, issueRepositoryId, priority)

	if err != nil {
		l.Error(err)
		return nil, NewServiceServiceError("Internal error while adding issue repository to service.")
	}

	s.eventRegistry.PushEvent(&AddIssueRepositoryToServiceEvent{ServiceID: serviceId, RepositoryID: issueRepositoryId})

	return s.GetService(serviceId)
}

func (s *serviceService) RemoveIssueRepositoryFromService(serviceId, issueRepositoryId int64) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":             RemoveIssueRepositoryFromServiceEventName,
		"serviceId":         serviceId,
		"issueRepositoryId": issueRepositoryId,
	})

	err := s.database.RemoveIssueRepositoryFromService(serviceId, issueRepositoryId)

	if err != nil {
		l.Error(err)
		return nil, NewServiceServiceError("Internal error while removing issue repository from service.")
	}

	s.eventRegistry.PushEvent(&RemoveIssueRepositoryFromServiceEvent{ServiceID: serviceId, RepositoryID: issueRepositoryId})

	return s.GetService(serviceId)
}

func (s *serviceService) ListServiceNames(filter *entity.ServiceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListServiceNamesEventName,
		"filter": filter,
	})

	serviceNames, err := s.database.GetServiceNames(filter)

	if err != nil {
		l.Error(err)
		return nil, NewServiceServiceError("Internal error while retrieving serviceNames.")
	}

	s.eventRegistry.PushEvent(&ListServiceNamesEvent{Filter: filter, Options: options, Names: serviceNames})

	return serviceNames, nil
}
