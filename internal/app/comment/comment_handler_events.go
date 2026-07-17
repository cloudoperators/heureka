// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package comment

import "github.com/cloudoperators/heureka/internal/app/event"

const (
	ListCommentsEventName  event.EventName = "ListComments"
	CreateCommentEventName event.EventName = "CreateComment"
)

type ListCommentsEvent struct {
	Filter  any
	Options any
	Results any
}

func (e *ListCommentsEvent) Name() event.EventName {
	return ListCommentsEventName
}

type CreateCommentEvent struct {
	Comment any
}

func (e *CreateCommentEvent) Name() event.EventName {
	return CreateCommentEventName
}
