// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

var (
	CacheTtlGetServiceAttrs             = 12 * time.Hour
	CacheTtlGetServicesWithAggregations = 12 * time.Hour
	CacheTtlGetServices                 = 12 * time.Hour
	CacheTtlGetAllSericeCursors         = 12 * time.Hour
	CacheTtlCountServices               = 12 * time.Hour
)

type serviceHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	cache         cache.Cache
	authz         openfga.Authorization
	logger        *logrus.Logger
}

func NewServiceHandler(handlerContext common.HandlerContext) ServiceHandler {
	return &serviceHandler{
		database:      handlerContext.DB,
		eventRegistry: handlerContext.EventReg,
		cache:         handlerContext.Cache,
		authz:         handlerContext.Authz,
		logger:        logrus.New(),
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

func (s *serviceHandler) GetService(ctx context.Context, serviceId int64) (*entity.Service, error) {
	op := appErrors.Op("serviceHandler.GetService")

	// get current user id
	currentUserId, err := common.GetCurrentUserId(ctx, s.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Services", fmt.Sprint(serviceId), err)
		applog.LogError(s.logger, wrappedErr, logrus.Fields{
			"serviceId": serviceId,
		})

		return nil, wrappedErr
	}

	// Authorization check
	hasPermission, err := s.authz.CheckPermission(openfga.RelationInput{
		UserType:   openfga.TypeUser,
		UserId:     openfga.UserId(fmt.Sprint(currentUserId)),
		Relation:   openfga.RelCanView,
		ObjectType: openfga.TypeService,
		ObjectId:   openfga.ObjectId(fmt.Sprint(serviceId)),
	})
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Services", fmt.Sprint(serviceId), err)
		applog.LogError(s.logger, wrappedErr, logrus.Fields{
			"serviceId": serviceId,
		})

		return nil, wrappedErr
	}

	if !hasPermission {
		wrappedErr := appErrors.PermissionDeniedError(string(op), "Service", fmt.Sprint(serviceId))
		applog.LogError(s.logger, wrappedErr, logrus.Fields{
			"serviceId": serviceId,
			"userId":    currentUserId,
		})

		return nil, wrappedErr
	}

	serviceFilter := entity.ServiceFilter{Id: []*int64{&serviceId}}
	lo := entity.NewListOptions()

	services, err := s.ListServices(ctx, &serviceFilter, lo)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Services", fmt.Sprint(serviceId), err)
		applog.LogError(s.logger, wrappedErr, logrus.Fields{
			"serviceId": serviceId,
		})

		return nil, wrappedErr
	}

	if len(services.Elements) != 1 {
		wrappedErr := appErrors.NotFoundError(string(op), "Service", fmt.Sprint(serviceId))
		applog.LogError(s.logger, wrappedErr, logrus.Fields{
			"serviceId": serviceId,
		})

		return nil, wrappedErr
	}

	s.eventRegistry.PushEvent(
		&GetServiceEvent{ServiceID: serviceId, Service: services.Elements[0].Service},
	)

	return services.Elements[0].Service, nil
}

func (s *serviceHandler) ListServices(ctx context.Context,
	filter *entity.ServiceFilter, options *entity.ListOptions,
) (*entity.List[entity.ServiceResult], error) {
	var (
		count    int64
		pageInfo *entity.PageInfo
		res      []entity.ServiceResult
		err      error
	)

	op := appErrors.Op("serviceHandler.ListServices")

	common.EnsurePaginated(&filter.Paginated)

	options = common.EnsureListOptions(options)

	// get current user id
	currentUserId, err := common.GetCurrentUserId(ctx, s.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Services", "", err)
		applog.LogError(s.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})

		return nil, wrappedErr
	}

	// Authorization check
	accessibleSupportGroupIds, err := s.authz.GetListOfAccessibleObjectIds(
		openfga.UserId(fmt.Sprint(currentUserId)),
		openfga.TypeSupportGroup,
	)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Services", "", err)
		applog.LogError(s.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})

		return nil, wrappedErr
	}

	// Update the filter.Id based on accessibleSupportGroupIds
	filter.SupportGroupId = common.CombineFilterWithAccessibleIds(
		filter.SupportGroupId,
		accessibleSupportGroupIds,
	)

	if options.IncludeAggregations {
		res, err = cache.CallCached[[]entity.ServiceResult](
			s.cache,
			CacheTtlGetServicesWithAggregations,
			"GetServicesWithAggregations",
			cache.WrapContext2(ctx, s.database.GetServicesWithAggregations),
			filter,
			options.Order,
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "Services", "", err)
			applog.LogError(s.logger, wrappedErr, logrus.Fields{
				"filter": filter,
			})

			return nil, wrappedErr
		}
	} else {
		res, err = cache.CallCached[[]entity.ServiceResult](
			s.cache,
			CacheTtlGetServices,
			"GetServices",
			cache.WrapContext2(ctx, s.database.GetServices),
			filter,
			options.Order,
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "Services", "", err)
			applog.LogError(s.logger, wrappedErr, logrus.Fields{
				"filter": filter,
			})

			return nil, wrappedErr
		}
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := cache.CallCached[[]string](
				s.cache,
				CacheTtlGetAllSericeCursors,
				"GetAllServiceCursors",
				cache.WrapContext2(ctx, s.database.GetAllServiceCursors),
				filter,
				options.Order,
			)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), "Services", "", err)
				applog.LogError(s.logger, wrappedErr, logrus.Fields{
					"filter": filter,
				})

				return nil, wrappedErr
			}

			pageInfo = common.GetPageInfo(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = cache.CallCached[int64](
			s.cache,
			CacheTtlCountServices,
			"CountServices",
			cache.WrapContext1(ctx, s.database.CountServices),
			filter,
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "Services", "", err)
			applog.LogError(s.logger, wrappedErr, logrus.Fields{
				"filter": filter,
			})

			return nil, wrappedErr
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

func (s *serviceHandler) CreateService(
	ctx context.Context,
	service *entity.Service,
) (*entity.Service, error) {
	f := &entity.ServiceFilter{
		CCRN: []*string{&service.CCRN},
	}

	l := logrus.WithFields(logrus.Fields{
		"event":  CreateServiceEventName,
		"object": service,
		"filter": f,
	})

	var err error

	service.BaseService.CreatedBy, err = common.GetCurrentUserId(ctx, s.database)
	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while creating service (GetUserId).")
	}

	service.BaseService.UpdatedBy = service.BaseService.CreatedBy
	lo := entity.NewListOptions()

	services, err := s.ListServices(ctx, f, lo)
	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while creating service.")
	}

	if len(services.Elements) > 0 {
		return nil, NewServiceHandlerError(
			fmt.Sprintf("Duplicated entry %s for name.", service.CCRN),
		)
	}

	newService, err := s.database.CreateService(service)
	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while creating service.")
	}

	s.eventRegistry.PushEvent(&CreateServiceEvent{Service: newService})

	return newService, nil
}

func (s *serviceHandler) UpdateService(
	ctx context.Context,
	service *entity.Service,
) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateServiceEventName,
		"object": service,
	})

	var err error

	service.BaseService.UpdatedBy, err = common.GetCurrentUserId(ctx, s.database)
	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while updating service (GetUserId).")
	}

	err = s.database.UpdateService(service)
	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while updating service.")
	}

	s.eventRegistry.PushEvent(&UpdateServiceEvent{Service: service})

	return s.GetService(ctx, service.Id)
}

func (s *serviceHandler) DeleteService(ctx context.Context, id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteServiceEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(ctx, s.database)
	if err != nil {
		l.Error(err)
		return NewServiceHandlerError("Internal error while deleting service (GetUserId).")
	}

	err = s.database.DeleteService(id, userId)
	if err != nil {
		l.Error(err)
		return NewServiceHandlerError("Internal error while deleting service.")
	}

	s.eventRegistry.PushEvent(&DeleteServiceEvent{ServiceID: id})

	return nil
}

func (s *serviceHandler) AddOwnerToService(
	ctx context.Context,
	serviceId, ownerId int64,
) (*entity.Service, error) {
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

	return s.GetService(ctx, serviceId)
}

func (s *serviceHandler) RemoveOwnerFromService(
	ctx context.Context,
	serviceId, ownerId int64,
) (*entity.Service, error) {
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

	return s.GetService(ctx, serviceId)
}

func (s *serviceHandler) AddIssueRepositoryToService(ctx context.Context, serviceId,
	issueRepositoryId int64, priority int64,
) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":             AddIssueRepositoryToServiceEventName,
		"serviceId":         serviceId,
		"issueRepositoryId": issueRepositoryId,
	})

	err := s.database.AddIssueRepositoryToService(serviceId, issueRepositoryId, priority)
	if err != nil {
		l.Error(err)

		return nil, NewServiceHandlerError(
			"Internal error while adding issue repository to service.",
		)
	}

	s.eventRegistry.PushEvent(
		&AddIssueRepositoryToServiceEvent{ServiceID: serviceId, RepositoryID: issueRepositoryId},
	)

	return s.GetService(ctx, serviceId)
}

func (s *serviceHandler) RemoveIssueRepositoryFromService(
	ctx context.Context,
	serviceId, issueRepositoryId int64,
) (*entity.Service, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":             RemoveIssueRepositoryFromServiceEventName,
		"serviceId":         serviceId,
		"issueRepositoryId": issueRepositoryId,
	})

	err := s.database.RemoveIssueRepositoryFromService(serviceId, issueRepositoryId)
	if err != nil {
		l.Error(err)

		return nil, NewServiceHandlerError(
			"Internal error while removing issue repository from service.",
		)
	}

	s.eventRegistry.PushEvent(
		&RemoveIssueRepositoryFromServiceEvent{
			ServiceID:    serviceId,
			RepositoryID: issueRepositoryId,
		},
	)

	return s.GetService(ctx, serviceId)
}

func (s *serviceHandler) ListServiceCcrns(
	ctx context.Context,
	filter *entity.ServiceFilter,
	options *entity.ListOptions,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListServiceCcrnsEventName,
		"filter": filter,
	})

	serviceCcrns, err := cache.CallCached[[]string](
		s.cache,
		CacheTtlGetServiceAttrs,
		"GetServiceCcrns",
		cache.WrapContext1(ctx, s.database.GetServiceCcrns),
		filter,
	)
	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while retrieving serviceCCRNs.")
	}

	s.eventRegistry.PushEvent(
		&ListServiceCcrnsEvent{Filter: filter, Options: options, Ccrns: serviceCcrns},
	)

	return serviceCcrns, nil
}

func (s *serviceHandler) ListServiceDomains(
	ctx context.Context,
	filter *entity.ServiceFilter,
	options *entity.ListOptions,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListServiceDomainsEventName,
		"filter": filter,
	})

	serviceDomains, err := cache.CallCached[[]string](
		s.cache,
		CacheTtlGetServiceAttrs,
		"GetServiceDomains",
		cache.WrapContext1(ctx, s.database.GetServiceDomains),
		filter,
	)
	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while retrieving serviceDomains.")
	}

	s.eventRegistry.PushEvent(
		&ListServiceDomainsEvent{Filter: filter, Options: options, Domains: serviceDomains},
	)

	return serviceDomains, nil
}

func (s *serviceHandler) ListServiceRegions(
	ctx context.Context,
	filter *entity.ServiceFilter,
	options *entity.ListOptions,
) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListServiceRegionsEventName,
		"filter": filter,
	})

	serviceRegions, err := cache.CallCached[[]string](
		s.cache,
		CacheTtlGetServiceAttrs,
		"GetServiceRegions",
		cache.WrapContext1(ctx, s.database.GetServiceRegions),
		filter,
	)
	if err != nil {
		l.Error(err)
		return nil, NewServiceHandlerError("Internal error while retrieving serviceRegions.")
	}

	s.eventRegistry.PushEvent(
		&ListServiceRegionsEvent{Filter: filter, Options: options, Regions: serviceRegions},
	)

	return serviceRegions, nil
}
