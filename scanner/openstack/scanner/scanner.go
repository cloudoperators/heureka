// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	log "github.com/sirupsen/logrus"
)

type Scanner struct {
	IdentityEndpoint string
	Username         string
	Password         string
	AuthToken        string
	Domain           string
	Project          string
}

func NewScanner(cfg Config) *Scanner {
	return &Scanner{
		Username:         cfg.OpenstackUsername,
		Password:         cfg.OpenstackPassword,
		Domain:           cfg.Domain,
		Project:          cfg.Project,
		IdentityEndpoint: cfg.IdentityEndpoint,
	}
}

func (s *Scanner) Setup() (*gophercloud.ServiceClient, error) {
	client, err := s.CreateServiceClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create auth token: %w", err)
	}

	return client, nil
}

func (s *Scanner) CreateServiceClient() (*gophercloud.ServiceClient, error) {
	provider, err := s.newAuthenticatedProviderClient()
	if err != nil {
		return nil, err
	}
	endpointOpts := gophercloud.EndpointOpts{}

	iClient, err := openstack.NewIdentityV3(provider, endpointOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity v3 client: %w", err)
	}

	return iClient, nil
}

func (s *Scanner) newAuthenticatedProviderClient() (*gophercloud.ProviderClient, error) {
	opts := &tokens.AuthOptions{
		IdentityEndpoint: s.IdentityEndpoint,
		Username:         s.Username,
		Password:         s.Password,
		DomainName:       s.Domain,
		AllowReauth:      true,
		Scope: tokens.Scope{
			ProjectName: s.Project,
			DomainName:  s.Domain,
		},
	}

	provider, err := openstack.NewClient(opts.IdentityEndpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"identityEndpoint": opts.IdentityEndpoint,
			"domain":           s.Domain,
			"project":          s.Project,
		}).WithError(err)
		return nil, err
	}

	err = openstack.AuthenticateV3(provider, opts, gophercloud.EndpointOpts{})
	return provider, err
}

func (s *Scanner) GetServers(service *gophercloud.ServiceClient) ([]servers.Server, error) {
	// Use Service client to get server info
	listService := servers.List(service, nil)
	allPages, err := listService.AllPages()
	if err != nil {
		log.WithFields(log.Fields{
			"url": service.Endpoint,
		}).WithError(err).Error("Error during request in ListManifests")
		return nil, err
	}

	serverList, err := servers.ExtractServers(allPages)
	if err != nil {
		log.WithFields(log.Fields{
			"url": service.Endpoint,
		}).WithError(err).Error("Error during request in ListManifests")
		return nil, err
	}

	return serverList, nil
}

func GetProjects(service *gophercloud.ServiceClient) ([]string, error) {
	return []string{"project1", "project2"}, nil
}
