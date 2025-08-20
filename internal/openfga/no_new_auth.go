package openfga

import (
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
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

// EventInjection adapts to func type type EventHandlerFunc func(database.Database, Event) in order to inject events.
func (a *NoAuthz) EventInjection(db database.Database, e event.Event) {
	// TODO
}

// CheckPermission checks if userId has permission on resourceId.
func (a *NoAuthz) CheckPermission(userId, resourceId, permission string) (bool, error) {
	// TODO
	return true, nil
}

// AddRelation adds a relationship between userId and resourceId.
func (a *NoAuthz) AddRelation(userId, resourceId string) error {
	// TODO
	return nil
}

// RemoveRelation removes a relationship between userId and resourceId.
func (a *NoAuthz) RemoveRelation(userId, resourceId string) error {
	// TODO
	return nil
}

// ListAccessibleResources returns a list of resource Ids that the user can access.
func (a *NoAuthz) ListAccessibleResources(userID string, resourceType string, permission string, relation string) ([]string, error) {
	// TODO
	return []string{}, nil
}
