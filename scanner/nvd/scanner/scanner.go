// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudoperators/heureka/scanner/nvd/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"k8s.io/client-go/util/retry"
)

type Scanner struct {
	BaseURL        string
	ApiKey         string
	ResultsPerPage string
	HTTPClient     *http.Client
	RateLimiter    *rate.Limiter
}

func (s *Scanner) Do(req *http.Request) (*http.Response, error) {
	err := s.RateLimiter.Wait(context.Background())
	if err != nil {
		return nil, err
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func NewScanner(cfg Config) *Scanner {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// The public rate limit (without an API key) is 5 requests in a rolling 30 second
	// window; the rate limit with an API key is 50 requests in a rolling 30 second window
	rl := rate.NewLimiter(rate.Every(30*time.Second/50), 50)

	return &Scanner{
		BaseURL:        cfg.NvdApiUrl,
		ApiKey:         cfg.NvdApiKey,
		ResultsPerPage: cfg.NvdResultsPerPage,
		HTTPClient:     &http.Client{Transport: tr},
		RateLimiter:    rl,
	}
}

// performRequest executes an HTTP request with retry logic using client-go's retry utility
func (s *Scanner) performRequest(url string) ([]byte, error) {
	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create new request: %w", err)
	}

	// Add API key to the request
	req.Header.Set("apiKey", s.ApiKey)

	// Create a custom backoff for the retry
	backoff := retry.DefaultBackoff
	backoff.Steps = 5                  // Maximum 5 retries
	backoff.Duration = 1 * time.Second // Start with 1 second delay
	backoff.Factor = 2.0               // Double the delay each retry
	backoff.Cap = 60 * time.Second     // Maximum delay of 60 seconds

	var responseBody []byte

	// Use client-go's retry utility with custom backoff
	err = retry.OnError(backoff,
		// Always retry on any error
		func(err error) bool { return true },
		// The operation to perform with retries
		func() error {
			// Execute the request through rate limiter
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

				return fmt.Errorf(errMsg)
			}

			// Successfully got a 2xx response, read the body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			// Store the response body in our outer variable
			responseBody = body
			return nil
		})

	if err != nil {
		return nil, fmt.Errorf("request failed after retries: %w", err)
	}

	return responseBody, nil
}

func (s *Scanner) GetCVEs(filter models.CveFilter) ([]models.CveItem, error) {
	index := 0
	totalResults := 1
	allCves := []models.CveItem{}

	for index < totalResults {
		url := s.createUrl(filter, index)

		// Execute the request with retry logic
		body, err := s.performRequest(url)
		if err != nil {
			return nil, err
		}

		// Parse the response
		var cveResponse models.CveResponse
		if err = json.Unmarshal(body, &cveResponse); err != nil {
			return nil, fmt.Errorf("couldn't unmarshal body into a CVE response: %w", err)
		}

		// Check if the response contains vulnerabilities
		if cveResponse.Vulnerabilities == nil {
			log.WithFields(log.Fields{
				"response": string(body),
			}).Warn("Response did not contain vulnerabilities array")
			return nil, fmt.Errorf("invalid response format: vulnerabilities array is missing")
		}

		// Append the vulnerabilities to our results
		allCves = append(allCves, cveResponse.Vulnerabilities...)

		// Update for pagination
		index += cveResponse.ResultsPerPage
		totalResults = cveResponse.TotalResults

		log.WithFields(log.Fields{
			"count":        len(cveResponse.Vulnerabilities),
			"index":        index,
			"totalResults": totalResults,
		}).Debug("Fetched CVE batch")
	}

	log.WithFields(log.Fields{
		"totalCVEs": len(allCves),
	}).Info("Successfully retrieved all CVEs")

	return allCves, nil
}

func (s *Scanner) createUrl(filter models.CveFilter, startIndex int) string {
	url := fmt.Sprintf("%s?resultsPerPage=%s&startIndex=%d", s.BaseURL, s.ResultsPerPage, startIndex)
	if filter.PubStartDate != "" {
		url += "&pubStartDate=" + filter.PubStartDate
	}
	if filter.PubEndDate != "" {
		url += "&pubEndDate=" + filter.PubEndDate
	}
	if filter.ModStartDate != "" {
		url += "&modStartDate=" + filter.ModStartDate
	}
	if filter.ModEndDate != "" {
		url += "&modEndDate=" + filter.ModEndDate
	}

	log.WithFields(log.Fields{
		"url": url,
	}).Debug("Created NVD URL")
	return url
}
