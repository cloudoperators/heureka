// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type IssueRepositoryService struct {
	ServiceId         int64     `json:"service_id"`
	IssueRepositoryId int64     `json:"issue_repository_id"`
	Priority          int64     `json:"priority"`
	CreatedAt         time.Time `json:"created_at"`
	DeletedAt         time.Time `json:"deleted_at,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}
