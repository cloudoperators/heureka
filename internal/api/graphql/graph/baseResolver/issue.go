// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
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
		return nil, NewResolverError("SingleIssueBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.IssueFilter{
		Id: parent.ChildIds,
	}

	opt := &entity.IssueListOptions{}

	issues, err := app.ListIssues(f, opt)

	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleIssueBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(issues.Elements) > 1 {
		return nil, NewResolverError("SingleIssueBaseResolver", "Internal Error - found multiple issues")
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

	var activityId []*int64
	var cvId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("IssueBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("IssueBaseResolver", "Bad Request - Error while parsing propagated ID")
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

	f := &entity.IssueFilter{
		PaginatedX:         entity.PaginatedX{First: first, After: after},
		ServiceCCRN:        filter.ServiceCcrn,
		SupportGroupCCRN:   filter.SupportGroupCcrn,
		ActivityId:         activityId,
		ComponentVersionId: cvId,
		PrimaryName:        filter.PrimaryName,
		Type:               lo.Map(filter.IssueType, func(item *model.IssueTypes, _ int) *string { return pointer.String(item.String()) }),

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

	//@todo propper error handling
	if err != nil {
		return nil, NewResolverError("IssueBaseResolver", err.Error())
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

	if filter == nil {
		filter = &model.IssueFilter{}
	}

	f := &entity.IssueFilter{
		PaginatedX:                      entity.PaginatedX{},
		ServiceCCRN:                     filter.ServiceCcrn,
		PrimaryName:                     filter.PrimaryName,
		SupportGroupCCRN:                filter.SupportGroupCcrn,
		Type:                            lo.Map(filter.IssueType, func(item *model.IssueTypes, _ int) *string { return pointer.String(item.String()) }),
		Search:                          filter.Search,
		IssueMatchStatus:                nil, //@todo Implement
		IssueMatchDiscoveryDate:         nil, //@todo Implement
		IssueMatchTargetRemediationDate: nil, //@todo Implement
		State:                           model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	names, err := app.ListIssueNames(f, opt)

	if err != nil {
		return nil, NewResolverError("IssueNamesBaseReolver", err.Error())
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
		return nil, NewResolverError("IssueCountsBaseResolver", err.Error())
	}

	var cvIds []*int64
	cvIds, err = util.ConvertStrToIntSlice(filter.ComponentVersionID)

	if err != nil {
		return nil, NewResolverError("IssueCountsBaseResolver", err.Error())
	}

	var serviceId []*int64
	var unique = false
	if parent != nil {
		var pid *int64
		if parent.Parent != nil {
			parentId := parent.Parent.GetID()
			pid, err = ParseCursor(&parentId)
			if err != nil {
				return nil, NewResolverError("IssueCountsBaseResolver", "Bad Request - Error while parsing propagated ID")
			}
		}

		switch parent.ParentName {
		case model.ComponentVersionNodeName:
			cvIds = []*int64{pid}
		case model.ServiceNodeName:
			serviceId = []*int64{pid}
		case model.VulnerabilityNodeName:
			unique = true
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
		State:              model.GetStateFilterType(filter.State),
		AllServices:        lo.FromPtr(filter.AllServices),
		Unique:             unique,
	}

	counts, err := app.GetIssueSeverityCounts(f)

	if err != nil {
		return nil, NewResolverError("IssueCountsBaseReolver", err.Error())
	}

	severityCounts := model.NewSeverityCounts(counts)

	return &severityCounts, nil
}
