// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package comment

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type CommentHandler interface {
	ListComments(ctx context.Context, filter *entity.CommentFilter, options *entity.ListOptions) (*entity.List[entity.CommentResult], error)
	CreateComment(ctx context.Context, comment *entity.Comment) (*entity.Comment, error)
}
