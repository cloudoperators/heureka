// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"

	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

type serviceHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewServiceHandler(db database.Database, er event.EventRegistry) ServiceHandler {
	return &serviceHandler{
		database:      db,
		eventRegistry: er,
	}
}

type ServiceHandlerError struct {
	msg string
}

func (e *ServiceHandlerError) Error() string {
	return fmt.Sprintf("ServiceHandlerError: %s", e.msg)
}

func NewServiceHandlerError(msg string) *ServiceHandlerError {
	return &ServiceHandlerError{msg: msg}
}

func (s *serviceHandler) getServiceResults(filter *entity.ServiceFilter) ([]entity.ServiceResult, error) {
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

func (s *serviceHandler) GetService(serviceId int64) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event": GetServiceEventName,
		"id":    serviceId,
	})
	serviceFilter := entity.ServiceFilter{Id: []*int64{&serviceId}}
	services, err := s.ListServices(&serviceFilter, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while retrieving services.")
	}

	if len(services.Elements) != 1 {
		return nil, NewServiceHandlerError(fmt.Sprintf("Service %d not found.", serviceId))
	}

	s.eventRegistry.PushEvent(&GetServiceEvent{ServiceID: serviceId, Service: services.Elements[0].Service})

	return services.Elements[0].Service, nil
}

func (s *serviceHandler) ListServices(filter *entity.ServiceFilter, options *entity.ListOptions) (*entity.List[entity.ServiceResult], error) {
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
		return nil, NewServiceHandlerError("Error while filtering for Services")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			ids, err := s.database.GetAllServiceIds(filter)
			if err != nil {
				l.Error(err)
				return nil, NewServiceHandlerError("Error while getting all Ids")
			}
			pageInfo = common.GetPageInfo(res, ids, *filter.First, *filter.After)
			count = int64(len(ids))
		}
	} else if options.ShowTotalCount {
		count, err = s.database.CountServices(filter)
		if err != nil {
			l.Error(err)
			return nil, NewServiceHandlerError("Error while total count of Services")
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

func (s *serviceHandler) CreateService(service *entity.Service) (*entity.Service, error) {
	f := &entity.ServiceFilter{
		CCRN: []*string{&service.CCRN},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateServiceEventName,
		"object": service,
		"filter": f,
	})

	services, err := s.ListServices(f, &entity.ListOptions{})

	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while creating service.")
	}

	if len(services.Elements) > 0 {
		return nil, NewServiceHandlerError(fmt.Sprintf("Duplicated entry %s for name.", service.CCRN))
	}

	newService, err := s.database.CreateService(service)

	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while creating service.")
	}

	s.eventRegistry.PushEvent(&CreateServiceEvent{Service: newService})

	return newService, nil
}

func (s *serviceHandler) UpdateService(service *entity.Service) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateServiceEventName,
		"object": service,
	})

	err := s.database.UpdateService(service)

	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while updating service.")
	}

	s.eventRegistry.PushEvent(&UpdateServiceEvent{Service: service})

	return s.GetService(service.Id)
}

func (s *serviceHandler) DeleteService(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteServiceEventName,
		"id":    id,
	})

	err := s.database.DeleteService(id)

	if err != nil {
		l.Error(err)
		return NewServiceHandlerError("Internal error while deleting service.")
	}

	s.eventRegistry.PushEvent(&DeleteServiceEvent{ServiceID: id})

	return nil
}

func (s *serviceHandler) AddOwnerToService(serviceId, ownerId int64) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":     AddOwnerToServiceEventName,
		"serviceId": serviceId,
		"ownerId":   ownerId,
	})

	err := s.database.AddOwnerToService(serviceId, ownerId)

	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while adding owner to service.")
	}

	s.eventRegistry.PushEvent(&AddOwnerToServiceEvent{ServiceID: serviceId, OwnerID: ownerId})

	return s.GetService(serviceId)
}

func (s *serviceHandler) RemoveOwnerFromService(serviceId, ownerId int64) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":     RemoveOwnerFromServiceEventName,
		"serviceId": serviceId,
		"ownerId":   ownerId,
	})

	err := s.database.RemoveOwnerFromService(serviceId, ownerId)

	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while removing owner from service.")
	}

	s.eventRegistry.PushEvent(&RemoveOwnerFromServiceEvent{ServiceID: serviceId, OwnerID: ownerId})

	return s.GetService(serviceId)
}

func (s *serviceHandler) AddIssueRepositoryToService(serviceId, issueRepositoryId int64, priority int64) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":             AddIssueRepositoryToServiceEventName,
		"serviceId":         serviceId,
		"issueRepositoryId": issueRepositoryId,
	})

	err := s.database.AddIssueRepositoryToService(serviceId, issueRepositoryId, priority)

	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while adding issue repository to service.")
	}

	s.eventRegistry.PushEvent(&AddIssueRepositoryToServiceEvent{ServiceID: serviceId, RepositoryID: issueRepositoryId})

	return s.GetService(serviceId)
}

func (s *serviceHandler) RemoveIssueRepositoryFromService(serviceId, issueRepositoryId int64) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":             RemoveIssueRepositoryFromServiceEventName,
		"serviceId":         serviceId,
		"issueRepositoryId": issueRepositoryId,
	})

	err := s.database.RemoveIssueRepositoryFromService(serviceId, issueRepositoryId)

	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while removing issue repository from service.")
	}

	s.eventRegistry.PushEvent(&RemoveIssueRepositoryFromServiceEvent{ServiceID: serviceId, RepositoryID: issueRepositoryId})

	return s.GetService(serviceId)
}

func (s *serviceHandler) ListServiceCcrns(filter *entity.ServiceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListServiceCcrnsEventName,
		"filter": filter,
	})

	serviceCcrns, err := s.database.GetServiceCcrns(filter)

	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while retrieving serviceCCRNs.")
	}

	s.eventRegistry.PushEvent(&ListServiceCcrnsEvent{Filter: filter, Options: options, Ccrns: serviceCcrns})

	return serviceCcrns, nil
}
