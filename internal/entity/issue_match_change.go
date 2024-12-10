// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type IssueMatchChangeAction string

const (
	IssueMatchChangeActionAdd    IssueMatchChangeAction = "add"
	IssueMatchChangeActionRemove IssueMatchChangeAction = "remove"
)

func (e IssueMatchChangeAction) String() string {
	return string(e)
}

func NewIssueMatchChangeAction(s string) IssueMatchChangeAction {
	switch s {
	case IssueMatchChangeActionAdd.String():
		return IssueMatchChangeActionAdd
	case IssueMatchChangeActionRemove.String():
		return IssueMatchChangeActionRemove
	}
	return IssueMatchChangeActionAdd
}

var AllIssueMatchChangeActions = []string{
	IssueMatchChangeActionAdd.String(),
	IssueMatchChangeActionRemove.String(),
}

type IssueMatchChange struct {
	Metadata
	Id           int64 `json:"id"`
	ActivityId   int64 `json:"activity_id"`
	Activity     *Activity
	IssueMatchId int64 `json:"issue_match_id"`
	IssueMatch   *IssueMatch
	Action       string `json:"action"`
}

type IssueMatchChangeFilter struct {
	Paginated
	Id           []*int64        `json:"id"`
	ActivityId   []*int64        `json:"activity_id"`
	IssueMatchId []*int64        `json:"issue_match_id"`
	Action       []*string       `json:"action"`
	State        StateFilterType `json:"state"`
}

type IssueMatchChangeResult struct {
	WithCursor
	*IssueMatchChange
}
