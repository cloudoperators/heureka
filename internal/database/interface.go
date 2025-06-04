// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package database

import "github.com/cloudoperators/heureka/internal/entity"

type Database interface {
	GetIssues(*entity.IssueFilter, []entity.Order) ([]entity.IssueResult, error)
	GetIssuesWithAggregations(*entity.IssueFilter, []entity.Order) ([]entity.IssueResult, error)
	CountIssues(*entity.IssueFilter) (int64, error)
	CountIssueTypes(*entity.IssueFilter) (*entity.IssueTypeCounts, error)
	CountIssueRatings(*entity.IssueFilter) (*entity.IssueSeverityCounts, error)
	GetAllIssueIds(*entity.IssueFilter) ([]int64, error)
	GetAllIssueCursors(*entity.IssueFilter, []entity.Order) ([]string, error)
	CreateIssue(*entity.Issue) (*entity.Issue, error)
	UpdateIssue(*entity.Issue) error
	DeleteIssue(int64, int64) error
	AddComponentVersionToIssue(int64, int64) error
	RemoveComponentVersionFromIssue(int64, int64) error
	GetIssueNames(*entity.IssueFilter) ([]string, error)

	GetServiceIssueVariants(*entity.ServiceIssueVariantFilter) ([]entity.ServiceIssueVariant, error)
	GetIssueVariants(*entity.IssueVariantFilter) ([]entity.IssueVariant, error)
	GetAllIssueVariantIds(*entity.IssueVariantFilter) ([]int64, error)
	CountIssueVariants(*entity.IssueVariantFilter) (int64, error)
	CreateIssueVariant(*entity.IssueVariant) (*entity.IssueVariant, error)
	UpdateIssueVariant(*entity.IssueVariant) error
	DeleteIssueVariant(int64, int64) error

	GetIssueRepositories(*entity.IssueRepositoryFilter) ([]entity.IssueRepository, error)
	GetAllIssueRepositoryIds(*entity.IssueRepositoryFilter) ([]int64, error)
	CountIssueRepositories(*entity.IssueRepositoryFilter) (int64, error)
	CreateIssueRepository(*entity.IssueRepository) (*entity.IssueRepository, error)
	UpdateIssueRepository(*entity.IssueRepository) error
	DeleteIssueRepository(int64, int64) error
	GetDefaultIssuePriority() int64
	GetDefaultRepositoryName() string

	GetIssueMatches(*entity.IssueMatchFilter, []entity.Order) ([]entity.IssueMatchResult, error)
	GetAllIssueMatchIds(*entity.IssueMatchFilter) ([]int64, error)
	GetAllIssueMatchCursors(*entity.IssueMatchFilter, []entity.Order) ([]string, error)
	CountIssueMatches(filter *entity.IssueMatchFilter) (int64, error)
	CreateIssueMatch(*entity.IssueMatch) (*entity.IssueMatch, error)
	UpdateIssueMatch(*entity.IssueMatch) error
	DeleteIssueMatch(int64, int64) error

	GetIssueMatchChanges(*entity.IssueMatchChangeFilter) ([]entity.IssueMatchChange, error)
	GetAllIssueMatchChangeIds(*entity.IssueMatchChangeFilter) ([]int64, error)
	CountIssueMatchChanges(filter *entity.IssueMatchChangeFilter) (int64, error)
	CreateIssueMatchChange(*entity.IssueMatchChange) (*entity.IssueMatchChange, error)
	UpdateIssueMatchChange(*entity.IssueMatchChange) error
	DeleteIssueMatchChange(int64, int64) error
	AddEvidenceToIssueMatch(int64, int64) error
	RemoveEvidenceFromIssueMatch(int64, int64) error

	GetServices(*entity.ServiceFilter, []entity.Order) ([]entity.ServiceResult, error)
	GetServicesWithAggregations(*entity.ServiceFilter, []entity.Order) ([]entity.ServiceResult, error)
	GetAllServiceCursors(*entity.ServiceFilter, []entity.Order) ([]string, error)
	GetAllServiceIds(*entity.ServiceFilter) ([]int64, error)
	CountServices(*entity.ServiceFilter) (int64, error)
	CreateService(*entity.Service) (*entity.Service, error)
	UpdateService(*entity.Service) error
	DeleteService(int64, int64) error
	AddOwnerToService(int64, int64) error
	RemoveOwnerFromService(int64, int64) error
	AddIssueRepositoryToService(int64, int64, int64) error
	RemoveIssueRepositoryFromService(int64, int64) error
	GetServiceCcrns(*entity.ServiceFilter) ([]string, error)

	GetUsers(*entity.UserFilter) ([]entity.User, error)
	GetAllUserIds(*entity.UserFilter) ([]int64, error)
	CountUsers(*entity.UserFilter) (int64, error)
	CreateUser(*entity.User) (*entity.User, error)
	UpdateUser(*entity.User) error
	DeleteUser(int64, int64) error
	GetUserNames(*entity.UserFilter) ([]string, error)
	GetUniqueUserIDs(*entity.UserFilter) ([]string, error)

	GetSupportGroups(*entity.SupportGroupFilter) ([]entity.SupportGroup, error)
	GetAllSupportGroupIds(*entity.SupportGroupFilter) ([]int64, error)
	CountSupportGroups(*entity.SupportGroupFilter) (int64, error)
	CreateSupportGroup(*entity.SupportGroup) (*entity.SupportGroup, error)
	UpdateSupportGroup(*entity.SupportGroup) error
	DeleteSupportGroup(int64, int64) error
	AddServiceToSupportGroup(int64, int64) error
	RemoveServiceFromSupportGroup(int64, int64) error
	AddUserToSupportGroup(int64, int64) error
	RemoveUserFromSupportGroup(int64, int64) error
	GetSupportGroupCcrns(*entity.SupportGroupFilter) ([]string, error)

	GetComponentInstances(*entity.ComponentInstanceFilter, []entity.Order) ([]entity.ComponentInstanceResult, error)
	GetAllComponentInstanceIds(*entity.ComponentInstanceFilter) ([]int64, error)
	GetAllComponentInstanceCursors(*entity.ComponentInstanceFilter, []entity.Order) ([]string, error)
	CountComponentInstances(*entity.ComponentInstanceFilter) (int64, error)
	CreateComponentInstance(*entity.ComponentInstance) (*entity.ComponentInstance, error)
	UpdateComponentInstance(*entity.ComponentInstance) error
	DeleteComponentInstance(int64, int64) error
	GetComponentCcrns(filter *entity.ComponentFilter) ([]string, error)
	GetCcrn(filter *entity.ComponentInstanceFilter) ([]string, error)
	GetRegion(filter *entity.ComponentInstanceFilter) ([]string, error)
	GetCluster(filter *entity.ComponentInstanceFilter) ([]string, error)
	GetNamespace(filter *entity.ComponentInstanceFilter) ([]string, error)
	GetDomain(filter *entity.ComponentInstanceFilter) ([]string, error)
	GetProject(filter *entity.ComponentInstanceFilter) ([]string, error)
	GetPod(filter *entity.ComponentInstanceFilter) ([]string, error)
	GetContainer(filter *entity.ComponentInstanceFilter) ([]string, error)
	GetType(filter *entity.ComponentInstanceFilter) ([]string, error)
	GetContext(filter *entity.ComponentInstanceFilter) ([]string, error)

	GetActivities(*entity.ActivityFilter) ([]entity.Activity, error)
	GetAllActivityIds(*entity.ActivityFilter) ([]int64, error)
	CountActivities(*entity.ActivityFilter) (int64, error)
	CreateActivity(*entity.Activity) (*entity.Activity, error)
	UpdateActivity(*entity.Activity) error
	DeleteActivity(int64, int64) error
	AddServiceToActivity(int64, int64) error
	RemoveServiceFromActivity(int64, int64) error
	AddIssueToActivity(int64, int64) error
	RemoveIssueFromActivity(int64, int64) error

	GetEvidences(*entity.EvidenceFilter) ([]entity.Evidence, error)
	GetAllEvidenceIds(*entity.EvidenceFilter) ([]int64, error)
	CountEvidences(*entity.EvidenceFilter) (int64, error)
	CreateEvidence(*entity.Evidence) (*entity.Evidence, error)
	UpdateEvidence(*entity.Evidence) error
	DeleteEvidence(int64, int64) error

	GetComponents(*entity.ComponentFilter) ([]entity.Component, error)
	GetAllComponentIds(*entity.ComponentFilter) ([]int64, error)
	CountComponents(*entity.ComponentFilter) (int64, error)
	CreateComponent(*entity.Component) (*entity.Component, error)
	UpdateComponent(*entity.Component) error
	DeleteComponent(int64, int64) error

	GetComponentVersions(*entity.ComponentVersionFilter, []entity.Order) ([]entity.ComponentVersionResult, error)
	GetAllComponentVersionCursors(*entity.ComponentVersionFilter, []entity.Order) ([]string, error)
	GetAllComponentVersionIds(*entity.ComponentVersionFilter) ([]int64, error)
	CountComponentVersions(*entity.ComponentVersionFilter) (int64, error)
	CreateComponentVersion(*entity.ComponentVersion) (*entity.ComponentVersion, error)
	UpdateComponentVersion(*entity.ComponentVersion) error
	DeleteComponentVersion(int64, int64) error

	CreateScannerRun(*entity.ScannerRun) (bool, error)
	CompleteScannerRun(string) (bool, error)
	FailScannerRun(string, string) (bool, error)
	ScannerRunByUUID(string) (*entity.ScannerRun, error)
	GetScannerRuns(*entity.ScannerRunFilter) ([]entity.ScannerRun, error)
	GetScannerRunTags() ([]string, error)
	CountScannerRuns(*entity.ScannerRunFilter) (int, error)

	CloseConnection() error

	CreateScannerRunComponentInstanceTracker(componentInstanceId int64, scannerRunUUID string) error

	Autoclose() (bool, error)
}
