// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type Component struct {
	Info
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ComponentResult struct {
	WithCursor
	*ComponentAggregations
	*Component
}

type ComponentFilter struct {
	Info
	Paginated
	Name               []*string `json:"name"`
	Id                 []*int64  `json:"id"`
	ComponentVersionId []*int64  `json:"component_version_id"`
}

type ComponentAggregations struct {
}
