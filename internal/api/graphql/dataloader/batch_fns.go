// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package dataloader

import (
	"context"
	"math"

	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
)

func newSeverityBatchFn(a app.Heureka) func(ctx context.Context, issueIDs []int64) ([]*entity.IssueMatch, []error) {
	return func(ctx context.Context, issueIDs []int64) ([]*entity.IssueMatch, []error) {
		keys := lo.Map(issueIDs, func(id int64, _ int) *int64 { v := id; return &v })

		first := math.MaxInt64
		f := &entity.IssueMatchFilter{
			Paginated: entity.Paginated{First: &first},
			IssueId:   keys,
		}
		opt := entity.NewListOptions()
		opt.Order = append(opt.Order, entity.Order{By: entity.IssueMatchRating, Direction: entity.OrderDirectionDesc})

		results, err := a.ListIssueMatches(ctx, f, opt)
		if err != nil {
			errs := make([]error, len(issueIDs))
			for i := range errs {
				errs[i] = err
			}

			return nil, errs
		}

		byIssueID := make(map[int64]*entity.IssueMatch, len(results.Elements))
		for i := range results.Elements {
			im := results.Elements[i].IssueMatch
			if _, exists := byIssueID[im.IssueId]; !exists {
				byIssueID[im.IssueId] = im
			}
		}

		out := make([]*entity.IssueMatch, len(issueIDs))

		errs := make([]error, len(issueIDs))
		for i, id := range issueIDs {
			out[i] = byIssueID[id]
		}

		return out, errs
	}
}

func newEarliestRemediationBatchFn(a app.Heureka) func(ctx context.Context, issueIDs []int64) ([]*entity.IssueMatch, []error) {
	return func(ctx context.Context, issueIDs []int64) ([]*entity.IssueMatch, []error) {
		keys := lo.Map(issueIDs, func(id int64, _ int) *int64 { v := id; return &v })

		first := math.MaxInt64
		f := &entity.IssueMatchFilter{
			Paginated: entity.Paginated{First: &first},
			IssueId:   keys,
		}
		opt := entity.NewListOptions()
		opt.Order = append(opt.Order, entity.Order{By: entity.IssueMatchTargetRemediationDate, Direction: entity.OrderDirectionAsc})

		results, err := a.ListIssueMatches(ctx, f, opt)
		if err != nil {
			errs := make([]error, len(issueIDs))
			for i := range errs {
				errs[i] = err
			}

			return nil, errs
		}

		byIssueID := make(map[int64]*entity.IssueMatch, len(results.Elements))
		for i := range results.Elements {
			im := results.Elements[i].IssueMatch
			if _, exists := byIssueID[im.IssueId]; !exists {
				byIssueID[im.IssueId] = im
			}
		}

		out := make([]*entity.IssueMatch, len(issueIDs))

		errs := make([]error, len(issueIDs))
		for i, id := range issueIDs {
			out[i] = byIssueID[id]
		}

		return out, errs
	}
}

func newSourceURLBatchFn(a app.Heureka) func(ctx context.Context, issueIDs []int64) ([]*entity.IssueVariant, []error) {
	return func(ctx context.Context, issueIDs []int64) ([]*entity.IssueVariant, []error) {
		keys := lo.Map(issueIDs, func(id int64, _ int) *int64 { v := id; return &v })

		first := math.MaxInt64
		f := &entity.IssueVariantFilter{
			Paginated: entity.Paginated{First: &first},
			IssueId:   keys,
		}
		opt := entity.NewListOptions()

		results, err := a.ListIssueVariants(ctx, f, opt)
		if err != nil {
			errs := make([]error, len(issueIDs))
			for i := range errs {
				errs[i] = err
			}

			return nil, errs
		}

		byIssueID := make(map[int64]*entity.IssueVariant, len(results.Elements))
		for i := range results.Elements {
			iv := results.Elements[i].IssueVariant
			if _, exists := byIssueID[iv.IssueId]; !exists {
				byIssueID[iv.IssueId] = iv
			}
		}

		out := make([]*entity.IssueVariant, len(issueIDs))

		errs := make([]error, len(issueIDs))
		for i, id := range issueIDs {
			out[i] = byIssueID[id]
		}

		return out, errs
	}
}

func newVulnCountsBatchFn(a app.Heureka) func(ctx context.Context, componentIDs []int64) ([]*entity.IssueSeverityCounts, []error) {
	return func(ctx context.Context, componentIDs []int64) ([]*entity.IssueSeverityCounts, []error) {
		keys := lo.Map(componentIDs, func(id int64, _ int) *int64 { v := id; return &v })

		f := &entity.ComponentFilter{
			Id: keys,
		}

		results, err := a.GetComponentVulnerabilityCounts(ctx, f)
		if err != nil {
			errs := make([]error, len(componentIDs))
			for i := range errs {
				errs[i] = err
			}

			return nil, errs
		}

		byComponentID := make(map[int64]*entity.IssueSeverityCounts, len(results))
		for i := range results {
			c := results[i]
			byComponentID[c.ComponentId] = &c
		}

		out := make([]*entity.IssueSeverityCounts, len(componentIDs))

		errs := make([]error, len(componentIDs))
		for i, id := range componentIDs {
			out[i] = byComponentID[id]
		}

		return out, errs
	}
}

func newComponentVersionByIDBatchFn(a app.Heureka) func(ctx context.Context, ids []int64) ([]*entity.ComponentVersion, []error) {
	return func(ctx context.Context, ids []int64) ([]*entity.ComponentVersion, []error) {
		keys := lo.Map(ids, func(id int64, _ int) *int64 { v := id; return &v })
		first := len(ids)

		results, err := a.ListComponentVersions(ctx, &entity.ComponentVersionFilter{Paginated: entity.Paginated{First: &first}, Id: keys}, entity.NewListOptions())
		if err != nil {
			errs := make([]error, len(ids))
			for i := range errs {
				errs[i] = err
			}

			return nil, errs
		}

		byID := make(map[int64]*entity.ComponentVersion, len(results.Elements))
		for i := range results.Elements {
			cv := results.Elements[i].ComponentVersion
			byID[cv.Id] = cv
		}

		out := make([]*entity.ComponentVersion, len(ids))
		for i, id := range ids {
			out[i] = byID[id]
		}

		return out, make([]error, len(ids))
	}
}

func newServiceByIDBatchFn(a app.Heureka) func(ctx context.Context, ids []int64) ([]*entity.Service, []error) {
	return func(ctx context.Context, ids []int64) ([]*entity.Service, []error) {
		keys := lo.Map(ids, func(id int64, _ int) *int64 { v := id; return &v })
		first := len(ids)

		results, err := a.ListServices(ctx, &entity.ServiceFilter{Paginated: entity.Paginated{First: &first}, Id: keys}, entity.NewListOptions())
		if err != nil {
			errs := make([]error, len(ids))
			for i := range errs {
				errs[i] = err
			}

			return nil, errs
		}

		byID := make(map[int64]*entity.Service, len(results.Elements))
		for i := range results.Elements {
			s := results.Elements[i].Service
			byID[s.Id] = s
		}

		out := make([]*entity.Service, len(ids))
		for i, id := range ids {
			out[i] = byID[id]
		}

		return out, make([]error, len(ids))
	}
}

func newComponentInstanceByIDBatchFn(a app.Heureka) func(ctx context.Context, ids []int64) ([]*entity.ComponentInstance, []error) {
	return func(ctx context.Context, ids []int64) ([]*entity.ComponentInstance, []error) {
		keys := lo.Map(ids, func(id int64, _ int) *int64 { v := id; return &v })
		first := len(ids)

		results, err := a.ListComponentInstances(ctx, &entity.ComponentInstanceFilter{Paginated: entity.Paginated{First: &first}, Id: keys}, entity.NewListOptions())
		if err != nil {
			errs := make([]error, len(ids))
			for i := range errs {
				errs[i] = err
			}

			return nil, errs
		}

		byID := make(map[int64]*entity.ComponentInstance, len(results.Elements))
		for i := range results.Elements {
			ci := results.Elements[i].ComponentInstance
			byID[ci.Id] = ci
		}

		out := make([]*entity.ComponentInstance, len(ids))
		for i, id := range ids {
			out[i] = byID[id]
		}

		return out, make([]error, len(ids))
	}
}

func newIssueByIDBatchFn(a app.Heureka) func(ctx context.Context, ids []int64) ([]*entity.Issue, []error) {
	return func(ctx context.Context, ids []int64) ([]*entity.Issue, []error) {
		keys := lo.Map(ids, func(id int64, _ int) *int64 { v := id; return &v })
		first := len(ids)

		results, err := a.ListIssues(ctx, &entity.IssueFilter{Paginated: entity.Paginated{First: &first}, Id: keys}, &entity.IssueListOptions{})
		if err != nil {
			errs := make([]error, len(ids))
			for i := range errs {
				errs[i] = err
			}

			return nil, errs
		}

		byID := make(map[int64]*entity.Issue, len(results.Elements))
		for i := range results.Elements {
			iss := results.Elements[i].Issue
			byID[iss.Id] = iss
		}

		out := make([]*entity.Issue, len(ids))
		for i, id := range ids {
			out[i] = byID[id]
		}

		return out, make([]error, len(ids))
	}
}

func newComponentByIDBatchFn(a app.Heureka) func(ctx context.Context, ids []int64) ([]*entity.Component, []error) {
	return func(ctx context.Context, ids []int64) ([]*entity.Component, []error) {
		keys := lo.Map(ids, func(id int64, _ int) *int64 { v := id; return &v })
		first := len(ids)

		results, err := a.ListComponents(ctx, &entity.ComponentFilter{Paginated: entity.Paginated{First: &first}, Id: keys}, entity.NewListOptions())
		if err != nil {
			errs := make([]error, len(ids))
			for i := range errs {
				errs[i] = err
			}

			return nil, errs
		}

		byID := make(map[int64]*entity.Component, len(results.Elements))
		for i := range results.Elements {
			c := results.Elements[i].Component
			byID[c.Id] = c
		}

		out := make([]*entity.Component, len(ids))
		for i, id := range ids {
			out[i] = byID[id]
		}

		return out, make([]error, len(ids))
	}
}

func newIssueRepositoryByIDBatchFn(a app.Heureka) func(ctx context.Context, ids []int64) ([]*entity.IssueRepository, []error) {
	return func(ctx context.Context, ids []int64) ([]*entity.IssueRepository, []error) {
		keys := lo.Map(ids, func(id int64, _ int) *int64 { v := id; return &v })
		first := len(ids)

		results, err := a.ListIssueRepositories(ctx, &entity.IssueRepositoryFilter{Paginated: entity.Paginated{First: &first}, Id: keys}, entity.NewListOptions())
		if err != nil {
			errs := make([]error, len(ids))
			for i := range errs {
				errs[i] = err
			}

			return nil, errs
		}

		byID := make(map[int64]*entity.IssueRepository, len(results.Elements))
		for i := range results.Elements {
			ir := results.Elements[i].IssueRepository
			byID[ir.Id] = ir
		}

		out := make([]*entity.IssueRepository, len(ids))
		for i, id := range ids {
			out[i] = byID[id]
		}

		return out, make([]error, len(ids))
	}
}
