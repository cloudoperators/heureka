// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type ComponentVersionFilter struct {
	Paginated
	Id          []*int64 `json:"id"`
	IssueId     []*int64 `json:"issue_id"`
	ComponentId []*int64 `json:"component_id"`
}

type ComponentVersionAggregations struct{}

type ComponentVersionResult struct {
	WithCursor
	*ComponentVersion
	*ComponentVersionAggregations
}

type ComponentVersion struct {
	Id                 int64               `json:"id"`
	Version            string              `json:"version"`
	Component          *Component          `json:"component,omitempty"`
	ComponentId        int64               `db:"componentversion_component_id"`
	ComponentInstances []ComponentInstance `json:"component_instances,omitempty"`
	Issues             []Issue             `json:"issues,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
	DeletedAt          time.Time           `json:"deleted_at,omitempty"`
	UpdatedAt          time.Time           `json:"updated_at"`
}
