// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
)

type componentVersionHandler struct {
	common.BaseHandler[entity.ComponentVersionResult, *entity.ComponentVersionFilter]
}

func NewComponentVersionHandler(hc common.HandlerContext) ComponentVersionHandler {
	return &componentVersionHandler{
		BaseHandler: common.NewBaseHandler(hc, common.BaseConfig[entity.ComponentVersionResult, *entity.ComponentVersionFilter]{
			Op:              appErrors.Op("componentVersionHandler"),
			Entity:          "ComponentVersions",
			CursorEntity:    "ComponentVersionCursors",
			CountEntity:     "ComponentVersionCount",
			GetFn:           hc.DB.GetComponentVersions,
			CursorsFn:       hc.DB.GetAllComponentVersionCursors,
			CountFn:         hc.DB.CountComponentVersions,
			Authz:           hc.Authz,
			AuthzObjectType: openfga.TypeComponent,
			AuthzApplyFn: func(f *entity.ComponentVersionFilter, ids []*int64) {
				f.ComponentId = common.CombineFilterWithAccessibleIds(f.ComponentId, ids)
			},
			ListEventFn: func(f *entity.ComponentVersionFilter, o *entity.ListOptions, r *entity.List[entity.ComponentVersionResult]) any {
				return &ListComponentVersionsEvent{Filter: f, Options: o, ComponentVersions: r}
			},
			DeleteFn:      hc.DB.DeleteComponentVersion,
			DeleteEventFn: func(id int64) any { return &DeleteComponentVersionEvent{ComponentVersionID: id} },
		}),
	}
}

func (cv *componentVersionHandler) ListComponentVersions(
	ctx context.Context,
	filter *entity.ComponentVersionFilter,
	options *entity.ListOptions,
) (*entity.List[entity.ComponentVersionResult], error) {
	return cv.List(ctx, appErrors.CallerOp(), filter, options)
}

func (cv *componentVersionHandler) CreateComponentVersion(
	ctx context.Context,
	componentVersion *entity.ComponentVersion,
) (*entity.ComponentVersion, error) {
	op := appErrors.CallerOp()

	var err error

	componentVersion.CreatedBy, err = common.GetCurrentUserId(ctx, cv.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentVersion", "", err)
	}

	componentVersion.UpdatedBy = componentVersion.CreatedBy

	// Derive and persist is_outdated from the running tag vs the latest registry
	// version. At create both are whatever the caller supplied.
	componentVersion.IsOutdated = entity.IsVersionOutdated(componentVersion.Tag, componentVersion.LatestVersion)

	newComponent, err := cv.DB().CreateComponentVersion(componentVersion)
	if err != nil {
		duplicateEntryError := &database.DuplicateEntryDatabaseError{}
		if errors.As(err, &duplicateEntryError) {
			return nil, appErrors.AlreadyExistsError(string(op), "ComponentVersion", "")
		}

		return nil, appErrors.InternalError(string(op), "ComponentVersion", "", err)
	}

	cv.PushEvent(&CreateComponentVersionEvent{ComponentVersion: newComponent})

	return newComponent, nil
}

func (cv *componentVersionHandler) UpdateComponentVersion(
	ctx context.Context,
	componentVersion *entity.ComponentVersion,
) (*entity.ComponentVersion, error) {
	op := appErrors.CallerOp()

	var err error

	componentVersion.UpdatedBy, err = common.GetCurrentUserId(ctx, cv.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentVersion", "", err)
	}

	// Recompute is_outdated from the MERGED state: a scanner update is sparse
	// (k8s-assets sets tag, keppel sets latestVersion in separate runs), so load
	// the stored row and fall back to its values for whichever field this update
	// omits before deriving the flag.
	existing, err := cv.ListComponentVersions(
		ctx,
		&entity.ComponentVersionFilter{Id: []*int64{&componentVersion.Id}},
		entity.NewListOptions(),
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentVersion", "", err)
	}

	effectiveTag := componentVersion.Tag

	effectiveLatest := componentVersion.LatestVersion

	if len(existing.Elements) == 1 {
		stored := existing.Elements[0].ComponentVersion
		if effectiveTag == "" {
			effectiveTag = stored.Tag
		}

		if effectiveLatest == nil {
			effectiveLatest = stored.LatestVersion
		}
	}

	componentVersion.IsOutdated = entity.IsVersionOutdated(effectiveTag, effectiveLatest)

	err = cv.DB().UpdateComponentVersion(componentVersion)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentVersion", "", err)
	}

	lo := entity.NewListOptions()

	componentVersionResult, err := cv.ListComponentVersions(
		ctx,
		&entity.ComponentVersionFilter{Id: []*int64{&componentVersion.Id}},
		lo,
	)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentVersion", "", err)
	}

	if len(componentVersionResult.Elements) != 1 {
		return nil, appErrors.InternalError(string(op), "ComponentVersion", "", fmt.Errorf("unexpected result count: %d", len(componentVersionResult.Elements)))
	}

	cv.PushEvent(&UpdateComponentVersionEvent{ComponentVersion: componentVersion})

	return componentVersionResult.Elements[0].ComponentVersion, nil
}

func (cv *componentVersionHandler) DeleteComponentVersion(ctx context.Context, id int64) error {
	return cv.Delete(ctx, id)
}
