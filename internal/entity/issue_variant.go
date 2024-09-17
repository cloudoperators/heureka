// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type IssueVariant struct {
	Info
	Id                int64            `json:"id"`
	IssueRepositoryId int64            `json:"issue_repository_id"`
	IssueRepository   *IssueRepository `json:"issue_repository"`
	SecondaryName     string           `json:"secondary_name"`
	IssueId           int64            `json:"issue_id"`
	Issue             *Issue           `json:"issue"`
	Severity          Severity         `json:"severity"`
	Description       string           `json:"description"`
}

type IssueVariantFilter struct {
	Info
	Paginated
	Id                []*int64  `json:"id"`
	SecondaryName     []*string `json:"secondary_name"`
	IssueId           []*int64  `json:"issue_id"`
	IssueRepositoryId []*int64  `json:"issue_repository_id"`
	ServiceId         []*int64  `json:"service_id"`
	IssueMatchId      []*int64  `json:"issue_match_id"`
}

func NewIssueVariantFilter() *IssueVariantFilter {
	return &IssueVariantFilter{
		Paginated: Paginated{
			First: nil,
			After: nil,
		},
		Id:                nil,
		SecondaryName:     nil,
		IssueId:           nil,
		IssueRepositoryId: nil,
		ServiceId:         nil,
		IssueMatchId:      nil,
	}
}

type IssueVariantAggregations struct {
	Info
}

type IssueVariantResult struct {
	WithCursor
	*IssueVariantAggregations
	*IssueVariant
}
