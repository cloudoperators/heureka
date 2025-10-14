// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package service

import "github.com/cloudoperators/heureka/internal/entity"

type ServiceHandler interface {
	GetService(serviceId int64) (*entity.Service, error)
	ListServices(filter *entity.ServiceFilter, options *entity.ListOptions) (*entity.List[entity.ServiceResult], error)
	CreateService(service *entity.Service) (*entity.Service, error)
	UpdateService(service *entity.Service) (*entity.Service, error)
	DeleteService(id int64) error
	AddOwnerToService(serviceId, ownerId int64) (*entity.Service, error)
	RemoveOwnerFromService(serviceId, ownerId int64) (*entity.Service, error)
	ListServiceCcrns(filter *entity.ServiceFilter, options *entity.ListOptions) ([]string, error)
	ListServiceDomains(filter *entity.ServiceFilter, options *entity.ListOptions) ([]string, error)
	ListServiceRegions(filter *entity.ServiceFilter, options *entity.ListOptions) ([]string, error)
	AddIssueRepositoryToService(int64, int64, int64) (*entity.Service, error)
	RemoveIssueRepositoryFromService(int64, int64) (*entity.Service, error)
}
