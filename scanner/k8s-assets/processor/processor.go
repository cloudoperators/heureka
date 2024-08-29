// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"
	"fmt"
	"net/http"

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
	Count int
}

func (c CCRN) String() string {
	// Define default CCRN
	return fmt.Sprintf("rn.cloud.sap/ccrn/kubernetes/v1/%s/%s/%s/kubernikus/%s/%s/%s/%s/%s",
		c.Region,
		c.Domain,
		c.Project,
		c.Cluster,
		c.Namespace,
		c.Pod,
		c.PodID,
		c.Container,
	)
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
func (p *Processor) ProcessService(ctx context.Context, namespace string, serviceInfo scanner.ServiceInfo) (string, error) {
	var serviceId string

	// Create new Service
	createServiceInput := &client.ServiceInput{
		Name: serviceInfo.Name,
	}

	createServiceResp, err := client.CreateService(ctx, *p.Client, createServiceInput)
	if err != nil {
		// The Service might already exist in the DB
		// Let's try to fetch one Service by name
		_serviceId, err := p.getService(ctx, serviceInfo)
		if err != nil {
			return "", fmt.Errorf("failed to create Service %s: %w", serviceInfo.Name, err)
		}
		serviceId = _serviceId
	} else {
		serviceId = createServiceResp.CreateService.Id
	}
	return serviceId, nil
}

// getService returns (if any) a ServiceID
func (p *Processor) getService(ctx context.Context, serviceInfo scanner.ServiceInfo) (string, error) {
	var serviceId string

	listServicesFilter := client.ServiceFilter{
		ServiceName: []string{serviceInfo.Name},
	}
	listServicesResp, err := client.ListServices(ctx, *p.Client, &listServicesFilter)
	if err != nil {
		return "", fmt.Errorf("Couldn't list services")
	}

	// Return the first item
	if listServicesResp.Services.TotalCount > 0 {
		for _, s := range listServicesResp.Services.Edges {
			serviceId = s.Node.Id
			break
		}
	} else {
		return "", fmt.Errorf("ListServices returned no ServiceID")
	}

	return serviceId, nil
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

func (p *Processor) getComponentVersion(ctx context.Context, manifest string) (string, error) {
	var componentVersionId string

	listComponentVersionFilter := client.ComponentVersionFilter{
		Version: []string{manifest},
	}
	listCompoVersResp, err := client.ListComponentVersions(ctx, *p.Client, &listComponentVersionFilter)
	if err != nil {
		return "", fmt.Errorf("Couldn't list ComponentVersion")
	}

	if listCompoVersResp.ComponentVersions.TotalCount > 0 {
		for _, cv := range listCompoVersResp.ComponentVersions.Edges {
			componentVersionId = cv.Node.Id
			break
		}
	} else {
		return "", fmt.Errorf("ListComponentVersion returned no ComponentVersion objects")
	}
	return componentVersionId, nil
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
		return fmt.Errorf("Couldn't find ComponentVersion")
	}

	// Create new CCRN
	ccrn := CCRN{
		Region:    p.config.RegionName,
		Domain:    "x", // TODO
		Project:   "x", // TODO
		Cluster:   p.config.ClusterName,
		Namespace: namespace,
		Pod:       podGroupName,
		PodID:     podGroupName, // Use podGroupName instead of individual pod name
		Container: containerInfo.Name,
	}

	// Create new ComponentInstance
	componentInstanceInput := &client.ComponentInstanceInput{
		Ccrn:               ccrn.String(),
		Count:              1, // TODO
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
	}).Debug("Created new entities")

	return nil
}
