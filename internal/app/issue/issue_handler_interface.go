// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type IssueHandler interface {
	ListIssues(context.Context, *entity.IssueFilter, *entity.IssueListOptions) (*entity.IssueList, error)
	GetIssue(context.Context, int64) (*entity.Issue, error)
	CreateIssue(context.Context, *entity.Issue) (*entity.Issue, error)
	UpdateIssue(context.Context, *entity.Issue) (*entity.Issue, error)
	DeleteIssue(context.Context, int64) error
	AddComponentVersionToIssue(context.Context, int64, int64) (*entity.Issue, error)
	RemoveComponentVersionFromIssue(context.Context, int64, int64) (*entity.Issue, error)
	ListIssueNames(context.Context, *entity.IssueFilter, *entity.ListOptions) ([]string, error)
	GetIssueSeverityCounts(context.Context, *entity.IssueFilter) (*entity.IssueSeverityCounts, error)
}
