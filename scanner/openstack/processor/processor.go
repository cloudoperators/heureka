// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/cloudoperators/heureka/scanner/openstack/client"
	log "github.com/sirupsen/logrus"
)

type Processor struct {
	Client *graphql.Client
}

type ServiceInfo struct {
	CCRN string
}

type SupportGroupInfo struct {
	CCRN string
}

type IssueRepositoryInfo struct {
	Name string
	Url  string
}

type ComponentInfo struct {
	CCRN string
	Type string
}

type ComponentVersionInfo struct {
	Version     string
	ComponentID string
}

type ComponentInstanceInfo struct {
	CCRN               string
	ComponentVersionID string
	ServiceID          string
	ServiceCCRN        string
}

type IssueInfo struct {
	PrimaryName string
	Description string
}

func NewProcessor(cfg Config) *Processor {
	httpClient := http.Client{}
	gClient := graphql.NewClient(cfg.HeurekaUrl, &httpClient)
	return &Processor{
		Client: &gClient,
	}
}

// ProcessService processes a service and creates a new service if it doesn't exist.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	serviceInfo ServiceInfo - The service info to be used for the request.
//
// Returns:
//
//	string - The ID of the service.
//	error - An error if something goes wrong during the request.
func (p *Processor) ProcessService(ctx context.Context, serviceInfo ServiceInfo) (string, error) {
	var serviceId string

	if serviceInfo.CCRN == "" {
		serviceInfo.CCRN = "none"
	}

	// The Service might already exist in the DB
	// Let's try to fetch one Service by name
	_serviceId, err := p.getService(ctx, serviceInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"serviceCcrn": serviceInfo.CCRN,
		}).Error("failed to fetch service")

		// Create new Service
		createServiceInput := &client.ServiceInput{
			Ccrn: serviceInfo.CCRN,
		}

		createServiceResp, err := client.CreateService(ctx, *p.Client, createServiceInput)
		if err != nil {
			return "", fmt.Errorf("failed to create Service %s: %w", serviceInfo.CCRN, err)
		} else {
			serviceId = createServiceResp.CreateService.Id
		}
	} else {
		serviceId = _serviceId
	}

	return serviceId, nil
}

// getService fetches a service by CCRN
// *Service has a unique constraint on CCRN,
// so this query should return at most one result.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	serviceInfo ServiceInfo - The service info to be used for the request.
//
// Returns:
//
//	string - The ID of the service.
//	error - An error if something goes wrong during the request.
func (p *Processor) getService(ctx context.Context, serviceInfo ServiceInfo) (string, error) {
	listServicesFilter := client.ServiceFilter{
		ServiceCcrn: []string{serviceInfo.CCRN},
	}
	listServicesResp, err := client.ListServices(ctx, *p.Client, &listServicesFilter)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("couldn't list services")
	}

	// Return the first item
	if listServicesResp.Services.TotalCount > 0 {
		return listServicesResp.Services.Edges[0].Node.Id, nil
	}

	// No Service found
	return "", fmt.Errorf("ListServices returned no ServiceID")
}

// ProcessSupportGroup processes a support group and creates a new support group if it doesn't exist.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	supportGroupInfo SupportGroupInfo - The support group info to be used for the request.
//
// Returns:
//
//	string - The ID of the support group.
//	error - An error if something goes wrong during the request.
func (p *Processor) ProcessSupportGroup(ctx context.Context, supportGroupInfo SupportGroupInfo) (string, error) {
	var supportGroupId string
	_supportGroupId, err := p.getSupportGroup(ctx, supportGroupInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"supportGroupCCRN": supportGroupInfo.CCRN,
		}).Error("failed to fetch support group")

		// Create new SupportGroup
		createSupportGroupInput := &client.SupportGroupInput{
			Ccrn: supportGroupInfo.CCRN,
		}
		createSupportGroupResp, err := client.CreateSupportGroup(ctx, *p.Client, createSupportGroupInput)
		if err != nil {
			return "", fmt.Errorf("failed to create SupportGroup %s: %w", supportGroupInfo.CCRN, err)
		} else {
			supportGroupId = createSupportGroupResp.CreateSupportGroup.Id
		}
	} else {
		supportGroupId = _supportGroupId
	}

	return supportGroupId, nil
}

// getSupportGroup fetches a support group by CCRN
// *SupportGroup has a unique constraint on CCRN,
// so this query should return at most one result.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	supportGroupInfo SupportGroupInfo - The support group info to be used for the request.
//
// Returns:
//
//	string - The ID of the support group.
//	error - An error if something goes wrong during the request.
func (p *Processor) getSupportGroup(ctx context.Context, supportGroupInfo SupportGroupInfo) (string, error) {
	var supportGroupId string

	listSupportGroupsFilter := client.SupportGroupFilter{
		SupportGroupCcrn: []string{supportGroupInfo.CCRN},
	}
	listSupportGroupsResp, err := client.ListSupportGroups(ctx, *p.Client, &listSupportGroupsFilter)
	if err != nil {
		return "", fmt.Errorf("couldn't list support groups")
	}

	// Return the first item
	if listSupportGroupsResp.SupportGroups.TotalCount > 0 && len(listSupportGroupsResp.SupportGroups.Edges) > 0 {
		supportGroupId = listSupportGroupsResp.SupportGroups.Edges[0].Node.Id
	} else {
		return "", fmt.Errorf("ListSupportGroups returned no SupportGroupID")
	}

	return supportGroupId, nil
}

// ConnectServiceToSupportGroup connects a service to a support group.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	serviceId string - The ID of the service.
//	supportGroupId string - The ID of the support group.
//
// Returns:
//
//	error - An error if something goes wrong during the request.
func (p *Processor) ConnectServiceToSupportGroup(ctx context.Context, serviceId string, supportGroupId string) error {
	_, err := client.AddServiceToSupportGroup(ctx, *p.Client, supportGroupId, serviceId)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"serviceId":      serviceId,
			"supportGroupId": supportGroupId,
		}).Warning("Failed adding service to support group")
		return err
	}

	return nil
}

// ProcessIssueRepository processes an issue repository and creates a new issue repository if it doesn't exist.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	issueRepositoryInfo IssueRepositoryInfo - The issue repository info to be used for the request.
//
// Returns:
//
//	string - The ID of the issue repository.
//	error - An error if something goes wrong during the request.
func (p *Processor) ProcessIssueRepository(ctx context.Context, issueRepositoryInfo IssueRepositoryInfo) (string, error) {
	var issueRepositoryId string

	if issueRepositoryInfo.Name == "" {
		issueRepositoryInfo.Name = "none"
	}

	// The IssueRepository might already exist in the DB
	// Let's try to fetch list of IssueRepository by name
	_issueRepositoryId, err := p.GetIssueRepository(ctx, issueRepositoryInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"issueRepositoryName": issueRepositoryInfo.Name,
		}).Error("failed to fetch issueRepository")

		// Create new IssueRepository
		createIssueRepositoryInput := &client.IssueRepositoryInput{
			Name: issueRepositoryInfo.Name,
			Url:  issueRepositoryInfo.Url,
		}

		createIssueRepositoryResp, err := client.CreateIssueRepository(ctx, *p.Client, createIssueRepositoryInput)
		if err != nil {
			return "", fmt.Errorf("failed to create IssueRepository %s: %w", issueRepositoryInfo.Name, err)
		} else {
			issueRepositoryId = createIssueRepositoryResp.CreateIssueRepository.Id
		}
	} else {
		issueRepositoryId = _issueRepositoryId
	}

	return issueRepositoryId, nil
}

// GetIssueRepository fetches an issue repository by name
// *IssueRepository has a unique constraint on name,
// so this query should return at most one result.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	issueRepositoryInfo IssueRepositoryInfo - The issue repository info to be used for the request.
//
// Returns:
//
//	string - The ID of the issue repository.
//	error - An error if something goes wrong during the request.
func (p *Processor) GetIssueRepository(ctx context.Context, issueRepositoryInfo IssueRepositoryInfo) (string, error) {
	listIssueRepositoryFilter := client.IssueRepositoryFilter{
		Name: []string{issueRepositoryInfo.Name},
	}

	listIssueRepositoryResp, err := client.ListIssueRepositories(ctx, *p.Client, &listIssueRepositoryFilter)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("couldn't list issue repositories")
	}

	// Return the first item
	if listIssueRepositoryResp.IssueRepositories.TotalCount > 0 {
		return listIssueRepositoryResp.IssueRepositories.Edges[0].Node.Id, nil
	}

	// No IssueRepository found
	return "", fmt.Errorf("ListIssueRepositories returned no IssueRepositoryID")
}

// ConnectIssueRepositoryToService connects an issue repository to a service.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	issueRepositoryId string - The ID of the issue repository.
//	serviceId string - The ID of the service.
//
// Returns:
//
//	error - An error if something goes wrong during the request.
func (p *Processor) ConnectIssueRepositoryToService(ctx context.Context, issueRepositoryId string, serviceId string) error {
	_, err := client.AddIssueRepositoryToService(ctx, *p.Client, issueRepositoryId, serviceId, 1)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"issueRepositoryId": issueRepositoryId,
			"serviceId":         serviceId,
		}).Warning("Failed adding issue repository to service")
		return err
	}

	return nil
}

// ProcessComponent processes a component and creates a new component if it doesn't exist.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	componentInfo ComponentInfo - The component info to be used for the request.
//
// Returns:
//
//	string - The ID of the component.
//	error - An error if something goes wrong during the request.
func (p *Processor) ProcessComponent(ctx context.Context, componentInfo ComponentInfo) (string, error) {
	var componentId string

	if componentInfo.CCRN == "" {
		componentInfo.CCRN = "none"
	}

	// The Component might already exist in the DB
	// Let's try to fetch list of Components by name
	_componentId, err := p.GetComponent(ctx, componentInfo)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"componentCcrn": componentInfo.CCRN,
		}).Error("failed to fetch component")

		// Create new Component
		createComponentInput := &client.ComponentInput{
			Ccrn: componentInfo.CCRN,
			Type: client.ComponentTypeValuesVirtualmachineimage,
		}

		createComponentResp, err := client.CreateComponent(ctx, *p.Client, createComponentInput)
		if err != nil {
			return "", fmt.Errorf("failed to create Component %s: %w", componentInfo.CCRN, err)
		} else {
			componentId = createComponentResp.CreateComponent.Id
		}
	} else {
		componentId = _componentId
	}

	return componentId, nil
}

// GetComponent fetches a component by CCRN
// *Component has a unique constraint on CCRN,
// so this query should return at most one result.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	componentInfo ComponentInfo - The component info to be used for the request.
//
// Returns:
//
//	string - The ID of the component.
//	error - An error if something goes wrong during the request.
func (p *Processor) GetComponent(ctx context.Context, componentInfo ComponentInfo) (string, error) {
	listComponentsFilter := client.ComponentFilter{
		ComponentCcrn: []string{componentInfo.CCRN},
	}
	listComponentsResp, err := client.ListComponents(ctx, *p.Client, &listComponentsFilter)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("couldn't list components")
	}

	// Return the first item
	if listComponentsResp.Components.TotalCount > 0 {
		return listComponentsResp.Components.Edges[0].Node.Id, nil
	}

	// No Component found
	return "", fmt.Errorf("ListComponents returned no ComponentID")
}

// ProcessComponentVersion processes a component version and creates a new component version if it doesn't exist.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	componentVersionInfo ComponentVersionInfo - The component version info to be used for the request.
//
// Returns:
//
//	string - The ID of the component version.
//	error - An error if something goes wrong during the request.
func (p *Processor) ProcessComponentVersion(ctx context.Context, componentVersionInfo ComponentVersionInfo) (string, error) {
	var componentVersionId string

	if componentVersionInfo.Version == "" {
		componentVersionInfo.Version = "none"
	}

	// The Component Version might already exist in the DB
	// Let's try to fetch list of Component Versions by name
	_componentVersionId, err := p.GetComponentVersion(ctx, componentVersionInfo)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"version":     componentVersionInfo.Version,
			"componentID": componentVersionInfo.ComponentID,
		}).Error("failed to fetch componentVersion")

		// Create new Component Version
		createComponentVersionInput := &client.ComponentVersionInput{
			Version:     componentVersionInfo.Version,
			ComponentId: componentVersionInfo.ComponentID,
		}

		createComponentVersionResp, err := client.CreateComponentVersion(ctx, *p.Client, createComponentVersionInput)
		if err != nil {
			return "", fmt.Errorf("failed to create ComponentVersion %s %s: %w", componentVersionInfo.Version, componentVersionInfo.ComponentID, err)
		} else {
			componentVersionId = createComponentVersionResp.CreateComponentVersion.Id
		}
	} else {
		componentVersionId = _componentVersionId
	}

	return componentVersionId, nil
}

// GetComponentVersion fetches a component version by version and component ID
// *ComponentVersion has a unique constraint on version + component ID,
// so this query should return at most one result.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	componentVersionInfo ComponentVersionInfo - The component version info to be used for the request.
//
// Returns:
//
//	string - The ID of the component version.
//	error - An error if something goes wrong during the request.
func (p *Processor) GetComponentVersion(ctx context.Context, componentVersionInfo ComponentVersionInfo) (string, error) {
	listComponentVersionFilter := client.ComponentVersionFilter{
		Version:     []string{componentVersionInfo.Version},
		ComponentId: []string{componentVersionInfo.ComponentID},
	}
	listComponentVersionssResp, err := client.ListComponentVersions(ctx, *p.Client, &listComponentVersionFilter)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("couldn't list component versions")
	}

	// Return the first item
	if listComponentVersionssResp.ComponentVersions.TotalCount > 0 {
		return listComponentVersionssResp.ComponentVersions.Edges[0].Node.Id, nil
	}

	// No Component Version found
	return "", fmt.Errorf("ListComponentVersions returned no ComponentVersionID")
}

// ProcessComponentInstance processes a component instance and creates a new component instance if it doesn't exist.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	componentInstanceInfo ComponentInstanceInfo - The component instance info to be used for the request.
//
// Returns:
//
//	string - The ID of the component instance.
//	error - An error if something goes wrong during the request.
func (p *Processor) ProcessComponentInstance(ctx context.Context, componentInstanceInfo ComponentInstanceInfo) (string, error) {
	var componentInstanceId string

	if componentInstanceInfo.CCRN == "" {
		componentInstanceInfo.CCRN = "none"
	}

	// The Component Instance might already exist in the DB
	// Let's try to fetch list of Component Instances by name
	_componentInstanceId, err := p.GetComponentInstance(ctx, componentInstanceInfo)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"ccrn":               componentInstanceInfo.CCRN,
			"componentVersionId": componentInstanceInfo.ComponentVersionID,
			"serviceId":          componentInstanceInfo.ServiceID,
		}).Error("failed to fetch componentInstance")

		// Create new Component Instance
		createComponentInstanceInput := &client.ComponentInstanceInput{
			Ccrn:               componentInstanceInfo.CCRN,
			Count:              1,
			ComponentVersionId: componentInstanceInfo.ComponentVersionID,
			ServiceId:          componentInstanceInfo.ServiceID,
		}

		createComponentInstanceResp, err := client.CreateComponentInstance(ctx, *p.Client, createComponentInstanceInput)
		if err != nil {
			return "", fmt.Errorf("failed to create ComponentInstance %s: %w", componentInstanceInfo.CCRN, err)
		} else {
			componentInstanceId = createComponentInstanceResp.CreateComponentInstance.Id
		}
	} else {
		componentInstanceId = _componentInstanceId
	}

	return componentInstanceId, nil
}

// GetComponentInstance fetches a component instance by CCRN + Service CCRN
// *ComponentInstance has a unique constraint on CCRN + Service CCRN,
// so this query should return at most one result.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	componentInstanceInfo ComponentInstanceInfo - The component instance info to be used for the request.
//
// Returns:
//
//	string - The ID of the component instance.
//	error - An error if something goes wrong during the request.
func (p *Processor) GetComponentInstance(ctx context.Context, componentInstanceInfo ComponentInstanceInfo) (string, error) {
	listComponentInstancesFilter := client.ComponentInstanceFilter{
		Ccrn:        []string{componentInstanceInfo.CCRN},
		ServiceCcrn: []string{componentInstanceInfo.ServiceCCRN},
	}
	listComponentInstancesResp, err := client.ListComponentInstances(ctx, *p.Client, &listComponentInstancesFilter)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("couldn't list component instances")
	}

	// Return the first item
	if listComponentInstancesResp.ComponentInstances.TotalCount > 0 {
		return listComponentInstancesResp.ComponentInstances.Edges[0].Node.Id, nil
	}

	// No Component Instance found
	return "", fmt.Errorf("ListComponentInstances returned no ComponentInstanceID")
}

// ProcessIssue processes an issue and creates a new issue if it doesn't exist.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	issueInfo IssueInfo - The issue info to be used for the request.
//
// Returns:
//
//	string - The ID of the issue.
//	error - An error if something goes wrong during the request.
func (p *Processor) ProcessIssue(ctx context.Context, issueInfo IssueInfo) (string, error) {
	var issueId string

	if issueInfo.PrimaryName == "" {
		issueInfo.PrimaryName = "none"
	}

	// The Issue might already exist in the DB
	// Let's try to fetch list of Issues by name
	_issueId, err := p.GetIssue(ctx, issueInfo)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"primaryName": issueInfo.PrimaryName,
			"description": issueInfo.Description,
		}).Error("failed to fetch issue")

		// Create new Issue
		createIssueInput := &client.IssueInput{
			PrimaryName: issueInfo.PrimaryName,
			Description: issueInfo.Description,
			Type:        client.IssueTypesPolicyviolation,
		}

		createIssueResp, err := client.CreateIssue(ctx, *p.Client, createIssueInput)
		if err != nil {
			return "", fmt.Errorf("failed to create Issue %s: %w", issueInfo.PrimaryName, err)
		} else {
			issueId = createIssueResp.CreateIssue.Id
		}
	} else {
		issueId = _issueId
	}

	return issueId, nil
}

// GetIssue fetches an issue by primary name
// *Issue has a unique constraint on primaryName,
// so this query should return at most one result.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	issueInfo IssueInfo - The issue info to be used for the request.
//
// Returns:
//
//	string - The ID of the issue.
//	error - An error if something goes wrong during the request.
func (p *Processor) GetIssue(ctx context.Context, issueInfo IssueInfo) (string, error) {
	listIssuesFilter := client.IssueFilter{
		PrimaryName: []string{issueInfo.PrimaryName},
	}
	listIssuesResp, err := client.ListIssues(ctx, *p.Client, &listIssuesFilter)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("couldn't list issues")
	}

	// Return the first item
	if listIssuesResp.Issues.TotalCount > 0 {
		return listIssuesResp.Issues.Edges[0].Node.Id, nil
	}

	// No Issue found
	return "", fmt.Errorf("ListIssues returned no IssueID")
}

// ConnectComponentVersionToIssue connects a component version to an issue.
//
// Parameters:
//
//	ctx context.Context - The context to be used for the request.
//	componentVersionId string - The ID of the component version.
//	issueId string - The ID of the issue.
//
// Returns:
//
//	error - An error if something goes wrong during the request.
func (p *Processor) ConnectComponentVersionToIssue(ctx context.Context, componentVersionId string, issueId string) error {
	_, err := client.AddComponentVersionToIssue(ctx, *p.Client, issueId, componentVersionId)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"componentVersionId": componentVersionId,
			"issueId":            issueId,
		}).Warning("Failed adding component version to issue")
		return err
	}

	return nil
}

// Policy4dot5Check checks if the given image name complies with policy 4.5.
// Policy 4.5 requires that the image name contains either "gardenlinux" or "SAP-compliant".
//
// Parameters:
//
//	img_name (string): The name of the image to be checked.
//
// Returns:
//
//	bool: Returns true if the image name complies with policy 4.5, otherwise false.
func (p *Processor) Policy4dot5Check(img_name string) bool {
	// This is a temporary hardcoded implementation of policy 4.5 for the OpenStack scanner PoC
	// This function will be replaced by the actual implementation of policy checks in the future
	// Policy 4.5 checks that the image name contains either "gardenlinux" or "SAP-compliant"

	if strings.Contains(img_name, "gardenlinux") || strings.Contains(img_name, "SAP-compliant") {
		return true
	}
	return false
}
