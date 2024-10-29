// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

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
	item, err := baseResolver.SupportGroupCcrn(r.App, ctx, filter)

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
	item.FilterName = &baseResolver.FilterDisplayCcrn
	return item, err
}

func (r *Resolver) ComponentInstanceFilterValue() graph.ComponentInstanceFilterValueResolver {
	return &componentInstanceFilterValueResolver{r}
}

type componentInstanceFilterValueResolver struct{ *Resolver }
