// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type SupportGroupService struct {
	Metadata
	SupportGroupId int64 `json:"support_group_id"`
	ServiceId      int64 `json:"service_id"`
}
