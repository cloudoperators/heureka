// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type SupportGroupService struct {
	SupportGroupId int64     `json:"support_group_id"`
	ServiceId      int64     `json:"service_id"`
	CreatedAt      time.Time `json:"created_at"`
	DeletedAt      time.Time `json:"deleted_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}
