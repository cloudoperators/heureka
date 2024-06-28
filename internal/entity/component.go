// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type Component struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ComponentResult struct {
	WithCursor
	*ComponentAggregations
	*Component
}

type ComponentFilter struct {
	Paginated
	Name               []*string `json:"name"`
	Id                 []*int64  `json:"id"`
	ComponentVersionId []*int64  `json:"component_version_id"`
}

type ComponentAggregations struct {
}
