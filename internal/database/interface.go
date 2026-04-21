// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"

	"github.com/cloudoperators/heureka/internal/entity"
)

type Database interface {
	GetIssues(context.Context, *entity.IssueFilter, []entity.Order) ([]entity.IssueResult, error)
	GetIssuesWithAggregations(context.Context, *entity.IssueFilter, []entity.Order) ([]entity.IssueResult, error)
	CountIssues(context.Context, *entity.IssueFilter) (int64, error)
	CountIssueTypes(context.Context, *entity.IssueFilter) (*entity.IssueTypeCounts, error)
	CountIssueRatings(context.Context, *entity.IssueFilter) (*entity.IssueSeverityCounts, error)
	GetAllIssueCursors(context.Context, *entity.IssueFilter, []entity.Order) ([]string, error)
	CreateIssue(*entity.Issue) (*entity.Issue, error)
	UpdateIssue(*entity.Issue) error
	DeleteIssue(int64, int64) error
	AddComponentVersionToIssue(int64, int64) error
	RemoveComponentVersionFromIssue(int64, int64) error
	GetIssueNames(context.Context, *entity.IssueFilter) ([]string, error)

	GetServiceIssueVariants(
		context.Context,
		*entity.ServiceIssueVariantFilter,
		[]entity.Order,
	) ([]entity.ServiceIssueVariantResult, error)
	GetIssueVariants(
		context.Context,
		*entity.IssueVariantFilter,
		[]entity.Order,
	) ([]entity.IssueVariantResult, error)
	GetAllIssueVariantCursors(context.Context, *entity.IssueVariantFilter, []entity.Order) ([]string, error)
	CountIssueVariants(context.Context, *entity.IssueVariantFilter) (int64, error)
	CreateIssueVariant(*entity.IssueVariant) (*entity.IssueVariant, error)
	UpdateIssueVariant(*entity.IssueVariant) error
	DeleteIssueVariant(int64, int64) error

	GetIssueRepositories(
		context.Context,
		*entity.IssueRepositoryFilter,
		[]entity.Order,
	) ([]entity.IssueRepositoryResult, error)
	GetAllIssueRepositoryCursors(context.Context, *entity.IssueRepositoryFilter, []entity.Order) ([]string, error)
	CountIssueRepositories(context.Context, *entity.IssueRepositoryFilter) (int64, error)
	CreateIssueRepository(*entity.IssueRepository) (*entity.IssueRepository, error)
	UpdateIssueRepository(*entity.IssueRepository) error
	DeleteIssueRepository(int64, int64) error
	GetDefaultIssuePriority() int64
	GetDefaultRepositoryName() string

	GetIssueMatches(context.Context, *entity.IssueMatchFilter, []entity.Order) ([]entity.IssueMatchResult, error)
	GetAllIssueMatchIds(context.Context, *entity.IssueMatchFilter) ([]int64, error)
	GetAllIssueMatchCursors(context.Context, *entity.IssueMatchFilter, []entity.Order) ([]string, error)
	CountIssueMatches(ctx context.Context, filter *entity.IssueMatchFilter) (int64, error)
	CreateIssueMatch(*entity.IssueMatch) (*entity.IssueMatch, error)
	UpdateIssueMatch(*entity.IssueMatch) error
	DeleteIssueMatch(int64, int64) error

	GetServices(context.Context, *entity.ServiceFilter, []entity.Order) ([]entity.ServiceResult, error)
	GetServicesWithAggregations(
		context.Context,
		*entity.ServiceFilter,
		[]entity.Order,
	) ([]entity.ServiceResult, error)
	GetAllServiceCursors(context.Context, *entity.ServiceFilter, []entity.Order) ([]string, error)
	CountServices(context.Context, *entity.ServiceFilter) (int64, error)
	CreateService(*entity.Service) (*entity.Service, error)
	UpdateService(*entity.Service) error
	DeleteService(int64, int64) error
	AddOwnerToService(int64, int64) error
	RemoveOwnerFromService(int64, int64) error
	AddIssueRepositoryToService(int64, int64, int64) error
	RemoveIssueRepositoryFromService(int64, int64) error
	GetServiceCcrns(context.Context, *entity.ServiceFilter) ([]string, error)
	GetServiceDomains(context.Context, *entity.ServiceFilter) ([]string, error)
	GetServiceRegions(context.Context, *entity.ServiceFilter) ([]string, error)

	GetUsers(context.Context, *entity.UserFilter) ([]entity.UserResult, error)
	GetAllUserIds(context.Context, *entity.UserFilter) ([]int64, error)
	GetAllUserCursors(context.Context, *entity.UserFilter, []entity.Order) ([]string, error)
	CountUsers(context.Context, *entity.UserFilter) (int64, error)
	CreateUser(*entity.User) (*entity.User, error)
	UpdateUser(*entity.User) error
	DeleteUser(int64, int64) error
	GetUserNames(context.Context, *entity.UserFilter) ([]string, error)
	GetUniqueUserIDs(context.Context, *entity.UserFilter) ([]string, error)

	GetSupportGroups(context.Context, *entity.SupportGroupFilter, []entity.Order) ([]entity.SupportGroupResult, error)
	GetAllSupportGroupCursors(context.Context, *entity.SupportGroupFilter, []entity.Order) ([]string, error)
	CountSupportGroups(context.Context, *entity.SupportGroupFilter) (int64, error)
	CreateSupportGroup(*entity.SupportGroup) (*entity.SupportGroup, error)
	UpdateSupportGroup(*entity.SupportGroup) error
	DeleteSupportGroup(int64, int64) error
	AddServiceToSupportGroup(int64, int64) error
	RemoveServiceFromSupportGroup(int64, int64) error
	AddUserToSupportGroup(int64, int64) error
	RemoveUserFromSupportGroup(int64, int64) error
	GetSupportGroupCcrns(context.Context, *entity.SupportGroupFilter) ([]string, error)

	GetComponentInstances(context.Context, *entity.ComponentInstanceFilter, []entity.Order) ([]entity.ComponentInstanceResult, error)
	GetAllComponentInstanceCursors(
		context.Context,
		*entity.ComponentInstanceFilter,
		[]entity.Order,
	) ([]string, error)
	CountComponentInstances(context.Context, *entity.ComponentInstanceFilter) (int64, error)
	CreateComponentInstance(*entity.ComponentInstance) (*entity.ComponentInstance, error)
	UpdateComponentInstance(*entity.ComponentInstance) error
	DeleteComponentInstance(int64, int64) error
	GetComponentCcrns(ctx context.Context, filter *entity.ComponentFilter) ([]string, error)
	GetCcrn(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)
	GetRegion(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)
	GetCluster(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)
	GetNamespace(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)
	GetDomain(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)
	GetProject(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)
	GetPod(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)
	GetContainer(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)
	GetType(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)
	GetContext(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)
	GetComponentInstanceParent(ctx context.Context, filter *entity.ComponentInstanceFilter) ([]string, error)

	GetComponents(context.Context, *entity.ComponentFilter, []entity.Order) ([]entity.ComponentResult, error)
	GetAllComponentCursors(context.Context, *entity.ComponentFilter, []entity.Order) ([]string, error)
	CountComponents(context.Context, *entity.ComponentFilter) (int64, error)
	CountComponentVulnerabilities(context.Context, *entity.ComponentFilter) ([]entity.IssueSeverityCounts, error)
	CreateComponent(*entity.Component) (*entity.Component, error)
	UpdateComponent(*entity.Component) error
	DeleteComponent(int64, int64) error

	GetComponentVersions(context.Context,
		*entity.ComponentVersionFilter,
		[]entity.Order,
	) ([]entity.ComponentVersionResult, error)
	GetAllComponentVersionCursors(context.Context, *entity.ComponentVersionFilter, []entity.Order) ([]string, error)
	CountComponentVersions(context.Context, *entity.ComponentVersionFilter) (int64, error)
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

	GetRemediations(context.Context, *entity.RemediationFilter, []entity.Order) ([]entity.RemediationResult, error)
	GetAllRemediationCursors(context.Context, *entity.RemediationFilter, []entity.Order) ([]string, error)
	CountRemediations(context.Context, *entity.RemediationFilter) (int64, error)
	CreateRemediation(*entity.Remediation) (*entity.Remediation, error)
	UpdateRemediation(*entity.Remediation) error
	DeleteRemediation(int64, int64) error

	GetPatches(context.Context, *entity.PatchFilter, []entity.Order) ([]entity.PatchResult, error)
	GetAllPatchCursors(context.Context, *entity.PatchFilter, []entity.Order) ([]string, error)
	CountPatches(context.Context, *entity.PatchFilter) (int64, error)

	CloseConnection() error

	CreateScannerRunComponentInstanceTracker(componentInstanceId int64, scannerRunUUID string) error

	Autopatch(context.Context) (bool, error)

	WaitPostMigrations() error
}
