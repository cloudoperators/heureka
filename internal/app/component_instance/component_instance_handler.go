// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import (
	"time"

	"errors"
	"fmt"
	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/cache"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/sirupsen/logrus"
	"strconv"
)

var CacheTtlGetComponentInstances = 12 * time.Hour
var CacheTtlGetAllComponentInstanceCursors = 12 * time.Hour
var CacheTtlCountComponentInstances = 12 * time.Hour

type componentInstanceHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
	cache         cache.Cache
	logger        *logrus.Logger
}

func NewComponentInstanceHandler(database database.Database, eventRegistry event.EventRegistry, cache cache.Cache) ComponentInstanceHandler {
	return &componentInstanceHandler{
		database:      database,
		eventRegistry: eventRegistry,
		cache:         cache,
		logger:        logrus.New(),
	}
}

func (ci *componentInstanceHandler) ListComponentInstances(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) (*entity.List[entity.ComponentInstanceResult], error) {
	op := appErrors.Op("componentInstanceHandler.ListComponentInstances")
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginatedX(&filter.PaginatedX)

	res, err := cache.CallCached[[]entity.ComponentInstanceResult](
		ci.cache,
		CacheTtlGetComponentInstances,
		"GetComponentInstances",
		ci.database.GetComponentInstances,
		filter,
		options.Order,
	)

	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstances", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := cache.CallCached[[]string](
				ci.cache,
				CacheTtlGetAllComponentInstanceCursors,
				"GetAllComponentInstanceCursors",
				ci.database.GetAllComponentInstanceCursors,
				filter,
				options.Order,
			)
			if err != nil {
				wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceCursors", "", err)
				applog.LogError(ci.logger, wrappedErr, logrus.Fields{
					"filter": filter,
				})
				return nil, wrappedErr
			}
			pageInfo = common.GetPageInfoX(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = cache.CallCached[int64](
			ci.cache,
			CacheTtlCountComponentInstances,
			"CountComponentInstances",
			ci.database.CountComponentInstances,
			filter,
		)
		if err != nil {
			wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceCount", "", err)
			applog.LogError(ci.logger, wrappedErr, logrus.Fields{
				"filter": filter,
			})
			return nil, wrappedErr
		}
	}

	result := &entity.List[entity.ComponentInstanceResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}

	ci.eventRegistry.PushEvent(&ListComponentInstancesEvent{
		Filter:             filter,
		Options:            options,
		ComponentInstances: result,
	})

	return result, nil
}

func (ci *componentInstanceHandler) CreateComponentInstance(componentInstance *entity.ComponentInstance, scannerRunUUID *string) (*entity.ComponentInstance, error) {
	op := appErrors.Op("componentInstanceHandler.CreateComponentInstance")

	// Input validation - check for required fields
	if componentInstance == nil {
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, "component instance cannot be nil")
		applog.LogError(ci.logger, err, logrus.Fields{})
		return nil, err
	}

	if componentInstance.CCRN == "" {
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, "CCRN is required")
		applog.LogError(ci.logger, err, logrus.Fields{
			"component_instance": componentInstance,
		})
		return nil, err
	}

	if componentInstance.ComponentVersionId <= 0 {
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, "valid component version ID is required")
		applog.LogError(ci.logger, err, logrus.Fields{
			"component_version_id": componentInstance.ComponentVersionId,
			"ccrn":                 componentInstance.CCRN,
		})
		return nil, err
	}

	if componentInstance.ServiceId <= 0 {
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, "valid service ID is required")
		applog.LogError(ci.logger, err, logrus.Fields{
			"service_id": componentInstance.ServiceId,
			"ccrn":       componentInstance.CCRN,
		})
		return nil, err
	}

	// Business rule validation - ParentId validation for specific types
	if err := validateParentIdForType(componentInstance.ParentId, componentInstance.Type.String()); err != nil {
		wrappedErr := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, err.Error())
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"parent_id":        componentInstance.ParentId,
			"type":             componentInstance.Type.String(),
			"ccrn":             componentInstance.CCRN,
			"validation_error": err.Error(),
		})
		return nil, wrappedErr
	}

	// Get current user for audit fields
	var err error
	componentInstance.CreatedBy, err = common.GetCurrentUserId(ci.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"ccrn": componentInstance.CCRN,
			"type": componentInstance.Type.String(),
		})
		return nil, wrappedErr
	}
	componentInstance.UpdatedBy = componentInstance.CreatedBy

	// Create the component instance in database
	newComponentInstance, err := ci.database.CreateComponentInstance(componentInstance)
	if err != nil {
		// Check for specific database errors
		duplicateEntryError := &database.DuplicateEntryDatabaseError{}
		if errors.As(err, &duplicateEntryError) {
			wrappedErr := appErrors.AlreadyExistsError(string(op), "ComponentInstance", componentInstance.CCRN)
			applog.LogError(ci.logger, wrappedErr, logrus.Fields{
				"ccrn":                    componentInstance.CCRN,
				"component_version_id":    componentInstance.ComponentVersionId,
				"service_id":              componentInstance.ServiceId,
				"duplicate_entry_details": duplicateEntryError.Error(),
			})
			return nil, wrappedErr
		}

		// Generic database error
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"ccrn":                 componentInstance.CCRN,
			"component_version_id": componentInstance.ComponentVersionId,
			"service_id":           componentInstance.ServiceId,
			"type":                 componentInstance.Type.String(),
		})
		return nil, wrappedErr
	}

	// Handle scanner run tracking if UUID provided
	if scannerRunUUID != nil {
		err = ci.database.CreateScannerRunComponentInstanceTracker(newComponentInstance.Id, *scannerRunUUID)
		if err != nil {
			// Log the error but don't fail the creation since the component instance was created successfully
			logErr := appErrors.InternalError(string(op), "ScannerRunComponentInstanceTracker",
				fmt.Sprintf("component_instance:%d-scanner_run:%s", newComponentInstance.Id, *scannerRunUUID), err)
			applog.LogError(ci.logger, logErr, logrus.Fields{
				"component_instance_id": newComponentInstance.Id,
				"scanner_run_uuid":      *scannerRunUUID,
				"ccrn":                  newComponentInstance.CCRN,
			})
			// Note: We don't return this error since the main operation succeeded
		}
	}

	// Emit success event
	ci.eventRegistry.PushEvent(&CreateComponentInstanceEvent{
		ComponentInstance: newComponentInstance,
	})

	return newComponentInstance, nil
}

func (ci *componentInstanceHandler) UpdateComponentInstance(componentInstance *entity.ComponentInstance, scannerRunUUID *string) (*entity.ComponentInstance, error) {
	op := appErrors.Op("componentInstanceHandler.UpdateComponentInstance")

	// Input validation
	if componentInstance == nil {
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, "component instance cannot be nil")
		applog.LogError(ci.logger, err, logrus.Fields{})
		return nil, err
	}

	if componentInstance.Id <= 0 {
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, fmt.Sprintf("invalid ID: %d", componentInstance.Id))
		applog.LogError(ci.logger, err, logrus.Fields{"id": componentInstance.Id})
		return nil, err
	}

	// Business rule validation - ParentId validation for specific types
	if err := validateParentIdForType(componentInstance.ParentId, componentInstance.Type.String()); err != nil {
		wrappedErr := appErrors.E(op, "ComponentInstance", strconv.FormatInt(componentInstance.Id, 10), appErrors.InvalidArgument, err.Error())
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"id":               componentInstance.Id,
			"parent_id":        componentInstance.ParentId,
			"type":             componentInstance.Type.String(),
			"validation_error": err.Error(),
		})
		return nil, wrappedErr
	}

	// Get current user for audit fields
	var err error
	componentInstance.UpdatedBy, err = common.GetCurrentUserId(ci.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", strconv.FormatInt(componentInstance.Id, 10), err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"id":   componentInstance.Id,
			"ccrn": componentInstance.CCRN,
		})
		return nil, wrappedErr
	}

	// Update the component instance in database
	err = ci.database.UpdateComponentInstance(componentInstance)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", strconv.FormatInt(componentInstance.Id, 10), err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"id":   componentInstance.Id,
			"ccrn": componentInstance.CCRN,
		})
		return nil, wrappedErr
	}

	// Handle scanner run tracking if UUID provided
	if scannerRunUUID != nil {
		err = ci.database.CreateScannerRunComponentInstanceTracker(componentInstance.Id, *scannerRunUUID)
		if err != nil {
			// Log the error but don't fail the update since the component instance was updated successfully
			logErr := appErrors.InternalError(string(op), "ScannerRunComponentInstanceTracker",
				fmt.Sprintf("component_instance:%d-scanner_run:%s", componentInstance.Id, *scannerRunUUID), err)
			applog.LogError(ci.logger, logErr, logrus.Fields{
				"component_instance_id": componentInstance.Id,
				"scanner_run_uuid":      *scannerRunUUID,
				"ccrn":                  componentInstance.CCRN,
			})
			// Note: We don't return this error since the main operation succeeded
		}
	}

	// Retrieve updated component instance to return fresh data
	lo := entity.NewListOptions()
	componentInstanceResult, err := ci.ListComponentInstances(&entity.ComponentInstanceFilter{Id: []*int64{&componentInstance.Id}}, lo)
	if err != nil {
		wrappedErr := appErrors.E(op, "ComponentInstance", strconv.FormatInt(componentInstance.Id, 10), appErrors.Internal, err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"id":   componentInstance.Id,
			"ccrn": componentInstance.CCRN,
		})
		return nil, wrappedErr
	}

	if len(componentInstanceResult.Elements) != 1 {
		err := appErrors.E(op, "ComponentInstance", strconv.FormatInt(componentInstance.Id, 10), appErrors.Internal,
			fmt.Sprintf("unexpected number of component instances found after update: expected 1, got %d", len(componentInstanceResult.Elements)))
		applog.LogError(ci.logger, err, logrus.Fields{
			"id":          componentInstance.Id,
			"found_count": len(componentInstanceResult.Elements),
			"ccrn":        componentInstance.CCRN,
		})
		return nil, err
	}

	updatedComponentInstance := componentInstanceResult.Elements[0].ComponentInstance

	// Emit success event
	ci.eventRegistry.PushEvent(&UpdateComponentInstanceEvent{
		ComponentInstance: updatedComponentInstance,
	})

	return updatedComponentInstance, nil
}

func (ci *componentInstanceHandler) DeleteComponentInstance(id int64) error {
	op := appErrors.Op("componentInstanceHandler.DeleteComponentInstance")

	// Input validation
	if id <= 0 {
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, fmt.Sprintf("invalid ID: %d", id))
		applog.LogError(ci.logger, err, logrus.Fields{"id": id})
		return err
	}

	// Get current user for audit fields
	userId, err := common.GetCurrentUserId(ci.database)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", strconv.FormatInt(id, 10), err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"id": id,
		})
		return wrappedErr
	}

	// Delete the component instance from database
	err = ci.database.DeleteComponentInstance(id, userId)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", strconv.FormatInt(id, 10), err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"id":      id,
			"user_id": userId,
		})
		return wrappedErr
	}

	// Emit success event
	ci.eventRegistry.PushEvent(&DeleteComponentInstanceEvent{
		ComponentInstanceID: id,
	})

	return nil
}

func (ci *componentInstanceHandler) ListRegions(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListRegions")

	regions, err := ci.database.GetRegion(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceRegions", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListRegionsEvent{
		Filter:  filter,
		Regions: regions,
	})

	return regions, nil
}

func (ci *componentInstanceHandler) ListCcrns(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListCcrns")

	ccrns, err := ci.database.GetCcrn(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceCcrns", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListCcrnEvent{
		Filter: filter,
		Ccrn:   ccrns,
	})

	return ccrns, nil
}

func (ci *componentInstanceHandler) ListClusters(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListClusters")

	clusters, err := ci.database.GetCluster(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceClusters", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListClustersEvent{
		Filter:   filter,
		Clusters: clusters,
	})

	return clusters, nil
}

func (ci *componentInstanceHandler) ListNamespaces(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListNamespaces")

	namespaces, err := ci.database.GetNamespace(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceNamespaces", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListNamespacesEvent{
		Filter:     filter,
		Namespaces: namespaces,
	})

	return namespaces, nil
}

func (ci *componentInstanceHandler) ListDomains(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListDomains")

	domains, err := ci.database.GetDomain(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceDomains", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListDomainsEvent{
		Filter:  filter,
		Domains: domains,
	})

	return domains, nil
}

func (ci *componentInstanceHandler) ListProjects(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListProjects")

	projects, err := ci.database.GetProject(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceProjects", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListProjectsEvent{
		Filter:   filter,
		Projects: projects,
	})

	return projects, nil
}

func (ci *componentInstanceHandler) ListPods(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListPods")

	pods, err := ci.database.GetPod(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstancePods", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListPodsEvent{
		Filter: filter,
		Pods:   pods,
	})

	return pods, nil
}

func (ci *componentInstanceHandler) ListContainers(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListContainers")

	containers, err := ci.database.GetContainer(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceContainers", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListContainersEvent{
		Filter:     filter,
		Containers: containers,
	})

	return containers, nil
}

func (ci *componentInstanceHandler) ListTypes(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListTypes")

	types, err := ci.database.GetType(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceTypes", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListTypesEvent{
		Filter: filter,
		Types:  types,
	})

	return types, nil
}

func (ci *componentInstanceHandler) ListParents(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListParents")

	parents, err := ci.database.GetComponentInstanceParent(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceParents", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListParentsEvent{
		Filter:  filter,
		Parents: parents,
	})

	return parents, nil
}

func (ci *componentInstanceHandler) ListContexts(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	op := appErrors.Op("componentInstanceHandler.ListContexts")

	contexts, err := ci.database.GetContext(filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstanceContexts", "", err)
		applog.LogError(ci.logger, wrappedErr, logrus.Fields{
			"filter": filter,
		})
		return nil, wrappedErr
	}

	ci.eventRegistry.PushEvent(&ListContextsEvent{
		Filter:   filter,
		Contexts: contexts,
	})

	return contexts, nil
}

// validateParentIdForType checks if ParentId is only set for allowed types.
func validateParentIdForType(parentId int64, typeStr string) error {
	if parentId != 0 && parentId != -1 {
		if typeStr != "RecordSet" && typeStr != "User" && typeStr != "SecurityGroupRule" {
			return NewComponentInstanceHandlerError(
				"ParentId can only be set for component instances of type 'RecordSet', 'User' or 'SecurityGroupRule', but got type '" + typeStr + "'")
		}
	}
	return nil
}
