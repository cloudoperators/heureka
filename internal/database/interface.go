// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package database

import "github.wdf.sap.corp/cc/heureka/internal/entity"

type Database interface {
	GetIssues(*entity.IssueFilter) ([]entity.Issue, error)
	GetIssuesWithAggregations(*entity.IssueFilter) ([]entity.IssueWithAggregations, error)
	CountIssues(*entity.IssueFilter) (int64, error)
	GetAllIssueIds(*entity.IssueFilter) ([]int64, error)

	GetIssueVariants(*entity.IssueVariantFilter) ([]entity.IssueVariant, error)
	GetAllIssueVariantIds(*entity.IssueVariantFilter) ([]int64, error)
	CountIssueVariants(*entity.IssueVariantFilter) (int64, error)
	CreateIssueVariant(*entity.IssueVariant) (*entity.IssueVariant, error)
	UpdateIssueVariant(*entity.IssueVariant) error
	DeleteIssueVariant(int64) error

	GetIssueRepositories(*entity.IssueRepositoryFilter) ([]entity.IssueRepository, error)
	GetAllIssueRepositoryIds(*entity.IssueRepositoryFilter) ([]int64, error)
	CountIssueRepositories(*entity.IssueRepositoryFilter) (int64, error)
	CreateIssueRepository(*entity.IssueRepository) (*entity.IssueRepository, error)
	UpdateIssueRepository(*entity.IssueRepository) error
	DeleteIssueRepository(int64) error

	GetIssueMatches(*entity.IssueMatchFilter) ([]entity.IssueMatch, error)
	GetAllIssueMatchIds(*entity.IssueMatchFilter) ([]int64, error)
	CountIssueMatches(filter *entity.IssueMatchFilter) (int64, error)
	CreateIssueMatch(*entity.IssueMatch) (*entity.IssueMatch, error)
	UpdateIssueMatch(*entity.IssueMatch) error
	DeleteIssueMatch(int64) error

	GetIssueMatchChanges(*entity.IssueMatchChangeFilter) ([]entity.IssueMatchChange, error)
	GetAllIssueMatchChangeIds(*entity.IssueMatchChangeFilter) ([]int64, error)
	CountIssueMatchChanges(filter *entity.IssueMatchChangeFilter) (int64, error)

	GetServices(*entity.ServiceFilter) ([]entity.Service, error)
	GetAllServiceIds(*entity.ServiceFilter) ([]int64, error)
	CountServices(*entity.ServiceFilter) (int64, error)
	CreateService(*entity.Service) (*entity.Service, error)
	UpdateService(*entity.Service) error
	DeleteService(int64) error

	GetUsers(*entity.UserFilter) ([]entity.User, error)
	GetAllUserIds(*entity.UserFilter) ([]int64, error)
	CountUsers(*entity.UserFilter) (int64, error)
	CreateUser(*entity.User) (*entity.User, error)
	UpdateUser(*entity.User) error
	DeleteUser(int64) error

	GetSupportGroups(*entity.SupportGroupFilter) ([]entity.SupportGroup, error)
	GetAllSupportGroupIds(*entity.SupportGroupFilter) ([]int64, error)
	CountSupportGroups(*entity.SupportGroupFilter) (int64, error)
	CreateSupportGroup(*entity.SupportGroup) (*entity.SupportGroup, error)
	UpdateSupportGroup(*entity.SupportGroup) error
	DeleteSupportGroup(int64) error

	GetComponentInstances(*entity.ComponentInstanceFilter) ([]entity.ComponentInstance, error)
	GetAllComponentInstanceIds(*entity.ComponentInstanceFilter) ([]int64, error)
	CountComponentInstances(*entity.ComponentInstanceFilter) (int64, error)
	CreateComponentInstance(*entity.ComponentInstance) (*entity.ComponentInstance, error)
	UpdateComponentInstance(*entity.ComponentInstance) error
	DeleteComponentInstance(int64) error

	GetActivities(*entity.ActivityFilter) ([]entity.Activity, error)
	GetAllActivityIds(*entity.ActivityFilter) ([]int64, error)
	CountActivities(*entity.ActivityFilter) (int64, error)
	CreateActivity(*entity.Activity) (*entity.Activity, error)
	UpdateActivity(*entity.Activity) error
	DeleteActivity(int64) error

	GetEvidences(*entity.EvidenceFilter) ([]entity.Evidence, error)
	GetAllEvidenceIds(*entity.EvidenceFilter) ([]int64, error)
	CountEvidences(*entity.EvidenceFilter) (int64, error)
	CreateEvidence(*entity.Evidence) (*entity.Evidence, error)
	UpdateEvidence(*entity.Evidence) error
	DeleteEvidence(int64) error

	GetComponents(*entity.ComponentFilter) ([]entity.Component, error)
	GetAllComponentIds(*entity.ComponentFilter) ([]int64, error)
	CountComponents(*entity.ComponentFilter) (int64, error)
	CreateComponent(*entity.Component) (*entity.Component, error)
	UpdateComponent(*entity.Component) error
	DeleteComponent(int64) error

	GetComponentVersions(*entity.ComponentVersionFilter) ([]entity.ComponentVersion, error)
	GetAllComponentVersionIds(*entity.ComponentVersionFilter) ([]int64, error)
	CountComponentVersions(*entity.ComponentVersionFilter) (int64, error)
	CreateComponentVersion(*entity.ComponentVersion) (*entity.ComponentVersion, error)
	UpdateComponentVersion(*entity.ComponentVersion) error
	DeleteComponentVersion(int64) error

	CloseConnection() error
}
