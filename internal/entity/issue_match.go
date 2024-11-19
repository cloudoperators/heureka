// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type IssueMatchStatusValue string

const (
	IssueMatchStatusValuesNew           IssueMatchStatusValue = "new"
	IssueMatchStatusValuesRiskAccepted  IssueMatchStatusValue = "risk_accepted"
	IssueMatchStatusValuesFalsePositive IssueMatchStatusValue = "false_positive"
	IssueMatchStatusValuesMitigated     IssueMatchStatusValue = "mitigated"
	IssueMatchStatusValuesNone          IssueMatchStatusValue = "none"
)

func (e IssueMatchStatusValue) String() string {
	return string(e)
}

func NewIssueMatchStatusValue(s string) IssueMatchStatusValue {
	switch s {
	case IssueMatchStatusValuesRiskAccepted.String():
		return IssueMatchStatusValuesRiskAccepted
	case IssueMatchStatusValuesFalsePositive.String():
		return IssueMatchStatusValuesFalsePositive
	case IssueMatchStatusValuesMitigated.String():
		return IssueMatchStatusValuesMitigated
	}
	return IssueMatchStatusValuesNew
}

var AllIssueMatchStatusValues = []string{
	IssueMatchStatusValuesNew.String(),
	IssueMatchStatusValuesRiskAccepted.String(),
	IssueMatchStatusValuesFalsePositive.String(),
	IssueMatchStatusValuesMitigated.String(),
}

type IssueMatch struct {
	Metadata
	Id                    int64                 `json:"id"`
	Status                IssueMatchStatusValue `json:"status"`
	User                  *User                 `json:"user,omitempty"`
	UserId                int64                 `json:"user_id"`
	Severity              Severity              `json:"severity,omitempty"`
	Evidences             []Evidence            `json:"evidence,omitempty"`
	ComponentInstance     *ComponentInstance    `json:"component_instance,omitempty"`
	ComponentInstanceId   int64                 `json:"component_instance_id"`
	Issue                 *Issue                `json:"issue,omitempty"`
	IssueId               int64                 `json:"issue_id"`
	RemediationDate       time.Time             `json:"remediation_date"`
	TargetRemediationDate time.Time             `json:"target_remediation_date"`
}

type IssueMatchFilter struct {
	Paginated
	Id                  []*int64  `json:"id"`
	AffectedServiceCCRN []*string `json:"affected_service_ccrn"`
	SeverityValue       []*string `json:"severity_value"`
	Status              []*string `json:"status"`
	IssueId             []*int64  `json:"issue_id"`
	EvidenceId          []*int64  `json:"evidence_id"`
	ComponentInstanceId []*int64  `json:"component_instance_id"`
	SupportGroupCCRN    []*string `json:"support_group_ccrn"`
	Search              []*string `json:"search"`
	ComponentCCRN       []*string `json:"component_ccrn"`
	PrimaryName         []*string `json:"primary_name"`
	IssueType           []*string `json:"issue_type"`
}

type IssueMatchResult struct {
	WithCursor
	*IssueMatch
}
