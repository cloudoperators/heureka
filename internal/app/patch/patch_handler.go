// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package patch

import (
	"context"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
)

type patchHandler struct {
	common.BaseHandler[entity.PatchResult, *entity.PatchFilter]
}

func NewPatchHandler(handlerContext common.HandlerContext) PatchHandler {
	return &patchHandler{
		BaseHandler: common.NewBaseHandler(handlerContext, common.BaseConfig[entity.PatchResult, *entity.PatchFilter]{
			Op:           appErrors.Op("patchHandler"),
			Entity:       "Patches",
			CursorEntity: "PatchCursors",
			CountEntity:  "PatchCount",
			GetFn:        handlerContext.DB.GetPatches,
			CursorsFn:    handlerContext.DB.GetAllPatchCursors,
			CountFn:      handlerContext.DB.CountPatches,
			ListEventFn: func(f *entity.PatchFilter, o *entity.ListOptions, r *entity.List[entity.PatchResult]) any {
				return &ListPatchesEvent{Filter: f, Options: o, Patches: r}
			},
		}),
	}
}

func (ph *patchHandler) ListPatches(
	ctx context.Context,
	filter *entity.PatchFilter,
	options *entity.ListOptions,
) (*entity.List[entity.PatchResult], error) {
	return ph.List(ctx, appErrors.CallerOp(), filter, options)
}
