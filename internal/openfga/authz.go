package openfga

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"

	"github.com/openfga/go-sdk/client"
	"github.com/openfga/language/pkg/go/transformer"
)

const (
	fgaStoreName = "heureka-store-Final"
)

type Authz struct {
	config      *util.Config
	logger      *logrus.Logger
	client      *client.OpenFgaClient
	currentUser string
}

func (a *Authz) GetCurrentUser() string {
	return a.config.CurrentUser
}

func (a *Authz) HandleCreateAuthzRelation(
	userFieldName string,
	user string,
	resourceId string,
	resourceType string,
	resourceRelation string,
) {
	l := logrus.WithFields(logrus.Fields{
		"event":            "HandleCreateAuthzRelation",
		"user":             user,
		"resourceId":       resourceId,
		"resourceType":     resourceType,
		"resourceRelation": resourceRelation,
	})

	err := a.AddRelation(userFieldName, user, resourceId, resourceType, resourceRelation)
	if err != nil {
		l.WithField("event-step", "OpenFGA AddRelation").WithError(err).Errorf("Error while adding relation tuple: (%s, %s, %s, %s)", user, resourceId, resourceType, resourceRelation)
	} else {
		l.WithField("event-step", "OpenFGA AddRelation").Infof("Added relation tuple: (%s, %s, %s, %s)", user, resourceId, resourceType, resourceRelation)
	}
}

func getAuthModelRequestFromFile(filePath string) (*client.ClientWriteAuthorizationModelRequest, error) {
	modelBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	modelJson, err := transformer.TransformDSLToJSON(string(modelBytes))
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON into the WriteAuthorizationModelRequest struct
	var modelRequest client.ClientWriteAuthorizationModelRequest
	err = json.Unmarshal([]byte(modelJson), &modelRequest)
	if err != nil {
		return nil, err
	}

	return &modelRequest, nil
}

func NewAuthz(l *logrus.Logger, cfg *util.Config) Authorization {
	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: cfg.OpenFGApiUrl,
	})
	if err != nil {
		l.Error("Could not initialize OpenFGA client: ", err)
		return nil
	}

	store, err := fgaClient.CreateStore(context.Background()).Body(client.ClientCreateStoreRequest{Name: fgaStoreName}).Execute()
	if err != nil {
		l.Error("Could not create OpenFGA store: ", err)
		return nil
	}

	// update the storeId of the current instance
	fgaClient.SetStoreId(store.Id)

	// Create the authorization model request from the model file
	modelRequest, err := getAuthModelRequestFromFile(cfg.AuthModelFilePath)
	if err != nil {
		l.Error("Could not parse OpenFGA model file: ", err)
		return nil
	}

	// Create the authorization model
	modelResponse, err := fgaClient.WriteAuthorizationModel(context.Background()).Body(*modelRequest).Execute()
	if err != nil {
		l.Error("Could not create OpenFGA authorization model: ", err)
		return nil
	}

	// update the modelId of the current instance
	fgaClient.SetAuthorizationModelId(modelResponse.AuthorizationModelId)

	l.Info("Initializing authorization with OpenFGA")
	return &Authz{config: cfg, logger: l, client: fgaClient, currentUser: cfg.CurrentUser}
}

// CheckPermission checks if userId has permission on resourceId.
func (a *Authz) CheckPermission(userFieldName string, userId string, resourceId string, resourceType string, permission string) (bool, error) {
	req := client.ClientCheckRequest{
		User:     userFieldName + ":" + userId,
		Relation: permission,
		Object:   resourceType + ":" + resourceId,
	}
	resp, err := a.client.Check(context.Background()).Body(req).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA Check error: %v", err)
		return false, err
	}
	return resp.GetAllowed(), nil
}

// AddRelation adds a relationship between userId and resourceId.
func (a *Authz) AddRelation(userFieldName string, userId string, resourceId string, resourceType string, relation string) error {
	tuple := client.ClientWriteRequest{
		Writes: []client.ClientTupleKey{
			{
				User:     userFieldName + ":" + userId,
				Relation: relation,
				Object:   resourceType + ":" + resourceId,
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
func (a *Authz) RemoveRelation(userFieldName string, userId string, resourceId string, resourceType string, relation string) error {
	tuple := client.ClientWriteRequest{
		Deletes: []client.ClientTupleKeyWithoutCondition{
			{
				User:     userFieldName + ":" + userId,
				Relation: relation,
				Object:   resourceType + ":" + resourceId,
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
func (a *Authz) ListAccessibleResources(userFieldName string, userId string, resourceType string, permission string, relation string) ([]string, error) {
	body := client.ClientListObjectsRequest{
		User:     userFieldName + ":" + userId,
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
