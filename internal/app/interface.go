// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.wdf.sap.corp/cc/heureka/internal/entity"
)

type Heureka interface {
	ListIssues(*entity.IssueFilter, *entity.ListOptions) (*entity.List[entity.IssueResult], error)
	ListIssueVariants(*entity.IssueVariantFilter, *entity.ListOptions) (*entity.List[entity.IssueVariantResult], error)
	CreateIssueVariant(*entity.IssueVariant) (*entity.IssueVariant, error)
	UpdateIssueVariant(*entity.IssueVariant) (*entity.IssueVariant, error)
	DeleteIssueVariant(int64) error

	ListEffectiveIssueVariants(*entity.IssueVariantFilter, *entity.ListOptions) (*entity.List[entity.IssueVariantResult], error)
	CreateIssueRepository(*entity.IssueRepository) (*entity.IssueRepository, error)
	UpdateIssueRepository(*entity.IssueRepository) (*entity.IssueRepository, error)
	DeleteIssueRepository(int64) error

	ListIssueMatches(filter *entity.IssueMatchFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchResult], error)
	CreateIssueMatch(*entity.IssueMatch) (*entity.IssueMatch, error)
	UpdateIssueMatch(*entity.IssueMatch) (*entity.IssueMatch, error)
	DeleteIssueMatch(int64) error

	ListIssueMatchChanges(filter *entity.IssueMatchChangeFilter, options *entity.ListOptions) (*entity.List[entity.IssueMatchChangeResult], error)

	ListServices(*entity.ServiceFilter, *entity.ListOptions) (*entity.List[entity.ServiceResult], error)
	CreateService(*entity.Service) (*entity.Service, error)
	UpdateService(*entity.Service) (*entity.Service, error)
	DeleteService(int64) error

	ListUsers(*entity.UserFilter, *entity.ListOptions) (*entity.List[entity.UserResult], error)
	CreateUser(*entity.User) (*entity.User, error)
	UpdateUser(*entity.User) (*entity.User, error)
	DeleteUser(int64) error

	ListSupportGroups(*entity.SupportGroupFilter, *entity.ListOptions) (*entity.List[entity.SupportGroupResult], error)
	CreateSupportGroup(*entity.SupportGroup) (*entity.SupportGroup, error)
	UpdateSupportGroup(*entity.SupportGroup) (*entity.SupportGroup, error)
	DeleteSupportGroup(int64) error

	ListComponentInstances(*entity.ComponentInstanceFilter, *entity.ListOptions) (*entity.List[entity.ComponentInstanceResult], error)
	CreateComponentInstance(*entity.ComponentInstance) (*entity.ComponentInstance, error)
	UpdateComponentInstance(*entity.ComponentInstance) (*entity.ComponentInstance, error)
	DeleteComponentInstance(int64) error

	ListActivities(*entity.ActivityFilter, *entity.ListOptions) (*entity.List[entity.ActivityResult], error)
	CreateActivity(*entity.Activity) (*entity.Activity, error)
	UpdateActivity(*entity.Activity) (*entity.Activity, error)
	DeleteActivity(int64) error

	ListEvidences(*entity.EvidenceFilter, *entity.ListOptions) (*entity.List[entity.EvidenceResult], error)
	CreateEvidence(*entity.Evidence) (*entity.Evidence, error)
	UpdateEvidence(*entity.Evidence) (*entity.Evidence, error)
	DeleteEvidence(int64) error

	ListComponents(*entity.ComponentFilter, *entity.ListOptions) (*entity.List[entity.ComponentResult], error)
	CreateComponent(*entity.Component) (*entity.Component, error)
	UpdateComponent(*entity.Component) (*entity.Component, error)
	DeleteComponent(int64) error

	ListComponentVersions(*entity.ComponentVersionFilter, *entity.ListOptions) (*entity.List[entity.ComponentVersionResult], error)
	CreateComponentVersion(*entity.ComponentVersion) (*entity.ComponentVersion, error)
	UpdateComponentVersion(*entity.ComponentVersion) (*entity.ComponentVersion, error)
	DeleteComponentVersion(int64) error

	ListIssueRepositories(*entity.IssueRepositoryFilter, *entity.ListOptions) (*entity.List[entity.IssueRepositoryResult], error)
	GetSeverity(*entity.SeverityFilter) (*entity.Severity, error)
	Shutdown() error
}
