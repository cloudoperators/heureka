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

// ProcessContainer creates a ComponentVersion and ComponentInstance for a container
func (p *Processor) ProcessContainer(ctx context.Context, namespace string, serviceID string, podInfo scanner.PodInfo, containerInfo scanner.ContainerInfo) error {

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
