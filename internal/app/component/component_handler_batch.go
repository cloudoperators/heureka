// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"context"
	"time"

	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/entity"
)

var (
	CacheTtlGetVersionsByComponentIDs            = 12 * time.Hour
	CacheTtlGetIssueCountsByComponentIDs         = 12 * time.Hour
	CacheTtlGetVulnerabilitiesByComponentIDs     = 12 * time.Hour
	CacheTtlGetVulnerabilityCountsByComponentIDs = 12 * time.Hour
)

func (ch *componentHandler) GetVersionsByComponentIDs(ctx context.Context, componentIDs []int64, serviceCCRN []*string) (map[int64][]entity.ComponentVersionResult, error) {
	return cache.CallCached[map[int64][]entity.ComponentVersionResult](
		ch.Cache(),
		cache.NewCacheCallParams(
			CacheTtlGetVersionsByComponentIDs,
			ctx,
			"GetVersionsByComponentIDs",
			ch.DB().GetVersionsByComponentIDs,
			componentIDs,
			serviceCCRN,
		),
	)
}

func (ch *componentHandler) GetIssueCountsByComponentIDs(ctx context.Context, componentIDs []int64, serviceCCRN []*string) (map[int64]entity.IssueSeverityCounts, error) {
	return cache.CallCached[map[int64]entity.IssueSeverityCounts](
		ch.Cache(),
		cache.NewCacheCallParams(
			CacheTtlGetIssueCountsByComponentIDs,
			ctx,
			"GetIssueCountsByComponentIDs",
			ch.DB().GetIssueCountsByComponentIDs,
			componentIDs,
			serviceCCRN,
		),
	)
}

func (ch *componentHandler) GetVulnerabilitiesByComponentIDs(ctx context.Context, componentIDs []int64) (map[int64][]entity.VulnerabilityResult, error) {
	return cache.CallCached[map[int64][]entity.VulnerabilityResult](
		ch.Cache(),
		cache.NewCacheCallParams(
			CacheTtlGetVulnerabilitiesByComponentIDs,
			ctx,
			"GetVulnerabilitiesByComponentIDs",
			ch.DB().GetVulnerabilitiesByComponentIDs,
			componentIDs,
		),
	)
}

func (ch *componentHandler) GetVulnerabilityCountsByComponentIDs(ctx context.Context, componentIDs []int64) (map[int64]int, error) {
	return cache.CallCached[map[int64]int](
		ch.Cache(),
		cache.NewCacheCallParams(
			CacheTtlGetVulnerabilityCountsByComponentIDs,
			ctx,
			"GetVulnerabilityCountsByComponentIDs",
			ch.DB().GetVulnerabilityCountsByComponentIDs,
			componentIDs,
		),
	)
}
