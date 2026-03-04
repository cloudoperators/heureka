// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
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

func PatchBaseResolver(app app.Heureka, ctx context.Context, filter *model.PatchFilter, first *int, after *string) (*model.PatchConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called PatchBaseResolver")

	if filter == nil {
		filter = &model.PatchFilter{}
	}

	ids, err := util.ConvertStrToIntSlice(filter.ID)
	if err != nil {
		return nil, NewResolverError("PatchBaseResolver", "Bad Request - Error while parsing filter ID")
	}

	serviceIds, err := util.ConvertStrToIntSlice(filter.ServiceID)
	if err != nil {
		return nil, NewResolverError("PatchBaseResolver", "Bad Request - Error while parsing filter Service ID")
	}

	componentVersionIds, err := util.ConvertStrToIntSlice(filter.ComponentVersionID)
	if err != nil {
		return nil, NewResolverError("PatchBaseResolver", "Bad Request - Error while parsing filter ComponentVersion ID")
	}

	f := &entity.PatchFilter{
		Paginated:            entity.Paginated{First: first, After: after},
		Id:                   ids,
		ServiceId:            serviceIds,
		ServiceName:          filter.ServiceName,
		ComponentVersionId:   componentVersionIds,
		ComponentVersionName: filter.ComponentVersionName,
		State:                model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)
	patches, err := app.ListPatches(f, opt)
	if err != nil {
		return nil, ToGraphQLError(err)
	}

	edges := []*model.PatchEdge{}
	for _, result := range patches.Elements {
		ci := model.NewPatch(result.Patch)
		edge := model.PatchEdge{
			Node:   &ci,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if patches.TotalCount != nil {
		tc = int(*patches.TotalCount)
	}

	connection := model.PatchConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(patches.PageInfo),
	}

	return &connection, nil
}
