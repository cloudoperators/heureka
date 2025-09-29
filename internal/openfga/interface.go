// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package openfga

import (
	"io/ioutil"

	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"
)

type UserType string
type UserId string
type ObjectType string
type RelationType string
type ObjectId string

type DeleteUserInput struct {
	UserType UserType
	UserId   UserId
}

type DeleteObjectInput struct {
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
	// check if userId has permission on resourceId
	CheckPermission(p PermissionInput) (bool, error)
	// add relationship between userId and resourceId
	AddRelation(r RelationInput) error
	// remove relationship between userId and resourceId
	RemoveRelation(r RelationInput) error
	// ListAccessibleResources returns a list of resource Ids that the user can access.
	ListAccessibleResources(p PermissionInput) ([]AccessibleResource, error)
	// Placeholder function that mimics getting user from User Context
	GetCurrentUser() string
	// Handler for generic event to create authz relation
	HandleCreateAuthzRelation(r RelationInput)
	// Handler for generic event to delete authz relation
	HandleDeleteAuthzRelation(r RelationInput)
	// Handler for generic event to update authz relation
	HandleUpdateAuthzRelation(r RelationInput)
	// Delete all relations for a given user
	DeleteUserRelations(r DeleteUserInput) error
	// Delete all relations for a given object
	DeleteObjectRelations(r DeleteObjectInput) error
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
