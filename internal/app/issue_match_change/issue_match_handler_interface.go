// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_match_change

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type IssueMatchChangeHandler interface {
	ListIssueMatchChanges(filter *entity.IssueMatchChangeFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchChangeResult], error)
	CreateIssueMatchChange(context.Context, *entity.IssueMatchChange) (*entity.IssueMatchChange, error)
	UpdateIssueMatchChange(context.Context, *entity.IssueMatchChange) (*entity.IssueMatchChange, error)
	DeleteIssueMatchChange(context.Context, int64) error
}
