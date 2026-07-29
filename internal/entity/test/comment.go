// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/cloudoperators/heureka/internal/entity"
)

func NewFakeCommentEntity() entity.Comment {
	return entity.Comment{
		Id:           int64(gofakeit.Number(1, 10000000)),
		Text:         gofakeit.Sentence(10),
		IssueMatchId: int64(gofakeit.Number(1, 10000000)),
		Metadata: entity.Metadata{
			CreatedAt: gofakeit.Date(),
			UpdatedAt: gofakeit.Date(),
		},
	}
}

func NNewFakeComments(n int) []entity.Comment {
	c := make([]entity.Comment, n)
	for i := range n {
		c[i] = NewFakeCommentEntity()
	}

	return c
}

func NewFakeCommentResult() entity.CommentResult {
	comment := NewFakeCommentEntity()

	return entity.CommentResult{
		Comment: &comment,
	}
}
