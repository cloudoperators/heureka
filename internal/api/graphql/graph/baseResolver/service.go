// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

func SingleServiceBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.Service, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleServiceBaseResolver")

	if parent == nil {
		return nil, NewResolverError("SingleServiceBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.ServiceFilter{
		Id: parent.ChildIds,
	}

	opt := entity.NewListOptions()

	services, err := app.ListServices(f, opt)

	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleServiceBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(services.Elements) > 1 {
		return nil, NewResolverError("SingleServiceBaseResolver", "Internal Error - found multiple services")
	}

	//not found
	if len(services.Elements) < 1 {
		return nil, nil
	}

	var sr entity.ServiceResult = services.Elements[0]
	service := model.NewService(sr.Service)

	return &service, nil
}

func ServiceBaseResolver(app app.Heureka, ctx context.Context, filter *model.ServiceFilter, first *int, after *string, orderBy []*model.ServiceOrderBy, parent *model.NodeParent) (*model.ServiceConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called ServiceBaseResolver")

	var activityId []*int64
	var irId []*int64
	var sgId []*int64
	var ownerId []*int64
	var issueId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("ServiceBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("ServiceBaseResolver", "Bad Request - Error while parsing propagated ID")
		}

		switch parent.ParentName {
		case model.ActivityNodeName:
			activityId = []*int64{pid}
		case model.IssueRepositoryNodeName:
			irId = []*int64{pid}
		case model.SupportGroupNodeName:
			sgId = []*int64{pid}
		case model.UserNodeName:
			ownerId = []*int64{pid}
		case model.VulnerabilityNodeName:
			issueId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.ServiceFilter{}
	}

	f := &entity.ServiceFilter{
		PaginatedX:        entity.PaginatedX{First: first, After: after},
		SupportGroupCCRN:  filter.SupportGroupCcrn,
		CCRN:              filter.ServiceCcrn,
		Domain:            filter.Domain,
		Region:            filter.Region,
		OwnerName:         filter.UserName,
		OwnerId:           ownerId,
		ActivityId:        activityId,
		IssueRepositoryId: irId,
		SupportGroupId:    sgId,
		IssueId:           issueId,
		Search:            filter.Search,
		State:             model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)
	for _, o := range orderBy {
		if *o.By == model.ServiceOrderByFieldSeverity {
			opt.Order = append(opt.Order, entity.Order{By: entity.CriticalCount, Direction: o.Direction.ToOrderDirectionEntity()})
			opt.Order = append(opt.Order, entity.Order{By: entity.HighCount, Direction: o.Direction.ToOrderDirectionEntity()})
			opt.Order = append(opt.Order, entity.Order{By: entity.MediumCount, Direction: o.Direction.ToOrderDirectionEntity()})
			opt.Order = append(opt.Order, entity.Order{By: entity.LowCount, Direction: o.Direction.ToOrderDirectionEntity()})
			opt.Order = append(opt.Order, entity.Order{By: entity.NoneCount, Direction: o.Direction.ToOrderDirectionEntity()})
			opt.Order = append(opt.Order, entity.Order{By: entity.ServiceId, Direction: o.Direction.ToOrderDirectionEntity()})
		} else {
			opt.Order = append(opt.Order, o.ToOrderEntity())
		}
	}

	services, err := app.ListServices(f, opt)

	if err != nil {
		return nil, NewResolverError("ServiceBaseResolver", err.Error())
	}

	edges := []*model.ServiceEdge{}
	for _, result := range services.Elements {
		s := model.NewServiceWithAggregations(&result)
		edge := model.ServiceEdge{
			Node:   &s,
			Cursor: result.Cursor(),
		}

		if lo.Contains(requestedFields, "edges.priority") {
			p := int(result.IssueRepositoryService.Priority)
			edge.Priority = &p
		}

		edges = append(edges, &edge)
	}

	tc := 0
	if services.TotalCount != nil {
		tc = int(*services.TotalCount)
	}

	connection := model.ServiceConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(services.PageInfo),
	}

	if lo.Contains(requestedFields, "issueCounts") {
		icFilter := &model.IssueFilter{
			SupportGroupCcrn: filter.SupportGroupCcrn,
		}
		if f.CCRN != nil {
			icFilter.ServiceCcrn = f.CCRN
		} else {
			icFilter.AllServices = lo.ToPtr(true)
		}
		severityCounts, err := IssueCountsBaseResolver(app, ctx, icFilter, nil)
		if err != nil {
			return nil, NewResolverError("ServiceBaseResolver", err.Error())
		}
		connection.IssueCounts = severityCounts
	}

	return &connection, nil

}

func ServiceCcrnBaseResolver(app app.Heureka, ctx context.Context, filter *model.ServiceFilter) (*model.FilterItem, error) {
	return ServiceFilterBaseResolver(app.ListServiceCcrns, ctx, filter, &FilterDisplayServiceCcrn)
}

func ServiceDomainBaseResolver(app app.Heureka, ctx context.Context, filter *model.ServiceFilter) (*model.FilterItem, error) {
	return ServiceFilterBaseResolver(app.ListServiceDomains, ctx, filter, &FilterDisplayDomain)
}

func ServiceRegionBaseResolver(app app.Heureka, ctx context.Context, filter *model.ServiceFilter) (*model.FilterItem, error) {
	return ServiceFilterBaseResolver(app.ListServiceRegions, ctx, filter, &FilterDisplayRegion)
}

func ServiceFilterBaseResolver(
	appCall func(filter *entity.ServiceFilter, opt *entity.ListOptions) ([]string, error),
	ctx context.Context,
	filter *model.ServiceFilter,
	filterDisplay *string,
) (*model.FilterItem, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called ServiceCcrnBaseResolver")

	if filter == nil {
		filter = &model.ServiceFilter{}
	}

	f := &entity.ServiceFilter{
		PaginatedX:       entity.PaginatedX{},
		SupportGroupCCRN: filter.SupportGroupCcrn,
		CCRN:             filter.ServiceCcrn,
		Domain:           filter.Domain,
		Region:           filter.Region,
		OwnerName:        filter.UserName,
		State:            model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	names, err := appCall(f, opt)

	if err != nil {
		return nil, NewResolverError("ServiceFilterBaseResolver", err.Error())
	}

	var pointerNames []*string

	for _, name := range names {
		pointerNames = append(pointerNames, &name)
	}

	filterItem := model.FilterItem{
		DisplayName: filterDisplay,
		Values:      pointerNames,
	}

	return &filterItem, nil
}
