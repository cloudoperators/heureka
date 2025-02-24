// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type SupportGroup struct {
	Metadata
	Id   int64  `json:"id"`
	CCRN string `json:"ccrn"`
}

type SupportGroupFilter struct {
	Paginated
	Id        []*int64          `json:"id"`
	ServiceId []*int64          `json:"service_id"`
	UserId    []*int64          `json:"user_id"`
	CCRN      []*string         `json:"ccrn"`
	State     []StateFilterType `json:"state"`
}

type SupportGroupAggregations struct {
}

type SupportGroupResult struct {
	WithCursor
	*SupportGroupAggregations
	*SupportGroup
}
