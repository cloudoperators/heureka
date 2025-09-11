package openfga

import (
	"github.com/cloudoperators/heureka/internal/util"
)

type NoAuthz struct {
	config *util.Config
}

func NewNoAuthz(cfg *util.Config) Authorization {
	return &NoAuthz{
		config: cfg,
	}
}

// CheckPermission checks if userId has permission on resourceId.
func (a *NoAuthz) CheckPermission(p PermissionInput) (bool, error) {
	return true, nil
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
