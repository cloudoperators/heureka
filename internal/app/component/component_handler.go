// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"context"
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
)

type componentHandler struct {
	common.BaseHandler[entity.ComponentResult, *entity.ComponentFilter]
}

func NewComponentHandler(hc common.HandlerContext) ComponentHandler {
	return &componentHandler{
		BaseHandler: common.NewBaseHandler(hc, common.BaseConfig[entity.ComponentResult, *entity.ComponentFilter]{
			Op:              appErrors.Op("componentHandler"),
			Entity:          "Components",
			CursorEntity:    "ComponentCursors",
			CountEntity:     "ComponentCount",
			GetFn:           hc.DB.GetComponents,
			CursorsFn:       hc.DB.GetAllComponentCursors,
			CountFn:         hc.DB.CountComponents,
			Authz:           hc.Authz,
			AuthzObjectType: openfga.TypeComponent,
			AuthzApplyFn: func(f *entity.ComponentFilter, ids []*int64) {
				f.Id = common.CombineFilterWithAccessibleIds(f.Id, ids)
			},
			ListEventFn: func(f *entity.ComponentFilter, o *entity.ListOptions, r *entity.List[entity.ComponentResult]) any {
				return &ListComponentsEvent{Filter: f, Options: o, Components: r}
			},
			DeleteFn:      hc.DB.DeleteComponent,
			DeleteEventFn: func(id int64) any { return &DeleteComponentEvent{ComponentID: id} },
		}),
	}
}

func (ch *componentHandler) ListComponents(
	ctx context.Context,
	filter *entity.ComponentFilter,
	options *entity.ListOptions,
) (*entity.List[entity.ComponentResult], error) {
	return ch.List(ctx, appErrors.CallerOp(), filter, options)
}

func (ch *componentHandler) CreateComponent(
	ctx context.Context,
	component *entity.Component,
) (*entity.Component, error) {
	op := appErrors.CallerOp()

	f := &entity.ComponentFilter{
		CCRN: []*string{&component.CCRN},
	}

	var err error

	component.CreatedBy, err = common.GetCurrentUserId(ctx, ch.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Component", "", err)
	}

	component.UpdatedBy = component.CreatedBy

	lo := entity.NewListOptions()

	components, err := ch.ListComponents(ctx, f, lo)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Component", "", err)
	}

	if len(components.Elements) > 0 {
		return nil, appErrors.AlreadyExistsError(string(op), "Component", component.CCRN)
	}

	newComponent, err := ch.DB().CreateComponent(component)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Component", "", err)
	}

	ch.PushEvent(&CreateComponentEvent{Component: newComponent})

	return newComponent, nil
}

func (ch *componentHandler) UpdateComponent(
	ctx context.Context,
	component *entity.Component,
) (*entity.Component, error) {
	op := appErrors.CallerOp()

	var err error

	component.UpdatedBy, err = common.GetCurrentUserId(ctx, ch.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Component", "", err)
	}

	err = ch.DB().UpdateComponent(component)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Component", "", err)
	}

	lo := entity.NewListOptions()

	componentResult, err := ch.ListComponents(
		ctx,
		&entity.ComponentFilter{Id: []*int64{&component.Id}},
		lo,
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Component", "", err)
	}

	if len(componentResult.Elements) != 1 {
		return nil, appErrors.InternalError(string(op), "Component", "", fmt.Errorf("unexpected result count: %d", len(componentResult.Elements)))
	}

	ch.PushEvent(&UpdateComponentEvent{Component: component})

	return componentResult.Elements[0].Component, nil
}

func (ch *componentHandler) DeleteComponent(ctx context.Context, id int64) error {
	return ch.Delete(ctx, id)
}

func (ch *componentHandler) ListComponentCcrns(
	ctx context.Context,
	filter *entity.ComponentFilter,
	options *entity.ListOptions,
) ([]string, error) {
	op := appErrors.CallerOp()

	componentCcrns, err := cache.CallCached[[]string](
		ch.Cache(),
		cache.NewCacheCallParams(
			common.DefaultCacheTTL,
			ctx,
			"GetComponentCcrns",
			ch.DB().GetComponentCcrns,
			filter,
		),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentCcrns", "", err)
	}

	ch.PushEvent(&ListComponentCcrnsEvent{Filter: filter, Options: options, CCRNs: componentCcrns})

	return componentCcrns, nil
}

func (ch *componentHandler) GetComponentVulnerabilityCounts(
	ctx context.Context,
	filter *entity.ComponentFilter,
) ([]entity.IssueSeverityCounts, error) {
	op := appErrors.CallerOp()

	counts, err := ch.DB().CountComponentVulnerabilities(ctx, filter)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentVulnerabilityCounts", "", err)
	}

	ch.PushEvent(&GetComponentIssueSeverityCountsEvent{Filter: filter, Counts: counts})

	return counts, nil
}
