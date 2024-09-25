// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type SupportGroupUser struct {
	Metadata
	SupportGroupId int64 `json:"support_group_id"`
	UserId         int64 `json:"user_id"`
}
