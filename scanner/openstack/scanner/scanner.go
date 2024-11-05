// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/roles"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/users"
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
	ProjectId        string
}

func NewScanner(cfg Config) *Scanner {
	return &Scanner{
		Username:         cfg.OpenstackUsername,
		Password:         cfg.OpenstackPassword,
		Domain:           cfg.Domain,
		Project:          cfg.Project,
		IdentityEndpoint: cfg.IdentityEndpoint,
		ProjectDomain:    cfg.ProjectDomain,
		ProjectId:        cfg.ProjectId,
		Region:           cfg.Region,
	}
}

func (s *Scanner) CreateComputeClient() (*gophercloud.ServiceClient, error) {
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

func (s *Scanner) CreateIdentityClient() (*gophercloud.ServiceClient, error) {
	provider, err := s.newAuthenticatedProviderClient()
	if err != nil {
		return nil, err
	}

	identityClient, err := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{
		Region: s.Region,
	})
	if err != nil {
		panic(err)
	}

	return identityClient, nil
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

func (s *Scanner) GetUsers(service *gophercloud.ServiceClient, projectID string) []map[string]interface{} {
	// Return users and roles for project
	roleNamesByID, err := GetRoleNames(service)
	if err != nil {
		panic(err)
	}

	listOpts := roles.ListAssignmentsOpts{
		ScopeProjectID: projectID,
	}

	allPages, err := roles.ListAssignments(service, listOpts).AllPages()
	if err != nil {
		panic(err)
	}

	allRoleAssignments, err := roles.ExtractRoleAssignments(allPages)
	if err != nil {
		panic(err)
	}

	userIDs := make(map[string][]string)
	for _, assignment := range allRoleAssignments {
		userIDs[assignment.User.ID] = append(userIDs[assignment.User.ID], assignment.Role.ID)
	}

	config := make(map[string][]string)

	for userID, roles := range userIDs {
		for _, role := range roles {
			RoleName := roleNamesByID[role]
			config[userID] = append(config[userID], RoleName)
		}
	}

	userNamesByRoles := GetUserNamesbyRoles(service, config)

	return FormatServerOutput(userNamesByRoles)
}

func FormatServerOutput(userList map[string][]string) []map[string]interface{} {
	// Format compliance results for OPA Policy input
	var Configs []map[string]interface{}

	for user, roles := range userList {
		newConfig := map[string]interface{}{
			"user":  user,
			"roles": roles,
		}

		Configs = append(Configs, newConfig)
	}

	return Configs
}

func GetRoleNames(service *gophercloud.ServiceClient) (map[string]string, error) {
	// Translate role IDs to readable names
	allPages, err := roles.List(service, roles.ListOpts{}).AllPages()
	if err != nil {
		panic(err)
	}

	allRoles, err := roles.ExtractRoles(allPages)
	if err != nil {
		panic(err)
	}

	roleNamesByID := make(map[string]string)
	for _, role := range allRoles {
		roleNamesByID[role.ID] = role.Name
	}

	return roleNamesByID, err
}

func GetUserNamesbyRoles(service *gophercloud.ServiceClient, IdsandRoles map[string][]string) map[string][]string {
	// Translate User IDs into readable usernames, and return map of readable usernames and roles
	userNamesByID := make(map[string][]string)

	for id, roles := range IdsandRoles {
		if id != "" {
			user := users.Get(service, id)
			userdata := user.Result.Body.(map[string]interface{})["user"].(map[string]interface{})
			username := userdata["name"].(string)
			userNamesByID[username] = roles
		}
	}

	return userNamesByID
}

func (s *Scanner) Md5Hash(toHash string) string {
	// Create a new MD5 hash
	hash := md5.New()
	hash.Write([]byte(toHash))

	// Calculate the MD5 checksum
	md5Hash := hash.Sum(nil)

	// Convert the hash to a hexadecimal string
	hashString := hex.EncodeToString(md5Hash)

	return hashString
}
