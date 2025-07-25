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

func SingleComponentInstanceBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.ComponentInstance, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleComponentInstanceBaseResolver")

	if parent == nil {
		return nil, NewResolverError("SingleComponentInstanceBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.ComponentInstanceFilter{
		Id: parent.ChildIds,
	}

	opt := &entity.ListOptions{}

	componentInstances, err := app.ListComponentInstances(f, opt)

	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleComponentInstanceBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(componentInstances.Elements) > 1 {
		return nil, NewResolverError("SingleComponentInstanceBaseResolver", "Internal Error - found multiple component instances")
	}

	//not found
	if len(componentInstances.Elements) < 1 {
		return nil, nil
	}

	var cir entity.ComponentInstanceResult = componentInstances.Elements[0]
	componentInstance := model.NewComponentInstance(cir.ComponentInstance)

	return &componentInstance, nil
}

func ComponentInstanceBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter, first *int, after *string, orderBy []*model.ComponentInstanceOrderBy, parent *model.NodeParent) (*model.ComponentInstanceConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called ComponentInstanceBaseResolver")

	var imId []*int64
	var serviceId []*int64
	var copmonentVersionId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("ComponentInstanceBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("ComponentInstanceBaseResolver", "Bad Request - Error while parsing propagated ID")
		}

		switch parent.ParentName {
		case model.IssueMatchNodeName:
			imId = []*int64{pid}
		case model.ServiceNodeName:
			serviceId = []*int64{pid}
		case model.ComponentVersionNodeName:
			copmonentVersionId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.ComponentInstanceFilter{}
	}

	parentIds, err := util.ConvertStrToIntSlice(filter.ParentID)
	if err != nil {
		return nil, NewResolverError("ComponentInstanceBaseResolver", "Invalid ParentID filter")
	}

	f := &entity.ComponentInstanceFilter{
		PaginatedX:              entity.PaginatedX{First: first, After: after},
		CCRN:                    filter.Ccrn,
		Region:                  filter.Region,
		Cluster:                 filter.Cluster,
		Namespace:               filter.Namespace,
		Domain:                  filter.Domain,
		Project:                 filter.Project,
		Pod:                     filter.Pod,
		Container:               filter.Container,
		Type:                    lo.Map(filter.Type, func(item *model.ComponentInstanceTypes, _ int) *string { return pointer.String(item.String()) }),
		IssueMatchId:            imId,
		ServiceId:               serviceId,
		ServiceCcrn:             filter.ServiceCcrn,
		ComponentVersionId:      copmonentVersionId,
		ComponentVersionVersion: filter.ComponentVersionDigest,
		Search:                  filter.Search,
		State:                   model.GetStateFilterType(filter.State),
		ParentId:                parentIds,
	}

	opt := GetListOptions(requestedFields)
	for _, o := range orderBy {
		opt.Order = append(opt.Order, o.ToOrderEntity())
	}

	componentInstances, err := app.ListComponentInstances(f, opt)

	//@todo propper error handling
	if err != nil {
		return nil, NewResolverError("ComponentInstanceBaseResolver", err.Error())
	}

	edges := []*model.ComponentInstanceEdge{}
	for _, result := range componentInstances.Elements {
		ci := model.NewComponentInstance(result.ComponentInstance)
		edge := model.ComponentInstanceEdge{
			Node:   &ci,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if componentInstances.TotalCount != nil {
		tc = int(*componentInstances.TotalCount)
	}

	connection := model.ComponentInstanceConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(componentInstances.PageInfo),
	}

	return &connection, nil
}

func CcrnBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	return ComponentInstanceFilterBaseResolver(app.ListCcrns, ctx, filter, &FilterDisplayCcrn)
}

func RegionBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	return ComponentInstanceFilterBaseResolver(app.ListRegions, ctx, filter, &FilterDisplayRegion)
}

func ClusterBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	return ComponentInstanceFilterBaseResolver(app.ListClusters, ctx, filter, &FilterDisplayCluster)
}

func NamespaceBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	return ComponentInstanceFilterBaseResolver(app.ListNamespaces, ctx, filter, &FilterDisplayNamespace)
}

func DomainBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	return ComponentInstanceFilterBaseResolver(app.ListDomains, ctx, filter, &FilterDisplayDomain)
}

func ProjectBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	return ComponentInstanceFilterBaseResolver(app.ListProjects, ctx, filter, &FilterDisplayProject)
}

func PodBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	return ComponentInstanceFilterBaseResolver(app.ListPods, ctx, filter, &FilterDisplayPod)
}

func ContainerBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	return ComponentInstanceFilterBaseResolver(app.ListContainers, ctx, filter, &FilterDisplayContainer)
}

func TypeBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	return ComponentInstanceFilterBaseResolver(app.ListTypes, ctx, filter, &FilterDisplayComponentInstanceType)
}

func ParentBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterItem, error) {
	return ComponentInstanceFilterBaseResolver(app.ListParents, ctx, filter, &ComponentInstanceFilterParentId)
}

func ContextBaseResolver(app app.Heureka, ctx context.Context, filter *model.ComponentInstanceFilter) (*model.FilterJSONItem, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called ComponentInstanceFilterBaseResolver (%s)", filter)

	if filter == nil {
		filter = &model.ComponentInstanceFilter{}
	}

	var contextFilters []*entity.Json
	for _, cont := range filter.Context {
		contextFilters = append(contextFilters, (*entity.Json)(&cont))
	}

	f := &entity.ComponentInstanceFilter{
		CCRN:      filter.Ccrn,
		Region:    filter.Region,
		Cluster:   filter.Cluster,
		Namespace: filter.Namespace,
		Domain:    filter.Domain,
		Project:   filter.Project,
		Pod:       filter.Pod,
		Container: filter.Container,
		Type:      lo.Map(filter.Type, func(item *model.ComponentInstanceTypes, _ int) *string { return pointer.String(item.String()) }),
		Context:   contextFilters,
		Search:    filter.Search,
		State:     model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	names, err := app.ListContexts(f, opt)

	if err != nil {
		return nil, NewResolverError("ComponentInstanceFilterBaseResolver", err.Error())
	}

	var pointerNames []map[string]any

	for _, name := range names {
		pointerNames = append(pointerNames, *util.ConvertStrToJsonNoError(&name))
	}

	filterItem := model.FilterJSONItem{
		DisplayName: &FilterDisplayContext,
		Values:      pointerNames,
	}

	return &filterItem, nil
}

func ComponentInstanceFilterBaseResolver(
	appCall func(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error),
	ctx context.Context,
	filter *model.ComponentInstanceFilter,
	filterDisplay *string) (*model.FilterItem, error) {

	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called ComponentInstanceFilterBaseResolver (%s)", filterDisplay)

	if filter == nil {
		filter = &model.ComponentInstanceFilter{}
	}

	f := &entity.ComponentInstanceFilter{
		CCRN:      filter.Ccrn,
		Region:    filter.Region,
		Cluster:   filter.Cluster,
		Namespace: filter.Namespace,
		Domain:    filter.Domain,
		Project:   filter.Project,
		Pod:       filter.Pod,
		Container: filter.Container,
		Type:      lo.Map(filter.Type, func(item *model.ComponentInstanceTypes, _ int) *string { return pointer.String(item.String()) }),
		Search:    filter.Search,
		State:     model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	names, err := appCall(f, opt)

	if err != nil {
		return nil, NewResolverError("ComponentInstanceFilterBaseResolver", err.Error())
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
