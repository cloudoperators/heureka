package openfga

import (
	"context"
	"os"

	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/openfga/go-sdk/client"
	"github.com/sirupsen/logrus"
)

type Authz struct {
	config *util.Config
	logger *logrus.Logger
	client *client.OpenFgaClient
}

func NewAuthz(l *logrus.Logger, cfg *util.Config) Authorization {
	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl:               os.Getenv("FGA_API_URL"),  // required
		StoreId:              os.Getenv("FGA_STORE_ID"), // optional
		AuthorizationModelId: os.Getenv("FGA_MODEL_ID"), // Optional
	})
	if err != nil {
		l.Error("Could not initialize OpenFGA client: ", err)
		return nil
	}

	if cfg.AuthzEnabled {
		l.Info("Initializing authorization with OpenFGA")
		return &Authz{config: cfg, logger: l, client: fgaClient}
	}
	return nil
}

// EventInjection adapts to func type type EventHandlerFunc func(database.Database, Event) in order to inject events.
func (a *Authz) EventInjection(db database.Database, e event.Event) {
	userid := "user:system"
	resourceType := "role"
	permission := "read"
	relation := "admin"

	resp, err := a.ListAccessibleResources(userid, resourceType, permission, relation)
	if err != nil {
		a.logger.Errorf("Error listing accessible resources: %v", err)
		return
	}

	a.logger.Infof("Accessible resources for user %s: %v", userid, resp)
}

// CheckPermission checks if userId has permission on resourceId.
func (a *Authz) CheckPermission(userId, resourceId, permission string) (bool, error) {
	// TODO: Implement actual authorization logic.
	return true, nil
}

// AddRelation adds a relationship between userId and resourceId.
func (a *Authz) AddRelation(userId, resourceId string) error {
	// TODO: Implement actual logic to add relation.
	return nil
}

// RemoveRelation removes a relationship between userId and resourceId.
func (a *Authz) RemoveRelation(userId, resourceId string) error {
	// TODO: Implement actual logic to remove relation.
	return nil
}

// ListAccessibleResources returns a list of resource Ids that the user can access.
func (a *Authz) ListAccessibleResources(userID string, resourceType string, permission string, relation string) ([]string, error) {
	body := client.ClientListObjectsRequest{
		User:     userID,
		Relation: relation,
		Type:     resourceType,
	}

	resp, err := a.client.ListObjects(context.Background()).Body(body).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA ListObjects error: %v", err)
		return nil, err
	}

	return resp.Objects, nil
}
