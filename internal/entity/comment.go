// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type Comment struct {
	Metadata
	Id           int64  `json:"id"`
	Text         string `json:"text"`
	IssueMatchId int64  `json:"issue_match_id"`
}

func (c *Comment) GetId() int64 {
	return c.Id
}

func (c *Comment) SetId(id int64) {
	c.Id = id
}

type CommentFilter struct {
	Paginated
	Id           []*int64          `json:"id"`
	IssueMatchId []*int64          `json:"issue_match_id"`
	State        []StateFilterType `json:"state"`
}

func (f *CommentFilter) Get() any {
	return f
}

func (f *CommentFilter) Ensure() Filter {
	if f == nil {
		return &CommentFilter{}
	}

	return f
}

type CommentResult struct {
	WithCursor
	*Comment
}
