// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package baseResolver

import (
	"context"
	"fmt"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

func SingleUserBaseResolver(
	app app.Heureka,
	ctx context.Context,
	parent *model.NodeParent,
) (*model.User, error) {
	return singleNodeByChildIds(
		ctx,
		"SingleUserBaseResolver",
		parent,
		NewResolverError("SingleUserBaseResolver", "Bad Request - No parent provided"),
		NewResolverError(
			"SingleUserBaseResolver",
			"Internal Error - found multiple users",
		),
		func(ctx context.Context, ids []*int64) ([]entity.UserResult, error) {
			users, err := app.ListUsers(ctx, &entity.UserFilter{Id: ids}, &entity.ListOptions{})
			if err != nil {
				return nil, NewResolverError("SingleUserBaseResolver", err.Error())
			}

			return users.Elements, nil
		},
		func(ur entity.UserResult) model.User {
			return model.NewUser(ur.User)
		},
	)
}

func UserBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.UserFilter,
	first *int,
	after *string,
	parent *model.NodeParent,
) (*model.UserConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called UserBaseResolver")

	var (
		supportGroupId []*int64
		serviceId      []*int64
	)

	if parent != nil {
		parentId := parent.Parent.GetID()

		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).
				Error("UserBaseResolver: Error while parsing propagated parent ID'")

			return nil, NewResolverError(
				"UserBaseResolver",
				"Bad Request - Error while parsing propagated ID",
			)
		}

		switch parent.ParentName {
		case model.SupportGroupNodeName:
			supportGroupId = []*int64{pid}
		case model.ServiceNodeName:
			serviceId = []*int64{pid}
		}
	}

	if filter == nil {
		filter = &model.UserFilter{}
	}

	f := &entity.UserFilter{
		Paginated:      entity.Paginated{First: first, After: after},
		SupportGroupId: supportGroupId,
		ServiceId:      serviceId,
		Name:           filter.UserName,
		UniqueUserID:   filter.UniqueUserID,
		State:          model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	users, err := app.ListUsers(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("UserBaseResolver", err.Error())
	}

	edges := buildEdges(users.Elements, func(result entity.UserResult) *model.UserEdge {
		user := model.NewUser(result.User)

		return &model.UserEdge{
			Node:   &user,
			Cursor: result.Cursor(),
		}
	})

	tc := totalCountOf(users.TotalCount)

	connection := model.UserConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(users.PageInfo),
	}

	return &connection, nil
}

func UserNameBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.UserFilter,
) (*model.FilterItem, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called UserNameBaseResolver")

	if filter == nil {
		filter = &model.UserFilter{}
	}

	f := &entity.UserFilter{
		Paginated:    entity.Paginated{},
		Name:         filter.UserName,
		UniqueUserID: filter.UniqueUserID,
		State:        model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	names, err := app.ListUserNames(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("UserNameBaseResolver", err.Error())
	}

	return toFilterItem(names, &FilterDisplayUserName), nil
}

func UniqueUserIDBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.UserFilter,
) (*model.FilterItem, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called UniqueUserIDBaseResolver")

	if filter == nil {
		filter = &model.UserFilter{}
	}

	f := &entity.UserFilter{
		Paginated:    entity.Paginated{},
		UniqueUserID: filter.UniqueUserID,
		Name:         filter.UserName,
		State:        model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	names, err := app.ListUniqueUserIDs(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("UniqueUserIDBaseResolver", err.Error())
	}

	return toFilterItem(names, &FilterDisplayUniqueUserId), nil
}

func UserNameWithIdBaseResolver(
	app app.Heureka,
	ctx context.Context,
	filter *model.UserFilter,
) (*model.FilterValueItem, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
	}).Debug("Called UserNameWithIdBaseResolver")

	if filter == nil {
		filter = &model.UserFilter{}
	}

	f := &entity.UserFilter{
		Paginated:    entity.Paginated{},
		Name:         filter.UserName,
		UniqueUserID: filter.UniqueUserID,
		State:        model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	names, ids, err := app.ListUserNamesAndIds(ctx, f, opt)
	if err != nil {
		return nil, NewResolverError("UserNameWithIdBaseResolver", err.Error())
	}

	var valueItems []*model.ValueItem

	for i := range ids {
		display := fmt.Sprintf("%s (%s)", names[i], ids[i])
		valueItem := model.ValueItem{Display: &display, Value: &ids[i]}
		valueItems = append(valueItems, &valueItem)
	}

	filterItem := model.FilterValueItem{
		DisplayName: &FilterDisplayUserNameWithId,
		Values:      valueItems,
	}

	return &filterItem, nil
}
