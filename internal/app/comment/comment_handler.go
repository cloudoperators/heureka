// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package comment

import (
	"context"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/sirupsen/logrus"
)

var (
	CacheTtlGetComments          = 12 * time.Hour
	CacheTtlGetAllCommentCursors = 12 * time.Hour
	CacheTtlCountComments        = 12 * time.Hour
)

type commentHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	cache         cache.Cache
	logger        *logrus.Logger
}

func NewCommentHandler(handlerContext common.HandlerContext) CommentHandler {
	return &commentHandler{
		database:      handlerContext.DB,
		eventRegistry: handlerContext.EventReg,
		cache:         handlerContext.Cache,
		logger:        logrus.New(),
	}
}

func (ch *commentHandler) ListComments(
	ctx context.Context,
	filter *entity.CommentFilter,
	options *entity.ListOptions,
) (*entity.List[entity.CommentResult], error) {
	op := appErrors.Op("commentHandler.ListComments")

	var (
		count    int64
		pageInfo *entity.PageInfo
	)

	common.EnsurePaginated(&filter.Paginated)

	res, err := cache.CallCached[[]entity.CommentResult](
		ch.cache,
		cache.NewCacheCallParams(
			CacheTtlGetComments,
			ctx,
			"GetComments",
			ch.database.GetComments,
			filter,
			options.Order,
		),
	)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "Comments", "", err)
		applog.LogError(ch.logger, wrappedErr, logrus.Fields{"filter": filter})

		return nil, wrappedErr
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := cache.CallCached[[]string](
				ch.cache,
				cache.NewCacheCallParams(
					CacheTtlGetAllCommentCursors,
					ctx,
					"GetAllCommentCursors",
					ch.database.GetAllCommentCursors,
					filter,
					options.Order,
				),
			)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), "CommentCursors", "", err)
				applog.LogError(ch.logger, wrappedErr, logrus.Fields{"filter": filter})

				return nil, wrappedErr
			}

			pageInfo = common.GetPageInfo(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = cache.CallCached[int64](
			ch.cache,
			cache.NewCacheCallParams(
				CacheTtlCountComments,
				ctx,
				"CountComments",
				ch.database.CountComments,
				filter,
			),
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "CommentCount", "", err)
			applog.LogError(ch.logger, wrappedErr, logrus.Fields{"filter": filter})

			return nil, wrappedErr
		}
	}

	result := &entity.List[entity.CommentResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	ch.eventRegistry.PushEvent(&ListCommentsEvent{
		Filter:  filter,
		Options: options,
		Results: result,
	})

	return result, nil
}

func (ch *commentHandler) CreateComment(
	ctx context.Context,
	comment *entity.Comment,
) (*entity.Comment, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateCommentEventName,
		"object": comment,
	})

	var err error

	comment.CreatedBy, err = common.GetCurrentUserId(ctx, ch.database)
	if err != nil {
		l.Error(err)
		return nil, newCommentHandlerError("Internal error while creating comment (GetUserId).")
	}

	now := time.Now().UTC()
	comment.CreatedAt = now
	comment.UpdatedAt = now
	comment.UpdatedBy = comment.CreatedBy

	newComment, err := ch.database.CreateComment(comment)
	if err != nil {
		l.Error(err)
		return nil, newCommentHandlerError("Internal error while creating comment.")
	}

	ch.eventRegistry.PushEvent(&CreateCommentEvent{Comment: newComment})

	return newComment, nil
}

type commentHandlerError struct {
	message string
}

func newCommentHandlerError(message string) *commentHandlerError {
	return &commentHandlerError{message: message}
}

func (e *commentHandlerError) Error() string {
	return e.message
}
