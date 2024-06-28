// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type SeverityFilter struct {
	IssueMatchId []*int64 `json:"issue_match_id"`
	IssueId      []*int64 `json:"issue_id"`
}
