// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type ActivityStatusValue string

const (
	ActivityStatusValuesOpen       ActivityStatusValue = "open"
	ActivityStatusValuesClosed     ActivityStatusValue = "closed"
	ActivityStatusValuesInProgress ActivityStatusValue = "in_progress"
)

func (e ActivityStatusValue) String() string {
	return string(e)
}

func NewActivityStatusValue(s string) ActivityStatusValue {
	switch s {
	case ActivityStatusValuesOpen.String():
		return ActivityStatusValuesOpen
	case ActivityStatusValuesClosed.String():
		return ActivityStatusValuesClosed
	case ActivityStatusValuesInProgress.String():
		return ActivityStatusValuesInProgress
	}
	return ActivityStatusValuesOpen
}

var AllActivityStatusValues = []string{
	ActivityStatusValuesOpen.String(),
	ActivityStatusValuesClosed.String(),
	ActivityStatusValuesInProgress.String(),
}

type Activity struct {
	Metadata
	Id        int64               `json:"id"`
	Status    ActivityStatusValue `json:"status"`
	Service   *Service            `json:"service,omitempty"`
	Issues    []Issue             `json:"issues,omitempty"`
	Evidences []Evidence          `json:"evidences,omitempty"`
}

func (a *Activity) GetId() int64 {
	return a.Id
}

type ActivityHasIssue struct {
	Metadata
	ActivityId int64 `json:"activity_id"`
	IssueId    int64 `json:"issue_id"`
}

type ActivityAggregations struct {
}

type ActivityFilter struct {
	Paginated
	Status      []*string         `json:"status"`
	ServiceCCRN []*string         `json:"service_ccrn"`
	Id          []*int64          `json:"id"`
	ServiceId   []*int64          `json:"service_id"`
	IssueId     []*int64          `json:"issue_id"`
	EvidenceId  []*int64          `json:"evidence_id"`
	State       []StateFilterType `json:"state"`
}

type ActivityResult struct {
	WithCursor
	*ActivityAggregations
	*Activity
}
