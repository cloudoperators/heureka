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

	ServiceFilterServiceCcrn      string = "serviceCcrn"
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

	VulnerabilityFilterSupportGroup string = "supportGroup"
	VulnerabilityFilterSeverity     string = "severity"
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
		IncludeAggregations: lo.Contains(requestedFields, "edges.node.objectMetadata"),
		Order:               []entity.Order{},
	}
}
