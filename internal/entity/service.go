// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type BaseService struct {
	Id             int64         `json:"id"`
	Name           string        `json:"name"`
	SupportGroup   *SupportGroup `json:"support_group,omitempty"`
	SupportGroupId int64         `db:"service_support_group_id"`
	Owners         []User        `json:"owners,omitempty"`
	Activities     []Activity    `json:"activities,omitempty"`
	Priority       int64         `json:"priority"`
	CreatedAt      time.Time     `json:"created_at"`
	DeletedAt      time.Time     `json:"deleted_at,omitempty"`
	UpdatedAt      time.Time     `json:"updated_at"`
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
	Paginated
	SupportGroupName    []*string `json:"support_group_name"`
	Id                  []*int64  `json:"id"`
	Name                []*string `json:"name"`
	OwnerName           []*string `json:"owner_name"`
	OwnerId             []*int64  `json:"owner_id"`
	ActivityId          []*int64  `json:"activity_id"`
	ComponentInstanceId []*int64  `json:"component_instance_id"`
	IssueRepositoryId   []*int64  `json:"issue_repository_id"`
	SupportGroupId      []*int64  `json:"support_group_id"`
	Search              []*string `json:"search"`
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
