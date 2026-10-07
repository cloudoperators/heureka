// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func SingleServiceBaseResolver(
	app app.Heureka,
	ctx context.Context,
	parent *model.NodeParent,
) (*model.Service, error) {
	return singleNodeByChildIds(
		ctx,
		"SingleServiceBaseResolver",
		parent,
		NewResolverError(
			"SingleServiceBaseResolver",
			"Bad Request - No parent provided",
		),
		NewResolverError(
			"SingleServiceBaseResolver",
			"Internal Error - found multiple services",
		),
		func(ctx context.Context, ids []*int64) ([]entity.ServiceResult, error) {
			services, err := app.ListServices(ctx, &entity.ServiceFilter{Id: ids}, entity.NewListOptions())
			if err != nil {
				return nil, NewResolverError("SingleServiceBaseResolver", err.Error())
			}

			return services.Elements, nil
		},
		func(sr entity.ServiceResult) model.Service {
			return model.NewService(sr.Service)
		},
	)
}

func ServiceBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.ServiceFilter,
	first *int,
	after *string,
	orderBy []*model.ServiceOrderBy,
	parent *model.NodeParent,
) (*model.ServiceConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called ServiceBaseResolver")

	var (
		irId    []*int64
		sgId    []*int64
		ownerId []*int64
		issueId []*int64
	)

	if parent != nil {
		parentId := parent.Parent.GetID()

		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).
				Error("ServiceBaseResolver: Error while parsing propagated parent ID'")

			return nil, NewResolverError(
				"ServiceBaseResolver",
				"Bad Request - Error while parsing propagated ID",
			)
		}

		switch parent.ParentName {
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
		Paginated:         entity.Paginated{First: first, After: after},
		SupportGroupCCRN:  filter.SupportGroupCcrn,
		CCRN:              filter.ServiceCcrn,
		Domain:            filter.Domain,
		Region:            filter.Region,
		OwnerName:         filter.UserName,
		OwnerId:           ownerId,
		IssueRepositoryId: irId,
		SupportGroupId:    sgId,
		IssueId:           issueId,
		Search:            filter.Search,
		State:             model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	for _, o := range orderBy {
		if *o.By == model.ServiceOrderByFieldSeverity {
			opt.Order = appendServiceSeverityOrder(opt.Order, o.Direction.ToOrderDirectionEntity())
		} else {
			opt.Order = append(opt.Order, o.ToOrderEntity())
		}
	}

	services, err := app.ListServices(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("ServiceBaseResolver", err.Error())
	}

	edges := buildEdges(services.Elements, func(result entity.ServiceResult) *model.ServiceEdge {
		s := model.NewServiceWithAggregations(&result)
		edge := &model.ServiceEdge{
			Node:   &s,
			Cursor: result.Cursor(),
		}

		if lo.Contains(requestedFields, "edges.priority") {
			p := int(result.IssueRepositoryService.Priority)
			edge.Priority = &p
		}

		return edge
	})

	// Batch pre-load for nested fields
	needOwners := lo.Contains(requestedFields, "edges.node.owners")
	needSupportGroups := lo.Contains(requestedFields, "edges.node.supportGroups")
	needIssueCounts := lo.Contains(requestedFields, "edges.node.issueCounts")

	if (needOwners || needSupportGroups || needIssueCounts) && len(edges) > 0 {
		serviceIDs := make([]int64, 0, len(edges))
		for _, edge := range edges {
			id, err := strconv.ParseInt(edge.Node.ID, 10, 64)
			if err != nil {
				logrus.WithField("id", edge.Node.ID).Warn("ServiceBaseResolver: failed to parse service ID for batch preload")
				continue
			}

			serviceIDs = append(serviceIDs, id)
		}

		var (
			ownersMap       map[int64][]entity.User
			supportGroupMap map[int64][]entity.SupportGroup
			issueCountsMap  map[int64]entity.IssueSeverityCounts
		)

		g, gCtx := errgroup.WithContext(ctx)

		if needOwners {
			g.Go(func() error {
				var err error

				ownersMap, err = app.ListOwnersByServiceIDs(gCtx, serviceIDs)
				if err != nil {
					logrus.WithField("error", err).Warn("ServiceBaseResolver: batch preload owners failed")
				}

				return nil // don't fail the whole request
			})
		}

		if needSupportGroups {
			g.Go(func() error {
				var err error

				supportGroupMap, err = app.ListSupportGroupsByServiceIDs(gCtx, serviceIDs)
				if err != nil {
					logrus.WithField("error", err).Warn("ServiceBaseResolver: batch preload support groups failed")
				}

				return nil
			})
		}

		if needIssueCounts {
			g.Go(func() error {
				var err error

				issueCountsMap, err = app.ListIssueCountsByServiceIDs(gCtx, serviceIDs)
				if err != nil {
					logrus.WithField("error", err).Warn("ServiceBaseResolver: batch preload issue counts failed")
				}

				return nil
			})
		}

		_ = g.Wait()

		// Attach pre-loaded data to service nodes
		for _, edge := range edges {
			id, err := strconv.ParseInt(edge.Node.ID, 10, 64)
			if err != nil {
				continue
			}

			if needOwners && ownersMap != nil {
				users := ownersMap[id]
				userEdges := make([]*model.UserEdge, 0, len(users))

				for i := range users {
					u := model.NewUser(&users[i])
					cursor := fmt.Sprintf("%d", users[i].Id)
					userEdges = append(userEdges, &model.UserEdge{
						Node:   &u,
						Cursor: &cursor,
					})
				}

				edge.Node.Owners = &model.UserConnection{
					TotalCount: len(userEdges),
					Edges:      userEdges,
				}
			}

			if needSupportGroups && supportGroupMap != nil {
				sgs := supportGroupMap[id]

				sgEdges := make([]*model.SupportGroupEdge, 0, len(sgs))
				for i := range sgs {
					sg := model.NewSupportGroup(&sgs[i])
					cursor := fmt.Sprintf("%d", sgs[i].Id)
					sgEdges = append(sgEdges, &model.SupportGroupEdge{
						Node:   &sg,
						Cursor: &cursor,
					})
				}

				edge.Node.SupportGroups = &model.SupportGroupConnection{
					TotalCount: len(sgEdges),
					Edges:      sgEdges,
				}
			}

			if needIssueCounts && issueCountsMap != nil {
				if counts, ok := issueCountsMap[id]; ok {
					sc := model.NewSeverityCounts(&counts)
					edge.Node.IssueCounts = &sc
				} else {
					sc := model.NewSeverityCounts(&entity.IssueSeverityCounts{})
					edge.Node.IssueCounts = &sc
				}
			}
		}
	}

	tc := totalCountOf(services.TotalCount)

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
			icFilter.AllServices = new(true)
		}

		severityCounts, err := IssueCountsBaseResolver(app, ctx, icFilter, nil)
		if err != nil {
			return nil, NewResolverError("ServiceBaseResolver", err.Error())
		}

		connection.IssueCounts = severityCounts
	}

	return &connection, nil
}

func ServiceCcrnBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.ServiceFilter,
) (*model.FilterItem, error) {
	return ServiceFilterBaseResolver(app.ListServiceCcrns, ctx, filter, &FilterDisplayServiceCcrn)
}

func ServiceDomainBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.ServiceFilter,
) (*model.FilterItem, error) {
	return ServiceFilterBaseResolver(app.ListServiceDomains, ctx, filter, &FilterDisplayDomain)
}

func ServiceRegionBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.ServiceFilter,
) (*model.FilterItem, error) {
	return ServiceFilterBaseResolver(app.ListServiceRegions, ctx, filter, &FilterDisplayRegion)
}

func ServiceFilterBaseResolver(
	appCall func(ctx context.Context, filter *entity.ServiceFilter, opt *entity.ListOptions) ([]string, error),
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
		Paginated:        entity.Paginated{},
		SupportGroupCCRN: filter.SupportGroupCcrn,
		CCRN:             filter.ServiceCcrn,
		Domain:           filter.Domain,
		Region:           filter.Region,
		OwnerName:        filter.UserName,
		State:            model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	names, err := appCall(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("ServiceFilterBaseResolver", err.Error())
	}

	return toFilterItem(names, filterDisplay), nil
}
