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

type DeleteRelationInput struct {
	UserType   UserType
	UserId     UserId
	Relation   RelationType
	ObjectType ObjectType
	ObjectId   ObjectId
}

type PermissionInput struct {
	UserType   UserType
	UserId     UserId
	Relation   RelationType
	ObjectType ObjectType
	ObjectId   ObjectId
}

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
	CheckPermission(p PermissionInput) (bool, error)
	// Add relationship between userId and resourceId
	AddRelation(r RelationInput) error
	// Remove a single relationship between userId and resourceId
	RemoveRelation(r RelationInput) error
	// Remove all relations that match any given RelationInput as filters
	RemoveRelationBulk(r []RelationInput) error
	// Handler for generic event to update authz relation
	UpdateRelation(r RelationInput)
	// List Relations based on multiple filters
	ListRelations(filters []RelationInput) ([]client.ClientTupleKeyWithoutCondition, error)
	// ListAccessibleResources returns a list of resource Ids that the user can access.
	ListAccessibleResources(p PermissionInput) ([]AccessibleResource, error)
	// Placeholder function that mimics getting user from User Context
	GetCurrentUser() string
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
