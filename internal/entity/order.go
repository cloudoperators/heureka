// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type OrderByField int

const (
	ComponentInstanceCcrn OrderByField = iota

	IssuePrimaryName

	IssueMatchId
	IssueMatchRating
	IssueMatchTargetRemediationDate

	SupportGroupName

	ServiceId
	ServiceCcrn
)

type OrderDirection int

const (
	OrderDirectionAsc OrderDirection = iota
	OrderDirectionDesc
)

type Order struct {
	By        OrderByField
	Direction OrderDirection
}
