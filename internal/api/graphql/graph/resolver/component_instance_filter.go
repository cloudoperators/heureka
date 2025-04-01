package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/baseResolver"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
)

// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

func (r *componentInstanceFilterValueResolver) ServiceCcrn(ctx context.Context, obj *model.ComponentInstanceFilterValue, filter *model.ServiceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.ServiceCcrnBaseResolver(r.App, ctx, filter)

	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ServiceFilterServiceCcrn
	return item, err
}

func (r *componentInstanceFilterValueResolver) SupportGroupCcrn(ctx context.Context, obj *model.ComponentInstanceFilterValue, filter *model.SupportGroupFilter) (*model.FilterItem, error) {
	item, err := baseResolver.SupportGroupCcrnBaseResolver(r.App, ctx, filter)

	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ServiceFilterSupportGroupCcrn
	return item, err
}

func (r *componentInstanceFilterValueResolver) Ccrn(ctx context.Context, obj *model.ComponentInstanceFilterValue, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.CcrnBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ComponentInstanceFilterComponentCcrn
	return item, err
}

func (r *componentInstanceFilterValueResolver) Region(ctx context.Context, obj *model.ComponentInstanceFilterValue, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.RegionBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ComponentInstanceFilterRegion
	return item, err
}

func (r *componentInstanceFilterValueResolver) Cluster(ctx context.Context, obj *model.ComponentInstanceFilterValue, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.ClusterBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ComponentInstanceFilterCluster
	return item, err
}

func (r *componentInstanceFilterValueResolver) Namespace(ctx context.Context, obj *model.ComponentInstanceFilterValue, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.NamespaceBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ComponentInstanceFilterNamespace
	return item, err
}

func (r *componentInstanceFilterValueResolver) Domain(ctx context.Context, obj *model.ComponentInstanceFilterValue, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.DomainBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ComponentInstanceFilterDomain
	return item, err
}

func (r *componentInstanceFilterValueResolver) Project(ctx context.Context, obj *model.ComponentInstanceFilterValue, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.ProjectBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ComponentInstanceFilterProject
	return item, err
}

func (r *componentInstanceFilterValueResolver) Pod(ctx context.Context, obj *model.ComponentInstanceFilterValue, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.PodBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ComponentInstanceFilterPod
	return item, err
}

func (r *componentInstanceFilterValueResolver) Container(ctx context.Context, obj *model.ComponentInstanceFilterValue, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	item, err := baseResolver.ContainerBaseResolver(r.App, ctx, filter)
	if err != nil {
		return nil, err
	}
	item.FilterName = &baseResolver.ComponentInstanceFilterContainer
	return item, err
}

func (r *Resolver) ComponentInstanceFilterValue() graph.ComponentInstanceFilterValueResolver {
	return &componentInstanceFilterValueResolver{r}
}

type componentInstanceFilterValueResolver struct{ *Resolver }
