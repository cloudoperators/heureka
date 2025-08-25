package openfga

import (
	"context"
	"os"

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

// CheckPermission checks if userId has permission on resourceId.
func (a *Authz) CheckPermission(userId, resourceId, permission string) (bool, error) {
	req := client.ClientCheckRequest{
		User:     userId,
		Object:   resourceId,
		Relation: permission,
	}
	resp, err := a.client.Check(context.Background()).Body(req).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA Check error: %v", err)
		return false, err
	}
	return resp.GetAllowed(), nil
}

// AddRelation adds a relationship between userId and resourceId.
func (a *Authz) AddRelation(userId, resourceId, relation string) error {
	tuple := client.ClientWriteRequest{
		Writes: []client.ClientTupleKey{
			{
				User:     userId,
				Relation: relation,
				Object:   resourceId,
			},
		},
	}
	resp, err := a.client.Write(context.Background()).Body(tuple).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA Write (AddRelation) error: %v", err)
	} else {
		a.logger.Infof("OpenFGA Write (AddRelation): %v | Added relation %s for user %s on resource %s", resp, relation, userId, resourceId)
	}
	return err

}

// RemoveRelation removes a relationship between userId and resourceId.
func (a *Authz) RemoveRelation(userId, resourceId, relation string) error {
	tuple := client.ClientWriteRequest{
		Deletes: []client.ClientTupleKeyWithoutCondition{
			{
				User:     userId,
				Relation: relation,
				Object:   resourceId,
			},
		},
	}
	_, err := a.client.Write(context.Background()).Body(tuple).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA Write (RemoveRelation) error: %v", err)
	}
	return err
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
