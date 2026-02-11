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

func SingleUserBaseResolver(app app.Heureka, ctx context.Context, parent *model.NodeParent) (*model.User, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called SingleUserBaseResolver")

	if parent == nil {
		return nil, NewResolverError("SingleUserBaseResolver", "Bad Request - No parent provided")
	}

	f := &entity.UserFilter{
		Id: parent.ChildIds,
	}

	opt := &entity.ListOptions{}

	users, err := app.ListUsers(f, opt)
	// error while fetching
	if err != nil {
		return nil, NewResolverError("SingleUserBaseResolver", err.Error())
	}

	// unexpected number of results (should at most be 1)
	if len(users.Elements) > 1 {
		return nil, NewResolverError("SingleUserBaseResolver", "Internal Error - found multiple users")
	}

	// not found
	if len(users.Elements) < 1 {
		return nil, nil
	}

	ur := users.Elements[0]
	user := model.NewUser(ur.User)

	return &user, nil
}

func UserBaseResolver(app app.Heureka, ctx context.Context, filter *model.UserFilter, first *int, after *string, parent *model.NodeParent) (*model.UserConnection, error) {
	requestedFields := GetPreloads(ctx)
	logrus.WithFields(logrus.Fields{
		"requestedFields": requestedFields,
		"parent":          parent,
	}).Debug("Called UserBaseResolver")

	afterId, err := ParseCursor(after)
	if err != nil {
		logrus.WithField("after", after).Error("UserBaseResolver: Error while parsing parameter 'after'")
		return nil, NewResolverError("UserBaseResolver", "Bad Request - unable to parse cursor 'after'")
	}

	var supportGroupId []*int64
	var serviceId []*int64
	if parent != nil {
		parentId := parent.Parent.GetID()
		pid, err := ParseCursor(&parentId)
		if err != nil {
			logrus.WithField("parent", parent).Error("UserBaseResolver: Error while parsing propagated parent ID'")
			return nil, NewResolverError("UserBaseResolver", "Bad Request - Error while parsing propagated ID")
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
		Paginated:      entity.Paginated{First: first, After: afterId},
		SupportGroupId: supportGroupId,
		ServiceId:      serviceId,
		Name:           filter.UserName,
		UniqueUserID:   filter.UniqueUserID,
		State:          model.GetStateFilterType(filter.State),
	}

	opt := GetListOptions(requestedFields)

	users, err := app.ListUsers(f, opt)
	if err != nil {
		return nil, NewResolverError("UserBaseResolver", err.Error())
	}

	edges := []*model.UserEdge{}
	for _, result := range users.Elements {
		user := model.NewUser(result.User)
		edge := model.UserEdge{
			Node:   &user,
			Cursor: result.Cursor(),
		}
		edges = append(edges, &edge)
	}

	tc := 0
	if users.TotalCount != nil {
		tc = int(*users.TotalCount)
	}

	connection := model.UserConnection{
		TotalCount: tc,
		Edges:      edges,
		PageInfo:   model.NewPageInfo(users.PageInfo),
	}

	return &connection, nil
}

func UserNameBaseResolver(app app.Heureka, ctx context.Context, filter *model.UserFilter) (*model.FilterItem, error) {
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

	names, err := app.ListUserNames(f, opt)
	if err != nil {
		return nil, NewResolverError("UserNameBaseResolver", err.Error())
	}

	var pointerNames []*string

	for _, name := range names {
		pointerNames = append(pointerNames, &name)
	}

	filterItem := model.FilterItem{
		DisplayName: &FilterDisplayUserName,
		Values:      pointerNames,
	}

	return &filterItem, nil
}

func UniqueUserIDBaseResolver(app app.Heureka, ctx context.Context, filter *model.UserFilter) (*model.FilterItem, error) {
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

	names, err := app.ListUniqueUserIDs(f, opt)
	if err != nil {
		return nil, NewResolverError("UniqueUserIDBaseResolver", err.Error())
	}

	var pointerNames []*string

	for _, name := range names {
		pointerNames = append(pointerNames, &name)
	}

	filterItem := model.FilterItem{
		DisplayName: &FilterDisplayUniqueUserId,
		Values:      pointerNames,
	}

	return &filterItem, nil
}

func UserNameWithIdBaseResolver(app app.Heureka, ctx context.Context, filter *model.UserFilter) (*model.FilterValueItem, error) {
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

	names, ids, err := app.ListUserNamesAndIds(f, opt)
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
