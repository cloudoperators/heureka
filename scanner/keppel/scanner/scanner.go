// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	"github.com/pkg/errors"
)

type Scanner struct {
	KeppelBaseUrl   string
	IdentityBaseUrl string
	Username        string
	Password        string
	AuthToken       string
	Region          string
	Domain          string
	Project         string
}

func NewScanner(cfg Config) *Scanner {
	return &Scanner{
		KeppelBaseUrl: cfg.KeppelBaseUrl,
		Username:      cfg.KeppelUsername,
		Password:      cfg.KeppelPassword,
		Region:        cfg.Region,
		Domain:        cfg.Domain,
		Project:       cfg.Project,
		// AuthToken: ,
	}
}

func (s *Scanner) Setup() error {
	token, err := s.CreateAuthToken()
	if err != nil {
		fmt.Println(err)
		return err
	}
	s.AuthToken = token
	return nil
}

func (s *Scanner) CreateAuthToken() (string, error) {
	provider, err := s.newAuthenticatedProviderClient()
	if err != nil {
		return "", errors.Wrap(err, "failed to authenticate")
	}
	endpointOpts := gophercloud.EndpointOpts{}

	iClient, err := openstack.NewIdentityV3(provider, endpointOpts)
	if err != nil {
		return "", errors.Wrap(err, "failed to create identity v3 client")
	}

	return iClient.Token(), nil
}

func (s *Scanner) newAuthenticatedProviderClient() (*gophercloud.ProviderClient, error) {

	opts := &tokens.AuthOptions{
		IdentityEndpoint: fmt.Sprintf("https://identity-3.%s.cloud.sap/v3", s.Region),
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
		return nil, err
	}

	err = openstack.AuthenticateV3(provider, opts, gophercloud.EndpointOpts{})
	return provider, err
}

func (s *Scanner) ListAccounts() ([]models.Account, error) {
	url := fmt.Sprintf("https://keppel.%s.cloud.sap/keppel/v1/accounts", s.Region)
	body, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		return nil, err
	}

	var accountResponse models.AccountResponse
	if err = json.Unmarshal(body, &accountResponse); err != nil {
		return nil, err
	}

	return accountResponse.Accounts, nil
}

func (s *Scanner) ListRepositories(account string) ([]models.Repository, error) {
	url := fmt.Sprintf("https://keppel.%s.cloud.sap/keppel/v1/accounts/%s/repositories", s.Region, account)
	body, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		return nil, err
	}

	var repositoryResponse models.RepositoryResponse
	if err = json.Unmarshal(body, &repositoryResponse); err != nil {
		return nil, err
	}

	return repositoryResponse.Repositories, nil
}

func (s *Scanner) ListManifests(account string, repository string) ([]models.Manifest, error) {
	url := fmt.Sprintf("https://keppel.%s.cloud.sap/keppel/v1/accounts/%s/repositories/%s/_manifests", s.Region, account, repository)
	body, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		return nil, err
	}

	var manifestResponse models.ManifestResponse
	if err = json.Unmarshal(body, &manifestResponse); err != nil {
		return nil, err
	}

	return manifestResponse.Manifests, nil
}

func (s *Scanner) GetTrivyReport(account string, repository string, manifest string) (*models.TrivyReport, error) {
	url := fmt.Sprintf("https://keppel.%s.cloud.sap/keppel/v1/accounts/%s/repositories/%s/_manifests/%s/trivy_report", s.Region, account, repository, manifest)
	body, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		return nil, err
	}

	var trivyReport models.TrivyReport
	if err = json.Unmarshal(body, &trivyReport); err != nil {
		return nil, err
	}

	return &trivyReport, nil

}

func (s *Scanner) sendRequest(url string, token string) ([]byte, error) {
	client := new(http.Client)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"X-Auth-Token": []string{token},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
