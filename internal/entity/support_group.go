// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type SupportGroup struct {
	Metadata
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type SupportGroupFilter struct {
	Metadata
	Paginated
	Id        []*int64  `json:"id"`
	ServiceId []*int64  `json:"service_id"`
	UserId    []*int64  `json:"user_id"`
	Name      []*string `json:"name"`
}

type SupportGroupAggregations struct {
	Metadata
}

type SupportGroupResult struct {
	WithCursor
	*SupportGroupAggregations
	*SupportGroup
}
