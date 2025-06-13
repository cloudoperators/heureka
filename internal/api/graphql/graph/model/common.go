// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package model

import "fmt"

type NodeName string

func (e NodeName) String() string {
	return string(e)
}

const (
	ActivityNodeName          NodeName = "Activity"
	IssueVariantNodeName      NodeName = "IssueVariant"
	IssueRepositoryNodeName   NodeName = "IssueRepository"
	ComponentNodeName         NodeName = "Component"
	ComponentInstanceNodeName NodeName = "ComponentInstance"
	ComponentVersionNodeName  NodeName = "ComponentVersion"
	EvidenceNodeName          NodeName = "Evidence"
	ServiceNodeName           NodeName = "Service"
	SupportGroupNodeName      NodeName = "SupportGroup"
	UserNodeName              NodeName = "User"
	IssueNodeName             NodeName = "Issue"
	IssueMatchNodeName        NodeName = "IssueMatch"
	IssueMatchChangeNodeName  NodeName = "IssueMatchChange"
	VulnerabilityNodeName     NodeName = "Vulnerability"
)

type NodeParent struct {
	Parent     Node
	ParentName NodeName
	ChildIds   []*int64
}

func IssueMatchChangeAction(s string) IssueMatchChangeActions {
	switch s {
	case IssueMatchChangeActionsAdd.String():
		return IssueMatchChangeActionsAdd
	case IssueMatchChangeActionsRemove.String():
		return IssueMatchChangeActionsRemove
	}
	return IssueMatchChangeActionsAdd
}

func IssueMatchStatusValue(s string) IssueMatchStatusValues {
	switch s {
	case IssueMatchStatusValuesFalsePositive.String():
		return IssueMatchStatusValuesFalsePositive
	case IssueMatchStatusValuesRiskAccepted.String():
		return IssueMatchStatusValuesRiskAccepted
	case IssueMatchStatusValuesMitigated.String():
		return IssueMatchStatusValuesMitigated
	}
	return IssueMatchStatusValuesNew
}

func SeverityValue(s string) (SeverityValues, error) {
	switch s {
	case SeverityValuesNone.String():
		return SeverityValuesNone, nil
	case SeverityValuesLow.String():
		return SeverityValuesLow, nil
	case SeverityValuesMedium.String():
		return SeverityValuesMedium, nil
	case SeverityValuesHigh.String():
		return SeverityValuesHigh, nil
	case SeverityValuesCritical.String():
		return SeverityValuesCritical, nil
	}
	return "unknown", fmt.Errorf("Invalid SeverityValues provided: %s", s)
}

func ComponentTypeValue(s string) (ComponentTypeValues, error) {
	switch s {
	case ComponentTypeValuesContainerImage.String():
		return ComponentTypeValuesContainerImage, nil
	case ComponentTypeValuesRepository.String():
		return ComponentTypeValuesRepository, nil
	case ComponentTypeValuesVirtualMachineImage.String():
		return ComponentTypeValuesVirtualMachineImage, nil
	}
	return "unknown", fmt.Errorf("Invalid ComponentTypeValues provided: %s", s)
}

func ComponentInstanceType(s string) (ComponentInstanceTypes, error) {
	switch s {
	case ComponentInstanceTypesUnknown.String():
		return ComponentInstanceTypesUnknown, nil
	case ComponentInstanceTypesProject.String():
		return ComponentInstanceTypesProject, nil
	case ComponentInstanceTypesServer.String():
		return ComponentInstanceTypesServer, nil
	case ComponentInstanceTypesSecurityGroup.String():
		return ComponentInstanceTypesSecurityGroup, nil
	case ComponentInstanceTypesDNSZone.String():
		return ComponentInstanceTypesDNSZone, nil
	case ComponentInstanceTypesFloatingIP.String():
		return ComponentInstanceTypesFloatingIP, nil
	case ComponentInstanceTypesRbacPolicy.String():
		return ComponentInstanceTypesRbacPolicy, nil
	case ComponentInstanceTypesUser.String():
		return ComponentInstanceTypesUser, nil
	case ComponentInstanceTypesContainer.String():
		return ComponentInstanceTypesContainer, nil
	case ComponentInstanceTypesRecordSet.String():
		return ComponentInstanceTypesRecordSet, nil
	case ComponentInstanceTypesSecurityGroupRule.String():
		return ComponentInstanceTypesSecurityGroupRule, nil
	}
	return "unknown", fmt.Errorf("Invalid ComponentInstanceType provided: %s", s)
}
