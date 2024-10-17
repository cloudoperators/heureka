// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	log "github.com/sirupsen/logrus"
)

type ImageInfo struct {
	Registry     string
	Account      string
	Organization string
	Repository   string
}

type Scanner struct {
	KeppelBaseUrl    string
	IdentityEndpoint string
	Username         string
	Password         string
	AuthToken        string
	Domain           string
	Project          string
}

func (i ImageInfo) FullRepository() string {
	if len(i.Organization) > 0 {
		return fmt.Sprintf("%s/%s", i.Organization, i.Repository)
	} else {
		return i.Repository
	}
}

func NewScanner(cfg Config) *Scanner {
	return &Scanner{
		KeppelBaseUrl:    cfg.KeppelBaseUrl(),
		Username:         cfg.KeppelUsername,
		Password:         cfg.KeppelPassword,
		Domain:           cfg.Domain,
		Project:          cfg.Project,
		IdentityEndpoint: cfg.IdentityEndpoint,
	}
}

func (s *Scanner) Setup() error {
	token, err := s.CreateAuthToken()
	if err != nil {
		return fmt.Errorf("failed to create auth token: %w", err)
	}
	s.AuthToken = token
	return nil
}

func (s *Scanner) CreateAuthToken() (string, error) {
	provider, err := s.newAuthenticatedProviderClient()
	if err != nil {
		return "", err
	}
	endpointOpts := gophercloud.EndpointOpts{}

	iClient, err := openstack.NewIdentityV3(provider, endpointOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create identity v3 client: %w", err)
	}

	return iClient.Token(), nil
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

func (s *Scanner) ListAccounts() ([]models.Account, error) {
	url := fmt.Sprintf("%s/keppel/v1/accounts", s.KeppelBaseUrl)
	body, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
		}).WithError(err).Error("Error during request in ListAccounts")
		return nil, err
	}

	var accountResponse models.AccountResponse
	if err = json.Unmarshal(body, &accountResponse); err != nil {
		log.WithFields(log.Fields{
			"url":  url,
			"body": body,
		}).WithError(err).Error("Error during unmarshal in ListAccounts")
		return nil, err
	}

	return accountResponse.Accounts, nil
}

func (s *Scanner) ListRepositories(account string) ([]models.Repository, error) {
	url := fmt.Sprintf("%s/keppel/v1/accounts/%s/repositories", s.KeppelBaseUrl, account)
	body, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
		}).WithError(err).Error("Error during request in ListRepositories")
		return nil, err
	}

	var repositoryResponse models.RepositoryResponse
	if err = json.Unmarshal(body, &repositoryResponse); err != nil {
		log.WithFields(log.Fields{
			"url":  url,
			"body": body,
		}).WithError(err).Error("Error during unmarshal in ListRepositories")
		return nil, err
	}

	return repositoryResponse.Repositories, nil
}

func (s *Scanner) ListManifests(account string, repository string) ([]models.Manifest, error) {
	url := fmt.Sprintf("%s/keppel/v1/accounts/%s/repositories/%s/_manifests", s.KeppelBaseUrl, account, repository)
	body, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
		}).WithError(err).Error("Error during request in ListManifests")
		return nil, err
	}

	var manifestResponse models.ManifestResponse
	if err = json.Unmarshal(body, &manifestResponse); err != nil {
		log.WithFields(log.Fields{
			"url":  url,
			"body": body,
		}).WithError(err).Error("Error during unmarshal in ListManifests")
		return nil, err
	}

	return manifestResponse.Manifests, nil
}

// ListChildManifests is requred asa on Keppel not all Images are including vulnerability scan results directly on the
// top layer of the image and rather have the scan results on the child manifests. An prime example of this are multi-arch
// images where the scan results are  available on the child manifests with the respective concrete architecture.
// This method is using the v2 API endpoint as on the v1 of the API the child manifests listing is not available.
//
// Note: The v2 API does return slightly different results and therefore some of the fileds of models.Manifest are unset.
// This fact is accepted and no additional struct for parsing all information is implemented at this point in time
// as the additional available information is currently not utilized.
func (s *Scanner) ListChildManifests(account string, repository string, manifest string) ([]models.Manifest, error) {
	url := fmt.Sprintf("%s/v2/%s/%s/manifests/%s", s.KeppelBaseUrl, account, repository, manifest)
	body, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
		}).WithError(err).Error("Error during request in ListManifests")
		return nil, err
	}

	var manifestResponse models.ManifestResponse
	if err = json.Unmarshal(body, &manifestResponse); err != nil {
		log.WithFields(log.Fields{
			"url":  url,
			"body": body,
		}).WithError(err).Error("Error during unmarshal in ListManifests")
		return nil, err
	}

	return manifestResponse.Manifests, nil
}

func (s *Scanner) GetTrivyReport(account string, repository string, manifest string) (*models.TrivyReport, error) {
	url := fmt.Sprintf("%s/keppel/v1/accounts/%s/repositories/%s/_manifests/%s/trivy_report", s.KeppelBaseUrl, account, repository, manifest)
	body, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url}).
			WithError(err).Error("Error during GetTrivyReport")
		return nil, err
	}

	if strings.Contains(string(body), "no vulnerability report found") {
		return nil, nil
	}

	var trivyReport models.TrivyReport
	if err = json.Unmarshal(body, &trivyReport); err != nil {
		if strings.Contains(string(body), "not") {
			log.WithFields(log.Fields{
				"url":  url,
				"body": body,
			}).Info("Trivy report not found")
			return nil, fmt.Errorf("Trivy report not found")
		}

		log.WithFields(log.Fields{
			"url":  url,
			"body": body,
		}).WithError(err).Error("Error during unmarshal in GetTrivyReport")
		return nil, err
	}

	return &trivyReport, nil

}

// extractImageInfo extracts image registry, image repository and the account name
// from a container image
func (s *Scanner) ExtractImageInfo(image string) (ImageInfo, error) {
	// Split the string to remove the tag
	parts := strings.Split(image, ":")
	if len(parts) < 1 {
		return ImageInfo{}, fmt.Errorf("invalid image")
	}

	// Split the remaining string by '/'
	components := strings.Split(parts[0], "/")
	if len(components) < 3 || len(components) > 4 {
		return ImageInfo{}, fmt.Errorf("invalid image string format: expected 3 or 4 components")
	}

	info := ImageInfo{
		Registry:   components[0],
		Repository: components[len(components)-1],
	}

	if len(components) == 3 {
		info.Account = components[1]
	} else { // len(components) == 4
		info.Account = components[1]
		info.Organization = components[2]
	}

	return info, nil
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
		log.WithFields(log.Fields{
			"url":  url,
			"body": body,
		}).WithError(err).Error("Error during reading response body")
		return nil, err
	}

	return body, nil
}
