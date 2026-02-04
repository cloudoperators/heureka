// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package openfga

import (
	"io/ioutil"

	"github.com/cloudoperators/heureka/internal/util"
	"github.com/openfga/go-sdk/client"
	"github.com/sirupsen/logrus"
)

type UserType string
type UserId string
type ObjectType string
type RelationType string
type ObjectId string

// IDs (shared across userId/objectId tuple definitions)
const (
	IDUser              = "userID"
	IDRole              = "roleID"
	IDSupportGroup      = "supportGroupID"
	IDService           = "serviceID"
	IDComponent         = "componentID"
	IDComponentVersion  = "componentVersionID"
	IDComponentInstance = "componentInstanceID"
	IDIssueMatch        = "issueMatchID"
)

// Types (shared across userType/objectType tuple definitions)
const (
	TypeUser              = "user"
	TypeRole              = "role"
	TypeSupportGroup      = "support_group"
	TypeService           = "service"
	TypeComponent         = "component"
	TypeComponentVersion  = "component_version"
	TypeComponentInstance = "component_instance"
	TypeIssueMatch        = "issue_match"
)

// Relations (shared across relations tuple definitions)
const (
	RelCanView           = "can_view"
	RelRole              = "role"
	RelSupportGroup      = "support_group"
	RelRelatedService    = "related_service"
	RelOwner             = "owner"
	RelAdmin             = "admin"
	RelMember            = "member"
	RelComponentInstance = "component_instance"
	RelComponentVersion  = "component_version"
)

type RelationInput struct {
	UserType   UserType
	UserId     UserId
	Relation   RelationType
	ObjectType ObjectType
	ObjectId   ObjectId
}

type AccessibleResource struct {
	ObjectType ObjectType
	ObjectId   ObjectId
}

type Authorization interface {
	// Check if userId has permission on resourceId
	CheckPermission(r RelationInput) (bool, error)
	// Add relationship between userId and resourceId
	AddRelation(r RelationInput) error
	// Add multiple relationships between userId and resourceId
	AddRelationBulk(r []RelationInput) error
	// Remove a single relationship between userId and resourceId
	RemoveRelation(r RelationInput) error
	// Remove all relations that match any given RelationInput as filters
	RemoveRelationBulk(r []RelationInput) error
	// Remove all relations in the authorization store
	RemoveAllRelations() error
	// Update relations based on filters provided
	UpdateRelation(r RelationInput, u RelationInput) error
	// List Relations based on multiple filters
	ListRelations(filters RelationInput) ([]client.ClientTupleKeyWithoutCondition, error)
	// ListAccessibleResources returns a list of resource Ids that the user can access.
	ListAccessibleResources(r RelationInput) ([]AccessibleResource, error)
}

func NewAuthorizationHandler(cfg *util.Config, enablelog bool) Authorization {
	l := newLogger(enablelog)

	if cfg.AuthzOpenFgaApiUrl != "" {
		return NewAuthz(l, cfg)
	}

	return NewNoAuthz(cfg)
}

func newLogger(enableLog bool) *logrus.Logger {
	l := logrus.New()
	if !enableLog {
		l.SetOutput(ioutil.Discard)
	}
	return l
}
