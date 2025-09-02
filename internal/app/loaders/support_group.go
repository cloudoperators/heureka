// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package loaders

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/vikstrous/dataloadgen"
	"log"
	"time"
)

var listSupportGroupsBatchCallback listSupportGroupsBatchFunc = func(f *entity.SupportGroupFilter, o *entity.ListOptions) (listSupportGroupsBatchResult, error) {
	panic("listSupportGroupsBatchCallback is not set")
}

type listSupportGroupsParams struct {
	Filter  *entity.SupportGroupFilter
	Options *entity.ListOptions
}
type listSupportGroupsResult = *entity.List[entity.SupportGroupResult]
type listSupportGroupsBatchResult = *entity.List[entity.SupportGroupBatchResult]
type listSupportGroupsFunc = func(f *entity.SupportGroupFilter, o *entity.ListOptions) (listSupportGroupsResult, error)
type listSupportGroupsBatchFunc = func(f *entity.SupportGroupFilter, o *entity.ListOptions) (listSupportGroupsBatchResult, error)
type listSupportGroupsLoader = dataloadgen.Loader[listSupportGroupsParams, listSupportGroupsResult]

func newListSupportGroupsLoader() *listSupportGroupsLoader {
	return dataloadgen.NewLoader(
		func(ctx context.Context, keys []listSupportGroupsParams) ([]listSupportGroupsResult, []error) {
			// Extract slices for batch call
			filters := make([]*entity.SupportGroupFilter, len(keys))
			options := make([]*entity.ListOptions, len(keys))
			for i, k := range keys {
				filters[i] = k.Filter
				options[i] = k.Options
			}

			return listSupportGroupsBatch(filters, options)
		},
		dataloadgen.WithWait(100*time.Millisecond),
	)
}

func ListSupportGroups(ctx context.Context, raw listSupportGroupsFunc, f *entity.SupportGroupFilter, o *entity.ListOptions) (listSupportGroupsResult, error) {
	if len(f.IssueId) > 0 {
		log.Println("AAAAAAAAAAAAAAAAAAAA")
		loaders, ok := getLoaders(ctx)
		if !ok || loaders == nil || loaders.listSupportGroups == nil {
			log.Println("RAW ListSupportGroups (no loader found in context)")
			return raw(f, o)
		}
		return loaders.listSupportGroups.Load(ctx, listSupportGroupsParams{
			Filter:  f,
			Options: o,
		})
	}
	log.Println("RAW ListSupportGroups (filtering not supported)")
	return raw(f, o)
}

func RegisterListSupportGroupsBatchCallback(cb listSupportGroupsBatchFunc) {
	listSupportGroupsBatchCallback = cb
}

func makeListSupportGroupsMergeKey(f *entity.SupportGroupFilter, o *entity.ListOptions) string {
	fCopy := *f
	fCopy.IssueId = nil // ignore Ids for grouping

	b, _ := json.Marshal(struct {
		F entity.SupportGroupFilter
		O entity.ListOptions
	}{F: fCopy, O: *o})

	return string(b)
}

func convertSupportGroupBatchResultToSupportGroupResult(sgbr entity.SupportGroupBatchResult) entity.SupportGroupResult { //TODO: ToSupportGroupResult
	return entity.SupportGroupResult{
		WithCursor:               sgbr.WithCursor,
		SupportGroupAggregations: sgbr.SupportGroupAggregations,
		SupportGroup:             sgbr.SupportGroup,
	}
}
func convertSupportGroupBatchResultListToSupportGrupResultList(list *entity.List[entity.SupportGroupBatchResult]) *entity.List[entity.SupportGroupResult] { //TODO: ToSupportGroupResultList
	filtered := lo.Map(list.Elements, func(sgbr entity.SupportGroupBatchResult, _ int) entity.SupportGroupResult {
		return convertSupportGroupBatchResultToSupportGroupResult(sgbr)
	})
	return &entity.List[entity.SupportGroupResult]{
		TotalCount: list.TotalCount,
		PageInfo:   list.PageInfo,
		Elements:   filtered,
	}
}

func filterByIssueIds(list *entity.List[entity.SupportGroupBatchResult], issueIds []*int64) *entity.List[entity.SupportGroupResult] {
	if len(issueIds) == 0 {
		convertSupportGroupBatchResultListToSupportGrupResultList(list)
	}

	idSet := make(map[int64]struct{}, len(issueIds))
	for _, id := range issueIds {
		if id != nil {
			idSet[*id] = struct{}{}
		}
	}

	filtered := []entity.SupportGroupResult{}
	for _, el := range list.Elements {
		if _, ok := idSet[el.WithIssueId.Value]; ok { //TODO: think about this??? el.Id??? to jest chyba problem bo w result nie ma issue_id, wiec nie ma jak przefiltrowac rezultatu
			filtered = append(filtered, convertSupportGroupBatchResultToSupportGroupResult(el))
		}
	}

	return &entity.List[entity.SupportGroupResult]{
		TotalCount: list.TotalCount,
		PageInfo:   list.PageInfo,
		Elements:   filtered,
	}
}

func listSupportGroupsBatch(
	filters []*entity.SupportGroupFilter,
	options []*entity.ListOptions,
) ([]*entity.List[entity.SupportGroupResult], []error) {

	// 1. Group requests by a "merge key" (filter without Id + options)
	// 2. Merge Ids for each group
	// 3. Execute a single DB call per group
	// 4. Distribute results back to corresponding requests

	results := make([]*entity.List[entity.SupportGroupResult], len(filters))
	errs := make([]error, len(filters))

	// Step 1: Group by merge key
	type group struct {
		filter  *entity.SupportGroupFilter
		option  *entity.ListOptions
		indexes []int
	}

	groups := map[string]*group{}

	for i, f := range filters {
		key := makeListSupportGroupsMergeKey(f, options[i]) // hash filter (without Id) + options

		fmt.Println("KEY:", key)
		g, ok := groups[key]
		if !ok {
			// clone filter to avoid modifying original
			fCopy := *f
			fCopy.IssueId = nil
			g = &group{
				filter:  &fCopy,
				option:  options[i],
				indexes: []int{},
			}
			groups[key] = g
		}

		// Merge IssueIds
		g.filter.IssueId = append(g.filter.IssueId, f.IssueId...)
		g.indexes = append(g.indexes, i)
	}
	fmt.Println("VAL: ", len(groups))

	// Step 2: Fetch from database per group
	for _, g := range groups {
		res, err := listSupportGroupsBatchCallback(g.filter, g.option)
		if err != nil {
			for _, idx := range g.indexes {
				errs[idx] = err
				results[idx] = nil
			}
			continue
		}

		// Step 3: Split results per original request
		for _, idx := range g.indexes {
			reqIds := filters[idx].IssueId
			results[idx] = filterByIssueIds(res, reqIds)
		}
	}

	return results, errs
}
