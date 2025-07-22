// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type SeverityValues string

const (
	SeverityValuesNone     SeverityValues = "None"
	SeverityValuesLow      SeverityValues = "Low"
	SeverityValuesMedium   SeverityValues = "Medium"
	SeverityValuesHigh     SeverityValues = "High"
	SeverityValuesCritical SeverityValues = "Critical"
)

func (s SeverityValues) String() string {
	return string(s)
}

type SeverityFilter struct {
	IssueMatchId []*int64 `json:"issue_match_id"`
	IssueId      []*int64 `json:"issue_id"`
}

var AllSeverityValues = []SeverityValues{
	SeverityValuesNone,
	SeverityValuesLow,
	SeverityValuesMedium,
	SeverityValuesHigh,
	SeverityValuesCritical,
}
