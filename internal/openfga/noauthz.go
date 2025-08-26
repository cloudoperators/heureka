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
func (a *NoAuthz) CheckPermission(userId, resourceId, permission string) (bool, error) {
	return true, nil
}

// AddRelation adds a relationship between userId and resourceId.
func (a *NoAuthz) AddRelation(userId, resourceId string, relation string) error {
	return nil
}

// RemoveRelation removes a relationship between userId and resourceId.
func (a *NoAuthz) RemoveRelation(userId, resourceId string, relation string) error {
	return nil
}

// ListAccessibleResources returns a list of resource Ids that the user can access.
func (a *NoAuthz) ListAccessibleResources(userID string, resourceType string, permission string, relation string) ([]string, error) {
	return []string{}, nil
}
