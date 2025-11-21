// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudoperators/heureka/scanners/keppel/models"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	"github.com/sapcc/go-api-declarations/bininfo"
	"github.com/sapcc/go-bits/httpext"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"k8s.io/client-go/util/retry"
)

type ImageInfo struct {
	Registry     string
	Account      string
	Organization string
	Repository   string
	Tag          string
}

type Scanner struct {
	KeppelBaseUrl    string
	IdentityEndpoint string
	Username         string
	UserDomain       string
	Password         string
	AuthToken        string
	Domain           string
	Project          string
	HTTPClient       *http.Client
	RateLimiter      *rate.Limiter
	TrivyRateLimiter *rate.Limiter
}

func (i ImageInfo) FullRepository() string {
	if len(i.Organization) > 0 {
		return fmt.Sprintf("%s/%s", i.Organization, i.Repository)
	} else {
		return i.Repository
	}
}

func NewScanner(cfg Config) *Scanner {
	orig := &http.Transport{}
	rt := http.RoundTripper(orig)
	wrap := httpext.WrapTransport(&rt)
	wrap.SetOverrideUserAgent("heureka-keppel-scanner", bininfo.CommitOr("unknown"))
	wrap.SetInsecureSkipVerify(true)

	rl := rate.NewLimiter(rate.Every(time.Minute/60), 10)    // 60 requests per minute
	trivyRl := rate.NewLimiter(rate.Every(time.Minute/5), 1) // 5 requests per minute

	return &Scanner{
		KeppelBaseUrl:    cfg.KeppelBaseUrl(),
		Username:         cfg.KeppelUsername,
		Password:         cfg.KeppelPassword,
		Domain:           cfg.Domain,
		UserDomain:       cfg.KeppelUserDomain,
		Project:          cfg.Project,
		IdentityEndpoint: cfg.IdentityEndpoint,
		HTTPClient:       &http.Client{Transport: rt},
		RateLimiter:      rl,
		TrivyRateLimiter: trivyRl,
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
		DomainName:       s.UserDomain,
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
	body, _, err := s.sendRequest(url, s.AuthToken)
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
	body, _, err := s.sendRequest(url, s.AuthToken)
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

// GetManifest returns a single manifest including child manifests from the image registry
func (s *Scanner) GetManifest(account string, repository string, digest string) (models.Manifest, error) {
	var manifest models.Manifest
	url := fmt.Sprintf("%s/v2/%s/%s/manifests/%s", s.KeppelBaseUrl, account, repository, digest)
	body, headers, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		return manifest, fmt.Errorf("couldn't get manifest for url: %s: %w", url, err)
	}

	var manifestResponse models.ManifestResponse
	if err = json.Unmarshal(body, &manifestResponse); err != nil {
		return manifest, fmt.Errorf("couldn't unmarshal body into a manifest response. url: %s, body: %s err: %w", url, body, err)
	}

	manifest.VulnerabilityStatus = headers.Get("X-Keppel-Vulnerability-Status")
	minLayerCreatedAt, err := strconv.ParseInt(headers.Get("X-Keppel-Min-Layer-Created-At"), 10, 64)
	if err != nil {
		minLayerCreatedAt = 0
	}
	maxLayerCreatedAt, err := strconv.ParseInt(headers.Get("X-Keppel-Max-Layer-Created-At"), 10, 64)
	if err != nil {
		maxLayerCreatedAt = 0
	}
	manifest.MinLayerCreatedAt = minLayerCreatedAt
	manifest.MaxLayerCreatedAt = maxLayerCreatedAt
	manifest.Digest = headers.Get("Docker-Content-Digest")

	manifest.Children = manifestResponse.Manifests

	return manifest, nil
}

func (s *Scanner) GetTrivyReport(account string, repository string, manifest string) (*models.TrivyReport, error) {
	url := fmt.Sprintf("%s/keppel/v1/accounts/%s/repositories/%s/_manifests/%s/trivy_report", s.KeppelBaseUrl, account, repository, manifest)

	body, _, err := s.sendRequest(url, s.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("couldn't get trivy report for url: %s: %w", url, err)
	}

	if strings.Contains(string(body), "no vulnerability report found") {
		return nil, nil
	}

	var trivyReport models.TrivyReport
	if err = json.Unmarshal(body, &trivyReport); err != nil {
		if strings.Contains(string(body), "not") {
			return nil, fmt.Errorf("trivy report not found for url: %s: %w", url, err)
		}

		return nil, fmt.Errorf("couldn't unmarshal body into a trivy report response. url: %s, body: %s err: %w", url, body, err)
	}

	return &trivyReport, nil

}

// extractImageInfo extracts image registry, image repository and the account name
// from a container image
func (s *Scanner) ExtractImageInfo(image string) (ImageInfo, error) {
	// Split the string to remove the tag
	imageAndTag := strings.Split(image, ":")
	if len(imageAndTag) < 1 {
		return ImageInfo{}, fmt.Errorf("invalid image")
	}

	// Split the remaining string by '/'
	parts := strings.Split(imageAndTag[0], "/")
	if len(parts) < 3 {
		return ImageInfo{}, fmt.Errorf("invalid image string format: at least registry, account and repository required")
	}

	info := ImageInfo{
		Registry:   parts[0],
		Account:    parts[1],
		Repository: strings.Join(parts[2:], "/"),
	}

	// Set tag if present
	if len(imageAndTag) > 1 {
		info.Tag = imageAndTag[1]
	}

	return info, nil
}

func (s *Scanner) sendRequest(url string, token string) ([]byte, http.Header, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't create new request: %w", err)
	}

	req.Header = http.Header{
		"X-Auth-Token":                          []string{token},
		"X-Keppel-No-Count-Towards-Last-Pulled": []string{"1"},
	}

	backoff := retry.DefaultBackoff
	backoff.Steps = 3                  // Maximum 3 retries
	backoff.Duration = 6 * time.Second // Start with 6 second delay
	backoff.Factor = 2.0               // Double the delay each retry
	backoff.Cap = 120 * time.Second    // Maximum delay of 120 seconds

	var responseBody []byte
	var responseHeaders http.Header

	// Use client-go's retry utility with custom backoff
	err = retry.OnError(backoff,
		// Retry on any error except for 401, 403, 404, 405 status codes
		func(err error) bool {
			if httpErr, ok := err.(*models.HTTPError); ok {
				return httpErr.StatusCode != http.StatusNotFound && httpErr.StatusCode != http.StatusMethodNotAllowed &&
					httpErr.StatusCode != http.StatusUnauthorized && httpErr.StatusCode != http.StatusForbidden
			}
			return true
		},
		// The operation to perform with retries
		func() error {
			resp, err := s.Do(req)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"url":   url,
				}).Warn("HTTP request failed")
				return err
			}

			// Ensure response body is closed after we're done with it
			defer resp.Body.Close()

			// Check if we got a successful response
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				// Read body before closing to get error details
				body, _ := io.ReadAll(resp.Body)
				errMsg := fmt.Sprintf("received HTTP %d: %s", resp.StatusCode, string(body))

				log.WithFields(log.Fields{
					"statusCode": resp.StatusCode,
					"url":        url,
					"response":   string(body),
				}).Warn(errMsg)

				return &models.HTTPError{
					StatusCode: resp.StatusCode,
					Body:       string(body),
				}
			}

			// Successfully got a 2xx response, read the body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			// Store the response body and headers in our outer variables
			responseBody = body
			responseHeaders = resp.Header
			return nil
		})

	if err != nil {
		return nil, nil, fmt.Errorf("request failed after retries: %w", err)
	}

	return responseBody, responseHeaders, nil
}

func (s *Scanner) Do(req *http.Request) (*http.Response, error) {
	var err error
	if strings.Contains(req.URL.String(), "trivy_report") {
		// Use the Trivy-specific rate limiter
		err = s.TrivyRateLimiter.Wait(context.Background())
	} else {
		// Use the general rate limiter
		err = s.RateLimiter.Wait(context.Background())
	}

	if err != nil {
		return nil, err
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
