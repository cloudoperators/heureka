// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package entity

type ComponentInstanceType string

const (
	ComponentInstanceTypeUnknown       ComponentInstanceType = "Unknown"
	ComponentInstanceTypeProject       ComponentInstanceType = "Project"
	ComponentInstanceTypeServer        ComponentInstanceType = "Server"
	ComponentInstanceTypeSecurityGroup ComponentInstanceType = "SecurityGroup"
	ComponentInstanceTypeDnsZone       ComponentInstanceType = "DnsZone"
	ComponentInstanceTypeFloatingIp    ComponentInstanceType = "FloatingIp"
	ComponentInstanceTypeRbacPolicy    ComponentInstanceType = "RbacPolicy"
	ComponentInstanceTypeUser          ComponentInstanceType = "User"
	ComponentInstanceTypeContainer     ComponentInstanceType = "Container"
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
	case ComponentInstanceTypeDnsZone:
		return 4
	case ComponentInstanceTypeFloatingIp:
		return 5
	case ComponentInstanceTypeRbacPolicy:
		return 6
	case ComponentInstanceTypeUser:
		return 7
	case ComponentInstanceTypeContainer:
		return 8
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
	}
	return ComponentInstanceTypeUnknown
}

var AllComponentInstanceType = []string{
	ComponentInstanceTypeUnknown.String(),
	ComponentInstanceTypeProject.String(),
	ComponentInstanceTypeServer.String(),
	ComponentInstanceTypeSecurityGroup.String(),
	ComponentInstanceTypeDnsZone.String(),
	ComponentInstanceTypeFloatingIp.String(),
	ComponentInstanceTypeRbacPolicy.String(),
	ComponentInstanceTypeUser.String(),
	ComponentInstanceTypeContainer.String(),
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
	Search                  []*string         `json:"search"`
	State                   []StateFilterType `json:"state"`
	ParentId                []*int64          `json:"parent_id"`
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
	Count              int16                 `json:"count"`
	ComponentVersion   *ComponentVersion     `json:"component_version,omitempty"`
	ComponentVersionId int64                 `db:"componentinstance_component_version_id"`
	Service            *Service              `json:"service,omitempty"`
	ServiceId          int64                 `db:"componentinstance_service_id"`
	ParentId           int64                 `db:"componentinstance_parent_id"`
}
