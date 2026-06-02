// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type SupportGroup struct {
	Metadata
	Id   int64  `json:"id"`
	CCRN string `json:"ccrn"`
}

func (sg *SupportGroup) GetId() int64 {
	return sg.Id
}

func (sg *SupportGroup) SetId(id int64) {
	sg.Id = id
}

type SupportGroupFilter struct {
	Paginated
	Id        []*int64          `json:"id"`
	ServiceId []*int64          `json:"service_id"`
	UserId    []*int64          `json:"user_id"`
	IssueId   []*int64          `json:"issue_id"`
	CCRN      []*string         `json:"ccrn"`
	State     []StateFilterType `json:"state"`
}

func (f *SupportGroupFilter) Get() any {
	return f
}

func (f *SupportGroupFilter) Ensure() Filter {
	if f == nil {
		return &SupportGroupFilter{}
	}

	return f
}

type SupportGroupAggregations struct{}

type SupportGroupResult struct {
	WithCursor
	*SupportGroupAggregations
	*SupportGroup
}
