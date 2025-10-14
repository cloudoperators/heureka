// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package openfga

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cloudoperators/heureka/internal/util"
	"github.com/sirupsen/logrus"

	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
	"github.com/openfga/language/pkg/go/transformer"
)

type Authz struct {
	config *util.Config
	logger *logrus.Logger
	client *client.OpenFgaClient
}

// Creates new Authorization implement using OpenFGA
func NewAuthz(l *logrus.Logger, cfg *util.Config) Authorization {
	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: cfg.AuthzOpenFgaApiUrl,
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodApiToken,
			Config: &credentials.Config{
				ApiToken: cfg.AuthzOpenFgaApiToken,
			},
		},
	})
	if err != nil {
		l.Error("Could not initialize OpenFGA client: ", err)
		return nil
	}

	// Check if the store already exists, otherwise create it
	storeId, err := CheckStore(fgaClient, cfg.AuthzOpenFgaStoreName)
	if err != nil {
		l.Error("Could not list OpenFGA stores: ", err)
		return nil
	}
	if storeId == "" {
		// store does not exist, create it
		store, err := fgaClient.CreateStore(context.Background()).Body(client.ClientCreateStoreRequest{Name: cfg.AuthzOpenFgaStoreName}).Execute()
		if err != nil {
			l.Error("Could not create OpenFGA store: ", err)
			return nil
		}
		storeId = store.Id
	}
	// update the storeId of the current instance
	fgaClient.SetStoreId(storeId)

	// Check if the model already exists, otherwise create it
	modelId, err := CheckModel(fgaClient, storeId)
	if err != nil {
		l.Error("Could not list OpenFGA models: ", err)
		return nil
	}
	if modelId == "" {
		// model does not exist, create it
		// Create the authorization model request from the model file
		cwd, err := os.Getwd()
		if err != nil {
			l.Error("Could not get current working directory: ", err)
			return nil
		}
		fmt.Print(cwd)
		modelRequest, err := getAuthModelRequestFromFile(cfg.AuthzModelFilePath)
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
		modelId = modelResponse.AuthorizationModelId
	}
	// update the modelId of the current instance
	fgaClient.SetAuthorizationModelId(modelId)

	l.Info("Initializing authorization with OpenFGA")
	return &Authz{config: cfg, logger: l, client: fgaClient}
}

func (a *Authz) GetCurrentUser() string {
	return a.config.CurrentUser
}

func (a *Authz) ListRelations(filters []RelationInput) ([]client.ClientTupleKeyWithoutCondition, error) {
	req := client.ClientReadRequest{} // Empty request returns all tuples
	resp, err := a.client.Read(context.Background()).Body(req).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA Read (ListRelations) error: %v", err)
		return nil, err
	}

	var tuples []client.ClientTupleKeyWithoutCondition
	for _, tuple := range resp.Tuples {
		userParts := strings.SplitN(tuple.Key.User, ":", 2)
		objectParts := strings.SplitN(tuple.Key.Object, ":", 2)

		for _, r := range filters {
			if r.UserType != "" && (len(userParts) < 1 || userParts[0] != string(r.UserType)) {
				continue
			}
			if r.UserId != "" && (len(userParts) < 2 || userParts[1] != string(r.UserId)) {
				continue
			}
			if r.Relation != "" && tuple.Key.Relation != string(r.Relation) {
				continue
			}
			if r.ObjectType != "" && (len(objectParts) < 1 || objectParts[0] != string(r.ObjectType)) {
				continue
			}
			if r.ObjectId != "" && (len(objectParts) < 2 || objectParts[1] != string(r.ObjectId)) {
				continue
			}

			// If all specified fields match for this filter, add the tuple and break to avoid duplicates
			tuples = append(tuples, client.ClientTupleKeyWithoutCondition{
				User:     tuple.Key.User,
				Relation: tuple.Key.Relation,
				Object:   tuple.Key.Object,
			})
			break
		}
	}
	return tuples, nil
}

func (a *Authz) RemoveRelationBulk(r []RelationInput) error {
	tuples, err := a.ListRelations(r)
	if err != nil {
		return err
	}

	if len(tuples) == 0 {
		return nil
	}

	writeReq := client.ClientWriteRequest{
		Deletes: tuples,
	}
	_, err = a.client.Write(context.Background()).Body(writeReq).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA Delete (DeleteRelations) error: %v", err)
	} else {
		a.logger.Infof("OpenFGA Delete (DeleteRelations): Deleted %d relations", len(tuples))
	}
	return err
}

func (a *Authz) UpdateRelation(r RelationInput) {
	l := logrus.WithFields(logrus.Fields{
		"event":            "HandleCreateAuthzRelation",
		"user":             r.UserId,
		"resourceId":       r.ObjectId,
		"resourceType":     r.ObjectType,
		"resourceRelation": r.Relation,
	})

	err := a.RemoveRelation(r)
	err = a.AddRelation(r)
	if err != nil {
		l.WithField("event-step", "OpenFGA AddRelation").WithError(err).Errorf("Error while adding relation tuple: (%s, %s, %s, %s)", r.UserId, r.ObjectId, r.ObjectType, r.Relation)
	} else {
		l.WithField("event-step", "OpenFGA AddRelation").Infof("Added relation tuple: (%s, %s, %s, %s)", r.UserId, r.ObjectId, r.ObjectType, r.Relation)
	}
}

// Reads the authorization model from a file, before creating the model in OpenFGA
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

// CheckStore checks if a store with the given name exists in OpenFGA.
func CheckStore(fgaClient *client.OpenFgaClient, storeName string) (string, error) {
	storesResponse, err := fgaClient.ListStores(context.Background()).Execute()
	if err != nil {
		return "", err
	}
	for _, s := range storesResponse.Stores {
		if s.Name == storeName {
			return s.Id, nil
		}
	}
	return "", nil
}

// CheckModel checks if an authorization model exists in OpenFGA for the given store.
func CheckModel(fgaClient *client.OpenFgaClient, storeId string) (string, error) {
	modelsResponse, err := fgaClient.ReadAuthorizationModels(context.Background()).Options(
		client.ClientReadAuthorizationModelsOptions{StoreId: &storeId},
	).Execute()
	if err != nil {
		return "", err
	}
	if len(modelsResponse.AuthorizationModels) > 0 {
		return modelsResponse.AuthorizationModels[0].Id, nil
	}
	return "", nil
}

// CheckTuple checks if a specific tuple exists in OpenFGA.
func (a *Authz) CheckTuple(r RelationInput) (bool, error) {
	userString := string(r.UserType) + ":" + string(r.UserId)
	relationString := string(r.Relation)
	objectString := string(r.ObjectType) + ":" + string(r.ObjectId)

	req := client.ClientReadRequest{
		User:     &userString,
		Relation: &relationString,
		Object:   &objectString,
	}
	resp, err := a.client.Read(context.Background()).Body(req).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA Read (CheckTuple) error: %v", err)
		return false, err
	}

	return len(resp.Tuples) > 0, nil
}

func (a *Authz) CheckPermission(p PermissionInput) (bool, error) {
	req := client.ClientCheckRequest{
		User:     string(p.UserType) + ":" + string(p.UserId),
		Relation: string(p.Relation),
		Object:   string(p.ObjectType) + ":" + string(p.ObjectId),
	}
	resp, err := a.client.Check(context.Background()).Body(req).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA Check error: %v", err)
		return false, err
	}
	return resp.GetAllowed(), nil
}

// AddRelation adds a relationship between userId and resourceId.
func (a *Authz) AddRelation(r RelationInput) error {
	if ok, err := a.CheckTuple(r); err != nil {
		return err
	} else if !ok {
		tuple := client.ClientWriteRequest{
			Writes: []client.ClientTupleKey{
				{
					User:     string(r.UserType) + ":" + string(r.UserId),
					Relation: string(r.Relation),
					Object:   string(r.ObjectType) + ":" + string(r.ObjectId),
				},
			},
		}
		resp, err := a.client.Write(context.Background()).Body(tuple).Execute()
		if err != nil {
			a.logger.Errorf("OpenFGA Write (AddRelation) error: %v", err)
		} else {
			a.logger.Infof("OpenFGA Write (AddRelation): %v | Added relation %s for user %s on resource %s", resp, r.Relation, r.UserId, r.ObjectId)
		}
		return err
	} else {
		a.logger.Infof("Relation %s for user %s on resource %s already exists", r.Relation, r.UserId, r.ObjectId)
	}
	return nil
}

// RemoveRelation removes a relationship between userId and resourceId.
func (a *Authz) RemoveRelation(r RelationInput) error {
	if ok, err := a.CheckTuple(r); err != nil {
		return err
	} else if ok {
		tuple := client.ClientWriteRequest{
			Deletes: []client.ClientTupleKeyWithoutCondition{
				{
					User:     string(r.UserType) + ":" + string(r.UserId),
					Relation: string(r.Relation),
					Object:   string(r.ObjectType) + ":" + string(r.ObjectId),
				},
			},
		}
		_, err := a.client.Write(context.Background()).Body(tuple).Execute()
		if err != nil {
			a.logger.Errorf("OpenFGA Write (RemoveRelation) error: %v", err)
		}
		return err
	} else {
		a.logger.Infof("Relation %s for user %s on resource %s doesn't exist", r.Relation, r.UserId, r.ObjectId)
	}
	return nil
}

// ListAccessibleResources returns a list of resource Ids that the user can access.
func (a *Authz) ListAccessibleResources(p PermissionInput) ([]AccessibleResource, error) {
	body := client.ClientListObjectsRequest{
		User:     string(p.UserType) + ":" + string(p.UserId),
		Relation: string(p.Relation),
		Type:     string(p.ObjectType),
	}

	resp, err := a.client.ListObjects(context.Background()).Body(body).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA ListObjects error: %v", err)
		return nil, err
	}

	var resources []AccessibleResource
	for _, obj := range resp.Objects {
		// Split the object string into type and id (e.g., "document:document1")
		parts := strings.SplitN(obj, ":", 2)
		if len(parts) == 2 {
			resources = append(resources, AccessibleResource{
				ObjectType: ObjectType(parts[0]),
				ObjectId:   ObjectId(parts[1]),
			})
		}
	}

	// if resources is empty, add a -1 resource to indicate no access
	if len(resources) == 0 {
		resources = append(resources, AccessibleResource{
			ObjectType: p.ObjectType,
			ObjectId:   ObjectId("-1"),
		})
	}

	return resources, nil
}

// GetListOfAccessibleObjectIds returns a list of object Ids of a given type that the user can access.
func (a *Authz) GetListOfAccessibleObjectIds(userId UserId, objectType ObjectType) ([]*int64, error) {
	permission := PermissionInput{
		UserType:   "user",
		UserId:     userId,
		Relation:   "can_view",
		ObjectType: objectType,
		ObjectId:   "*",
	}

	// Get all services the user has access to
	accessibleServices, err := a.ListAccessibleResources(permission)
	if err != nil {
		return nil, err
	}

	// Convert []openfga.ObjectId to []int64
	var ids []*int64
	for _, resource := range accessibleServices {
		if intId, err := strconv.ParseInt(string(resource.ObjectId), 10, 64); err == nil {
			ids = append(ids, &intId)
		}
	}

	return ids, nil
}
