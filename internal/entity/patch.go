// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type Patch struct {
	Metadata
	Id                   int64  `json:"id"`
	ServiceId            int64  `json:"service_id"`
	ServiceName          string `json:"service_name"`
	ComponentVersionId   int64  `json:"component_version_id"`
	ComponentVersionName string `json:"component_version_name"`
}

func (p *Patch) GetId() int64 {
	return p.Id
}

func (p *Patch) SetId(id int64) {
	p.Id = id
}

type PatchFilter struct {
	Paginated
	Id                   []*int64          `json:"id"`
	ServiceId            []*int64          `json:"service_id"`
	ServiceName          []*string         `json:"service_name"`
	ComponentVersionId   []*int64          `json:"component_version_id"`
	ComponentVersionName []*string         `json:"component_version_name"`
	State                []StateFilterType `json:"state"`
}

type PatchResult struct {
	WithCursor
	*Patch
}
