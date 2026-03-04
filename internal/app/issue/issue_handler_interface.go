// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type IssueHandler interface {
	ListIssues(*entity.IssueFilter, *entity.IssueListOptions) (*entity.IssueList, error)
	GetIssue(int64) (*entity.Issue, error)
	CreateIssue(context.Context, *entity.Issue) (*entity.Issue, error)
	UpdateIssue(context.Context, *entity.Issue) (*entity.Issue, error)
	DeleteIssue(context.Context, int64) error
	AddComponentVersionToIssue(int64, int64) (*entity.Issue, error)
	RemoveComponentVersionFromIssue(int64, int64) (*entity.Issue, error)
	ListIssueNames(*entity.IssueFilter, *entity.ListOptions) ([]string, error)
	GetIssueSeverityCounts(*entity.IssueFilter) (*entity.IssueSeverityCounts, error)
}
