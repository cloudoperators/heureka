// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package openfga

import (
	"context"
	"encoding/json"
	"errors"
	"os"
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
	storeId, err := checkStore(fgaClient, cfg.AuthzOpenFgaStoreName)
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
	modelId, err := checkModel(fgaClient, storeId)
	if err != nil {
		l.Error("Could not list OpenFGA models: ", err)
		return nil
	}
	if modelId == "" {
		// model does not exist, create it
		// Create the authorization model request from the model file
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

// checkStore checks if a store with the given name exists in OpenFGA.
func checkStore(fgaClient *client.OpenFgaClient, storeName string) (string, error) {
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

// checkModel checks if an authorization model exists in OpenFGA for the given store.
func checkModel(fgaClient *client.OpenFgaClient, storeId string) (string, error) {
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

// checkTuple checks if a specific tuple exists in OpenFGA.
func (a *Authz) checkTuple(r RelationInput) (bool, error) {
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
		a.logger.Errorf("OpenFGA Read (checkTuple) error: %v", err)
		return false, err
	}

	return len(resp.Tuples) > 0, nil
}

// CheckPermission checks if userId has permission on objectId.
func (a *Authz) CheckPermission(r RelationInput) (bool, error) {
	req := client.ClientCheckRequest{
		User:     string(r.UserType) + ":" + string(r.UserId),
		Relation: string(r.Relation),
		Object:   string(r.ObjectType) + ":" + string(r.ObjectId),
	}
	resp, err := a.client.Check(context.Background()).Body(req).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA Check error: %v", err)
		return false, err
	}
	return resp.GetAllowed(), nil
}

// AddRelation adds a specified relationship between userId and objectId.
func (a *Authz) AddRelation(r RelationInput) error {
	l := a.logRel("HandleAddAuthzRelation", r)

	// Check to avoid duplicate writes
	ok, err := a.checkTuple(r)
	if err != nil {
		l.WithField("event-step", "OpenFGA Read").
			WithError(err).
			Error("Failed to read relation before add")
		return err
	}
	if ok {
		l.WithField("event-step", "OpenFGA Read").
			Info("Relation already exists; skipping add")
		return nil
	}

	// Write relation
	tuple := client.ClientWriteRequest{
		Writes: []client.ClientTupleKey{
			{
				User:     string(r.UserType) + ":" + string(r.UserId),
				Relation: string(r.Relation),
				Object:   string(r.ObjectType) + ":" + string(r.ObjectId),
			},
		},
	}
	_, err = a.client.Write(context.Background()).Body(tuple).Execute()
	if err != nil {
		l.WithField("event-step", "OpenFGA AddRelation").
			WithError(err).
			Errorf("Error while adding relation tuple: (%s, %s, %s, %s)", r.UserId, r.ObjectId, r.ObjectType, r.Relation)
		return err
	}

	l.WithField("event-step", "OpenFGA AddRelation").
		Infof("Added relation tuple: (%s, %s, %s, %s)", r.UserId, r.ObjectId, r.ObjectType, r.Relation)
	return nil
}

// AddRelationBulk adds multiple specified relationships between userId(s) and objectId(s).
func (a *Authz) AddRelationBulk(relations []RelationInput) error {
	l := a.logger.WithFields(logrus.Fields{
		"event":         "HandleAddAuthzRelationBulk",
		"relationCount": len(relations),
	})

	options := client.ClientWriteOptions{
		Conflict: client.ClientWriteConflictOptions{
			OnDuplicateWrites: client.CLIENT_WRITE_REQUEST_ON_DUPLICATE_WRITES_IGNORE,
		},
	}

	tupleStrings := make([]client.ClientTupleKey, 0, len(relations))
	for _, rel := range relations {
		tupleStrings = append(tupleStrings, client.ClientTupleKey{
			User:     string(rel.UserType) + ":" + string(rel.UserId),
			Relation: string(rel.Relation),
			Object:   string(rel.ObjectType) + ":" + string(rel.ObjectId),
		})
	}

	tuple := client.ClientWriteRequest{
		Writes: tupleStrings,
	}

	_, err := a.client.Write(context.Background()).Body(tuple).Options(options).Execute()
	if err != nil {
		l.WithField("event-step", "OpenFGA AddRelationsBulk").
			WithError(err).
			Error("Failed to add relations")
		return err
	}

	l.WithField("event-step", "OpenFGA AddRelationsBulk").
		WithField("added", len(tupleStrings)).
		Info("Added relations")
	return nil
}

// RemoveRelation removes a relationship between userId and objectId.
func (a *Authz) RemoveRelation(r RelationInput) error {
	l := a.logRel("HandleRemoveAuthzRelation", r)

	// Check existence before delete
	ok, err := a.checkTuple(r)
	if err != nil {
		l.WithField("event-step", "OpenFGA Read").
			WithError(err).
			Error("Failed to read relation for deletion")
		return err
	}
	if !ok {
		l.WithField("event-step", "OpenFGA Read").
			Info("No matching relation to delete")
		return nil
	}

	// Delete the relation
	writeReq := client.ClientWriteRequest{
		Deletes: []client.ClientTupleKeyWithoutCondition{
			{
				User:     string(r.UserType) + ":" + string(r.UserId),
				Relation: string(r.Relation),
				Object:   string(r.ObjectType) + ":" + string(r.ObjectId),
			},
		},
	}
	options := client.ClientWriteOptions{
		Conflict: client.ClientWriteConflictOptions{
			OnMissingDeletes: client.CLIENT_WRITE_REQUEST_ON_MISSING_DELETES_IGNORE,
		},
	}
	_, err = a.client.Write(context.Background()).Body(writeReq).Options(options).Execute()
	if err != nil {
		l.WithField("event-step", "OpenFGA DeleteRelations").
			WithError(err).
			Errorf("Error while deleting relation tuple: (%s, %s, %s, %s)", r.UserId, r.ObjectId, r.ObjectType, r.Relation)
		return err
	}

	l.WithField("event-step", "OpenFGA DeleteRelations").
		WithField("deleted", 1).
		Infof("Deleted relation tuple: (%s, %s, %s, %s)", r.UserId, r.ObjectId, r.ObjectType, r.Relation)
	return nil
}

// RemoveRelationBulk removes all relations that match the given RelationInput as filters.
func (a *Authz) RemoveRelationBulk(r []RelationInput) error {
	l := a.logger.WithFields(logrus.Fields{
		"event":       "HandleRemoveAuthzRelationBulk",
		"filterCount": len(r),
	})

	// Collect all matching tuples for given filters
	tuples := []client.ClientTupleKeyWithoutCondition{}
	for _, rel := range r {
		found, err := a.ListRelations(rel)
		if err != nil {
			l.WithField("event-step", "OpenFGA ListRelations").
				WithError(err).
				Error("Failed to read relations for deletion")
			return err
		}
		tuples = append(tuples, found...)
	}

	if len(tuples) == 0 {
		l.WithField("event-step", "OpenFGA ListRelations").
			Info("No matching relations to delete")
		return nil
	}

	writeReq := client.ClientWriteRequest{
		Deletes: tuples,
	}
	options := client.ClientWriteOptions{
		Conflict: client.ClientWriteConflictOptions{
			OnMissingDeletes: client.CLIENT_WRITE_REQUEST_ON_MISSING_DELETES_IGNORE,
		},
	}
	_, err := a.client.Write(context.Background()).Body(writeReq).Options(options).Execute()
	if err != nil {
		l.WithField("event-step", "OpenFGA DeleteRelations").
			WithField("attemptedDeletes", len(tuples)).
			WithError(err).
			Error("Failed to delete relations")
		return err
	}

	l.WithField("event-step", "OpenFGA DeleteRelations").
		WithField("deleted", len(tuples)).
		Info("Deleted relations")
	return nil
}

// UpdateRelation updates relations by removing relations that match the filter for the old relation and adding the new relation.
func (a *Authz) UpdateRelation(add RelationInput, rem RelationInput) error {
	l := a.logRel("HandleUpdateAuthzRelation", rem)

	if err := a.RemoveRelationBulk([]RelationInput{rem}); err != nil {
		l.WithField("event-step", "OpenFGA RemoveRelationBulk").
			WithError(err).
			Errorf("Error while removing relation tuple: (%s, %s, %s, %s)", rem.UserId, rem.ObjectId, rem.ObjectType, rem.Relation)
		return err
	}
	l.WithField("event-step", "OpenFGA RemoveRelationBulk").
		Infof("Removed relation tuple: (%s, %s, %s, %s)", rem.UserId, rem.ObjectId, rem.ObjectType, rem.Relation)

	if err := a.AddRelation(add); err != nil {
		// switch log context to the add relation
		a.logRel("HandleUpdateAuthzRelation", add).
			WithField("event-step", "OpenFGA AddRelation").
			WithError(err).
			Errorf("Error while adding relation tuple: (%s, %s, %s, %s)", add.UserId, add.ObjectId, add.ObjectType, add.Relation)
		return err
	}
	a.logRel("HandleUpdateAuthzRelation", add).
		WithField("event-step", "OpenFGA AddRelation").
		Infof("Added relation tuple: (%s, %s, %s, %s)", add.UserId, add.ObjectId, add.ObjectType, add.Relation)

	return nil
}

// ListRelations lists all relations that match any given RelationInput as filter(s)
func (a *Authz) ListRelations(filter RelationInput) ([]client.ClientTupleKeyWithoutCondition, error) {
	// openfga POST read relation tuple requirements to be checked for:
	// tuple_key (the filter object itself) is optional, if not provided, all tuples are returned
	// object is mandatory if a tuple_key is provided, but objectId is not necessary, just a type can be specified
	// user is mandatory only if object is specified in type only (if object type and id are both specified, user is optional)
	// if user is specified, it must have both type and id, not just a type or id alone

	// convert relationinput filters to a client read request
	var userStr, relationStr, objectStr string

	if filter.UserType != "" && filter.UserId != "" {
		userStr = string(filter.UserType) + ":" + string(filter.UserId)
	}
	if filter.Relation != "" {
		relationStr = string(filter.Relation)
	}
	if filter.ObjectType != "" {
		objectStr = string(filter.ObjectType) + ":"
		if filter.ObjectId != "" {
			objectStr += string(filter.ObjectId)
		}
	} else {
		return nil, errors.New("objectType must be specified in the filter")
	}

	req := client.ClientReadRequest{
		User:     &userStr,
		Relation: &relationStr,
		Object:   &objectStr,
	}
	resp, err := a.client.Read(context.Background()).Body(req).Execute()
	if err != nil {
		a.logger.Errorf("OpenFGA Read (ListRelations) error: %v", err)
		return nil, err
	}

	// convert response tuples to []client.ClientTupleKeyWithoutCondition
	var tuples []client.ClientTupleKeyWithoutCondition
	for _, tuple := range resp.Tuples {
		tuples = append(tuples, client.ClientTupleKeyWithoutCondition{
			User:     tuple.Key.User,
			Relation: tuple.Key.Relation,
			Object:   tuple.Key.Object,
		})
	}
	return tuples, nil
}

// ListAccessibleResources returns a list of objectIds of a certain objectType that the user can access.
func (a *Authz) ListAccessibleResources(r RelationInput) ([]AccessibleResource, error) {
	body := client.ClientListObjectsRequest{
		User:     string(r.UserType) + ":" + string(r.UserId),
		Relation: string(r.Relation),
		Type:     string(r.ObjectType),
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

	return resources, nil
}

// RemoveAllRelations removes all tuples for the store backing the provided Authorization.
func (a *Authz) RemoveAllRelations() error {
	if a == nil {
		return nil
	}

	ctx := context.Background()
	options := client.ClientWriteOptions{
		Conflict: client.ClientWriteConflictOptions{
			OnMissingDeletes: client.CLIENT_WRITE_REQUEST_ON_MISSING_DELETES_IGNORE,
		},
	}

	resp, err := a.client.Read(ctx).Body(client.ClientReadRequest{}).Execute()
	if err != nil {
		return err
	}
	if len(resp.Tuples) == 0 {
		return nil
	}
	deletes := make([]client.ClientTupleKeyWithoutCondition, 0, len(resp.Tuples))
	for _, t := range resp.Tuples {
		deletes = append(deletes, client.ClientTupleKeyWithoutCondition{
			User:     t.Key.User,
			Relation: t.Key.Relation,
			Object:   t.Key.Object,
		})
	}
	_, err = a.client.Write(ctx).Body(client.ClientWriteRequest{Deletes: deletes}).Options(options).Execute()
	if err != nil {
		return err
	}
	return nil
}
