// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type IssueMatchChange struct {
	Id           int64 `json:"id"`
	ActivityId   int64 `json:"activity_id"`
	Activity     *Activity
	IssueMatchId int64 `json:"issue_match_id"`
	IssueMatch   *IssueMatch
	Action       string    `json:"action"`
	CreatedAt    time.Time `json:"created_at"`
	DeletedAt    time.Time `json:"deleted_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type IssueMatchChangeFilter struct {
	Paginated
	Id           []*int64  `json:"id"`
	ActivityId   []*int64  `json:"activity_id"`
	IssueMatchId []*int64  `json:"issue_match_id"`
	Action       []*string `json:"action"`
}

type IssueMatchChangeResult struct {
	WithCursor
	*IssueMatchChange
}
