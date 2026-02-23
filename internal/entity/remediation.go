// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type RemediationType string

const (
	RemediationTypeFalsePositive RemediationType = "false_positive"
	RemediationTypeRiskAccepted  RemediationType = "risk_accepted"
	RemediationTypeMitigation    RemediationType = "mitigation"
	RemediationTypeRescore       RemediationType = "rescore"
	RemediationTypeUnknown       RemediationType = "unknown"
)

func (e RemediationType) String() string {
	return string(e)
}

func NewRemediationType(s string) RemediationType {
	switch s {
	case RemediationTypeFalsePositive.String():
		return RemediationTypeFalsePositive
	case RemediationTypeRiskAccepted.String():
		return RemediationTypeRiskAccepted
	case RemediationTypeMitigation.String():
		return RemediationTypeMitigation
	case RemediationTypeRescore.String():
		return RemediationTypeRescore
	}

	return RemediationTypeUnknown
}

var AllRemediationTypes = []string{
	RemediationTypeFalsePositive.String(),
	RemediationTypeRiskAccepted.String(),
	RemediationTypeMitigation.String(),
	RemediationTypeRescore.String(),
}

type Remediation struct {
	Metadata
	Id              int64           `json:"id"`
	Type            RemediationType `json:"type"`
	Description     string          `json:"description"`
	RemediationDate time.Time       `json:"remediation_date"`
	ExpirationDate  time.Time       `json:"expiration_date"`
	Severity        SeverityValues  `json:"severity"`
	Service         string          `json:"service"`
	ServiceId       int64           `json:"service_id"`
	Component       string          `json:"component"`
	ComponentId     int64           `json:"component_id"`
	Issue           string          `json:"issue"`
	IssueId         int64           `json:"issue_id"`
	RemediatedBy    string          `json:"remediated_by"`
	RemediatedById  int64           `json:"remediated_by_id"`
}

type RemediationFilter struct {
	PaginatedX
	Id          []*int64          `json:"id"`
	Severity    []*string         `json:"severity"`
	Service     []*string         `json:"service"`
	ServiceId   []*int64          `json:"service_id"`
	Component   []*string         `json:"component"`
	ComponentId []*int64          `json:"component_id"`
	Issue       []*string         `json:"issue"`
	IssueId     []*int64          `json:"issue_id"`
	Type        []*string         `json:"type"`
	State       []StateFilterType `json:"state"`
	Search      []*string         `json:"search"`
}

func (rf *RemediationFilter) GetPaginatedX() *PaginatedX {
	return &rf.PaginatedX
}

type RemediationResult struct {
	WithCursor
	*Remediation
}
