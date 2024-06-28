// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type User struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	SapID     string    `json:"sapId"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserFilter struct {
	Paginated
	Name           []*string `json:"name"`
	SapID          []*string `json:"sapId"`
	Id             []*int64  `json:"id"`
	SupportGroupId []*int64  `json:"support_group_id"`
	ServiceId      []*int64  `json:"service_id"`
}

type UserAggregations struct {
}

type UserResult struct {
	WithCursor
	*UserAggregations
	*User
}
