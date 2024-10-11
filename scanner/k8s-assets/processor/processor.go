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
	Client *graphql.Client
	config Config
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
	ccrnTemplate := `rn.cloud.sap/ccrn/kubernetes/v1/{{.Region}}/-/-/-/{{.Cluster}}/{{.Namespace}}/{{.Pod}}/{{.Container}}`

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
	return &Processor{
		config: cfg,
		Client: &gClient,
	}
}

// ProcessService creates a service and processes all its pods
func (p *Processor) ProcessService(ctx context.Context, serviceInfo scanner.ServiceInfo) (string, error) {
	var serviceId string

	if serviceInfo.Name == "" {
		serviceInfo.Name = "none"
	}

	// The Service might already exist in the DB
	// Let's try to fetch one Service by name
	_serviceId, err := p.getService(ctx, serviceInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"serviceName": serviceInfo.Name,
		}).Error("failed to fetch service")

		// Create new Service
		createServiceInput := &client.ServiceInput{
			Name: serviceInfo.Name,
		}

		createServiceResp, err := client.CreateService(ctx, *p.Client, createServiceInput)
		if err != nil {
			return "", fmt.Errorf("failed to create Service %s: %w", serviceInfo.Name, err)
		} else {
			serviceId = createServiceResp.CreateService.Id
		}
	} else {
		serviceId = _serviceId
	}

	var supportGroupId string
	_supportGroupId, err := p.getSupportGroup(ctx, serviceInfo)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"serviceName": serviceInfo.Name,
		}).Error("failed to fetch service")

		// Create new SupportGroup
		createSupportGroupInput := &client.SupportGroupInput{
			Name: serviceInfo.SupportGroup,
		}
		createSupportGroupResp, err := client.CreateSupportGroup(ctx, *p.Client, createSupportGroupInput)
		if err != nil {
			return "", fmt.Errorf("failed to create SupportGroup %s: %w", serviceInfo.SupportGroup, err)
		} else {
			supportGroupId = createSupportGroupResp.CreateSupportGroup.Id
		}
	} else {
		supportGroupId = _supportGroupId
	}

	_, err = client.AddServiceToSupportGroup(ctx, *p.Client, supportGroupId, serviceId)

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"serviceName":  serviceInfo.Name,
			"supportGroup": serviceInfo.SupportGroup,
		}).Warning("Failed adding service to support group")
	}

	return serviceId, nil
}

func (p *Processor) getSupportGroup(ctx context.Context, serviceInfo scanner.ServiceInfo) (string, error) {
	var supportGroupId string

	listSupportGroupsFilter := client.SupportGroupFilter{
		SupportGroupName: []string{serviceInfo.SupportGroup},
	}
	listSupportGroupsResp, err := client.ListSupportGroups(ctx, *p.Client, &listSupportGroupsFilter)
	if err != nil {
		return "", fmt.Errorf("Couldn't list support groups")
	}

	// Return the first item
	if listSupportGroupsResp.SupportGroups.TotalCount > 0 && len(listSupportGroupsResp.SupportGroups.Edges) > 0 {
		supportGroupId = listSupportGroupsResp.SupportGroups.Edges[0].Node.Id
	} else {
		return "", fmt.Errorf("ListSupportGroups returned no SupportGroupID")
	}

	return supportGroupId, nil
}

// getService returns (if any) a ServiceID
func (p *Processor) getService(ctx context.Context, serviceInfo scanner.ServiceInfo) (string, error) {
	listServicesFilter := client.ServiceFilter{
		ServiceName: []string{serviceInfo.Name},
	}
	listServicesResp, err := client.ListServices(ctx, *p.Client, &listServicesFilter)
	if err != nil {
		return "", fmt.Errorf("Couldn't list services")
	}

	// Return the first item
	if listServicesResp.Services.TotalCount > 0 {
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

func (p *Processor) getComponentVersion(ctx context.Context, versionHash string) (string, error) {
	//separating image name and version hash
	imageAndVersion := strings.SplitN(versionHash, "@", 2)
	if len(imageAndVersion) < 2 {
		return "", fmt.Errorf("Couldn't split image and version")
	}
	image := imageAndVersion[0]
	version := imageAndVersion[1]

	//@todo Temporary Start
	// AS we do not scan all the individual registries, we replace the registry string
	// this is a temporary "hack" we need to move this to the heureka core with a registry configuration
	//so that the respective versions are created correctly during version creation

	var myMap map[string]string = make(map[string]string)
	myMap["keppel.global.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.qa-de-1.cloud.sap/ccloud-mirror"] = "keppel.eu-de-1.cloud.sap/ccloud"
	myMap["keppel.eu-de-2.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.s-eu-de-1.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.na-us-1.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.na-us-2.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.ap-jp-2.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.ap-jp-1.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.na-us-3.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.na-ca-1.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.eu-nl-1.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.ap-ae-1.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.ap-sa-1.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.ap-sa-2.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.ap-cn-1.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	myMap["keppel.ap-au-1.cloud.sap"] = "keppel.eu-de-1.cloud.sap"
	var images []string = make([]string, 1)

	for replace, with := range myMap {
		if strings.Contains(image, replace) {
			image = strings.Replace(image, replace, with, 1)
		}
	}
	//@todo Temporary End

	images[0] = image

	listComponentVersionFilter := client.ComponentVersionFilter{
		ComponentName: []string{image},
		Version:       []string{version},
	}
	listCompoVersResp, err := client.ListComponentVersions(ctx, *p.Client, &listComponentVersionFilter)
	if err != nil {
		return "", fmt.Errorf("Couldn't list ComponentVersion")
	}

	if listCompoVersResp.ComponentVersions.TotalCount > 0 {
		return listCompoVersResp.ComponentVersions.Edges[0].Node.Id, nil
	}

	return "", fmt.Errorf("ListComponentVersion returned no ComponentVersion objects")
}

// ProcessContainer creates a ComponentVersion and ComponentInstance for a container
func (p *Processor) ProcessContainer(
	ctx context.Context,
	namespace string,
	serviceID string,
	podGroupName string,
	containerInfo UniqueContainerInfo,
) error {
	// Find component version by container image hash
	componentVersionId, err := p.getComponentVersion(ctx, containerInfo.ImageHash)
	if err != nil {
		return fmt.Errorf("Couldn't find ComponentVersion (imageHash: %s): %w", containerInfo.ImageHash, err)
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

	// Create new Component
	componentInput := &client.ComponentInput{
		// TODO: Put name from container info
		Name: fmt.Sprintf("%s/%s/%s", "registry", "account.Name", "repository.Name"),
		Type: client.ComponentTypeValuesContainerimage,
	}
	createComponentResp, err := client.CreateComponent(ctx, *p.Client, componentInput)
	if err != nil {
		return fmt.Errorf("failed to create Component: %w", err)
	}
	log.WithFields(log.Fields{
		"componentId": createComponentResp.CreateComponent.Id,
	}).Info("Component created")

	// Create new ComponentVersion
	componentVersionInput := &client.ComponentVersionInput{
		Version:     containerInfo.ImageHash,
		ComponentId: createComponentResp.CreateComponent.Id,
	}
	createCompVersionResp, err := client.CreateComponentVersion(ctx, *p.Client, componentVersionInput)
	if err != nil {
		return fmt.Errorf("failed to create ComponentVersion: %w", err)
	}
	log.WithFields(log.Fields{
		"componentId": createCompVersionResp.CreateComponentVersion.Id,
	}).Info("ComponentVersion created")

	// Create new ComponentInstance
	componentInstanceInput := &client.ComponentInstanceInput{
		Ccrn:               ccrn.String(),
		Count:              containerInfo.Count,
		ComponentVersionId: componentVersionId,
		ServiceId:          serviceID,
	}
	createCompInstResp, err := client.CreateComponentInstance(ctx, *p.Client, componentInstanceInput)
	if err != nil {
		return fmt.Errorf("failed to create ComponentInstance: %w", err)
	}
	componentInstanceID := createCompInstResp.CreateComponentInstance.Id

	// Do logging
	log.WithFields(log.Fields{
		"componentVersionID":  componentVersionId,
		"componentInstanceID": componentInstanceID,
		"podGroup":            podGroupName,
		"container":           containerInfo.Name,
		"count":               containerInfo.Count,
	}).Info("Created new entities")

	return nil
}
