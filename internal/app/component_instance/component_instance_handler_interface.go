// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type ComponentInstanceHandler interface {
	ListComponentInstances(
		context.Context,
		*entity.ComponentInstanceFilter,
		*entity.ListOptions,
	) (*entity.List[entity.ComponentInstanceResult], error)
	CreateComponentInstance(
		context.Context,
		*entity.ComponentInstance,
		*string,
	) (*entity.ComponentInstance, error)
	UpdateComponentInstance(
		context.Context,
		*entity.ComponentInstance,
		*string,
	) (*entity.ComponentInstance, error)
	DeleteComponentInstance(context.Context, int64) error
	ListCcrns(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListRegions(
		ctx context.Context,
		filter *entity.ComponentInstanceFilter,
		options *entity.ListOptions,
	) ([]string, error)
	ListClusters(
		ctx context.Context,
		filter *entity.ComponentInstanceFilter,
		options *entity.ListOptions,
	) ([]string, error)
	ListNamespaces(
		ctx context.Context,
		filter *entity.ComponentInstanceFilter,
		options *entity.ListOptions,
	) ([]string, error)
	ListDomains(
		ctx context.Context,
		filter *entity.ComponentInstanceFilter,
		options *entity.ListOptions,
	) ([]string, error)
	ListProjects(
		ctx context.Context,
		filter *entity.ComponentInstanceFilter,
		options *entity.ListOptions,
	) ([]string, error)
	ListPods(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListContainers(
		ctx context.Context,
		filter *entity.ComponentInstanceFilter,
		options *entity.ListOptions,
	) ([]string, error)
	ListTypes(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error)
	ListParents(
		ctx context.Context,
		filter *entity.ComponentInstanceFilter,
		options *entity.ListOptions,
	) ([]string, error)
	ListContexts(
		ctx context.Context,
		filter *entity.ComponentInstanceFilter,
		options *entity.ListOptions,
	) ([]string, error)
}
