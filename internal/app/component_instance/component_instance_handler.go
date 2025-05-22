// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package component_instance

import (
	"github.com/cloudoperators/heureka/internal/app/common"
	"github.com/cloudoperators/heureka/internal/app/event"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/sirupsen/logrus"
)

type componentInstanceHandler struct {
	database      database.Database
	eventRegistry event.EventRegistry
}

func NewComponentInstanceHandler(database database.Database, eventRegistry event.EventRegistry) ComponentInstanceHandler {
	return &componentInstanceHandler{
		database:      database,
		eventRegistry: eventRegistry,
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
	var count int64
	var pageInfo *entity.PageInfo

	common.EnsurePaginatedX(&filter.PaginatedX)

	l := logrus.WithFields(logrus.Fields{
		"event":  ListComponentInstancesEventName,
		"filter": filter,
	})

	res, err := ci.database.GetComponentInstances(filter, options.Order)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Error while filtering for ComponentInstances")
	}

	if options.ShowPageInfo {
		if len(res) > 0 {
			cursors, err := ci.database.GetAllComponentInstanceCursors(filter, options.Order)
			if err != nil {
				return nil, NewComponentInstanceHandlerError("Error while getting all cursors")
			}
			pageInfo = common.GetPageInfoX(res, cursors, *filter.First, filter.After)
			count = int64(len(cursors))
		}
	} else if options.ShowTotalCount {
		count, err = ci.database.CountComponentInstances(filter)
		if err != nil {
			l.Error(err)
			return nil, NewComponentInstanceHandlerError("Error while total count of ComponentInstances")
		}
	}

	ci.eventRegistry.PushEvent(&ListComponentInstancesEvent{
		Filter:  filter,
		Options: options,
		ComponentInstances: &entity.List[entity.ComponentInstanceResult]{
			TotalCount: &count,
			PageInfo:   pageInfo,
			Elements:   res,
		},
	})

	return &entity.List[entity.ComponentInstanceResult]{
		TotalCount: &count,
		PageInfo:   pageInfo,
		Elements:   res,
	}, nil
}

func (ci *componentInstanceHandler) CreateComponentInstance(componentInstance *entity.ComponentInstance, scannerRunUUID *string) (*entity.ComponentInstance, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  CreateComponentInstanceEventName,
		"object": componentInstance,
	})

	var err error
	componentInstance.CreatedBy, err = common.GetCurrentUserId(ci.database)
	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while creating componentInstance (GetUserId).")
	}
	componentInstance.UpdatedBy = componentInstance.CreatedBy

	newComponentInstance, err := ci.database.CreateComponentInstance(componentInstance)

	if err != nil {
		return nil, NewComponentInstanceHandlerError("Internal error while creating componentInstance.")
	}

	if scannerRunUUID != nil {
		err = ci.database.CreateScannerRunComponentInstanceTracker(newComponentInstance.Id, *scannerRunUUID)

		if err != nil {
			return nil, NewComponentInstanceHandlerError("Internal error while creating ScannerRunComponentInstanceTracker.")
		}
	}

	ci.eventRegistry.PushEvent(&CreateComponentInstanceEvent{
		ComponentInstance: newComponentInstance,
	})

	return newComponentInstance, nil
}

func (ci *componentInstanceHandler) UpdateComponentInstance(componentInstance *entity.ComponentInstance, scannerRunUUID *string) (*entity.ComponentInstance, error) {
	l := logrus.WithFields(logrus.Fields{
		"event":  UpdateComponentInstanceEventName,
		"object": componentInstance,
	})

	var err error
	componentInstance.UpdatedBy, err = common.GetCurrentUserId(ci.database)
	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while updating componentInstance (GetUserId).")
	}

	err = ci.database.UpdateComponentInstance(componentInstance)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while updating componentInstance.")
	}

	if scannerRunUUID != nil {
		err = ci.database.CreateScannerRunComponentInstanceTracker(componentInstance.Id, *scannerRunUUID)

		if err != nil {
			return nil, NewComponentInstanceHandlerError("Internal error while creating ScannerRunComponentInstanceTracker.")
		}
	}

	lo := entity.NewListOptions()

	componentInstanceResult, err := ci.ListComponentInstances(&entity.ComponentInstanceFilter{Id: []*int64{&componentInstance.Id}}, lo)

	if err != nil {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Internal error while retrieving updated componentInstance.")
	}

	if len(componentInstanceResult.Elements) != 1 {
		l.Error(err)
		return nil, NewComponentInstanceHandlerError("Multiple componentInstances found.")
	}

	ci.eventRegistry.PushEvent(&UpdateComponentInstanceEvent{
		ComponentInstance: componentInstance,
	})

	return componentInstanceResult.Elements[0].ComponentInstance, nil
}

func (ci *componentInstanceHandler) DeleteComponentInstance(id int64) error {
	l := logrus.WithFields(logrus.Fields{
		"event": DeleteComponentInstanceEventName,
		"id":    id,
	})

	userId, err := common.GetCurrentUserId(ci.database)
	if err != nil {
		l.Error(err)
		return NewComponentInstanceHandlerError("Internal error while deleting componentInstance (GetUserId).")
	}

	err = ci.database.DeleteComponentInstance(id, userId)

	if err != nil {
		l.Error(err)
		return NewComponentInstanceHandlerError("Internal error while deleting componentInstance.")
	}

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
