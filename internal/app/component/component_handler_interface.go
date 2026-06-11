// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type ComponentHandler interface {
	ListComponents(
		context.Context,
		*entity.ComponentFilter,
		*entity.ListOptions,
	) (*entity.List[entity.ComponentResult], error)
	CreateComponent(context.Context, *entity.Component) (*entity.Component, error)
	UpdateComponent(context.Context, *entity.Component) (*entity.Component, error)
	DeleteComponent(context.Context, int64) error
	ListComponentCcrns(context.Context, *entity.ComponentFilter, *entity.ListOptions) ([]string, error)
	GetComponentVulnerabilityCounts(context.Context, *entity.ComponentFilter) ([]entity.IssueSeverityCounts, error)
	GetVersionsByComponentIDs(ctx context.Context, componentIDs []int64, serviceCCRN []*string) (map[int64][]entity.ComponentVersionResult, error)
	GetIssueCountsByComponentIDs(ctx context.Context, componentIDs []int64, serviceCCRN []*string) (map[int64]entity.IssueSeverityCounts, error)
	GetVulnerabilitiesByComponentIDs(ctx context.Context, componentIDs []int64) (map[int64][]entity.VulnerabilityResult, error)
	GetVulnerabilityCountsByComponentIDs(ctx context.Context, componentIDs []int64) (map[int64]int, error)
}
