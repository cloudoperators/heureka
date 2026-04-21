// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package patch

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type PatchHandler interface {
	ListPatches(context.Context, *entity.PatchFilter, *entity.ListOptions) (*entity.List[entity.PatchResult], error)
}
