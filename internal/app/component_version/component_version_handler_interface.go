// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type ComponentVersionHandler interface {
	ListComponentVersions(*entity.ComponentVersionFilter, *entity.ListOptions) (*entity.List[entity.ComponentVersionResult], error)
	CreateComponentVersion(context.Context, *entity.ComponentVersion) (*entity.ComponentVersion, error)
	UpdateComponentVersion(context.Context, *entity.ComponentVersion) (*entity.ComponentVersion, error)
	DeleteComponentVersion(context.Context, int64) error
}
