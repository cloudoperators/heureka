// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"bytes"
	"text/template"

	"github.com/Khan/genqlient/graphql"
	"github.com/cloudoperators/heureka/scanners/k8s-assets/client"
	"github.com/cloudoperators/heureka/scanners/k8s-assets/scanner"
	log "github.com/sirupsen/logrus"
)

type Processor struct {
	Client    *graphql.Client
	config    Config
	advConfig *AdvancedConfig
}

type CCRN struct {
	Region    string
	Domain    string
	Project   string
	Cluster   string
	Namespace string
	Pod       string
	PodID     string
	Container string
}

type ImageVersion struct {
	Image   string
	Version string
}

// UniqueContainerInfo extends scanner.ContainerInfo to represent a unique container
// configuration within a pod replica set, adding a count of occurrences.
// It is used by the CollectUniqueContainers function to aggregate information about
// distinct containers across multiple pods.
type UniqueContainerInfo struct {
	scanner.ContainerInfo
	PodName      string
	GenerateName string
	Count        int
}

func (c CCRN) String() string {
	//actual CCRN template as per the CCRN spec of k8s_regsitry
	// Reference: https://github.wdf.sap.corp/PlusOne/resource-name/blob/main/ccrn-chart/templates/crds/k8s_registry/container.yaml
	ccrnTemplate := `ccrn: apiVersion=k8s-registry.ccrn.sap.cloud/v1, kind=container, cluster={{.Cluster}}, namespace={{.Namespace}}, pod={{.Pod}}, name={{.Container}}`

	tmpl, err := template.New("ccrn").Parse(ccrnTemplate)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, c)
	if err != nil {
		log.Error("Couldn't create CCRN string")
		return ""
	}
	return buf.String()
}

func NewProcessor(cfg Config) *Processor {
	httpClient := http.Client{}
	gClient := graphql.NewClient(cfg.HeurekaUrl, &httpClient)
	advCfg, err := cfg.LoadAdvancedConfig()
	if err != nil {
		log.WithError(err).Error("failed to load advanced config")
	}
	return &Processor{
		config:    cfg,
		Client:    &gClient,
		advConfig: advCfg,
	}
}

// ProcessService creates a service and processes all its pods
func (p *Processor) ProcessService(ctx context.Context, serviceInfo scanner.ServiceInfo) (string, error) {
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
		} else if createServiceResp.CreateService != nil && createServiceResp.CreateService.Id != "" {
			serviceId = createServiceResp.CreateService.Id
		}
	} else {
		serviceId = _serviceId
	}

	var supportGroupId string
	_supportGroupId, err := p.getSupportGroup(ctx, serviceInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"serviceCcrn": serviceInfo.CCRN,
		}).Error("failed to fetch service")

		// Create new SupportGroup
		createSupportGroupInput := &client.SupportGroupInput{
			Ccrn: serviceInfo.SupportGroup,
		}
		createSupportGroupResp, err := client.CreateSupportGroup(ctx, *p.Client, createSupportGroupInput)
		if err != nil {
			return "", fmt.Errorf("failed to create SupportGroup %s: %w", serviceInfo.SupportGroup, err)
		} else if createSupportGroupResp.CreateSupportGroup == nil {
			return "", fmt.Errorf("failed to create SupportGroup as CreateSupportGroup response is nil")
		} else {
			supportGroupId = createSupportGroupResp.CreateSupportGroup.Id
		}
	} else {
		supportGroupId = _supportGroupId
	}

	_, err = client.AddServiceToSupportGroup(ctx, *p.Client, supportGroupId, serviceId)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"serviceCcrn":  serviceInfo.CCRN,
			"supportGroup": serviceInfo.SupportGroup,
		}).Warning("Failed adding service to support group")
	}

	return serviceId, nil
}

func (p *Processor) getComponent(ctx context.Context, componentCcrn string) (string, error) {
	componetFilter := client.ComponentFilter{ComponentCcrn: []string{componentCcrn}}

	listComponentResp, err := client.ListComponents(ctx, *p.Client, &componetFilter)
	if err != nil {
		return "", fmt.Errorf("Couldn't list components")
	}

	if listComponentResp != nil && listComponentResp.Components != nil && listComponentResp.Components.TotalCount > 0 && len(listComponentResp.Components.Edges) > 0 {
		return listComponentResp.Components.Edges[0].Node.Id, nil
	}

	return "", fmt.Errorf("Component not found.")
}

func (p *Processor) getSupportGroup(ctx context.Context, serviceInfo scanner.ServiceInfo) (string, error) {
	var supportGroupId string

	listSupportGroupsFilter := client.SupportGroupFilter{
		SupportGroupCcrn: []string{serviceInfo.SupportGroup},
	}
	listSupportGroupsResp, err := client.ListSupportGroups(ctx, *p.Client, &listSupportGroupsFilter)
	if err != nil {
		return "", fmt.Errorf("Couldn't list support groups")
	}

	// Return the first item
	if listSupportGroupsResp != nil && listSupportGroupsResp.SupportGroups != nil && listSupportGroupsResp.SupportGroups.TotalCount > 0 && len(listSupportGroupsResp.SupportGroups.Edges) > 0 {
		supportGroupId = listSupportGroupsResp.SupportGroups.Edges[0].Node.Id
	} else {
		return "", fmt.Errorf("ListSupportGroups returned no SupportGroupID")
	}

	return supportGroupId, nil
}

// getService returns (if any) a ServiceID
func (p *Processor) getService(ctx context.Context, serviceInfo scanner.ServiceInfo) (string, error) {
	listServicesFilter := client.ServiceFilter{
		ServiceCcrn: []string{serviceInfo.CCRN},
	}
	listServicesResp, err := client.ListServices(ctx, *p.Client, &listServicesFilter)
	if err != nil {
		return "", fmt.Errorf("Couldn't list services")
	}

	// Return the first item
	if listServicesResp != nil && listServicesResp.Services != nil && listServicesResp.Services.TotalCount > 0 && len(listServicesResp.Services.Edges) > 0 {
		return listServicesResp.Services.Edges[0].Node.Id, nil
	}

	return "", fmt.Errorf("ListServices returned no ServiceID")
}

// CollectUniqueContainers processes a PodReplicaSet and returns a slice of
// unique container name, image name, and image ID combinations with their
// respective counts across all pods in the replica set.
//
// This function identifies unique combinations based on container name, image
// name, and image ID (referred to as ImageHash in the code). It counts how many
// times each unique combination appears across all pods in the replica set.
//
// Parameters:
//   - podReplicaSet: A scanner.PodReplicaSet object representing a group of related pods.
//
// Returns:
//   - []UniqueContainerInfo: A slice of UniqueContainerInfo structs. Each struct contains:
//   - ContainerInfo: The original container information (Name, Image, ImageHash).
//   - Count: The number of times this unique combination appears in the PodReplicaSet.
//
// The returned slice will contain one entry for each unique combination of
// container name, image name, and image ID found in the PodReplicaSet, along
// with a count of its occurrences.
func (p *Processor) CollectUniqueContainers(podReplicaSet scanner.PodReplicaSet) []UniqueContainerInfo {
	uniqueContainers := make(map[string]*UniqueContainerInfo)

	for _, pod := range podReplicaSet.Pods {
		for _, container := range pod.Containers {
			key := fmt.Sprintf("%s-%s", container.Name, container.ImageHash)
			if _, exists := uniqueContainers[key]; !exists {
				uniqueContainers[key] = &UniqueContainerInfo{
					ContainerInfo: container,
					PodName:       pod.Name,
					GenerateName:  strings.TrimSuffix(pod.GenerateName, "-"),
					Count:         0,
				}
			}
			uniqueContainers[key].Count++
		}
	}

	result := make([]UniqueContainerInfo, 0, len(uniqueContainers))
	for _, container := range uniqueContainers {
		result = append(result, *container)
	}

	return result
}

func (p *Processor) ProcessPodReplicaSet(ctx context.Context, namespace string, serviceID string, podReplicaSet scanner.PodReplicaSet) error {
	uniqueContainers := p.CollectUniqueContainers(podReplicaSet)

	for _, containerInfo := range uniqueContainers {
		if err := p.ProcessContainer(ctx, namespace, serviceID, podReplicaSet.GenerateName, containerInfo); err != nil {
			return fmt.Errorf("failed to process container %s in pod replica set %s: %w", containerInfo.Name, podReplicaSet.GenerateName, err)
		}
	}
	return nil
}

func (p *Processor) getComponentInstance(ctx context.Context, ccrn string) (string, error) {

	f := client.ComponentInstanceFilter{
		Ccrn: []string{ccrn},
	}

	listComponentInstancesResp, err := client.ListComponentInstances(ctx, *p.Client, &f)

	if err != nil {
		return "", fmt.Errorf("Couldn't list ComponentInstances")
	}

	if listComponentInstancesResp != nil && listComponentInstancesResp.ComponentInstances != nil && listComponentInstancesResp.ComponentInstances.TotalCount > 0 {
		return listComponentInstancesResp.ComponentInstances.Edges[0].Node.Id, nil
	}

	return "", fmt.Errorf("ListComponentInstances returned no ComponentInstance objects")
}

func (p *Processor) getComponentVersion(ctx context.Context, image string, version string) (string, error) {
	listComponentVersionFilter := client.ComponentVersionFilter{
		ComponentCcrn: []string{image},
		Version:       []string{version},
	}
	listCompoVersResp, err := client.ListComponentVersions(ctx, *p.Client, &listComponentVersionFilter)
	if err != nil {
		return "", fmt.Errorf("Couldn't list ComponentVersion")
	}

	if listCompoVersResp != nil && listCompoVersResp.ComponentVersions != nil && listCompoVersResp.ComponentVersions.TotalCount > 0 {
		return listCompoVersResp.ComponentVersions.Edges[0].Node.Id, nil
	}

	return "", fmt.Errorf("ListComponentVersion returned no ComponentVersion objects")
}

// extractVersion returns the hash part ia container image
func (p *Processor) extractImageVersion(versionHash string) (*ImageVersion, error) {
	//separating image name and version hash
	imageAndVersion := strings.SplitN(versionHash, "@", 2)
	if len(imageAndVersion) < 2 {
		return nil, fmt.Errorf("Couldn't split image and version")
	}
	imageVersion := &ImageVersion{
		Image:   imageAndVersion[0],
		Version: imageAndVersion[1],
	}
	return imageVersion, nil
}

// ProcessContainer is responsible for creating several entities based on the
// information at container level
func (p *Processor) ProcessContainer(
	ctx context.Context,
	namespace string,
	serviceID string,
	podGroupName string,
	containerInfo UniqueContainerInfo,
) error {
	var (
		componentId         string
		componentVersionId  string
		componentInstanceId string
	)

	if p.advConfig != nil {

		if podSideCar, ok := p.advConfig.GetSideCar(containerInfo.Name); ok {
			// get the service id
			sid, err := p.ProcessService(ctx, scanner.ServiceInfo{
				SupportGroup: podSideCar.SupportGroup,
				CCRN:         podSideCar.ServiceName,
			})
			if err != nil {
				log.WithError(err).Error("failed to process service")
			} else {
				//overwrite the ServiceID for the component instance
				serviceID = sid
			}
		}
	}

	// Create new CCRN
	ccrn := CCRN{
		Region:    p.config.RegionName,
		Domain:    "-", // Not used at the moment
		Project:   "-", // Not used at the moment
		Cluster:   p.config.ClusterName,
		Namespace: namespace,
		Pod:       containerInfo.GenerateName,
		PodID:     containerInfo.PodName,
		Container: containerInfo.Name,
	}

	//
	// Create new Component
	//

	// Check if we have everything we need
	if len(containerInfo.ImageRegistry) == 0 || len(containerInfo.ImageAccount) == 0 || len(containerInfo.ImageRepository) == 0 {
		return fmt.Errorf("cannot create Component (one or more containerInfo fields are empty)")
	}

	// we only consider main registry for now
	if strings.HasPrefix(containerInfo.ImageRegistry, "keppel") {
		containerInfo.ImageRegistry = "keppel.eu-de-1.cloud.sap"
	}

	componentCcrn := fmt.Sprintf("%s/%s/%s", containerInfo.ImageRegistry, containerInfo.ImageAccount, containerInfo.ImageRepository)

	componentId, err := p.getComponent(ctx, componentCcrn)

	if err != nil {
		componentId, err = p.createComponent(ctx, &client.ComponentInput{
			Ccrn: componentCcrn,
			Type: client.ComponentTypeValuesContainerimage,
		})
		if err != nil {
			return fmt.Errorf("failed to create Component. %w", err)
		}
	}

	if componentId == "" {
		return fmt.Errorf("failed to create Component. ComponentId is empty")
	}

	//
	// Create new ComponentVersion
	//
	iv, err := p.extractImageVersion(containerInfo.ImageHash)
	if err != nil {
		log.WithFields(log.Fields{
			"imageHash": containerInfo.ImageHash,
		}).Error("cannot extract image and version from imagehash")
		return fmt.Errorf("cannot extract image version: %w", err)
	}
	componentVersionId, err = p.getComponentVersion(ctx, componentCcrn, iv.Version)
	if err != nil {
		log.WithFields(log.Fields{
			"id": containerInfo.ImageHash,
		}).Info("ComponentVersion not found")

		componentVersionId, err = p.createComponentVersion(
			ctx,
			iv.Version,
			componentId,
			containerInfo.ImageTag,
			containerInfo.ImageRepository,
			containerInfo.ImageAccount,
		)
		if err != nil {
			return fmt.Errorf("failed to create ComponentVersion: %w", err)
		}
	}

	componentInstanceId, err = p.getComponentInstance(ctx, ccrn.String())
	input := &client.ComponentInstanceInput{
		Ccrn:               ccrn.String(),
		Count:              containerInfo.Count,
		ComponentVersionId: componentVersionId,
		ServiceId:          serviceID,
	}

	if err == nil {
		componentInstanceId, err = p.updateComponentInstance(ctx, componentInstanceId, input)
		if err != nil {
			return fmt.Errorf("failed to update ComponentInstance: %w", err)
		}
	} else {
		//
		// Create or update ComponentInstance
		//
		componentInstanceId, err = p.createComponentInstance(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to create ComponentInstance: %w", err)
		}
	}
	// Do logging
	log.WithFields(log.Fields{
		"componentVersionID":  componentVersionId,
		"componentInstanceID": componentInstanceId,
		"podGroup":            podGroupName,
		"container":           containerInfo.Name,
		"count":               containerInfo.Count,
	}).Info("Created new entities")

	return nil
}

// createComponent creates a new Component
func (p *Processor) createComponent(ctx context.Context, input *client.ComponentInput) (string, error) {
	createComponentResp, err := client.CreateComponent(ctx, *p.Client, input)

	if err != nil {
		return "", fmt.Errorf("failed to create Component: %w", err)
	} else if createComponentResp.CreateComponent == nil {
		return "", fmt.Errorf("failed to create Component as CreateComponent response is nil")
	}

	log.WithFields(log.Fields{
		"componentId": createComponentResp.CreateComponent.Id,
	}).Info("Component created")

	return createComponentResp.CreateComponent.Id, nil
}

// createComponentVersion create a new ComponentVersion based on a container image hash
func (p *Processor) createComponentVersion(
	ctx context.Context,
	version string,
	componentId string,
	tag string,
	repository string,
	organization string,
) (string, error) {
	componentVersionInput := &client.ComponentVersionInput{
		Version:      version,
		ComponentId:  componentId,
		Tag:          tag,
		Repository:   repository,
		Organization: organization,
	}
	createCompVersionResp, err := client.CreateComponentVersion(ctx, *p.Client, componentVersionInput)
	if err != nil {
		return "", fmt.Errorf("failed to create ComponentVersion: %w", err)
	} else if createCompVersionResp.CreateComponentVersion == nil {
		return "", fmt.Errorf("failed to create ComponentVersion as CreateComponentVersion response is nil")
	}

	log.WithFields(log.Fields{
		"componentVersionId": createCompVersionResp.CreateComponentVersion.Id,
		"tag":                tag,
		"repository":         repository,
		"organization":       organization,
	}).Info("ComponentVersion created")

	return createCompVersionResp.CreateComponentVersion.Id, nil
}

// createComponentInstance creates a new ComponentInstance
func (p *Processor) createComponentInstance(ctx context.Context, input *client.ComponentInstanceInput) (string, error) {
	createCompInstResp, err := client.CreateComponentInstance(ctx, *p.Client, input)
	if err != nil {
		return "", fmt.Errorf("failed to create ComponentInstance: %w", err)
	} else if createCompInstResp.CreateComponentInstance == nil {
		return "", fmt.Errorf("failed to create ComponentInstance as CreateComponentInstance response is nil")
	}

	log.WithFields(log.Fields{
		"componentInstanceId": createCompInstResp.CreateComponentInstance.Id,
	}).Info("ComponentInstance created")

	return createCompInstResp.CreateComponentInstance.Id, nil
}

func (p *Processor) updateComponentInstance(ctx context.Context, id string, input *client.ComponentInstanceInput) (string, error) {
	updateCompInstResp, err := client.UpdateComponentInstance(ctx, *p.Client, id, input)
	if err != nil {
		return "", fmt.Errorf("failed to update ComponentInstance: %w", err)
	} else if updateCompInstResp.UpdateComponentInstance == nil {
		return "", fmt.Errorf("failed to update ComponentInstance as CreateComponentInstance response is nil")
	}

	log.WithFields(log.Fields{
		"componentInstanceId": updateCompInstResp.UpdateComponentInstance.Id,
	}).Info("ComponentInstance updated")

	return updateCompInstResp.UpdateComponentInstance.Id, nil
}
