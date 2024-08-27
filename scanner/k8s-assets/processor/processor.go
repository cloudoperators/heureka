// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"net/http"

	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/cloudoperators/heureka/scanners/k8s-assets/client"
	"github.com/cloudoperators/heureka/scanners/k8s-assets/scanner"
	log "github.com/sirupsen/logrus"
)

type Processor struct {
	Client *graphql.Client
}

type CCRN struct {
	Region    string
	Cluster   string
	Namespace string
	Pod       string
	PodID     string
	Container string
}

func (c CCRN) String() string {
	// Define default CCRN
	return fmt.Sprintf("rn.cloud.sap/ccrn/kubernetes/v1/%s/-/-/kubernikus/%s/%s/%s/%s/%s",
		c.Region,
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

// ProcessPod processes a single pod and its containers
func (p *Processor) ProcessPod(ctx context.Context, namespace string, serviceID string, podInfo scanner.PodInfo) error {
	for _, containerInfo := range podInfo.Containers {
		if err := p.ProcessContainer(ctx, namespace, serviceID, podInfo, containerInfo); err != nil {
			return fmt.Errorf("failed to process container %s in pod %s: %w", containerInfo.Name, podInfo.Name, err)
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

func (p *Processor) getComponent(ctx context.Context, name string) (string, error) {
	var componentId string

	listComponentFilter := client.ComponentFilter{
		ComponentName: []string{name},
	}
	listComponentResp, err := client.ListComponents(ctx, *p.Client, &listComponentFilter)
	if err != nil {
		return "", fmt.Errorf("Couldn't list Components")
	}

	if listComponentResp.Components.TotalCount > 0 {
		for _, c := range listComponentResp.Components.Edges {
			componentId = c.Node.Id
			break
		}
	} else {
		return "", fmt.Errorf("ListComponents returned no Component objects")
	}
	return componentId, nil
}

// func (p *Processor) getComponentInstance(ctx context.Context, serviceId string) (string, error) {
// 	listComponentInstanceFilter := client.ComponentInstanceFilter{
// 		ServiceId: []string{serviceId},
// 	}
// 	listCompInstResp, err := client.ListComponentInstances(ctx, *p.Client, &listComponentInstanceFilter)
// 	if err != nil {
// 		return "", fmt.Errorf("couldn't list ComponentInstances: %w", err)
// 	}

// 	if listCompInstResp.ComponentInstances.TotalCount > 0 {
// 		return listCompInstResp.ComponentInstances.Edges[0].Node.Id, nil
// 	}
// 	return "", fmt.Errorf("no ComponentInstance found with CCRN: and ServiceId: %s", serviceId)
// }

// ProcessContainer creates a ComponentVersion and ComponentInstance for a container
func (p *Processor) ProcessContainer(ctx context.Context, namespace string, serviceID string, podInfo scanner.PodInfo, containerInfo scanner.ContainerInfo) error {

	// Find component version by container image hash
	componentVersionId, err := p.getComponentVersion(ctx, containerInfo.ImageHash)
	if err != nil {
		return fmt.Errorf("Couldn't find ComponentVersion")
	}

	// Create new CCRN
	ccrn := CCRN{
		Region:    "de",
		Cluster:   "testing",
		Namespace: namespace,
		Pod:       podInfo.Name,
		PodID:     podInfo.Name, // Change this!
		Container: containerInfo.Name,
	}

	// Create new ComponentInstance
	componentInstanceInput := &client.ComponentInstanceInput{
		Ccrn:               ccrn.String(),
		Count:              len(podInfo.Containers),
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
	}).Info("Created new entities")

	return nil
}
