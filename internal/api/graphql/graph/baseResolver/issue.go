// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"k8s.io/utils/pointer"
)

func GetIssueListOptions(requestedFields []string) *entity.IssueListOptions {
	listOptions := GetListOptions(requestedFields)
	return &entity.IssueListOptions{
		ListOptions:         *listOptions,
		ShowIssueTypeCounts: lo.Contains(requestedFields, "issueTypeCounts"),
	}
}

func SingleIssueBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.Issue, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleIssueBaseResolver")

	if parent == nil {
		return nil, ToGraphQLError(appErrors.E(appErrors.Op("SingleIssueBaseResolver"), "Issue", appErrors.InvalidArgument, "No parent provided"))
	}

	f := &entity.IssueFilter{
		Id: parent.ChildIds,
	}

	opt := &entity.IssueListOptions{}

	issues, err := app.ListIssues(f, opt)
	if err != nil {
		return nil, ToGraphQLError(err)
	}

	// unexpected number of results (should at most be 1)
	if len(issues.Elements) > 1 {
		return nil, ToGraphQLError(appErrors.E(appErrors.Op("SingleIssueBaseResolver"), "Issue", appErrors.Internal, "found multiple issues"))
	}

	//not found
	if len(issues.Elements) < 1 {
		return nil, nil
	}

	var ir entity.IssueResult = issues.Elements[0]
	issue := model.NewIssueWithAggregations(&ir)

	return &issue, nil
}

func IssueBaseResolver(app app.Heureka, ctx context.Context, filter *model.IssueFilter, first *int, after *string, orderBy []*model.IssueOrderBy, parent *model.NodeParent) (*model.IssueConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called IssueBaseResolver")

	var irId []*int64
	var activityId []*int64
	var cvId []*int64
	var err error
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("IssueBaseResolver: Error while parsing propagated parent ID'")
			return nil, ToGraphQLError(appErrors.E(appErrors.Op("IssueBaseResolver"), "Issue", appErrors.InvalidArgument, "Error while parsing propagated ID"))
		}

		switch parent.ParentName {
		case model.ActivityNodeName:
			activityId = []*int64{pid}
		case model.ComponentVersionNodeName:
			cvId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.IssueFilter{}
	}

	if filter.IssueRepositoryID != nil {
		irId, err = util.ConvertStrToIntSlice(filter.IssueRepositoryID)
		if err != nil {
			return nil, NewResolverError("IssueBaseResolver", "Bad Request - unable to parse filter, the value of the filter IssueRepositoryId is invalid")
		}
	}

	f := &entity.IssueFilter{
		PaginatedX:         entity.PaginatedX{First: first, After: after},
		ServiceCCRN:        filter.ServiceCcrn,
		SupportGroupCCRN:   filter.SupportGroupCcrn,
		ActivityId:         activityId,
		ComponentVersionId: cvId,
		PrimaryName:        filter.PrimaryName,
		Type:               lo.Map(filter.IssueType, func(item *model.IssueTypes, _ int) *string { return pointer.String(item.String()) }),
		IssueRepositoryId:  irId,

		Search: filter.Search,

		IssueMatchStatus:                nil, //@todo Implement
		IssueMatchDiscoveryDate:         nil, //@todo Implement
		IssueMatchTargetRemediationDate: nil, //@todo Implement
		State:                           model.GetStateFilterType(filter.State),
	}

	opt := GetIssueListOptions(requestedFields)
	for _, o := range orderBy {
		if *o.By == model.IssueOrderByFieldSeverity {
			opt.Order = append(opt.Order, o.ToOrderEntity())
			opt.Order = append(opt.Order, entity.Order{By: entity.IssueId, Direction: o.Direction.ToOrderDirectionEntity()})
		} else {
			opt.Order = append(opt.Order, o.ToOrderEntity())
		}
	}

	issues, err := app.ListIssues(f, opt)
	if err != nil {
		return nil, ToGraphQLError(err)
	}

	edges := []*model.IssueEdge{}
	for _, result := range issues.Elements {
		iss := model.NewIssueWithAggregations(&result)
		edge := model.IssueEdge{
			Node:   &iss,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	totalCount := 0
	if issues.TotalCount != nil {
		totalCount = int(*issues.TotalCount)
	}

	vulnerabilityCount := 0
	policiyViolationCount := 0
	securityEventCount := 0

	if issues.VulnerabilityCount != nil && issues.PolicyViolationCount != nil && issues.SecurityEventCount != nil {
		vulnerabilityCount = int(*issues.VulnerabilityCount)
		policiyViolationCount = int(*issues.PolicyViolationCount)
		securityEventCount = int(*issues.SecurityEventCount)
	}

	connection := model.IssueConnection{
		TotalCount:           totalCount,
		VulnerabilityCount:   vulnerabilityCount,
		PolicyViolationCount: policiyViolationCount,
		SecurityEventCount:   securityEventCount,
		Edges:                edges,
		PageInfo:             model.NewPageInfo(issues.PageInfo),
	}

	return &connection, nil
}

func IssueNameBaseResolver(app app.Heureka, ctx context.Context, filter *model.IssueFilter) (*model.FilterItem, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called IssueNameBaseResolver")

	var irId []*int64
	var err error

	if filter == nil {
		filter = &model.IssueFilter{}
	}

	if filter.IssueRepositoryID != nil {
		irId, err = util.ConvertStrToIntSlice(filter.IssueRepositoryID)
		if err != nil {
			return nil, NewResolverError("IssueBaseResolver", "Bad Request - unable to parse filter, the value of the filter IssueRepositoryId is invalid")
		}
	}

	f := &entity.IssueFilter{
		PaginatedX:                      entity.PaginatedX{},
		ServiceCCRN:                     filter.ServiceCcrn,
		PrimaryName:                     filter.PrimaryName,
		SupportGroupCCRN:                filter.SupportGroupCcrn,
		Type:                            lo.Map(filter.IssueType, func(item *model.IssueTypes, _ int) *string { return pointer.String(item.String()) }),
		IssueRepositoryId:               irId,
		Search:                          filter.Search,
		IssueMatchStatus:                nil, //@todo Implement
		IssueMatchDiscoveryDate:         nil, //@todo Implement
		IssueMatchTargetRemediationDate: nil, //@todo Implement
		State:                           model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	names, err := app.ListIssueNames(f, opt)
	if err != nil {
		return nil, ToGraphQLError(err)
	}

	var pointerNames []*string

	for _, name := range names {
		pointerNames = append(pointerNames, &name)
	}

	filterItem := model.FilterItem{
		DisplayName: &FilterDisplayIssuePrimaryName,
		Values:      pointerNames,
	}

	return &filterItem, nil
}

func IssueCountsBaseResolver(app app.Heureka, ctx context.Context, filter *model.IssueFilter, parent *model.NodeParent) (*model.SeverityCounts, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called IssueCountsBaseResolver")

	if filter == nil {
		filter = &model.IssueFilter{}
	}

	irIds, err := util.ConvertStrToIntSlice(filter.IssueRepositoryID)
	if err != nil {
		return nil, ToGraphQLError(err)
	}

	var cvIds []*int64
	cvIds, err = util.ConvertStrToIntSlice(filter.ComponentVersionID)
	if err != nil {
		return nil, ToGraphQLError(err)
	}

	var serviceId []*int64
	var componentId []*int64
	var unique = false
	if parent != nil {
		var pid *int64
		if parent.Parent != nil {
			parentId := parent.Parent.GetID()
			pid, err = ParseCursor(&parentId)
			if err != nil {
				return nil, ToGraphQLError(appErrors.E(appErrors.Op("IssueCountsBaseResolver"), "Issue", appErrors.InvalidArgument, "Error while parsing propagated ID"))
			}
		}

		switch parent.ParentName {
		case model.ComponentVersionNodeName:
			cvIds = []*int64{pid}
		case model.ServiceNodeName:
			serviceId = []*int64{pid}
		case model.VulnerabilityNodeName:
			if len(filter.SupportGroupCcrn) == 0 {
				unique = true
			}
		}
	}

	f := &entity.IssueFilter{
		PaginatedX:         entity.PaginatedX{},
		ServiceCCRN:        filter.ServiceCcrn,
		SupportGroupCCRN:   filter.SupportGroupCcrn,
		PrimaryName:        filter.PrimaryName,
		Type:               lo.Map(filter.IssueType, func(item *model.IssueTypes, _ int) *string { return pointer.String(item.String()) }),
		Search:             filter.Search,
		IssueRepositoryId:  irIds,
		ComponentVersionId: cvIds,
		ServiceId:          serviceId,
		ComponentId:        componentId,
		State:              model.GetStateFilterType(filter.State),
		AllServices:        lo.FromPtr(filter.AllServices),
		Unique:             unique,
	}

	counts, err := app.GetIssueSeverityCounts(f)
	if err != nil {
		return nil, ToGraphQLError(err)
	}

	severityCounts := model.NewSeverityCounts(counts)

	return &severityCounts, nil
}
