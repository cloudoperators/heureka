// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package openfga

import (
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/openfga/go-sdk/client"
)

type NoAuthz struct {
	config *util.Config
}

func NewNoAuthz(cfg *util.Config) Authorization {
	return &NoAuthz{
		config: cfg,
	}
}

func (a *NoAuthz) UpdateRelation(r RelationInput) {
}

func (a *NoAuthz) GetCurrentUser() string {
	return ""
}

// CheckPermission checks if userId has permission on resourceId.
func (a *NoAuthz) CheckPermission(p PermissionInput) (bool, error) {
	return false, nil
}

// AddRelation adds a relationship between userId and resourceId.
func (a *NoAuthz) AddRelation(r RelationInput) error {
	return nil
}

// RemoveRelation removes a relationship between userId and resourceId.
func (a *NoAuthz) RemoveRelation(r RelationInput) error {
	return nil
}

// ListAccessibleResources returns a list of resource Ids that the user can access.
func (a *NoAuthz) ListAccessibleResources(p PermissionInput) ([]AccessibleResource, error) {
	resources := []AccessibleResource{}
	return resources, nil
}

// DeleteObjectRelations deletes all relations for a given object.
func (a *NoAuthz) RemoveRelationBulk(input []RelationInput) error {
	return nil
}

// ListRelations lists all relations for a given input.
func (a *NoAuthz) ListRelations(input []RelationInput) ([]client.ClientTupleKeyWithoutCondition, error) {
	return []client.ClientTupleKeyWithoutCondition{}, nil
}
