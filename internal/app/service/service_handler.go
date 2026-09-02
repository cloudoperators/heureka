// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
)

type serviceHandler struct {
	common.BaseHandler[entity.ServiceResult, *entity.ServiceFilter]
}

func NewServiceHandler(handlerContext common.HandlerContext) ServiceHandler {
	return &serviceHandler{
		BaseHandler: common.NewBaseHandler(handlerContext, common.BaseConfig[entity.ServiceResult, *entity.ServiceFilter]{
			Op:                    appErrors.Op("serviceHandler"),
			Entity:                "Services",
			CursorEntity:          "ServiceCursors",
			CountEntity:           "ServiceCount",
			GetFn:                 handlerContext.DB.GetServices,
			GetWithAggregationsFn: handlerContext.DB.GetServicesWithAggregations,
			CursorsFn:             handlerContext.DB.GetAllServiceCursors,
			CountFn:               handlerContext.DB.CountServices,
			Authz:                 handlerContext.Authz,
			AuthzObjectType:       openfga.TypeSupportGroup,
			AuthzApplyFn: func(f *entity.ServiceFilter, ids []*int64) {
				f.SupportGroupId = common.CombineFilterWithAccessibleIds(f.SupportGroupId, ids)
			},
			ListEventFn: func(f *entity.ServiceFilter, o *entity.ListOptions, r *entity.List[entity.ServiceResult]) any {
				return &ListServicesEvent{Filter: f, Options: o, Services: r}
			},
			DeleteFn:      handlerContext.DB.DeleteService,
			DeleteEventFn: func(id int64) any { return &DeleteServiceEvent{ServiceID: id} },
		}),
	}
}

func (s *serviceHandler) GetService(ctx context.Context, serviceId int64) (*entity.Service, error) {
	op := appErrors.CallerOp()

	currentUserId, err := common.GetCurrentUserId(ctx, s.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Service", fmt.Sprint(serviceId), err)
	}

	hasPermission, err := s.Authz().CheckPermission(openfga.RelationInput{
		UserType:   openfga.TypeUser,
		UserId:     openfga.UserId(fmt.Sprint(currentUserId)),
		Relation:   openfga.RelCanView,
		ObjectType: openfga.TypeService,
		ObjectId:   openfga.ObjectId(fmt.Sprint(serviceId)),
	})
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Service", fmt.Sprint(serviceId), err)
	}

	if !hasPermission {
		return nil, appErrors.PermissionDeniedError(string(op), "Service", fmt.Sprint(serviceId))
	}

	result, err := s.ListServices(ctx, &entity.ServiceFilter{Id: []*int64{&serviceId}}, entity.NewListOptions())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Service", fmt.Sprint(serviceId), err)
	}

	if len(result.Elements) != 1 {
		return nil, appErrors.NotFoundError(string(op), "Service", fmt.Sprint(serviceId))
	}

	s.PushEvent(&GetServiceEvent{ServiceID: serviceId, Service: result.Elements[0].Service})

	return result.Elements[0].Service, nil
}

func (s *serviceHandler) ListServices(
	ctx context.Context,
	filter *entity.ServiceFilter,
	options *entity.ListOptions,
) (*entity.List[entity.ServiceResult], error) {
	return s.List(ctx, appErrors.CallerOp(), filter, options)
}

func (s *serviceHandler) CreateService(ctx context.Context, service *entity.Service) (*entity.Service, error) {
	op := appErrors.CallerOp()

	var err error

	service.BaseService.CreatedBy, err = common.GetCurrentUserId(ctx, s.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Service", "", err)
	}

	service.BaseService.UpdatedBy = service.BaseService.CreatedBy

	existing, err := s.ListServices(ctx, &entity.ServiceFilter{CCRN: []*string{&service.CCRN}}, entity.NewListOptions())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Service", "", err)
	}

	if len(existing.Elements) > 0 {
		return nil, appErrors.AlreadyExistsError(string(op), "Service", service.CCRN)
	}

	newService, err := s.DB().CreateService(service)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Service", "", err)
	}

	s.PushEvent(&CreateServiceEvent{Service: newService})

	return newService, nil
}

func (s *serviceHandler) UpdateService(ctx context.Context, service *entity.Service) (*entity.Service, error) {
	op := appErrors.CallerOp()
	id := strconv.FormatInt(service.Id, 10)

	var err error

	service.BaseService.UpdatedBy, err = common.GetCurrentUserId(ctx, s.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Service", id, err)
	}

	if err = s.DB().UpdateService(service); err != nil {
		return nil, appErrors.InternalError(string(op), "Service", id, err)
	}

	s.PushEvent(&UpdateServiceEvent{Service: service})

	return s.GetService(ctx, service.Id)
}

func (s *serviceHandler) DeleteService(ctx context.Context, id int64) error {
	return s.Delete(ctx, id)
}

func (s *serviceHandler) AddOwnerToService(ctx context.Context, serviceId, ownerId int64) (*entity.Service, error) {
	op := appErrors.CallerOp()

	if err := s.DB().AddOwnerToService(serviceId, ownerId); err != nil {
		return nil, appErrors.InternalError(string(op), "Service", fmt.Sprint(serviceId), err)
	}

	s.PushEvent(&AddOwnerToServiceEvent{ServiceID: serviceId, OwnerID: ownerId})

	return s.GetService(ctx, serviceId)
}

func (s *serviceHandler) RemoveOwnerFromService(ctx context.Context, serviceId, ownerId int64) (*entity.Service, error) {
	op := appErrors.CallerOp()

	if err := s.DB().RemoveOwnerFromService(serviceId, ownerId); err != nil {
		return nil, appErrors.InternalError(string(op), "Service", fmt.Sprint(serviceId), err)
	}

	s.PushEvent(&RemoveOwnerFromServiceEvent{ServiceID: serviceId, OwnerID: ownerId})

	return s.GetService(ctx, serviceId)
}

func (s *serviceHandler) AddIssueRepositoryToService(
	ctx context.Context,
	serviceId, issueRepositoryId, priority int64,
) (*entity.Service, error) {
	op := appErrors.CallerOp()

	if err := s.DB().AddIssueRepositoryToService(serviceId, issueRepositoryId, priority); err != nil {
		return nil, appErrors.InternalError(string(op), "Service", fmt.Sprint(serviceId), err)
	}

	s.PushEvent(&AddIssueRepositoryToServiceEvent{ServiceID: serviceId, RepositoryID: issueRepositoryId})

	return s.GetService(ctx, serviceId)
}

func (s *serviceHandler) RemoveIssueRepositoryFromService(
	ctx context.Context,
	serviceId, issueRepositoryId int64,
) (*entity.Service, error) {
	op := appErrors.CallerOp()

	if err := s.DB().RemoveIssueRepositoryFromService(serviceId, issueRepositoryId); err != nil {
		return nil, appErrors.InternalError(string(op), "Service", fmt.Sprint(serviceId), err)
	}

	s.PushEvent(&RemoveIssueRepositoryFromServiceEvent{ServiceID: serviceId, RepositoryID: issueRepositoryId})

	return s.GetService(ctx, serviceId)
}

func (s *serviceHandler) ListServiceCcrns(
	ctx context.Context,
	filter *entity.ServiceFilter,
	options *entity.ListOptions,
) ([]string, error) {
	op := appErrors.CallerOp()

	ccrns, err := cache.CallCached[[]string](
		s.Cache(),
		cache.NewCacheCallParams(common.DefaultCacheTTL, ctx, "GetServiceCcrns", s.DB().GetServiceCcrns, filter),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ServiceCcrns", "", err)
	}

	s.PushEvent(&ListServiceCcrnsEvent{Filter: filter, Options: options, Ccrns: ccrns})

	return ccrns, nil
}

func (s *serviceHandler) ListServiceDomains(
	ctx context.Context,
	filter *entity.ServiceFilter,
	options *entity.ListOptions,
) ([]string, error) {
	op := appErrors.CallerOp()

	domains, err := cache.CallCached[[]string](
		s.Cache(),
		cache.NewCacheCallParams(common.DefaultCacheTTL, ctx, "GetServiceDomains", s.DB().GetServiceDomains, filter),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ServiceDomains", "", err)
	}

	s.PushEvent(&ListServiceDomainsEvent{Filter: filter, Options: options, Domains: domains})

	return domains, nil
}

func (s *serviceHandler) ListServiceRegions(
	ctx context.Context,
	filter *entity.ServiceFilter,
	options *entity.ListOptions,
) ([]string, error) {
	op := appErrors.CallerOp()

	regions, err := cache.CallCached[[]string](
		s.Cache(),
		cache.NewCacheCallParams(common.DefaultCacheTTL, ctx, "GetServiceRegions", s.DB().GetServiceRegions, filter),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ServiceRegions", "", err)
	}

	s.PushEvent(&ListServiceRegionsEvent{Filter: filter, Options: options, Regions: regions})

	return regions, nil
}

func (s *serviceHandler) ListOwnersByServiceIDs(
	ctx context.Context,
	serviceIDs []int64,
) (map[int64][]entity.User, error) {
	op := appErrors.CallerOp()

	owners, err := cache.CallCached[map[int64][]entity.User](
		s.Cache(),
		cache.NewCacheCallParams(common.DefaultCacheTTL, ctx, "GetOwnersByServiceIDs", s.DB().GetOwnersByServiceIDs, serviceIDs),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Owners", "", err)
	}

	return owners, nil
}

func (s *serviceHandler) ListSupportGroupsByServiceIDs(
	ctx context.Context,
	serviceIDs []int64,
) (map[int64][]entity.SupportGroup, error) {
	op := appErrors.CallerOp()

	supportGroups, err := cache.CallCached[map[int64][]entity.SupportGroup](
		s.Cache(),
		cache.NewCacheCallParams(common.DefaultCacheTTL, ctx, "GetSupportGroupsByServiceIDs", s.DB().GetSupportGroupsByServiceIDs, serviceIDs),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "SupportGroups", "", err)
	}

	return supportGroups, nil
}

func (s *serviceHandler) ListIssueCountsByServiceIDs(
	ctx context.Context,
	serviceIDs []int64,
) (map[int64]entity.IssueSeverityCounts, error) {
	op := appErrors.CallerOp()

	issueCounts, err := cache.CallCached[map[int64]entity.IssueSeverityCounts](
		s.Cache(),
		cache.NewCacheCallParams(common.DefaultCacheTTL, ctx, "GetIssueCountsByServiceIDs", s.DB().GetIssueCountsByServiceIDs, serviceIDs),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "IssueCounts", "", err)
	}

	return issueCounts, nil
}
