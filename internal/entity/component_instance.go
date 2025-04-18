// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type ComponentInstanceFilter struct {
	PaginatedX
	IssueMatchId            []*int64          `json:"issue_match_id"`
	ServiceId               []*int64          `json:"service_id"`
	ServiceCcrn             []*string         `json:"service_ccrn"`
	ComponentVersionId      []*int64          `json:"component_version_id"`
	ComponentVersionVersion []*string         `json:"component_version_version"`
	Id                      []*int64          `json:"id"`
	CCRN                    []*string         `json:"ccrn"`
	Region                  []*string         `json:"region"`
	Cluster                 []*string         `json:"cluster"`
	Namespace               []*string         `json:"namespace"`
	Domain                  []*string         `json:"domain"`
	Project                 []*string         `json:"project"`
	Pod                     []*string         `json:"pod"`
	Container               []*string         `json:"container"`
	Search                  []*string         `json:"search"`
	State                   []StateFilterType `json:"state"`
}

type ComponentInstanceAggregations struct {
}

type ComponentInstanceResult struct {
	WithCursor
	*ComponentInstance
	*ComponentInstanceAggregations
}

type ComponentInstance struct {
	Metadata
	Id                 int64             `json:"id"`
	CCRN               string            `json:"ccrn"`
	Region             string            `json:"region"`
	Cluster            string            `json:"cluster"`
	Namespace          string            `json:"namespace"`
	Domain             string            `json:"domain"`
	Project            string            `json:"project"`
	Pod                string            `json:"pod"`
	Container          string            `json:"container"`
	Count              int16             `json:"count"`
	ComponentVersion   *ComponentVersion `json:"component_version,omitempty"`
	ComponentVersionId int64             `db:"componentinstance_component_version_id"`
	Service            *Service          `json:"service,omitempty"`
	ServiceId          int64             `db:"componentinstance_service_id"`
}
