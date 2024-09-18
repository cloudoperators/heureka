// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import "github.com/cloudoperators/heureka/internal/entity"

type ComponentInstanceHandler interface {
	ListComponentInstances(*entity.ComponentInstanceFilter, *entity.ListOptions) (*entity.List[entity.ComponentInstanceResult], error)
	CreateComponentInstance(*entity.ComponentInstance) (*entity.ComponentInstance, error)
	UpdateComponentInstance(*entity.ComponentInstance) (*entity.ComponentInstance, error)
	DeleteComponentInstance(int64) error
}
