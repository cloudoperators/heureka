// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	applog "github.com/cloudoperators/heureka/internal/app/logging"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	appErrors "github.com/cloudoperators/heureka/internal/errors"
	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/sirupsen/logrus"
)

type componentInstanceHandler struct {
	common.BaseHandler[entity.ComponentInstanceResult, *entity.ComponentInstanceFilter]
}

func NewComponentInstanceHandler(hc common.HandlerContext) ComponentInstanceHandler {
	return &componentInstanceHandler{
		BaseHandler: common.NewBaseHandler(hc, common.BaseConfig[entity.ComponentInstanceResult, *entity.ComponentInstanceFilter]{
			Op:              appErrors.Op("componentInstanceHandler"),
			Entity:          "ComponentInstances",
			CursorEntity:    "ComponentInstanceCursors",
			CountEntity:     "ComponentInstanceCount",
			GetFn:           hc.DB.GetComponentInstances,
			CursorsFn:       hc.DB.GetAllComponentInstanceCursors,
			CountFn:         hc.DB.CountComponentInstances,
			Authz:           hc.Authz,
			AuthzObjectType: openfga.TypeService,
			AuthzApplyFn: func(f *entity.ComponentInstanceFilter, ids []*int64) {
				f.ServiceId = common.CombineFilterWithAccessibleIds(f.ServiceId, ids)
			},
			ListEventFn: func(f *entity.ComponentInstanceFilter, o *entity.ListOptions, r *entity.List[entity.ComponentInstanceResult]) any {
				return &ListComponentInstancesEvent{Filter: f, Options: o, ComponentInstances: r}
			},
			DeleteFn:      hc.DB.DeleteComponentInstance,
			DeleteEventFn: func(id int64) any { return &DeleteComponentInstanceEvent{ComponentInstanceID: id} },
		}),
	}
}

func (ci *componentInstanceHandler) ListComponentInstances(
	ctx context.Context,
	filter *entity.ComponentInstanceFilter,
	options *entity.ListOptions,
) (*entity.List[entity.ComponentInstanceResult], error) {
	return ci.List(ctx, appErrors.CallerOp(), filter, options)
}

func (ci *componentInstanceHandler) CreateComponentInstance(
	ctx context.Context,
	componentInstance *entity.ComponentInstance,
	scannerRunUUID *string,
) (*entity.ComponentInstance, error) {
	op := appErrors.CallerOp()

	if componentInstance == nil {
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, "component instance cannot be nil")
		applog.LogError(logrus.StandardLogger(), err, logrus.Fields{})

		return nil, err
	}

	if componentInstance.CCRN == "" {
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, "CCRN is required")
		applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"component_instance": componentInstance})

		return nil, err
	}

	if componentInstance.ServiceId <= 0 {
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, "valid service ID is required")
		applog.LogError(logrus.StandardLogger(), err, logrus.Fields{
			"service_id": componentInstance.ServiceId,
			"ccrn":       componentInstance.CCRN,
		})

		return nil, err
	}

	if err := validateParentIdForType(componentInstance.ParentId, componentInstance.Type.String()); err != nil {
		wrappedErr := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, err.Error())
		applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{
			"parent_id":        componentInstance.ParentId,
			"type":             componentInstance.Type.String(),
			"ccrn":             componentInstance.CCRN,
			"validation_error": err.Error(),
		})

		return nil, wrappedErr
	}

	var err error

	componentInstance.CreatedBy, err = common.GetCurrentUserId(ctx, ci.DB())
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
		applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{
			"ccrn": componentInstance.CCRN,
			"type": componentInstance.Type.String(),
		})

		return nil, wrappedErr
	}

	componentInstance.UpdatedBy = componentInstance.CreatedBy

	newComponentInstance, err := ci.DB().CreateComponentInstance(componentInstance)
	if err != nil {
		duplicateEntryError := &database.DuplicateEntryDatabaseError{}
		if errors.As(err, &duplicateEntryError) {
			wrappedErr := appErrors.AlreadyExistsError(string(op), "ComponentInstance", componentInstance.CCRN)
			applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{
				"ccrn":                    componentInstance.CCRN,
				"component_version_id":    componentInstance.ComponentVersionId,
				"service_id":              componentInstance.ServiceId,
				"duplicate_entry_details": duplicateEntryError.Error(),
			})

			return nil, wrappedErr
		}

		wrappedErr := appErrors.InternalError(string(op), "ComponentInstance", "", err)
		applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{
			"ccrn":                 componentInstance.CCRN,
			"component_version_id": componentInstance.ComponentVersionId,
			"service_id":           componentInstance.ServiceId,
			"type":                 componentInstance.Type.String(),
		})

		return nil, wrappedErr
	}

	if scannerRunUUID != nil {
		err = ci.DB().CreateScannerRunComponentInstanceTracker(newComponentInstance.Id, *scannerRunUUID)
		if err != nil {
			logErr := appErrors.InternalError(
				string(op),
				"ScannerRunComponentInstanceTracker",
				fmt.Sprintf("component_instance:%d-scanner_run:%s", newComponentInstance.Id, *scannerRunUUID),
				err,
			)

			applog.LogError(logrus.StandardLogger(), logErr, logrus.Fields{
				"component_instance_id": newComponentInstance.Id,
				"scanner_run_uuid":      *scannerRunUUID,
				"ccrn":                  newComponentInstance.CCRN,
			})
		}
	}

	ci.PushEvent(&CreateComponentInstanceEvent{ComponentInstance: newComponentInstance})

	return newComponentInstance, nil
}

func (ci *componentInstanceHandler) UpdateComponentInstance(
	ctx context.Context,
	componentInstance *entity.ComponentInstance,
	scannerRunUUID *string,
) (*entity.ComponentInstance, error) {
	op := appErrors.CallerOp()

	if componentInstance == nil {
		return nil, appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, "component instance cannot be nil")
	}

	if componentInstance.Id <= 0 {
		return nil, appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, fmt.Sprintf("invalid ID: %d", componentInstance.Id))
	}

	if err := validateParentIdForType(componentInstance.ParentId, componentInstance.Type.String()); err != nil {
		return nil, appErrors.E(op, "ComponentInstance", strconv.FormatInt(componentInstance.Id, 10), appErrors.InvalidArgument, err.Error())
	}

	id := strconv.FormatInt(componentInstance.Id, 10)

	var err error

	componentInstance.UpdatedBy, err = common.GetCurrentUserId(ctx, ci.DB())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentInstance", id, err)
	}

	if err = ci.DB().UpdateComponentInstance(componentInstance); err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentInstance", id, err)
	}

	if scannerRunUUID != nil {
		if err = ci.DB().CreateScannerRunComponentInstanceTracker(componentInstance.Id, *scannerRunUUID); err != nil {
			applog.LogError(logrus.StandardLogger(), appErrors.InternalError(string(op), "ScannerRunComponentInstanceTracker",
				fmt.Sprintf("component_instance:%d-scanner_run:%s", componentInstance.Id, *scannerRunUUID), err), logrus.Fields{})
		}
	}

	result, err := ci.ListComponentInstances(ctx, &entity.ComponentInstanceFilter{Id: []*int64{&componentInstance.Id}}, entity.NewListOptions())
	if err != nil {
		return nil, appErrors.InternalError(string(op), "ComponentInstance", id, err)
	}

	if len(result.Elements) != 1 {
		return nil, appErrors.E(op, "ComponentInstance", id, appErrors.Internal,
			fmt.Sprintf("unexpected number of component instances found after update: expected 1, got %d", len(result.Elements)))
	}

	updatedComponentInstance := result.Elements[0].ComponentInstance
	ci.PushEvent(&UpdateComponentInstanceEvent{ComponentInstance: updatedComponentInstance})

	return updatedComponentInstance, nil
}

func (ci *componentInstanceHandler) DeleteComponentInstance(ctx context.Context, id int64) error {
	if id <= 0 {
		op := appErrors.CallerOp()
		err := appErrors.E(op, "ComponentInstance", appErrors.InvalidArgument, fmt.Sprintf("invalid ID: %d", id))
		applog.LogError(logrus.StandardLogger(), err, logrus.Fields{"id": id})

		return err
	}

	return ci.Delete(ctx, id)
}

func (ci *componentInstanceHandler) listStrings(
	op appErrors.Op,
	ctx context.Context,
	filter *entity.ComponentInstanceFilter,
	entityName string,
	fetchFn func(context.Context, *entity.ComponentInstanceFilter) ([]string, error),
	eventFn func([]string) event.Event,
) ([]string, error) {
	values, err := fetchFn(ctx, filter)
	if err != nil {
		wrappedErr := appErrors.InternalError(string(op), entityName, "", err)
		applog.LogError(logrus.StandardLogger(), wrappedErr, logrus.Fields{"filter": filter})

		return nil, wrappedErr
	}

	ci.PushEvent(eventFn(values))

	return values, nil
}

func (ci *componentInstanceHandler) ListRegions(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstanceRegions", ci.DB().GetRegion,
		func(v []string) event.Event { return &ListRegionsEvent{Filter: filter, Regions: v} })
}

func (ci *componentInstanceHandler) ListCcrns(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstanceCcrns", ci.DB().GetCcrn,
		func(v []string) event.Event { return &ListCcrnEvent{Filter: filter, Ccrn: v} })
}

func (ci *componentInstanceHandler) ListClusters(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstanceClusters", ci.DB().GetCluster,
		func(v []string) event.Event { return &ListClustersEvent{Filter: filter, Clusters: v} })
}

func (ci *componentInstanceHandler) ListNamespaces(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstanceNamespaces", ci.DB().GetNamespace,
		func(v []string) event.Event { return &ListNamespacesEvent{Filter: filter, Namespaces: v} })
}

func (ci *componentInstanceHandler) ListDomains(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstanceDomains", ci.DB().GetDomain,
		func(v []string) event.Event { return &ListDomainsEvent{Filter: filter, Domains: v} })
}

func (ci *componentInstanceHandler) ListProjects(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstanceProjects", ci.DB().GetProject,
		func(v []string) event.Event { return &ListProjectsEvent{Filter: filter, Projects: v} })
}

func (ci *componentInstanceHandler) ListPods(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstancePods", ci.DB().GetPod,
		func(v []string) event.Event { return &ListPodsEvent{Filter: filter, Pods: v} })
}

func (ci *componentInstanceHandler) ListContainers(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstanceContainers", ci.DB().GetContainer,
		func(v []string) event.Event { return &ListContainersEvent{Filter: filter, Containers: v} })
}

func (ci *componentInstanceHandler) ListTypes(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstanceTypes", ci.DB().GetType,
		func(v []string) event.Event { return &ListTypesEvent{Filter: filter, Types: v} })
}

func (ci *componentInstanceHandler) ListParents(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstanceParents", ci.DB().GetComponentInstanceParent,
		func(v []string) event.Event { return &ListParentsEvent{Filter: filter, Parents: v} })
}

func (ci *componentInstanceHandler) ListContexts(ctx context.Context, filter *entity.ComponentInstanceFilter, options *entity.ListOptions) ([]string, error) {
	return ci.listStrings(appErrors.CallerOp(), ctx, filter, "ComponentInstanceContexts", ci.DB().GetContext,
		func(v []string) event.Event { return &ListContextsEvent{Filter: filter, Contexts: v} })
}

// validateParentIdForType checks if ParentId is only set for allowed types.
func validateParentIdForType(parentId int64, typeStr string) error {
	if parentId != 0 && parentId != -1 {
		if typeStr != "RecordSet" && typeStr != "User" && typeStr != "SecurityGroupRule" {
			return fmt.Errorf(
				"ParentId can only be set for component instances of type 'RecordSet', 'User' or 'SecurityGroupRule', but got type '%s'",
				typeStr,
			)
		}
	}

	return nil
}
