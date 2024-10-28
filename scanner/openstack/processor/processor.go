// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"net/http"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
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

func (p *Processor) ProcessServers(serverList []servers.Server) ([]map[string]interface{}, error) {
	// This function processes the list of servers and checks if they are compliant with policy 4.5

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

func policy4dot5Check(img_name string) bool {
	// This is a temporary hardcoded implementation of policy 4.5 for the OpenStack scanner PoC
	// This function will be replaced by the actual implementation of policy checks in the future
	// Policy 4.5 checks that the image name contains either "gardenlinux" or "SAP-compliant"

	if strings.Contains(img_name, "gardenlinux") || strings.Contains(img_name, "SAP-compliant") {
		return true
	}
	return false
}
