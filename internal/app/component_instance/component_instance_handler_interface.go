// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import "github.com/cloudoperators/heureka/internal/entity"

type ComponentInstanceHandler interface {
	ListComponentInstances(*entity.ComponentInstanceFilter, *entity.ListOptions) (*entity.List[entity.ComponentInstanceResult], error)
	CreateComponentInstance(*entity.ComponentInstance, string) (*entity.ComponentInstance, error)
	UpdateComponentInstance(*entity.ComponentInstance) (*entity.ComponentInstance, error)
	DeleteComponentInstance(int64) error
	ListCcrns(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListRegions(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListClusters(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListNamespaces(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListDomains(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListProjects(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListPods(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListContainers(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListTypes(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListContexts(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
}
