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
)

type Processor struct {
	Client *graphql.Client
}

func NewProcessor(cfg Config) *Processor {
	httpClient := http.Client{}
	gClient := graphql.NewClient(cfg.HeurekaUrl, &httpClient)
	return &Processor{
		Client: &gClient,
	}
}

// ProcessNamespace processes all services within a namespace
func (p *Processor) ProcessNamespace(ctx context.Context, namespace string, services []scanner.ServiceInfo) error {
	for _, serviceInfo := range services {
		if err := p.ProcessService(ctx, namespace, serviceInfo); err != nil {
			return fmt.Errorf("failed to process service %s in namespace %s: %w", serviceInfo.Name, namespace, err)
		}
	}
	return nil
}

// ProcessService creates a service and processes all its pods
func (p *Processor) ProcessService(ctx context.Context, namespace string, serviceInfo scanner.ServiceInfo) error {
	serviceID, err := p.Client.CreateService(ctx, serviceInfo.Name, namespace)
	if err != nil {
		return fmt.Errorf("failed to create Service %s: %w", serviceInfo.Name, err)
	}

	for _, podInfo := range serviceInfo.Pods {
		if err := p.ProcessPod(ctx, namespace, serviceID, podInfo); err != nil {
			return fmt.Errorf("failed to process pod %s: %w", podInfo.Name, err)
		}
	}

	return nil
}

// ProcessPod processes a single pod and its containers
func (p *Processor) ProcessPod(ctx context.Context, namespace string, serviceID string, podInfo scanner.PodInfo) error {
	for _, containerInfo := range podInfo.Containers {
		if err := p.ProcessContainer(ctx, namespace, serviceID, podInfo.Name, containerInfo); err != nil {
			return fmt.Errorf("failed to process container %s in pod %s: %w", containerInfo.Name, podInfo.Name, err)
		}
	}
	return nil
}

// ProcessContainer creates a ComponentVersion and ComponentInstance for a container
func (p *Processor) ProcessContainer(ctx context.Context, namespace string, serviceID string, podName string, containerInfo scanner.ContainerInfo) error {
	compo

	nentVersion, err := scanner.MapImageHashToComponentVersion(containerInfo.ImageHash)
	if err != nil {
		return fmt.Errorf("failed to map image hash to component version: %w", err)
	}

	componentVersionID, err := p.Client.CreateComponentVersion(ctx, componentVersion.Version, containerInfo.ImageHash)
	if err != nil {
		return fmt.Errorf("failed to create ComponentVersion: %w", err)
	}

	_, err = p.Client.CreateComponentInstance(ctx, componentVersionID, serviceID, podName)
	if err != nil {
		return fmt.Errorf("failed to create ComponentInstance: %w", err)
	}

	return nil
}
