// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"
	"fmt"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
)

var (
	FilterDisplayServiceCcrn           string = "Service"
	FilterDisplaySupportGroupCcrn      string = "Support Group"
	FilterDisplayUserName              string = "User Name"
	FilterDisplayUserNameWithId        string = "User"
	FilterDisplayUniqueUserId          string = "Unique User ID"
	FilterDisplayComponentCcrn         string = "Pod"
	FilterDisplayIssueType             string = "Issue Type"
	FilterDisplayIssueMatchStatus      string = "Issue Match Status"
	FilterDisplayIssueMatchID          string = "Issue Match ID"
	FilterDisplayIssuePrimaryName      string = "Issue Name"
	FilterDisplayIssueSeverity         string = "Severity"
	FilterDisplayCcrn                  string = "CCRN"
	FilterDisplayRegion                string = "Region"
	FilterDisplayCluster               string = "Cluster"
	FilterDisplayNamespace             string = "Namespace"
	FilterDisplayDomain                string = "Domain"
	FilterDisplayProject               string = "Project"
	FilterDisplayPod                   string = "Pod"
	FilterDisplayContainer             string = "Container"
	FilterDisplayComponentInstanceType string = "Component Instance Type"
	FilterDisplayContext               string = "Context"

	ServiceFilterServiceCcrn      string = "serviceCcrn"
	ServiceFilterDomain           string = "domain"
	ServiceFilterRegion           string = "region"
	ServiceFilterUniqueUserId     string = "uniqueUserId"
	ServiceFilterType             string = "type"
	ServiceFilterUserName         string = "userName"
	ServiceFilterSupportGroupCcrn string = "supportGroupCcrn"
	ServiceFilterUserNameWithId   string = "uniqueUserId"

	IssueMatchFilterPrimaryName      string = "primaryName"
	IssueMatchFilterComponentCcrn    string = "componentCcrn"
	IssueMatchFilterIssueType        string = "issueType"
	IssueMatchFilterStatus           string = "status"
	IssueMatchFilterSeverity         string = "severity"
	IssueMatchFilterServiceCcrn      string = "serviceCcrn"
	IssueMatchFilterSupportGroupCcrn string = "supportGroupCcrn"

	ComponentInstanceFilterComponentCcrn string = "componentCcrn"
	ComponentInstanceFilterRegion        string = "region"
	ComponentInstanceFilterCluster       string = "cluster"
	ComponentInstanceFilterNamespace     string = "namespace"
	ComponentInstanceFilterDomain        string = "domain"
	ComponentInstanceFilterProject       string = "project"
	ComponentInstanceFilterPod           string = "pod"
	ComponentInstanceFilterContainer     string = "container"
	ComponentInstanceFilterType          string = "type"
	ComponentInstanceFilterParentId      string = "parentId"

	VulnerabilityFilterSupportGroup string = "supportGroup"
	VulnerabilityFilterSeverity     string = "severity"
	VulnerabilityFilterService      string = "service"
	VulnerabilityFilterRegion       string = "region"
)

type ResolverError struct {
	resolver string
	msg      string
}

func (re *ResolverError) Error() string {
	return fmt.Sprintf("%s: %s", re.resolver, re.msg)
}

func NewResolverError(resolver string, msg string) *ResolverError {
	return &ResolverError{
		resolver: resolver,
		msg:      msg,
	}
}

func ParseCursor(cursor *string) (*int64, error) {
	if cursor == nil {
		var tmp int64 = 0
		return &tmp, nil
	}

	id, err := strconv.ParseInt(*cursor, 10, 64)
	if err != nil {
		return nil, err
	}

	return &id, err
}

func GetPreloads(ctx context.Context) []string {
	return GetNestedPreloads(
		graphql.GetOperationContext(ctx),
		graphql.CollectFieldsCtx(ctx, nil),
		"",
	)
}

func GetNestedPreloads(
	ctx *graphql.OperationContext,
	fields []graphql.CollectedField,
	prefix string,
) (preloads []string) {
	for _, column := range fields {
		prefixColumn := GetPreloadString(prefix, column.Name)
		preloads = append(preloads, prefixColumn)
		preloads = append(
			preloads,
			GetNestedPreloads(
				ctx,
				graphql.CollectFields(ctx, column.Selections, nil),
				prefixColumn,
			)...,
		)
	}

	return
}

func GetPreloadString(prefix, name string) string {
	if len(prefix) > 0 {
		return prefix + "." + name
	}

	return name
}

func GetListOptions(requestedFields []string) *entity.ListOptions {
	return &entity.ListOptions{
		ShowTotalCount:      lo.Contains(requestedFields, "totalCount"),
		ShowPageInfo:        lo.Contains(requestedFields, "pageInfo"),
		IncludeAggregations: lo.Contains(requestedFields, "edges.node.objectMetadata"),
		Order:               []entity.Order{},
	}
}

func GetRoot(fctx *graphql.FieldContext) *graphql.FieldContext {
	if fctx.Object == "Query" {
		return fctx
	}

	return GetRoot(fctx.Parent)
}

// totalCountOf returns 0 when the total-count pointer is nil, otherwise the
// dereferenced value as an int. It replaces the `tc := 0; if ptr != nil { ... }`
// block duplicated across every list base resolver.
func totalCountOf(totalCount *int64) int {
	if totalCount == nil {
		return 0
	}

	return int(*totalCount)
}

// toFilterItem builds a FilterItem from a flat list of string values. A nil or
// empty input yields a FilterItem with nil Values, matching the prior inline
// behavior of the per-entity filter resolvers.
func toFilterItem(names []string, displayName *string) *model.FilterItem {
	var values []*string

	for _, name := range names {
		name := name
		values = append(values, &name)
	}

	return &model.FilterItem{
		DisplayName: displayName,
		Values:      values,
	}
}

// appendServiceSeverityOrder appends the fixed multi-column ordering applied when
// a Service query is ordered by severity: Critical, High, Medium, Low and None
// counts, followed by the ServiceId tiebreaker, all in the given direction.
func appendServiceSeverityOrder(order []entity.Order, direction entity.OrderDirection) []entity.Order {
	return append(order,
		entity.Order{By: entity.CriticalCount, Direction: direction},
		entity.Order{By: entity.HighCount, Direction: direction},
		entity.Order{By: entity.MediumCount, Direction: direction},
		entity.Order{By: entity.LowCount, Direction: direction},
		entity.Order{By: entity.NoneCount, Direction: direction},
		entity.Order{By: entity.ServiceId, Direction: direction},
	)
}

// appendIssueSeverityOrder appends the mapped severity order followed by the
// IssueId tiebreaker, matching the prior inline behavior for Issue severity order.
func appendIssueSeverityOrder(order []entity.Order, o *model.IssueOrderBy) []entity.Order {
	return append(order,
		o.ToOrderEntity(),
		entity.Order{By: entity.IssueId, Direction: o.Direction.ToOrderDirectionEntity()},
	)
}
