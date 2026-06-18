// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type Component struct {
	Metadata
	Id           int64  `json:"id"`
	CCRN         string `json:"ccrn"`
	Type         string `json:"type"`
	Repository   string `json:"repository"`
	Organization string `json:"organization"`
	Url          string `json:"url"`
}

func (c *Component) GetId() int64 {
	return c.Id
}

func (c *Component) SetId(id int64) {
	c.Id = id
}

type ComponentResult struct {
	WithCursor
	*ComponentAggregations
	*Component
}

type ComponentFilter struct {
	Paginated
	CCRN               []*string         `json:"ccrn"`
	Repository         []*string         `json:"repository"`
	Organization       []*string         `json:"organization"`
	ServiceCCRN        []*string         `json:"service_ccrn"`
	Id                 []*int64          `json:"id"`
	ComponentVersionId []*int64          `json:"component_version_id"`
	State              []StateFilterType `json:"state"`

	// UseMvComponentService enables the optimized query path via the mvComponentService
	// materialized view instead of the expensive CV→CI→S join chain. Set by ImageBaseResolver
	// to avoid scanning millions of ComponentInstance rows when filtering by service.
	UseMvComponentService bool `json:"use_mv_component_service"`
}

func (f *ComponentFilter) Get() any {
	return f
}

func (f *ComponentFilter) Ensure() Filter {
	if f == nil {
		return &ComponentFilter{}
	}

	return f
}

type ComponentAggregations struct{}
