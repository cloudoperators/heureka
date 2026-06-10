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

func (cs *componentHandler) GetVersionsByComponentIDs(ctx context.Context, componentIDs []int64, serviceCCRN []*string) (map[int64][]entity.ComponentVersionResult, error) {
	return cache.CallCached[map[int64][]entity.ComponentVersionResult](
		cs.cache,
		CacheTtlGetVersionsByComponentIDs,
		"GetVersionsByComponentIDs",
		cache.WrapContext2(ctx, cs.database.GetVersionsByComponentIDs),
		componentIDs,
		serviceCCRN,
	)
}

func (cs *componentHandler) GetIssueCountsByComponentIDs(ctx context.Context, componentIDs []int64, serviceCCRN []*string) (map[int64]entity.IssueSeverityCounts, error) {
	return cache.CallCached[map[int64]entity.IssueSeverityCounts](
		cs.cache,
		CacheTtlGetIssueCountsByComponentIDs,
		"GetIssueCountsByComponentIDs",
		cache.WrapContext2(ctx, cs.database.GetIssueCountsByComponentIDs),
		componentIDs,
		serviceCCRN,
	)
}

func (cs *componentHandler) GetVulnerabilitiesByComponentIDs(ctx context.Context, componentIDs []int64) (map[int64][]entity.VulnerabilityResult, error) {
	return cache.CallCached[map[int64][]entity.VulnerabilityResult](
		cs.cache,
		CacheTtlGetVulnerabilitiesByComponentIDs,
		"GetVulnerabilitiesByComponentIDs",
		cache.WrapContext1(ctx, cs.database.GetVulnerabilitiesByComponentIDs),
		componentIDs,
	)
}

func (cs *componentHandler) GetVulnerabilityCountsByComponentIDs(ctx context.Context, componentIDs []int64) (map[int64]int, error) {
	return cache.CallCached[map[int64]int](
		cs.cache,
		CacheTtlGetVulnerabilityCountsByComponentIDs,
		"GetVulnerabilityCountsByComponentIDs",
		cache.WrapContext1(ctx, cs.database.GetVulnerabilityCountsByComponentIDs),
		componentIDs,
	)
}
