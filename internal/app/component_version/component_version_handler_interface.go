// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_version

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type ComponentVersionHandler interface {
	ListComponentVersions(*entity.ComponentVersionFilter, *entity.ListOptions) (*entity.List[entity.ComponentVersionResult], error)
	CreateComponentVersion(*entity.ComponentVersion) (*entity.ComponentVersion, error)
	UpdateComponentVersion(*entity.ComponentVersion) (*entity.ComponentVersion, error)
	DeleteComponentVersion(int64) error
}
