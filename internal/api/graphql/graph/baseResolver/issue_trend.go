// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"
	"time"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

func strSliceToPtrSlice(ss []string) []*string {
	out := make([]*string, len(ss))
	for i := range ss {
		s := ss[i]
		out[i] = &s
	}

	return out
}

func IssueTrendBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.IssueTrendFilter,
) (*model.IssueTrend, error) {
	logrus.WithField("filter", filter).Debug("Called IssueTrendBaseResolver")

	after, err := time.Parse(time.RFC3339, filter.After)
	if err != nil {
		return nil, NewResolverError("IssueTrendBaseResolver", "Bad Request - invalid 'after' date")
	}

	before, err := time.Parse(time.RFC3339, filter.Before)
	if err != nil {
		return nil, NewResolverError("IssueTrendBaseResolver", "Bad Request - invalid 'before' date")
	}

	granularity := entity.TrendGranularityDaily

	switch filter.Granularity {
	case model.TrendGranularityWeekly:
		granularity = entity.TrendGranularityWeekly
	case model.TrendGranularityMonthly:
		granularity = entity.TrendGranularityMonthly
	}

	entityFilter := entity.IssueTrendFilter{
		ServiceCCRN:      strSliceToPtrSlice(filter.ServiceCcrn),
		SupportGroupCCRN: strSliceToPtrSlice(filter.SupportGroupCcrn),
		After:            after,
		Before:           before,
		Granularity:      granularity,
	}

	trend, err := app.GetIssueTrend(ctx, &entityFilter)
	if err != nil {
		return nil, ToGraphQLError(err)
	}

	buckets := make([]*model.IssueTrendBucket, len(trend.Buckets))

	for i, b := range trend.Buckets {
		buckets[i] = &model.IssueTrendBucket{
			Date:       b.Date.Format(time.RFC3339),
			Critical:   int(b.Critical),
			High:       int(b.High),
			Medium:     int(b.Medium),
			Low:        int(b.Low),
			None:       int(b.None),
			Total:      int(b.Total),
			Remediated: int(b.Remediated),
		}
	}

	return &model.IssueTrend{Buckets: buckets}, nil
}
