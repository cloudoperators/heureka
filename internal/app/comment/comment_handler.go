// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package comment

import (
	"context"
	"time"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
)

type commentHandler struct {
	common.BaseHandler[entity.CommentResult, *entity.CommentFilter]
}

func NewCommentHandler(hc common.HandlerContext) CommentHandler {
	return &commentHandler{
		BaseHandler: common.NewBaseHandler(
			hc,
			common.BaseConfig[entity.CommentResult, *entity.CommentFilter]{
				Op:           appErrors.Op("commentHandler"),
				Entity:       "Comments",
				CursorEntity: "CommentCursors",
				CountEntity:  "CommentCount",
				GetFn:        hc.DB.GetComments,
				CursorsFn:    hc.DB.GetAllCommentCursors,
				CountFn:      hc.DB.CountComments,
				ListEventFn: func(f *entity.CommentFilter, o *entity.ListOptions, r *entity.List[entity.CommentResult]) any {
					return &ListCommentsEvent{Filter: f, Options: o, Results: r}
				},
			},
		),
	}
}

func (ch *commentHandler) ListComments(
	ctx context.Context, filter *entity.CommentFilter, options *entity.ListOptions,
) (*entity.List[entity.CommentResult], error) {
	return ch.List(ctx, appErrors.CallerOp(), filter, options)
}

func (ch *commentHandler) CreateComment(
	ctx context.Context, comment *entity.Comment,
) (*entity.Comment, error) {
	op := appErrors.CallerOp()

	var err error

	comment.CreatedBy, err = common.GetCurrentUserId(ctx, ch.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Comment", "", err)
	}

	now := time.Now().UTC()
	comment.CreatedAt = now
	comment.UpdatedAt = now
	comment.UpdatedBy = comment.CreatedBy

	newComment, err := ch.DB().CreateComment(comment)
	if err != nil {
		return nil, appErrors.InternalError(string(op), "Comment", "", err)
	}

	ch.PushEvent(&CreateCommentEvent{Comment: newComment})

	return newComment, nil
}
