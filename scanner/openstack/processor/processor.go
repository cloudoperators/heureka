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
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	log "github.com/sirupsen/logrus"
)

type Processor struct {
	Client *graphql.Client
}

type ServiceInfo struct {
	CCRN         string
	SupportGroup string
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

// ProcessServers processes a list of servers and checks if they are compliant with policy 4.5.
// It returns a slice of maps, where each map contains the server name, server image name,
// and the compliance result.
//
// Parameters:
//
//	serverList []servers.Server - A list of servers to be processed.
//
// Returns:
//
//	[]map[string]interface{} - A slice of maps containing server details and compliance results.
//	error - An error if something goes wrong during processing.
func (p *Processor) ProcessServers(serverList []servers.Server) ([]map[string]interface{}, error) {

	output := []map[string]interface{}{}

	for _, server := range serverList {

		imgName := server.Metadata["image_name"]

		resultObj := map[string]interface{}{
			"server_name":       server.Name,
			"server_image_name": imgName,
		}

		if policy4dot5Check(imgName) {
			resultObj["result"] = "compliant"
		} else {
			resultObj["result"] = "non-compliant"
		}

		output = append(output, resultObj)
	}

	return output, nil
}

// policy4dot5Check checks if the given image name complies with policy 4.5.
// Policy 4.5 requires that the image name contains either "gardenlinux" or "SAP-compliant".
//
// Parameters:
//
//	img_name (string): The name of the image to be checked.
//
// Returns:
//
//	bool: Returns true if the image name complies with policy 4.5, otherwise false.
func policy4dot5Check(img_name string) bool {
	// This is a temporary hardcoded implementation of policy 4.5 for the OpenStack scanner PoC
	// This function will be replaced by the actual implementation of policy checks in the future
	// Policy 4.5 checks that the image name contains either "gardenlinux" or "SAP-compliant"

	if strings.Contains(img_name, "gardenlinux") || strings.Contains(img_name, "SAP-compliant") {
		return true
	}
	return false
}
