// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue

import (
	"context"
	"time"

	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/entity"
)

var (
	CacheTtlGetMaxSeverityByIssueIDs             = 12 * time.Hour
	CacheTtlGetEarliestRemediationByIssueIDs     = 12 * time.Hour
	CacheTtlGetSourceURLsByIssueIDs              = 12 * time.Hour
	CacheTtlGetServicesByIssueIDs                = 12 * time.Hour
	CacheTtlGetSupportGroupsByIssueIDs           = 12 * time.Hour
	CacheTtlGetVulnerabilityAggregatesByIssueIDs = 12 * time.Hour
)

func (is *issueHandler) GetMaxSeverityByIssueIDs(ctx context.Context, issueIDs []int64) (map[int64]string, error) {
	return cache.CallCached[map[int64]string](
		is.Cache(),
		cache.NewCacheCallParams(
			CacheTtlGetMaxSeverityByIssueIDs,
			ctx,
			"GetMaxSeverityByIssueIDs",
			is.DB().GetMaxSeverityByIssueIDs,
			issueIDs,
		),
	)
}

func (is *issueHandler) GetEarliestRemediationByIssueIDs(ctx context.Context, issueIDs []int64) (map[int64]time.Time, error) {
	return cache.CallCached[map[int64]time.Time](
		is.Cache(),
		cache.NewCacheCallParams(
			CacheTtlGetEarliestRemediationByIssueIDs,
			ctx,
			"GetEarliestRemediationByIssueIDs",
			is.DB().GetEarliestRemediationByIssueIDs,
			issueIDs,
		),
	)
}

func (is *issueHandler) GetSourceURLsByIssueIDs(ctx context.Context, issueIDs []int64) (map[int64]string, error) {
	return cache.CallCached[map[int64]string](
		is.Cache(),
		cache.NewCacheCallParams(
			CacheTtlGetSourceURLsByIssueIDs,
			ctx,
			"GetSourceURLsByIssueIDs",
			is.DB().GetSourceURLsByIssueIDs,
			issueIDs,
		),
	)
}

func (is *issueHandler) GetServicesByIssueIDs(ctx context.Context, issueIDs []int64) (map[int64][]entity.ServiceResult, error) {
	return cache.CallCached[map[int64][]entity.ServiceResult](
		is.Cache(),
		cache.NewCacheCallParams(
			CacheTtlGetServicesByIssueIDs,
			ctx,
			"GetServicesByIssueIDs",
			is.DB().GetServicesByIssueIDs,
			issueIDs,
		),
	)
}

func (is *issueHandler) GetSupportGroupsByIssueIDs(ctx context.Context, issueIDs []int64) (map[int64][]entity.SupportGroupResult, error) {
	return cache.CallCached[map[int64][]entity.SupportGroupResult](
		is.Cache(),
		cache.NewCacheCallParams(
			CacheTtlGetSupportGroupsByIssueIDs,
			ctx,
			"GetSupportGroupsByIssueIDs",
			is.DB().GetSupportGroupsByIssueIDs,
			issueIDs,
		),
	)
}

func (is *issueHandler) GetVulnerabilityAggregatesByIssueIDs(ctx context.Context, issueIDs []int64) (map[int64]entity.VulnerabilityAggregate, error) {
	return cache.CallCached[map[int64]entity.VulnerabilityAggregate](
		is.Cache(),
		cache.NewCacheCallParams(
			CacheTtlGetVulnerabilityAggregatesByIssueIDs,
			ctx,
			"GetVulnerabilityAggregatesByIssueIDs",
			is.DB().GetVulnerabilityAggregatesByIssueIDs,
			issueIDs,
		),
	)
}
