// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type BaseIssueRepository struct {
	Id            int64          `json:"id"`
	Name          string         `json:"name"`
	Url           string         `json:"url"`
	IssueVariants []IssueVariant `json:"issue_variants,omitempty"`
	Services      []Service      `json:"services,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     time.Time      `json:"deleted_at,omitempty"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type IssueRepositoryFilter struct {
	Paginated
	Id          []*int64  `json:"id"`
	ServiceId   []*int64  `json:"service_id"`
	Name        []*string `json:"name"`
	ServiceName []*string `json:"service_name"`
}

type IssueRepositoryAggregations struct {
}

type IssueRepository struct {
	BaseIssueRepository
	IssueRepositoryService
}

type IssueRepositoryResult struct {
	WithCursor
	*IssueRepositoryAggregations
	*IssueRepository
}
