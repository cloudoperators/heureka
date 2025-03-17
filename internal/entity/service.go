// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type BaseService struct {
	Metadata
	Id             int64         `json:"id"`
	CCRN           string        `json:"ccrn"`
	SupportGroup   *SupportGroup `json:"support_group,omitempty"`
	SupportGroupId int64         `db:"service_support_group_id"`
	Owners         []User        `json:"owners,omitempty"`
	Activities     []Activity    `json:"activities,omitempty"`
	Priority       int64         `json:"priority"`
}

type ServiceAggregations struct {
	ComponentInstances int64
	IssueMatches       int64
}

type ServiceWithAggregations struct {
	Service
	ServiceAggregations
}

type ServiceFilter struct {
	PaginatedX
	SupportGroupCCRN    []*string         `json:"support_group_ccrn"`
	Id                  []*int64          `json:"id"`
	CCRN                []*string         `json:"ccrn"`
	OwnerName           []*string         `json:"owner_name"`
	OwnerId             []*int64          `json:"owner_id"`
	ActivityId          []*int64          `json:"activity_id"`
	ComponentInstanceId []*int64          `json:"component_instance_id"`
	IssueRepositoryId   []*int64          `json:"issue_repository_id"`
	SupportGroupId      []*int64          `json:"support_group_id"`
	Search              []*string         `json:"search"`
	State               []StateFilterType `json:"state"`
}

type Service struct {
	BaseService
	IssueRepositoryService
}

type ServiceResult struct {
	WithCursor
	*ServiceAggregations
	*Service
}
