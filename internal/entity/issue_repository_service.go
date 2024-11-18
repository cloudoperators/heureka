// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type IssueRepositoryService struct {
	Metadata
	ServiceId         int64 `json:"service_id"`
	IssueRepositoryId int64 `json:"issue_repository_id"`
	Priority          int64 `json:"priority"`
}
