// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import "github.com/cloudoperators/heureka/internal/entity"

type ComponentHandler interface {
	ListComponents(*entity.ComponentFilter, *entity.ListOptions) (*entity.List[entity.ComponentResult], error)
	CreateComponent(*entity.Component) (*entity.Component, error)
	UpdateComponent(*entity.Component) (*entity.Component, error)
	DeleteComponent(int64) error
	ListComponentCcrns(*entity.ComponentFilter, *entity.ListOptions) ([]string, error)
	GetComponentVulnerabilityCounts(*entity.ComponentFilter) ([]entity.IssueSeverityCounts, error)
}
