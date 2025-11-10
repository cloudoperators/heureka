// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type RemediationType string

const (
	RemediationTypeFalsePositive RemediationType = "false_positive"
	RemediationTypeUnknown       RemediationType = "unknown"
)

func (e RemediationType) String() string {
	return string(e)
}

func NewRemediationType(s string) RemediationType {
	switch s {
	case RemediationTypeFalsePositive.String():
		return RemediationTypeFalsePositive
	}
	return RemediationTypeUnknown
}

var AllRemediationTypes = []string{
	RemediationTypeFalsePositive.String(),
}

type Remediation struct {
	Metadata
	Id              int64           `json:"id"`
	Type            RemediationType `json:"type"`
	Description     string          `json:"description"`
	RemediationDate time.Time       `json:"remediation_date"`
	ExpirationDate  time.Time       `json:"expiration_date"`
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
	Service     []*string         `json:"service"`
	ServiceId   []*int64          `json:"service_id"`
	Component   []*string         `json:"component"`
	ComponentId []*int64          `json:"component_id"`
	Issue       []*string         `json:"issue"`
	IssueId     []*int64          `json:"issue_id"`
	Type        []*string         `json:"type"`
	State       []StateFilterType `json:"state"`
}

type RemediationResult struct {
	WithCursor
	*Remediation
}
