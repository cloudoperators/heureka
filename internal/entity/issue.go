// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"time"
)

type IssueWithAggregations struct {
	IssueAggregations
	Issue
}

type IssueType string

const (
	IssueTypeVulnerability   IssueType = "Vulnerability"
	IssueTypePolicyViolation IssueType = "PolicyViolation"
	IssueTypeSecurityEvent   IssueType = "SecurityEvent"
)

func (e IssueType) String() string {
	return string(e)
}

func NewIssueType(s string) IssueType {
	switch s {
	case IssueTypeVulnerability.String():
		return IssueTypeVulnerability
	case IssueTypePolicyViolation.String():
		return IssueTypePolicyViolation
	case IssueTypeSecurityEvent.String():
		return IssueTypeSecurityEvent
	}
	return IssueTypeVulnerability
}

var AllIssueTypes = []string{
	IssueTypeVulnerability.String(),
	IssueTypePolicyViolation.String(),
	IssueTypeSecurityEvent.String(),
}

type IssueResult struct {
	WithCursor
	*IssueAggregations `json:"issue_aggregations"`
	*Issue             `json:"issue"`
	*IssueVariant      `json:"issue_variant"`
}

type IssueStatus int

const (
	IssueStatusAll IssueStatus = iota
	IssueStatusOpen
	IssueStatusRemediated
)

type IssueFilter struct {
	PaginatedX
	PrimaryName                     []*string         `json:"primary_name"`
	ServiceCCRN                     []*string         `json:"service_ccrn"`
	AllServices                     bool              `json:"all_services"`
	HasIssueMatches                 bool              `json:"has_issue_matches"`
	SupportGroupCCRN                []*string         `json:"support_group_ccrn"`
	Type                            []*string         `json:"type"`
	Id                              []*int64          `json:"id"`
	ActivityId                      []*int64          `json:"activity_id"`
	IssueMatchId                    []*int64          `json:"issue_match_id"`
	ComponentId                     []*int64          `json:"component_id"`
	ComponentVersionId              []*int64          `json:"component_version_id"`
	IssueVariantId                  []*int64          `json:"issue_variant_id"`
	IssueRepositoryId               []*int64          `json:"issue_repository_id"`
	ServiceId                       []*int64          `json:"service_id"`
	Search                          []*string         `json:"search"`
	IssueMatchStatus                []*string         `json:"issue_match_status"`
	IssueMatchSeverity              []*string         `json:"issue_match_severity"`
	IssueMatchDiscoveryDate         *TimeFilter       `json:"issue_match_discovery_date"`
	IssueMatchTargetRemediationDate *TimeFilter       `json:"issue_match_target_remediation_date"`
	Unique                          bool              `json:"unique"`
	Status                          IssueStatus       `json:"status"`
	State                           []StateFilterType `json:"state"`
}

type IssueAggregations struct {
	Activities                    int64
	IssueMatches                  int64
	AffectedServices              int64
	AffectedComponentInstances    int64
	ComponentVersions             int64
	EarliestTargetRemediationDate time.Time
	EarliestDiscoveryDate         time.Time
}

type Issue struct {
	Metadata
	Id                int64              `json:"id"`
	Type              IssueType          `json:"type"`
	PrimaryName       string             `json:"primary_name"`
	Description       string             `json:"description"`
	IssueVariants     []IssueVariant     `json:"issue_variants,omitempty"`
	IssueMatches      []IssueMatch       `json:"issue_matches,omitempty"`
	ComponentVersions []ComponentVersion `json:"component_versions,omitempty"`
	Activity          []Activity         `json:"activity,omitempty"`
}

type IssueCount struct {
	Count int64  `json:"issue_count"`
	Value string `json:"issue_value"`
}

type IssueTypeCounts struct {
	VulnerabilityCount   int64 `json:"vulnerability_count"`
	PolicyViolationCount int64 `json:"policy_violation_count"`
	SecurityEventCount   int64 `json:"security_event_count"`
}

type IssueSeverityCounts struct {
	Critical int64 `json:"critical"`
	High     int64 `json:"high"`
	Medium   int64 `json:"medium"`
	Low      int64 `json:"low"`
	None     int64 `json:"none"`
	Total    int64 `json:"total"`
}

func (itc *IssueTypeCounts) TotalIssueCount() int64 {
	return itc.VulnerabilityCount + itc.PolicyViolationCount + itc.SecurityEventCount
}

type IssueList struct {
	*List[IssueResult]
	VulnerabilityCount   *int64
	PolicyViolationCount *int64
	SecurityEventCount   *int64
}

type IssueListOptions struct {
	ListOptions
	ShowIssueTypeCounts bool
}
