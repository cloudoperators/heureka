// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package issue_repository

import "github.com/cloudoperators/heureka/internal/entity"

type IssueRepositoryHandler interface {
	ListIssueRepositories(*entity.IssueRepositoryFilter, *entity.ListOptions) (*entity.List[entity.IssueRepositoryResult], error)
	CreateIssueRepository(*entity.IssueRepository) (*entity.IssueRepository, error)
	UpdateIssueRepository(*entity.IssueRepository) (*entity.IssueRepository, error)
	DeleteIssueRepository(int64) error
}
