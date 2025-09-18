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

type ComponentInstanceHandlerError struct {
	message string
}

func NewComponentInstanceHandlerError(message string) *ComponentInstanceHandlerError {
	return &ComponentInstanceHandlerError{message: message}
}

func (e *ComponentInstanceHandlerError) Error() string {
	return e.message
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

func (s *componentInstanceHandler) ListCcrns(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListCcrnEventName,
		"filter": filter,
	})

	ccrn, err := s.database.GetCcrn(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Ccrn.")
	}

	s.eventRegistry.PushEvent(&ListCcrnEvent{Filter: filter, Ccrn: ccrn})

	return ccrn, nil
}
func (s *componentInstanceHandler) ListRegions(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListRegionsEventName,
		"filter": filter,
	})

	regions, err := s.database.GetRegion(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Region.")
	}

	s.eventRegistry.PushEvent(&ListRegionsEvent{Filter: filter, Regions: regions})

	return regions, nil
}
func (s *componentInstanceHandler) ListClusters(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListClustersEventName,
		"filter": filter,
	})

	clusters, err := s.database.GetCluster(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Cluster.")
	}

	s.eventRegistry.PushEvent(&ListClustersEvent{Filter: filter, Clusters: clusters})

	return clusters, nil
}
func (s *componentInstanceHandler) ListNamespaces(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListNamespacesEventName,
		"filter": filter,
	})

	namespaces, err := s.database.GetNamespace(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Namespace.")
	}

	s.eventRegistry.PushEvent(&ListNamespacesEvent{Filter: filter, Namespaces: namespaces})

	return namespaces, nil
}
func (s *componentInstanceHandler) ListDomains(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListDomainsEventName,
		"filter": filter,
	})

	domains, err := s.database.GetDomain(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Domain.")
	}

	s.eventRegistry.PushEvent(&ListDomainsEvent{Filter: filter, Domains: domains})

	return domains, nil
}
func (s *componentInstanceHandler) ListProjects(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListProjectsEventName,
		"filter": filter,
	})

	projects, err := s.database.GetProject(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Project.")
	}

	s.eventRegistry.PushEvent(&ListProjectsEvent{Filter: filter, Projects: projects})

	return projects, nil
}
func (s *componentInstanceHandler) ListPods(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListPodsEventName,
		"filter": filter,
	})

	pods, err := s.database.GetPod(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Pod.")
	}

	s.eventRegistry.PushEvent(&ListPodsEvent{Filter: filter, Pods: pods})

	return pods, nil
}
func (s *componentInstanceHandler) ListContainers(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListContainersEventName,
		"filter": filter,
	})

	containers, err := s.database.GetContainer(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Container.")
	}

	s.eventRegistry.PushEvent(&ListContainersEvent{Filter: filter, Containers: containers})

	return containers, nil
}
func (s *componentInstanceHandler) ListTypes(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListTypesEventName,
		"filter": filter,
	})

	types, err := s.database.GetType(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Type.")
	}

	s.eventRegistry.PushEvent(&ListTypesEvent{Filter: filter, Types: types})

	return types, nil
}

func (s *componentInstanceHandler) ListParents(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListParentsEventName,
		"filter": filter,
	})

	parents, err := s.database.GetComponentInstanceParent(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Parent.")
	}

	s.eventRegistry.PushEvent(&ListParentsEvent{Filter: filter, Parents: parents})

	return parents, nil
}

func (s *componentInstanceHandler) ListContexts(filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  ListContextsEventName,
		"filter": filter,
	})

	contexts, err := s.database.GetContext(filter)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving Type.")
	}

	s.eventRegistry.PushEvent(&ListContextsEvent{Filter: filter, Contexts: contexts})

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
