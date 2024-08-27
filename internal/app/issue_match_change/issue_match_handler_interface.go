// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match_change

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type IssueMatchChangeService interface {
	ListIssueMatchChanges(filter *entity.IssueMatchChangeFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchChangeResult], error)
	CreateIssueMatchChange(*entity.IssueMatchChange) (*entity.IssueMatchChange, error)
	UpdateIssueMatchChange(*entity.IssueMatchChange) (*entity.IssueMatchChange, error)
	DeleteIssueMatchChange(int64) error
}
