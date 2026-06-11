// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

var patchObject = DbObject[*entity.Patch, *entity.PatchFilter, entity.PatchResult]{
	Prefix:       "patch",
	TableName:    "Patch",
	TableKey:     "P",
	DefaultOrder: entity.Order{By: entity.PatchId, Direction: entity.OrderDirectionAsc},
	FilterProperties: []*FilterProperty[*entity.PatchFilter]{
		NewFilterProperty("P.patch_id = ?", func(filter *entity.PatchFilter) any { return filter.Id }),
		NewFilterProperty("P.patch_service_id = ?", func(filter *entity.PatchFilter) any { return filter.ServiceId }),
		NewFilterProperty("P.patch_service_name = ?", func(filter *entity.PatchFilter) any { return filter.ServiceName }),
		NewFilterProperty("P.patch_component_version_id = ?", func(filter *entity.PatchFilter) any { return filter.ComponentVersionId }),
		NewFilterProperty("P.patch_component_version_name = ?", func(filter *entity.PatchFilter) any { return filter.ComponentVersionName }),
		NewStateFilterProperty("P.patch", func(filter *entity.PatchFilter) any { return filter.State }),
	},
	GetItemAppender: func(l []entity.PatchResult, e RowComposite, order []entity.Order) []entity.PatchResult {
		p := e.AsPatch()
		cursor, _ := EncodeCursor(WithPatch(order, p))

		pr := entity.PatchResult{
			WithCursor: entity.WithCursor{
				Value: cursor,
			},
			Patch: &p,
		}

		return append(l, pr)
	},
	GetAllCursorItemAppender: func(l []string, e RowComposite, order []entity.Order) []string {
		r := e.AsPatch()

		cursor, _ := EncodeCursor(WithPatch(order, r))

		return append(l, cursor)
	},
}

func (s *SqlDatabase) GetPatches(
	ctx context.Context,
	filter *entity.PatchFilter,
	order []entity.Order,
) ([]entity.PatchResult, error) {
	return patchObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountPatches(ctx context.Context, filter *entity.PatchFilter) (int64, error) {
	return patchObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) GetAllPatchCursors(
	ctx context.Context,
	filter *entity.PatchFilter,
	order []entity.Order,
) ([]string, error) {
	return patchObject.GetAllCursors(ctx, s.db, filter, order)
}
