// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type ServiceHandler interface {
	GetService(ctx context.Context, serviceId int64) (*entity.Service, error)
	ListServices(ctx context.Context, filter *entity.ServiceFilter, options *entity.ListOptions) (*entity.List[entity.ServiceResult], error)
	CreateService(ctx context.Context, service *entity.Service) (*entity.Service, error)
	UpdateService(ctx context.Context, service *entity.Service) (*entity.Service, error)
	DeleteService(ctx context.Context, id int64) error
	AddOwnerToService(ctx context.Context, serviceId, ownerId int64) (*entity.Service, error)
	RemoveOwnerFromService(ctx context.Context, serviceId, ownerId int64) (*entity.Service, error)
	ListServiceCcrns(filter *entity.ServiceFilter, options *entity.ListOptions) ([]string, error)
	ListServiceDomains(filter *entity.ServiceFilter, options *entity.ListOptions) ([]string, error)
	ListServiceRegions(filter *entity.ServiceFilter, options *entity.ListOptions) ([]string, error)
	AddIssueRepositoryToService(context.Context, int64, int64, int64) (*entity.Service, error)
	RemoveIssueRepositoryFromService(context.Context, int64, int64) (*entity.Service, error)
}
