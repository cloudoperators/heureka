// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type ComponentVersionFilter struct {
	Paginated
	Id            []*int64  `json:"id"`
	IssueId       []*int64  `json:"issue_id"`
	ComponentCCRN []*string `json:"component_ccrn"`
	ComponentId   []*int64  `json:"component_id"`
	Version       []*string `json:"version"`
}

type ComponentVersionAggregations struct {
}

type ComponentVersionResult struct {
	WithCursor
	*ComponentVersion
	*ComponentVersionAggregations
}

type ComponentVersion struct {
	Metadata
	Id                 int64               `json:"id"`
	Version            string              `json:"version"`
	Component          *Component          `json:"component,omitempty"`
	ComponentId        int64               `db:"componentversion_component_id"`
	ComponentInstances []ComponentInstance `json:"component_instances,omitempty"`
	Issues             []Issue             `json:"issues,omitempty"`
}
