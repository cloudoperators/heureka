// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type OrderByField int

const (
	ComponentInstanceCcrn OrderByField = iota
	ComponentInstanceId
	ComponentInstanceRegion
	ComponentInstanceCluster
	ComponentInstanceNamespace
	ComponentInstanceDomain
	ComponentInstanceProject
	ComponentInstancePod
	ComponentInstanceContainer
	ComponentInstanceTypeOrder

	ComponentVersionId
	ComponentVersionRepository

	ComponentId
	ComponentCcrn
	ComponentRepository

	IssueId
	IssuePrimaryName

	IssueVariantRating

	IssueMatchId
	IssueMatchRating
	IssueMatchTargetRemediationDate

	CriticalCount
	HighCount
	MediumCount
	LowCount
	NoneCount

	RemediationId
	PatchId

	RemediationIssue
	RemediationSeverity

	SupportGroupId
	SupportGroupCcrn

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
