// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type ComponentInstanceFilter struct {
	Info
	Paginated
	IssueMatchId       []*int64 `json:"issue_match_id"`
	ServiceId          []*int64 `json:"service_id"`
	ComponentVersionId []*int64 `json:"component_version_id"`
	Id                 []*int64 `json:"id"`
}

type ComponentInstanceAggregations struct {
	Info
}

type ComponentInstanceResult struct {
	WithCursor
	*ComponentInstance
	*ComponentInstanceAggregations
}

type ComponentInstance struct {
	Info
	Id                 int64             `json:"id"`
	CCRN               string            `json:"ccrn"`
	Count              int16             `json:"count"`
	ComponentVersion   *ComponentVersion `json:"component_version,omitempty"`
	ComponentVersionId int64             `db:"componentinstance_component_version_id"`
	Service            *Service          `json:"service,omitempty"`
	ServiceId          int64             `db:"componentinstance_service_id"`
}
