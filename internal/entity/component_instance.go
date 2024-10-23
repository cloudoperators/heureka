// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type ComponentInstanceFilter struct {
	Paginated
	IssueMatchId       []*int64  `json:"issue_match_id"`
	ServiceId          []*int64  `json:"service_id"`
	ComponentVersionId []*int64  `json:"component_version_id"`
	Id                 []*int64  `json:"id"`
	CCRN               []*string `json:"ccrn"`
	Search             []*string `json:"search"`
}

type ComponentInstanceAggregations struct{}

type ComponentInstanceResult struct {
	WithCursor
	*ComponentInstance
	*ComponentInstanceAggregations
}

type ComponentInstance struct {
	Id                 int64             `json:"id"`
	CCRN               string            `json:"ccrn"`
	Count              int16             `json:"count"`
	ComponentVersion   *ComponentVersion `json:"component_version,omitempty"`
	ComponentVersionId int64             `db:"componentinstance_component_version_id"`
	Service            *Service          `json:"service,omitempty"`
	ServiceId          int64             `db:"componentinstance_service_id"`
	CreatedAt          time.Time         `json:"created_at"`
	DeletedAt          time.Time         `json:"deleted_at,omitempty"`
	UpdatedAt          time.Time         `json:"updated_at"`
}
