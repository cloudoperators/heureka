// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"
	"fmt"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
)

var (
	FilterDisplayServiceName      string = "Service Name"
	FilterDisplaySupportGroupName string = "Support Group Name"
	FilterDisplayUserName         string = "User Name"
	FilterDisplayUniqueUserId     string = "Unique User ID"
	FilterDisplayComponentName    string = "Component Name"
	FilterDisplayIssueType        string = "Issue Type"
	FilterDisplayIssueMatchStatus string = "Issue Match Status"
	FilterDisplayIssuePrimaryName string = "Issue Name"
	FilterDisplayIssueSeverity    string = "Severity"

	ServiceFilterServiceName      string = "serviceName"
	ServiceFilterUniqueUserId     string = "uniqueUserId"
	ServiceFilterType             string = "type"
	ServiceFilterUserName         string = "userName"
	ServiceFilterSupportGroupName string = "supportGroupName"

	IssueMatchFilterPrimaryName      string = "primaryName"
	IssueMatchFilterComponentName    string = "componentName"
	IssueMatchFilterIssueType        string = "issueType"
	IssueMatchFilterStatus           string = "status"
	IssueMatchFilterSeverity         string = "severity"
	IssueMatchFilterAffectedService  string = "affectedService"
	IssueMatchFilterSupportGroupName string = "supportGroupName"
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

func GetNestedPreloads(ctx *graphql.OperationContext, fields []graphql.CollectedField, prefix string) (preloads []string) {
	for _, column := range fields {
		prefixColumn := GetPreloadString(prefix, column.Name)
		preloads = append(preloads, prefixColumn)
		preloads = append(preloads, GetNestedPreloads(ctx, graphql.CollectFields(ctx, column.Selections, nil), prefixColumn)...)
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
		IncludeAggregations: lo.Contains(requestedFields, "edges.node.metadata"),
	}
}
