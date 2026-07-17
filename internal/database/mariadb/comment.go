// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package mariadb

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

var commentObject = DbObject[*entity.Comment, *entity.CommentFilter, entity.CommentResult, *any]{
	Prefix:       "comment",
	TableName:    "Comment",
	TableKey:     "C",
	DefaultOrder: entity.Order{By: entity.CommentId, Direction: entity.OrderDirectionAsc},
	Properties: []*Property[*entity.Comment]{
		NewProperty("comment_text", func(c *entity.Comment) (any, bool) {
			return c.Text, c.Text != ""
		}),
		NewProperty("comment_issuematch_id", func(c *entity.Comment) (any, bool) {
			return c.IssueMatchId, c.IssueMatchId != 0
		}),
		NewProperty("comment_created_by", func(c *entity.Comment) (any, bool) { return c.CreatedBy, NoUpdate }),
		NewProperty("comment_updated_by", func(c *entity.Comment) (any, bool) { return c.UpdatedBy, c.UpdatedBy != 0 }),
	},
	FilterProperties: []*FilterProperty[*entity.CommentFilter]{
		NewFilterProperty("C.comment_id = ?", func(filter *entity.CommentFilter) any { return filter.Id }),
		NewFilterProperty("C.comment_issuematch_id = ?", func(filter *entity.CommentFilter) any { return filter.IssueMatchId }),
		NewStateFilterProperty("C.comment", func(filter *entity.CommentFilter) any { return filter.State }),
	},
	RowToData: func(e RowComposite, _ []entity.Order) (*entity.Comment, string) {
		c := e.AsComment()
		cursor, _ := EncodeCursor(WithComment(c))

		return &c, cursor
	},
	NewResult: func(c *entity.Comment, _ *any, cursor string) entity.CommentResult {
		return entity.CommentResult{
			WithCursor: entity.WithCursor{Value: cursor},
			Comment:    c,
		}
	},
}

func (s *SqlDatabase) GetComments(ctx context.Context, filter *entity.CommentFilter, order []entity.Order) ([]entity.CommentResult, error) {
	return commentObject.Get(ctx, s.db, filter, order)
}

func (s *SqlDatabase) GetAllCommentCursors(ctx context.Context, filter *entity.CommentFilter, order []entity.Order) ([]string, error) {
	return commentObject.GetAllCursors(ctx, s.db, filter, order)
}

func (s *SqlDatabase) CountComments(ctx context.Context, filter *entity.CommentFilter) (int64, error) {
	return commentObject.Count(ctx, s.db, filter)
}

func (s *SqlDatabase) CreateComment(comment *entity.Comment) (*entity.Comment, error) {
	return commentObject.Create(s.db, comment)
}
