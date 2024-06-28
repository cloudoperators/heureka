// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type IssueVariant struct {
	Id                int64            `json:"id"`
	IssueRepositoryId int64            `json:"issue_repository_id"`
	IssueRepository   *IssueRepository `json:"issue_repository"`
	SecondaryName     string           `json:"secondary_name"`
	IssueId           int64            `json:"issue_id"`
	Issue             *Issue           `json:"issue"`
	Severity          Severity         `json:"severity"`
	Description       string           `json:"description"`
	CreatedAt         time.Time        `json:"created_at"`
	DeletedAt         time.Time        `json:"deleted_at,omitempty"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

type IssueVariantFilter struct {
	Paginated
	Id                []*int64  `json:"id"`
	SecondaryName     []*string `json:"secondary_name"`
	IssueId           []*int64  `json:"issue_id"`
	IssueRepositoryId []*int64  `json:"issue_repository_id"`
	ServiceId         []*int64  `json:"service_id"`
	IssueMatchId      []*int64  `json:"issue_match_id"`
}

type IssueVariantAggregations struct {
}

type IssueVariantResult struct {
	WithCursor
	*IssueVariantAggregations
	*IssueVariant
}
