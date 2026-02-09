// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

import "time"

type EvidenceType string

const (
	EvidenceTypeValuesRiskAccepted       EvidenceType = "risk_accepted"
	EvidenceTypeValuesMitigated          EvidenceType = "mitigated"
	EvidenceTypeValuesSeverityAdjustment EvidenceType = "severity_adjustment"
	EvidenceTypeValuesFalsePositive      EvidenceType = "false_positive"
	EvidenceTypeValuesReOpen             EvidenceType = "reopen"
)

func (e EvidenceType) String() string {
	return string(e)
}

func NewEvidenceTypeValue(s string) EvidenceType {
	switch s {
	case EvidenceTypeValuesRiskAccepted.String():
		return EvidenceTypeValuesRiskAccepted
	case EvidenceTypeValuesFalsePositive.String():
		return EvidenceTypeValuesFalsePositive
	case EvidenceTypeValuesMitigated.String():
		return EvidenceTypeValuesMitigated
	case EvidenceTypeValuesSeverityAdjustment.String():
		return EvidenceTypeValuesSeverityAdjustment
	}
	return EvidenceTypeValuesReOpen
}

var AllEvidenceTypeValues = []string{
	EvidenceTypeValuesReOpen.String(),
	EvidenceTypeValuesRiskAccepted.String(),
	EvidenceTypeValuesFalsePositive.String(),
	EvidenceTypeValuesMitigated.String(),
	EvidenceTypeValuesSeverityAdjustment.String(),
}

type Evidence struct {
	Metadata
	Id          int64        `json:"id"`
	Description string       `json:"description"`
	Type        EvidenceType `json:"type"`
	RaaEnd      time.Time    `json:"raa_end"`
	Severity    Severity     `json:"severity"`
	User        *User        `json:"user,omitempty"`
	UserId      int64        `db:"evidence_author_id"`
}

type EvidenceFilter struct {
	Paginated
	Id           []*int64          `json:"id"`
	IssueMatchId []*int64          `json:"issue_match_id"`
	UserId       []*int64          `json:"user_id"`
	State        []StateFilterType `json:"state"`
}
type EvidenceAggregations struct{}

type EvidenceResult struct {
	WithCursor
	*Evidence
	*EvidenceAggregations
}
