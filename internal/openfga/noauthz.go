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

func (a *NoAuthz) HandleCreateAuthzRelation(
	userFieldName string,
	user string,
	resourceId string,
	resourceType string,
	resourceRelation string,
) {

}

func (a *NoAuthz) GetCurrentUser() string {
	return ""
}

// CheckPermission checks if userId has permission on resourceId.
func (a *NoAuthz) CheckPermission(userFieldName string, userId string, resourceId string, resourceType string, permission string) (bool, error) {
	return true, nil
}

// AddRelation adds a relationship between userId and resourceId.
func (a *NoAuthz) AddRelation(userFieldName string, userId string, resourceId string, resourceType string, relation string) error {
	return nil
}

// RemoveRelation removes a relationship between userId and resourceId.
func (a *NoAuthz) RemoveRelation(userFieldName string, userId string, resourceId string, resourceType string, relation string) error {
	return nil
}

// ListAccessibleResources returns a list of resource Ids that the user can access.
func (a *NoAuthz) ListAccessibleResources(userFieldName string, userId string, resourceType string, permission string, relation string) ([]string, error) {
	return []string{}, nil
}
