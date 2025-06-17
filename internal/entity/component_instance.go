// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type ComponentInstanceType string

const (
	ComponentInstanceTypeUnknown           ComponentInstanceType = "Unknown"
	ComponentInstanceTypeProject           ComponentInstanceType = "Project"
	ComponentInstanceTypeServer            ComponentInstanceType = "Server"
	ComponentInstanceTypeSecurityGroup     ComponentInstanceType = "SecurityGroup"
	ComponentInstanceTypeSecurityGroupRule ComponentInstanceType = "SecurityGroupRule"
	ComponentInstanceTypeDnsZone           ComponentInstanceType = "DnsZone"
	ComponentInstanceTypeFloatingIp        ComponentInstanceType = "FloatingIp"
	ComponentInstanceTypeRbacPolicy        ComponentInstanceType = "RbacPolicy"
	ComponentInstanceTypeUser              ComponentInstanceType = "User"
	ComponentInstanceTypeContainer         ComponentInstanceType = "Container"
	ComponentInstanceTypeRecordSet         ComponentInstanceType = "RecordSet"
)

func (e ComponentInstanceType) String() string {
	return string(e)
}

func (e ComponentInstanceType) Index() int {
	switch e {
	case ComponentInstanceTypeUnknown:
		return 0
	case ComponentInstanceTypeProject:
		return 1
	case ComponentInstanceTypeServer:
		return 2
	case ComponentInstanceTypeSecurityGroup:
		return 3
	case ComponentInstanceTypeSecurityGroupRule:
		return 4
	case ComponentInstanceTypeDnsZone:
		return 5
	case ComponentInstanceTypeFloatingIp:
		return 6
	case ComponentInstanceTypeRbacPolicy:
		return 7
	case ComponentInstanceTypeUser:
		return 8
	case ComponentInstanceTypeContainer:
		return 9
	case ComponentInstanceTypeRecordSet:
		return 10
	default:
		return -1
	}
}

func NewComponentInstanceType(s string) ComponentInstanceType {
	switch s {
	case ComponentInstanceTypeUnknown.String():
		return ComponentInstanceTypeUnknown
	case ComponentInstanceTypeProject.String():
		return ComponentInstanceTypeProject
	case ComponentInstanceTypeServer.String():
		return ComponentInstanceTypeServer
	case ComponentInstanceTypeSecurityGroup.String():
		return ComponentInstanceTypeSecurityGroup
	case ComponentInstanceTypeSecurityGroupRule.String():
		return ComponentInstanceTypeSecurityGroupRule
	case ComponentInstanceTypeDnsZone.String():
		return ComponentInstanceTypeDnsZone
	case ComponentInstanceTypeFloatingIp.String():
		return ComponentInstanceTypeFloatingIp
	case ComponentInstanceTypeRbacPolicy.String():
		return ComponentInstanceTypeRbacPolicy
	case ComponentInstanceTypeUser.String():
		return ComponentInstanceTypeUser
	case ComponentInstanceTypeContainer.String():
		return ComponentInstanceTypeContainer
	case ComponentInstanceTypeRecordSet.String():
		return ComponentInstanceTypeRecordSet
	}
	return ComponentInstanceTypeUnknown
}

var AllComponentInstanceType = []string{
	ComponentInstanceTypeUnknown.String(),
	ComponentInstanceTypeProject.String(),
	ComponentInstanceTypeServer.String(),
	ComponentInstanceTypeSecurityGroup.String(),
	ComponentInstanceTypeSecurityGroupRule.String(),
	ComponentInstanceTypeDnsZone.String(),
	ComponentInstanceTypeFloatingIp.String(),
	ComponentInstanceTypeRbacPolicy.String(),
	ComponentInstanceTypeUser.String(),
	ComponentInstanceTypeContainer.String(),
	ComponentInstanceTypeRecordSet.String(),
}

type ComponentInstanceFilter struct {
	PaginatedX
	IssueMatchId            []*int64          `json:"issue_match_id"`
	ServiceId               []*int64          `json:"service_id"`
	ServiceCcrn             []*string         `json:"service_ccrn"`
	ComponentVersionId      []*int64          `json:"component_version_id"`
	ComponentVersionVersion []*string         `json:"component_version_version"`
	Id                      []*int64          `json:"id"`
	CCRN                    []*string         `json:"ccrn"`
	Region                  []*string         `json:"region"`
	Cluster                 []*string         `json:"cluster"`
	Namespace               []*string         `json:"namespace"`
	Domain                  []*string         `json:"domain"`
	Project                 []*string         `json:"project"`
	Pod                     []*string         `json:"pod"`
	Container               []*string         `json:"container"`
	Type                    []*string         `json:"type"`
	ParentId                []*int64          `json:"parent_id"`
	Context                 []*Json           `json:"context"`
	Search                  []*string         `json:"search"`
	State                   []StateFilterType `json:"state"`
}

type ComponentInstanceAggregations struct {
}

type ComponentInstanceResult struct {
	WithCursor
	*ComponentInstance
	*ComponentInstanceAggregations
}

type ComponentInstance struct {
	Metadata
	Id                 int64                 `json:"id"`
	CCRN               string                `json:"ccrn"`
	Region             string                `json:"region"`
	Cluster            string                `json:"cluster"`
	Namespace          string                `json:"namespace"`
	Domain             string                `json:"domain"`
	Project            string                `json:"project"`
	Pod                string                `json:"pod"`
	Container          string                `json:"container"`
	Type               ComponentInstanceType `json:"type"`
	ParentId           int64                 `json:"parent_id,omitempty"`
	Context            *Json                 `json:"context"`
	Count              int16                 `json:"count"`
	ComponentVersion   *ComponentVersion     `json:"component_version,omitempty"`
	ComponentVersionId int64                 `db:"componentinstance_component_version_id"`
	Service            *Service              `json:"service,omitempty"`
	ServiceId          int64                 `db:"componentinstance_service_id"`
}
