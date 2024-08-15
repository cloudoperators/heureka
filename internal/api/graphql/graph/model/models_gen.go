// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
)

type Connection interface {
	IsConnection()
	GetTotalCount() int
	GetPageInfo() *PageInfo
}

type Edge interface {
	IsEdge()
	GetNode() Node
	GetCursor() *string
}

type Node interface {
	IsNode()
	GetID() string
}

type Activity struct {
	ID                string                      `json:"id"`
	Status            *ActivityStatusValues       `json:"status,omitempty"`
	Services          *ServiceConnection          `json:"services,omitempty"`
	Issues            *IssueConnection            `json:"issues,omitempty"`
	Evidences         *EvidenceConnection         `json:"evidences,omitempty"`
	IssueMatchChanges *IssueMatchChangeConnection `json:"issueMatchChanges,omitempty"`
}

func (Activity) IsNode()            {}
func (this Activity) GetID() string { return this.ID }

type ActivityConnection struct {
	TotalCount int             `json:"totalCount"`
	Edges      []*ActivityEdge `json:"edges,omitempty"`
	PageInfo   *PageInfo       `json:"pageInfo,omitempty"`
}

func (ActivityConnection) IsConnection()               {}
func (this ActivityConnection) GetTotalCount() int     { return this.TotalCount }
func (this ActivityConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type ActivityEdge struct {
	Node   *Activity `json:"node"`
	Cursor *string   `json:"cursor,omitempty"`
}

func (ActivityEdge) IsEdge()                 {}
func (this ActivityEdge) GetNode() Node      { return *this.Node }
func (this ActivityEdge) GetCursor() *string { return this.Cursor }

type ActivityFilter struct {
	ServiceName []*string               `json:"serviceName,omitempty"`
	Status      []*ActivityStatusValues `json:"status,omitempty"`
}

type ActivityInput struct {
	Status *ActivityStatusValues `json:"status,omitempty"`
}

type Cvss struct {
	Vector        *string            `json:"vector,omitempty"`
	Base          *CVSSBase          `json:"base,omitempty"`
	Temporal      *CVSSTemporal      `json:"temporal,omitempty"`
	Environmental *CVSSEnvironmental `json:"environmental,omitempty"`
}

type CVSSBase struct {
	Score                 *float64 `json:"score,omitempty"`
	AttackVector          *string  `json:"attackVector,omitempty"`
	AttackComplexity      *string  `json:"attackComplexity,omitempty"`
	PrivilegesRequired    *string  `json:"privilegesRequired,omitempty"`
	UserInteraction       *string  `json:"userInteraction,omitempty"`
	Scope                 *string  `json:"scope,omitempty"`
	ConfidentialityImpact *string  `json:"confidentialityImpact,omitempty"`
	IntegrityImpact       *string  `json:"integrityImpact,omitempty"`
	AvailabilityImpact    *string  `json:"availabilityImpact,omitempty"`
}

type CVSSEnvironmental struct {
	Score                         *float64 `json:"score,omitempty"`
	ModifiedAttackVector          *string  `json:"modifiedAttackVector,omitempty"`
	ModifiedAttackComplexity      *string  `json:"modifiedAttackComplexity,omitempty"`
	ModifiedPrivilegesRequired    *string  `json:"modifiedPrivilegesRequired,omitempty"`
	ModifiedUserInteraction       *string  `json:"modifiedUserInteraction,omitempty"`
	ModifiedScope                 *string  `json:"modifiedScope,omitempty"`
	ModifiedConfidentialityImpact *string  `json:"modifiedConfidentialityImpact,omitempty"`
	ModifiedIntegrityImpact       *string  `json:"modifiedIntegrityImpact,omitempty"`
	ModifiedAvailabilityImpact    *string  `json:"modifiedAvailabilityImpact,omitempty"`
	ConfidentialityRequirement    *string  `json:"confidentialityRequirement,omitempty"`
	AvailabilityRequirement       *string  `json:"availabilityRequirement,omitempty"`
	IntegrityRequirement          *string  `json:"integrityRequirement,omitempty"`
}

type CVSSParameter struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
}

type CVSSTemporal struct {
	Score               *float64 `json:"score,omitempty"`
	ExploitCodeMaturity *string  `json:"exploitCodeMaturity,omitempty"`
	RemediationLevel    *string  `json:"remediationLevel,omitempty"`
	ReportConfidence    *string  `json:"reportConfidence,omitempty"`
}

type Component struct {
	ID                string                      `json:"id"`
	Name              *string                     `json:"name,omitempty"`
	Type              *ComponentTypeValues        `json:"type,omitempty"`
	ComponentVersions *ComponentVersionConnection `json:"componentVersions,omitempty"`
}

func (Component) IsNode()            {}
func (this Component) GetID() string { return this.ID }

type ComponentConnection struct {
	TotalCount int              `json:"totalCount"`
	Edges      []*ComponentEdge `json:"edges,omitempty"`
	PageInfo   *PageInfo        `json:"pageInfo,omitempty"`
}

func (ComponentConnection) IsConnection()               {}
func (this ComponentConnection) GetTotalCount() int     { return this.TotalCount }
func (this ComponentConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type ComponentEdge struct {
	Node   *Component `json:"node"`
	Cursor *string    `json:"cursor,omitempty"`
}

func (ComponentEdge) IsEdge()                 {}
func (this ComponentEdge) GetNode() Node      { return *this.Node }
func (this ComponentEdge) GetCursor() *string { return this.Cursor }

type ComponentFilter struct {
	ComponentName []*string `json:"componentName,omitempty"`
}

type ComponentInput struct {
	Name *string              `json:"name,omitempty"`
	Type *ComponentTypeValues `json:"type,omitempty"`
}

type ComponentInstance struct {
	ID                 string                `json:"id"`
	Ccrn               *string               `json:"ccrn,omitempty"`
	Count              *int                  `json:"count,omitempty"`
	ComponentVersionID *string               `json:"componentVersionId,omitempty"`
	ComponentVersion   *ComponentVersion     `json:"componentVersion,omitempty"`
	IssueMatches       *IssueMatchConnection `json:"issueMatches,omitempty"`
	ServiceID          *string               `json:"serviceId,omitempty"`
	Service            *Service              `json:"service,omitempty"`
	CreatedAt          *string               `json:"createdAt,omitempty"`
	UpdatedAt          *string               `json:"updatedAt,omitempty"`
}

func (ComponentInstance) IsNode()            {}
func (this ComponentInstance) GetID() string { return this.ID }

type ComponentInstanceConnection struct {
	TotalCount int                      `json:"totalCount"`
	Edges      []*ComponentInstanceEdge `json:"edges"`
	PageInfo   *PageInfo                `json:"pageInfo,omitempty"`
}

func (ComponentInstanceConnection) IsConnection()               {}
func (this ComponentInstanceConnection) GetTotalCount() int     { return this.TotalCount }
func (this ComponentInstanceConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type ComponentInstanceEdge struct {
	Node   *ComponentInstance `json:"node"`
	Cursor *string            `json:"cursor,omitempty"`
}

func (ComponentInstanceEdge) IsEdge()                 {}
func (this ComponentInstanceEdge) GetNode() Node      { return *this.Node }
func (this ComponentInstanceEdge) GetCursor() *string { return this.Cursor }

type ComponentInstanceFilter struct {
	IssueMatchID []*string `json:"issueMatchId,omitempty"`
}

type ComponentInstanceInput struct {
	Ccrn               *string `json:"ccrn,omitempty"`
	Count              *int    `json:"count,omitempty"`
	ComponentVersionID *string `json:"componentVersionId,omitempty"`
	ServiceID          *string `json:"serviceId,omitempty"`
}

type ComponentVersion struct {
	ID                 string                       `json:"id"`
	Version            *string                      `json:"version,omitempty"`
	ComponentID        *string                      `json:"componentId,omitempty"`
	Component          *Component                   `json:"component,omitempty"`
	Issues             *IssueConnection             `json:"issues,omitempty"`
	ComponentInstances *ComponentInstanceConnection `json:"componentInstances,omitempty"`
}

func (ComponentVersion) IsNode()            {}
func (this ComponentVersion) GetID() string { return this.ID }

type ComponentVersionConnection struct {
	TotalCount int                     `json:"totalCount"`
	Edges      []*ComponentVersionEdge `json:"edges"`
	PageInfo   *PageInfo               `json:"pageInfo,omitempty"`
}

func (ComponentVersionConnection) IsConnection()               {}
func (this ComponentVersionConnection) GetTotalCount() int     { return this.TotalCount }
func (this ComponentVersionConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type ComponentVersionEdge struct {
	Node   *ComponentVersion `json:"node"`
	Cursor *string           `json:"cursor,omitempty"`
}

func (ComponentVersionEdge) IsEdge()                 {}
func (this ComponentVersionEdge) GetNode() Node      { return *this.Node }
func (this ComponentVersionEdge) GetCursor() *string { return this.Cursor }

type ComponentVersionFilter struct {
	IssueID []*string `json:"issueId,omitempty"`
	Version []*string `json:"version,omitempty"`
}

type ComponentVersionInput struct {
	Version     *string `json:"version,omitempty"`
	ComponentID *string `json:"componentId,omitempty"`
}

type DateTimeFilter struct {
	After  *string `json:"after,omitempty"`
	Before *string `json:"before,omitempty"`
}

type Evidence struct {
	ID           string                `json:"id"`
	Description  *string               `json:"description,omitempty"`
	Type         *string               `json:"type,omitempty"`
	Vector       *string               `json:"vector,omitempty"`
	RaaEnd       *string               `json:"raaEnd,omitempty"`
	AuthorID     *string               `json:"authorId,omitempty"`
	Author       *User                 `json:"author,omitempty"`
	ActivityID   *string               `json:"activityId,omitempty"`
	Activity     *Activity             `json:"activity,omitempty"`
	IssueMatches *IssueMatchConnection `json:"issueMatches,omitempty"`
}

func (Evidence) IsNode()            {}
func (this Evidence) GetID() string { return this.ID }

type EvidenceConnection struct {
	TotalCount int             `json:"totalCount"`
	Edges      []*EvidenceEdge `json:"edges,omitempty"`
	PageInfo   *PageInfo       `json:"pageInfo,omitempty"`
}

func (EvidenceConnection) IsConnection()               {}
func (this EvidenceConnection) GetTotalCount() int     { return this.TotalCount }
func (this EvidenceConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type EvidenceEdge struct {
	Node   *Evidence `json:"node"`
	Cursor *string   `json:"cursor,omitempty"`
}

func (EvidenceEdge) IsEdge()                 {}
func (this EvidenceEdge) GetNode() Node      { return *this.Node }
func (this EvidenceEdge) GetCursor() *string { return this.Cursor }

type EvidenceFilter struct {
	Placeholder []*bool `json:"placeholder,omitempty"`
}

type EvidenceInput struct {
	Description *string        `json:"description,omitempty"`
	Type        *string        `json:"type,omitempty"`
	RaaEnd      *string        `json:"raaEnd,omitempty"`
	AuthorID    *string        `json:"authorId,omitempty"`
	ActivityID  *string        `json:"activityId,omitempty"`
	Severity    *SeverityInput `json:"severity,omitempty"`
}

type FilterItem struct {
	DisplayName *string   `json:"displayName,omitempty"`
	FilterName  *string   `json:"filterName,omitempty"`
	Values      []*string `json:"values,omitempty"`
}

type Issue struct {
	ID                string                      `json:"id"`
	Type              *IssueTypes                 `json:"type,omitempty"`
	PrimaryName       *string                     `json:"primaryName,omitempty"`
	Description       *string                     `json:"description,omitempty"`
	LastModified      *string                     `json:"lastModified,omitempty"`
	IssueVariants     *IssueVariantConnection     `json:"issueVariants,omitempty"`
	Activities        *ActivityConnection         `json:"activities,omitempty"`
	IssueMatches      *IssueMatchConnection       `json:"issueMatches,omitempty"`
	ComponentVersions *ComponentVersionConnection `json:"componentVersions,omitempty"`
	Metadata          *IssueMetadata              `json:"metadata,omitempty"`
}

func (Issue) IsNode()            {}
func (this Issue) GetID() string { return this.ID }

type IssueConnection struct {
	TotalCount           int          `json:"totalCount"`
	VulnerabilityCount   int          `json:"vulnerabilityCount"`
	PolicyViolationCount int          `json:"policyViolationCount"`
	SecurityEventCount   int          `json:"securityEventCount"`
	Edges                []*IssueEdge `json:"edges"`
	PageInfo             *PageInfo    `json:"pageInfo,omitempty"`
}

func (IssueConnection) IsConnection()               {}
func (this IssueConnection) GetTotalCount() int     { return this.TotalCount }
func (this IssueConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type IssueEdge struct {
	Node   *Issue  `json:"node"`
	Cursor *string `json:"cursor,omitempty"`
}

func (IssueEdge) IsEdge()                 {}
func (this IssueEdge) GetNode() Node      { return *this.Node }
func (this IssueEdge) GetCursor() *string { return this.Cursor }

type IssueFilter struct {
	AffectedService    []*string                 `json:"affectedService,omitempty"`
	PrimaryName        []*string                 `json:"primaryName,omitempty"`
	IssueMatchStatus   []*IssueMatchStatusValues `json:"issueMatchStatus,omitempty"`
	IssueType          []*IssueTypes             `json:"issueType,omitempty"`
	ComponentVersionID []*string                 `json:"componentVersionId,omitempty"`
	Search             []*string                 `json:"search,omitempty"`
}

type IssueInput struct {
	PrimaryName *string     `json:"primaryName,omitempty"`
	Description *string     `json:"description,omitempty"`
	Type        *IssueTypes `json:"type,omitempty"`
}

type IssueMatch struct {
	ID                     string                      `json:"id"`
	Status                 *IssueMatchStatusValues     `json:"status,omitempty"`
	RemediationDate        *string                     `json:"remediationDate,omitempty"`
	DiscoveryDate          *string                     `json:"discoveryDate,omitempty"`
	TargetRemediationDate  *string                     `json:"targetRemediationDate,omitempty"`
	Severity               *Severity                   `json:"severity,omitempty"`
	EffectiveIssueVariants *IssueVariantConnection     `json:"effectiveIssueVariants,omitempty"`
	Evidences              *EvidenceConnection         `json:"evidences,omitempty"`
	IssueID                *string                     `json:"issueId,omitempty"`
	Issue                  *Issue                      `json:"issue"`
	UserID                 *string                     `json:"userId,omitempty"`
	User                   *User                       `json:"user,omitempty"`
	ComponentInstanceID    *string                     `json:"componentInstanceId,omitempty"`
	ComponentInstance      *ComponentInstance          `json:"componentInstance"`
	IssueMatchChanges      *IssueMatchChangeConnection `json:"issueMatchChanges,omitempty"`
}

func (IssueMatch) IsNode()            {}
func (this IssueMatch) GetID() string { return this.ID }

type IssueMatchChange struct {
	ID           string                   `json:"id"`
	Action       *IssueMatchChangeActions `json:"action,omitempty"`
	IssueMatchID *string                  `json:"issueMatchId,omitempty"`
	IssueMatch   *IssueMatch              `json:"issueMatch"`
	ActivityID   *string                  `json:"activityId,omitempty"`
	Activity     *Activity                `json:"activity"`
}

func (IssueMatchChange) IsNode()            {}
func (this IssueMatchChange) GetID() string { return this.ID }

type IssueMatchChangeConnection struct {
	TotalCount int                     `json:"totalCount"`
	Edges      []*IssueMatchChangeEdge `json:"edges,omitempty"`
	PageInfo   *PageInfo               `json:"pageInfo,omitempty"`
}

func (IssueMatchChangeConnection) IsConnection()               {}
func (this IssueMatchChangeConnection) GetTotalCount() int     { return this.TotalCount }
func (this IssueMatchChangeConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type IssueMatchChangeEdge struct {
	Node   *IssueMatchChange `json:"node"`
	Cursor *string           `json:"cursor,omitempty"`
}

func (IssueMatchChangeEdge) IsEdge()                 {}
func (this IssueMatchChangeEdge) GetNode() Node      { return *this.Node }
func (this IssueMatchChangeEdge) GetCursor() *string { return this.Cursor }

type IssueMatchChangeFilter struct {
	Action []*IssueMatchChangeActions `json:"action,omitempty"`
}

type IssueMatchChangeInput struct {
	Action       *IssueMatchChangeActions `json:"action,omitempty"`
	IssueMatchID *string                  `json:"issueMatchId,omitempty"`
	ActivityID   *string                  `json:"activityId,omitempty"`
}

type IssueMatchConnection struct {
	TotalCount int               `json:"totalCount"`
	Edges      []*IssueMatchEdge `json:"edges,omitempty"`
	PageInfo   *PageInfo         `json:"pageInfo,omitempty"`
}

func (IssueMatchConnection) IsConnection()               {}
func (this IssueMatchConnection) GetTotalCount() int     { return this.TotalCount }
func (this IssueMatchConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type IssueMatchEdge struct {
	Node   *IssueMatch `json:"node"`
	Cursor *string     `json:"cursor,omitempty"`
}

func (IssueMatchEdge) IsEdge()                 {}
func (this IssueMatchEdge) GetNode() Node      { return *this.Node }
func (this IssueMatchEdge) GetCursor() *string { return this.Cursor }

type IssueMatchFilter struct {
	ID               []*string                 `json:"id,omitempty"`
	Search           []*string                 `json:"search,omitempty"`
	PrimaryName      []*string                 `json:"primaryName,omitempty"`
	ComponentName    []*string                 `json:"componentName,omitempty"`
	IssueType        []*IssueTypes             `json:"issueType,omitempty"`
	Status           []*IssueMatchStatusValues `json:"status,omitempty"`
	Severity         []*SeverityValues         `json:"severity,omitempty"`
	AffectedService  []*string                 `json:"affectedService,omitempty"`
	SupportGroupName []*string                 `json:"supportGroupName,omitempty"`
}

type IssueMatchFilterValue struct {
	Status           *FilterItem `json:"status,omitempty"`
	Severity         *FilterItem `json:"severity,omitempty"`
	IssueType        *FilterItem `json:"issueType,omitempty"`
	PrimaryName      *FilterItem `json:"primaryName,omitempty"`
	AffectedService  *FilterItem `json:"affectedService,omitempty"`
	ComponentName    *FilterItem `json:"componentName,omitempty"`
	SupportGroupName *FilterItem `json:"supportGroupName,omitempty"`
}

type IssueMatchInput struct {
	Status                *IssueMatchStatusValues `json:"status,omitempty"`
	RemediationDate       *string                 `json:"remediationDate,omitempty"`
	DiscoveryDate         *string                 `json:"discoveryDate,omitempty"`
	TargetRemediationDate *string                 `json:"targetRemediationDate,omitempty"`
	IssueID               *string                 `json:"issueId,omitempty"`
	ComponentInstanceID   *string                 `json:"componentInstanceId,omitempty"`
	UserID                *string                 `json:"userId,omitempty"`
}

type IssueMetadata struct {
	ServiceCount                  int    `json:"serviceCount"`
	ActivityCount                 int    `json:"activityCount"`
	IssueMatchCount               int    `json:"issueMatchCount"`
	ComponentInstanceCount        int    `json:"componentInstanceCount"`
	ComponentVersionCount         int    `json:"componentVersionCount"`
	EarliestDiscoveryDate         string `json:"earliestDiscoveryDate"`
	EarliestTargetRemediationDate string `json:"earliestTargetRemediationDate"`
}

type IssueRepository struct {
	ID            string                  `json:"id"`
	Name          *string                 `json:"name,omitempty"`
	URL           *string                 `json:"url,omitempty"`
	IssueVariants *IssueVariantConnection `json:"issueVariants,omitempty"`
	Services      *ServiceConnection      `json:"services,omitempty"`
	CreatedAt     *string                 `json:"created_at,omitempty"`
	UpdatedAt     *string                 `json:"updated_at,omitempty"`
}

func (IssueRepository) IsNode()            {}
func (this IssueRepository) GetID() string { return this.ID }

type IssueRepositoryConnection struct {
	TotalCount int                    `json:"totalCount"`
	Edges      []*IssueRepositoryEdge `json:"edges,omitempty"`
	PageInfo   *PageInfo              `json:"pageInfo,omitempty"`
}

func (IssueRepositoryConnection) IsConnection()               {}
func (this IssueRepositoryConnection) GetTotalCount() int     { return this.TotalCount }
func (this IssueRepositoryConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type IssueRepositoryEdge struct {
	Node      *IssueRepository `json:"node"`
	Cursor    *string          `json:"cursor,omitempty"`
	Priority  *int             `json:"priority,omitempty"`
	CreatedAt *string          `json:"created_at,omitempty"`
	UpdatedAt *string          `json:"updated_at,omitempty"`
}

func (IssueRepositoryEdge) IsEdge()                 {}
func (this IssueRepositoryEdge) GetNode() Node      { return *this.Node }
func (this IssueRepositoryEdge) GetCursor() *string { return this.Cursor }

type IssueRepositoryFilter struct {
	ServiceName []*string `json:"serviceName,omitempty"`
	ServiceID   []*string `json:"serviceId,omitempty"`
	Name        []*string `json:"name,omitempty"`
}

type IssueRepositoryInput struct {
	Name *string `json:"name,omitempty"`
	URL  *string `json:"url,omitempty"`
}

type IssueVariant struct {
	ID                string           `json:"id"`
	SecondaryName     *string          `json:"secondaryName,omitempty"`
	Description       *string          `json:"description,omitempty"`
	Severity          *Severity        `json:"severity,omitempty"`
	IssueRepositoryID *string          `json:"issueRepositoryId,omitempty"`
	IssueRepository   *IssueRepository `json:"issueRepository,omitempty"`
	IssueID           *string          `json:"issueId,omitempty"`
	Issue             *Issue           `json:"issue,omitempty"`
	CreatedAt         *string          `json:"created_at,omitempty"`
	UpdatedAt         *string          `json:"updated_at,omitempty"`
}

func (IssueVariant) IsNode()            {}
func (this IssueVariant) GetID() string { return this.ID }

type IssueVariantConnection struct {
	TotalCount int                 `json:"totalCount"`
	Edges      []*IssueVariantEdge `json:"edges,omitempty"`
	PageInfo   *PageInfo           `json:"pageInfo,omitempty"`
}

func (IssueVariantConnection) IsConnection()               {}
func (this IssueVariantConnection) GetTotalCount() int     { return this.TotalCount }
func (this IssueVariantConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type IssueVariantEdge struct {
	Node      *IssueVariant `json:"node"`
	Cursor    *string       `json:"cursor,omitempty"`
	CreatedAt *string       `json:"created_at,omitempty"`
	UpdatedAt *string       `json:"updated_at,omitempty"`
}

func (IssueVariantEdge) IsEdge()                 {}
func (this IssueVariantEdge) GetNode() Node      { return *this.Node }
func (this IssueVariantEdge) GetCursor() *string { return this.Cursor }

type IssueVariantFilter struct {
	SecondaryName []*string `json:"secondaryName,omitempty"`
}

type IssueVariantInput struct {
	SecondaryName     *string        `json:"secondaryName,omitempty"`
	Description       *string        `json:"description,omitempty"`
	IssueRepositoryID *string        `json:"issueRepositoryId,omitempty"`
	IssueID           *string        `json:"issueId,omitempty"`
	Severity          *SeverityInput `json:"severity,omitempty"`
}

type Mutation struct {
}

type Page struct {
	After      *string `json:"after,omitempty"`
	IsCurrent  *bool   `json:"isCurrent,omitempty"`
	PageNumber *int    `json:"pageNumber,omitempty"`
	PageCount  *int    `json:"pageCount,omitempty"`
}

type PageInfo struct {
	HasNextPage     *bool   `json:"hasNextPage,omitempty"`
	HasPreviousPage *bool   `json:"hasPreviousPage,omitempty"`
	IsValidPage     *bool   `json:"isValidPage,omitempty"`
	PageNumber      *int    `json:"pageNumber,omitempty"`
	NextPageAfter   *string `json:"nextPageAfter,omitempty"`
	Pages           []*Page `json:"pages,omitempty"`
}

type Query struct {
}

type Service struct {
	ID                 string                       `json:"id"`
	Name               *string                      `json:"name,omitempty"`
	Owners             *UserConnection              `json:"owners,omitempty"`
	SupportGroups      *SupportGroupConnection      `json:"supportGroups,omitempty"`
	Activities         *ActivityConnection          `json:"activities,omitempty"`
	IssueRepositories  *IssueRepositoryConnection   `json:"issueRepositories,omitempty"`
	ComponentInstances *ComponentInstanceConnection `json:"componentInstances,omitempty"`
}

func (Service) IsNode()            {}
func (this Service) GetID() string { return this.ID }

type ServiceConnection struct {
	TotalCount int            `json:"totalCount"`
	Edges      []*ServiceEdge `json:"edges,omitempty"`
	PageInfo   *PageInfo      `json:"pageInfo,omitempty"`
}

func (ServiceConnection) IsConnection()               {}
func (this ServiceConnection) GetTotalCount() int     { return this.TotalCount }
func (this ServiceConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type ServiceEdge struct {
	Node     *Service `json:"node"`
	Cursor   *string  `json:"cursor,omitempty"`
	Priority *int     `json:"priority,omitempty"`
}

func (ServiceEdge) IsEdge()                 {}
func (this ServiceEdge) GetNode() Node      { return *this.Node }
func (this ServiceEdge) GetCursor() *string { return this.Cursor }

type ServiceFilter struct {
	ServiceName      []*string `json:"serviceName,omitempty"`
	UniqueUserID     []*string `json:"uniqueUserId,omitempty"`
	Type             []*int    `json:"type,omitempty"`
	UserName         []*string `json:"userName,omitempty"`
	SupportGroupName []*string `json:"supportGroupName,omitempty"`
}

type ServiceFilterValue struct {
	ServiceName      *FilterItem `json:"serviceName,omitempty"`
	UniqueUserID     *FilterItem `json:"uniqueUserId,omitempty"`
	UserName         *FilterItem `json:"userName,omitempty"`
	SupportGroupName *FilterItem `json:"supportGroupName,omitempty"`
}

type ServiceInput struct {
	Name *string `json:"name,omitempty"`
}

type Severity struct {
	Value *SeverityValues `json:"value,omitempty"`
	Score *float64        `json:"score,omitempty"`
	Cvss  *Cvss           `json:"cvss,omitempty"`
}

type SeverityInput struct {
	Vector *string `json:"vector,omitempty"`
}

type SupportGroup struct {
	ID       string             `json:"id"`
	Name     *string            `json:"name,omitempty"`
	Users    *UserConnection    `json:"users,omitempty"`
	Services *ServiceConnection `json:"services,omitempty"`
}

func (SupportGroup) IsNode()            {}
func (this SupportGroup) GetID() string { return this.ID }

type SupportGroupConnection struct {
	TotalCount int                 `json:"totalCount"`
	Edges      []*SupportGroupEdge `json:"edges,omitempty"`
	PageInfo   *PageInfo           `json:"pageInfo,omitempty"`
}

func (SupportGroupConnection) IsConnection()               {}
func (this SupportGroupConnection) GetTotalCount() int     { return this.TotalCount }
func (this SupportGroupConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type SupportGroupEdge struct {
	Node   *SupportGroup `json:"node"`
	Cursor *string       `json:"cursor,omitempty"`
}

func (SupportGroupEdge) IsEdge()                 {}
func (this SupportGroupEdge) GetNode() Node      { return *this.Node }
func (this SupportGroupEdge) GetCursor() *string { return this.Cursor }

type SupportGroupFilter struct {
	SupportGroupName []*string `json:"supportGroupName,omitempty"`
	UserIds          []*string `json:"userIds,omitempty"`
}

type SupportGroupInput struct {
	Name *string `json:"name,omitempty"`
}

type User struct {
	ID            string                  `json:"id"`
	UniqueUserID  *string                 `json:"uniqueUserId,omitempty"`
	Type          int                     `json:"type"`
	Name          *string                 `json:"name,omitempty"`
	SupportGroups *SupportGroupConnection `json:"supportGroups,omitempty"`
	Services      *ServiceConnection      `json:"services,omitempty"`
}

func (User) IsNode()            {}
func (this User) GetID() string { return this.ID }

type UserConnection struct {
	TotalCount int         `json:"totalCount"`
	Edges      []*UserEdge `json:"edges,omitempty"`
	PageInfo   *PageInfo   `json:"pageInfo,omitempty"`
}

func (UserConnection) IsConnection()               {}
func (this UserConnection) GetTotalCount() int     { return this.TotalCount }
func (this UserConnection) GetPageInfo() *PageInfo { return this.PageInfo }

type UserEdge struct {
	Node   *User   `json:"node"`
	Cursor *string `json:"cursor,omitempty"`
}

func (UserEdge) IsEdge()                 {}
func (this UserEdge) GetNode() Node      { return *this.Node }
func (this UserEdge) GetCursor() *string { return this.Cursor }

type UserFilter struct {
	UserName        []*string `json:"userName,omitempty"`
	SupportGroupIds []*string `json:"supportGroupIds,omitempty"`
	UniqueUserID    []*string `json:"uniqueUserId,omitempty"`
}

type UserInput struct {
	UniqueUserID *string `json:"uniqueUserId,omitempty"`
	Type         *string `json:"type,omitempty"`
	Name         *string `json:"name,omitempty"`
}

type ActivityStatusValues string

const (
	ActivityStatusValuesOpen       ActivityStatusValues = "open"
	ActivityStatusValuesClosed     ActivityStatusValues = "closed"
	ActivityStatusValuesInProgress ActivityStatusValues = "in_progress"
)

var AllActivityStatusValues = []ActivityStatusValues{
	ActivityStatusValuesOpen,
	ActivityStatusValuesClosed,
	ActivityStatusValuesInProgress,
}

func (e ActivityStatusValues) IsValid() bool {
	switch e {
	case ActivityStatusValuesOpen, ActivityStatusValuesClosed, ActivityStatusValuesInProgress:
		return true
	}
	return false
}

func (e ActivityStatusValues) String() string {
	return string(e)
}

func (e *ActivityStatusValues) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ActivityStatusValues(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ActivityStatusValues", str)
	}
	return nil
}

func (e ActivityStatusValues) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ComponentTypeValues string

const (
	ComponentTypeValuesContainerImage      ComponentTypeValues = "containerImage"
	ComponentTypeValuesVirtualMachineImage ComponentTypeValues = "virtualMachineImage"
	ComponentTypeValuesRepository          ComponentTypeValues = "repository"
)

var AllComponentTypeValues = []ComponentTypeValues{
	ComponentTypeValuesContainerImage,
	ComponentTypeValuesVirtualMachineImage,
	ComponentTypeValuesRepository,
}

func (e ComponentTypeValues) IsValid() bool {
	switch e {
	case ComponentTypeValuesContainerImage, ComponentTypeValuesVirtualMachineImage, ComponentTypeValuesRepository:
		return true
	}
	return false
}

func (e ComponentTypeValues) String() string {
	return string(e)
}

func (e *ComponentTypeValues) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ComponentTypeValues(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ComponentTypeValues", str)
	}
	return nil
}

func (e ComponentTypeValues) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type IssueMatchChangeActions string

const (
	IssueMatchChangeActionsAdd    IssueMatchChangeActions = "add"
	IssueMatchChangeActionsRemove IssueMatchChangeActions = "remove"
)

var AllIssueMatchChangeActions = []IssueMatchChangeActions{
	IssueMatchChangeActionsAdd,
	IssueMatchChangeActionsRemove,
}

func (e IssueMatchChangeActions) IsValid() bool {
	switch e {
	case IssueMatchChangeActionsAdd, IssueMatchChangeActionsRemove:
		return true
	}
	return false
}

func (e IssueMatchChangeActions) String() string {
	return string(e)
}

func (e *IssueMatchChangeActions) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = IssueMatchChangeActions(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid IssueMatchChangeActions", str)
	}
	return nil
}

func (e IssueMatchChangeActions) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type IssueMatchStatusValues string

const (
	IssueMatchStatusValuesNew           IssueMatchStatusValues = "new"
	IssueMatchStatusValuesRiskAccepted  IssueMatchStatusValues = "risk_accepted"
	IssueMatchStatusValuesFalsePositive IssueMatchStatusValues = "false_positive"
	IssueMatchStatusValuesMitigated     IssueMatchStatusValues = "mitigated"
)

var AllIssueMatchStatusValues = []IssueMatchStatusValues{
	IssueMatchStatusValuesNew,
	IssueMatchStatusValuesRiskAccepted,
	IssueMatchStatusValuesFalsePositive,
	IssueMatchStatusValuesMitigated,
}

func (e IssueMatchStatusValues) IsValid() bool {
	switch e {
	case IssueMatchStatusValuesNew, IssueMatchStatusValuesRiskAccepted, IssueMatchStatusValuesFalsePositive, IssueMatchStatusValuesMitigated:
		return true
	}
	return false
}

func (e IssueMatchStatusValues) String() string {
	return string(e)
}

func (e *IssueMatchStatusValues) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = IssueMatchStatusValues(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid IssueMatchStatusValues", str)
	}
	return nil
}

func (e IssueMatchStatusValues) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type IssueStatusValues string

const (
	IssueStatusValuesUnaffected IssueStatusValues = "unaffected"
	IssueStatusValuesOpen       IssueStatusValues = "open"
	IssueStatusValuesRemediated IssueStatusValues = "remediated"
	IssueStatusValuesOverdue    IssueStatusValues = "overdue"
)

var AllIssueStatusValues = []IssueStatusValues{
	IssueStatusValuesUnaffected,
	IssueStatusValuesOpen,
	IssueStatusValuesRemediated,
	IssueStatusValuesOverdue,
}

func (e IssueStatusValues) IsValid() bool {
	switch e {
	case IssueStatusValuesUnaffected, IssueStatusValuesOpen, IssueStatusValuesRemediated, IssueStatusValuesOverdue:
		return true
	}
	return false
}

func (e IssueStatusValues) String() string {
	return string(e)
}

func (e *IssueStatusValues) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = IssueStatusValues(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid IssueStatusValues", str)
	}
	return nil
}

func (e IssueStatusValues) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type IssueTypes string

const (
	IssueTypesVulnerability   IssueTypes = "Vulnerability"
	IssueTypesPolicyViolation IssueTypes = "PolicyViolation"
	IssueTypesSecurityEvent   IssueTypes = "SecurityEvent"
)

var AllIssueTypes = []IssueTypes{
	IssueTypesVulnerability,
	IssueTypesPolicyViolation,
	IssueTypesSecurityEvent,
}

func (e IssueTypes) IsValid() bool {
	switch e {
	case IssueTypesVulnerability, IssueTypesPolicyViolation, IssueTypesSecurityEvent:
		return true
	}
	return false
}

func (e IssueTypes) String() string {
	return string(e)
}

func (e *IssueTypes) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = IssueTypes(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid IssueTypes", str)
	}
	return nil
}

func (e IssueTypes) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type SeverityValues string

const (
	SeverityValuesNone     SeverityValues = "None"
	SeverityValuesLow      SeverityValues = "Low"
	SeverityValuesMedium   SeverityValues = "Medium"
	SeverityValuesHigh     SeverityValues = "High"
	SeverityValuesCritical SeverityValues = "Critical"
)

var AllSeverityValues = []SeverityValues{
	SeverityValuesNone,
	SeverityValuesLow,
	SeverityValuesMedium,
	SeverityValuesHigh,
	SeverityValuesCritical,
}

func (e SeverityValues) IsValid() bool {
	switch e {
	case SeverityValuesNone, SeverityValuesLow, SeverityValuesMedium, SeverityValuesHigh, SeverityValuesCritical:
		return true
	}
	return false
}

func (e SeverityValues) String() string {
	return string(e)
}

func (e *SeverityValues) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SeverityValues(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SeverityValues", str)
	}
	return nil
}

func (e SeverityValues) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
