// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type BaseIssueRepository struct {
	Info
	Id            int64          `json:"id"`
	Name          string         `json:"name"`
	Url           string         `json:"url"`
	IssueVariants []IssueVariant `json:"issue_variants,omitempty"`
	Services      []Service      `json:"services,omitempty"`
}

type IssueRepositoryFilter struct {
	Info
	Paginated
	Id          []*int64  `json:"id"`
	ServiceId   []*int64  `json:"service_id"`
	Name        []*string `json:"name"`
	ServiceName []*string `json:"service_name"`
}

func NewIssueRepositoryFilter() *IssueRepositoryFilter {
	return &IssueRepositoryFilter{
		Paginated: Paginated{
			First: nil,
			After: nil,
		},
		Id:          nil,
		ServiceId:   nil,
		Name:        nil,
		ServiceName: nil,
	}
}

type IssueRepositoryAggregations struct {
	Info
}

type IssueRepository struct {
	Info
	BaseIssueRepository
	IssueRepositoryService
}

type IssueRepositoryResult struct {
	WithCursor
	*IssueRepositoryAggregations
	*IssueRepository
}
