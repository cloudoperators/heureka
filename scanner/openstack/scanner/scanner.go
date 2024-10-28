// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	log "github.com/sirupsen/logrus"
)

type Scanner struct {
	IdentityEndpoint string
	Username         string
	ProjectDomain    string
	Password         string
	AuthToken        string
	Domain           string
	Project          string
	Region           string
}

func NewScanner(cfg Config) *Scanner {
	return &Scanner{
		Username:         cfg.OpenstackUsername,
		Password:         cfg.OpenstackPassword,
		Domain:           cfg.Domain,
		Project:          cfg.Project,
		IdentityEndpoint: cfg.IdentityEndpoint,
		ProjectDomain:    cfg.ProjectDomain,
		Region:           cfg.Region,
	}
}

func (s *Scanner) Setup() (*gophercloud.ServiceClient, error) {
	client, err := s.CreateServiceClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create service client: %w", err)
	}

	return client, nil
}

func (s *Scanner) CreateServiceClient() (*gophercloud.ServiceClient, error) {
	provider, err := s.newAuthenticatedProviderClient()
	if err != nil {
		return nil, err
	}

	computeClient, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: s.Region,
	})
	if err != nil {
		panic(err)
	}

	return computeClient, nil
}

func (s *Scanner) newAuthenticatedProviderClient() (*gophercloud.ProviderClient, error) {
	authOpts := gophercloud.AuthOptions{}
	authOpts.AllowReauth = true
	authOpts.Username = s.Username
	authOpts.Password = s.Password
	authOpts.IdentityEndpoint = s.IdentityEndpoint
	authOpts.DomainName = s.Domain

	if s.ProjectDomain != "" {
		authOpts.Scope = &gophercloud.AuthScope{
			ProjectName: s.Project,
			DomainName:  s.Domain,
		}
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, err
	}

	// Set a 60-second timeout
	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	provider.HTTPClient = *httpClient

	return provider, err
}

func (s *Scanner) GetServers(service *gophercloud.ServiceClient) ([]servers.Server, error) {
	// Use Service client to get server info
	listService := servers.List(service, nil)
	allPages, err := listService.AllPages()
	if err != nil {
		log.WithFields(log.Fields{
			"url": service.Endpoint,
		}).WithError(err).Error("Error during request in servers list")
		return nil, err
	}

	serverList, err := servers.ExtractServers(allPages)
	if err != nil {
		log.WithFields(log.Fields{
			"url": service.Endpoint,
		}).WithError(err).Error("Error during extracting pagination")
		return nil, err
	}

	return serverList, nil
}

func GetProjects(service *gophercloud.ServiceClient) ([]string, error) {
	return []string{"project1", "project2"}, nil
}
