// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"
)

func SingleComponentVersionBaseResolver(
	app app.Heureka,
	ctx context.Context,
	parent *model.NodeParent,
) (*model.ComponentVersion, error) {
	return singleNodeByChildIds(
		ctx,
		"SingleComponentVersionBaseResolver",
		parent,
		NewResolverError(
			"SingleComponentVersionBaseResolver",
			"Bad Request - No parent provided",
		),
		NewResolverError(
			"SingleComponentVersionBaseResolver",
			"Internal Error - found multiple component versions",
		),
		func(ctx context.Context, ids []*int64) ([]entity.ComponentVersionResult, error) {
			componentVersions, err := app.ListComponentVersions(ctx, &entity.ComponentVersionFilter{Id: ids}, &entity.ListOptions{})
			if err != nil {
				return nil, NewResolverError("SingleComponentVersionBaseResolver", err.Error())
			}

			return componentVersions.Elements, nil
		},
		func(cvr entity.ComponentVersionResult) model.ComponentVersion {
			return model.NewComponentVersion(cvr.ComponentVersion)
		},
	)
}

func ComponentVersionBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.ComponentVersionFilter,
	first *int,
	after *string,
	orderBy []*model.ComponentVersionOrderBy,
	parent *model.NodeParent,
) (*model.ComponentVersionConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called ComponentVersionBaseResolver")

	if filter == nil {
		filter = &model.ComponentVersionFilter{}
	}

	var (
		issueId     []*int64
		componentId []*int64
		err         error
	)

	if parent != nil {
		parentId := parent.Parent.GetID()

		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).
				Error("ComponentVersionBaseResolver: Error while parsing propagated parent ID'")

			return nil, NewResolverError(
				"ComponentVersionBaseResolver",
				"Bad Request - Error while parsing propagated ID",
			)
		}

		switch parent.ParentName {
		case model.IssueNodeName:
			issueId = []*int64{pid}
		case model.ComponentNodeName:
			componentId = []*int64{pid}
		case model.ImageNodeName:
			componentId = []*int64{pid}
		}
	} else {
		componentId, err = util.ConvertStrToIntSlice(filter.ComponentID)
		if err != nil {
			return nil, NewResolverError(
				"ComponentVersionBaseResolver",
				"Bad Request - Error while parsing filter component ID",
			)
		}
	}

	serviceIds, err := util.ConvertStrToIntSlice(filter.ServiceID)
	if err != nil {
		return nil, NewResolverError(
			"ComponentVersionBaseResolver",
			"Bad Request - Error while parsing filter service ID",
		)
	}

	repositoryIds, err := util.ConvertStrToIntSlice(filter.IssueRepositoryID)
	if err != nil {
		return nil, NewResolverError(
			"ComponentVersionBaseResolver",
			"Bad Request - Error while parsing filter issue repository ID",
		)
	}

	f := &entity.ComponentVersionFilter{
		Paginated:         entity.Paginated{First: first, After: after},
		IssueId:           issueId,
		ComponentId:       componentId,
		ComponentCCRN:     filter.ComponentCcrn,
		ServiceCCRN:       filter.ServiceCcrn,
		ServiceId:         serviceIds,
		IssueRepositoryId: repositoryIds,
		Version:           filter.Version,
		State:             model.GetStateFilterType(filter.State),
		EndOfLife:         filter.EndOfLife,
	}

	opt := GetListOptions(requestedFields)

	for _, o := range orderBy {
		if *o.By == model.ComponentVersionOrderByFieldSeverity {
			opt.Order = append(
				opt.Order,
				entity.Order{
					By:        entity.CriticalCount,
					Direction: o.Direction.ToOrderDirectionEntity(),
				},
			)
			opt.Order = append(
				opt.Order,
				entity.Order{By: entity.HighCount, Direction: o.Direction.ToOrderDirectionEntity()},
			)
			opt.Order = append(
				opt.Order,
				entity.Order{
					By:        entity.MediumCount,
					Direction: o.Direction.ToOrderDirectionEntity(),
				},
			)
			opt.Order = append(
				opt.Order,
				entity.Order{By: entity.LowCount, Direction: o.Direction.ToOrderDirectionEntity()},
			)
			opt.Order = append(
				opt.Order,
				entity.Order{By: entity.NoneCount, Direction: o.Direction.ToOrderDirectionEntity()},
			)
			opt.Order = append(
				opt.Order,
				entity.Order{
					By:        entity.ComponentVersionId,
					Direction: o.Direction.ToOrderDirectionEntity(),
				},
			)
		} else {
			opt.Order = append(opt.Order, o.ToOrderEntity())
		}
	}

	componentVersions, err := app.ListComponentVersions(ctx, f, opt)
	//@todo propper error handling
	if err != nil {
		return nil, NewResolverError("ComponentVersionBaseResolver", err.Error())
	}

	edges := buildEdges(componentVersions.Elements, func(result entity.ComponentVersionResult) *model.ComponentVersionEdge {
		cv := model.NewComponentVersion(result.ComponentVersion)

		return &model.ComponentVersionEdge{
			Node:   &cv,
			Cursor: result.Cursor(),
		}
	})

	tc := totalCountOf(componentVersions.TotalCount)

	connection := model.ComponentVersionConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(componentVersions.PageInfo),
	}

	return &connection, nil
}
