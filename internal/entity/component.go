// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type Component struct {
	Metadata
	Id   int64  `json:"id"`
	CCRN string `json:"ccrn"`
	Type string `json:"type"`
}

type ComponentResult struct {
	WithCursor
	*ComponentAggregations
	*Component
}

type ComponentFilter struct {
	Paginated
	CCRN               []*string         `json:"ccrn"`
	Id                 []*int64          `json:"id"`
	ComponentVersionId []*int64          `json:"component_version_id"`
	State              []StateFilterType `json:"state"`
}

type ComponentAggregations struct {
}
