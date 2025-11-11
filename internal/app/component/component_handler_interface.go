// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type ComponentHandler interface {
	ListComponents(*entity.ComponentFilter, *entity.ListOptions) (*entity.List[entity.ComponentResult], error)
	CreateComponent(context.Context, *entity.Component) (*entity.Component, error)
	UpdateComponent(context.Context, *entity.Component) (*entity.Component, error)
	DeleteComponent(context.Context, int64) error
	ListComponentCcrns(*entity.ComponentFilter, *entity.ListOptions) ([]string, error)
	GetComponentVulnerabilityCounts(*entity.ComponentFilter) ([]entity.IssueSeverityCounts, error)
}
