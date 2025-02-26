// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/cloudoperators/heureka/scanner/nvd/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const (
	// Retry configuration
	maxRetries     = 5
	initialBackoff = 4 * time.Second
	maxBackoff     = 60 * time.Second
	backoffFactor  = 2.0
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

	// Initialize scanner

	return &Scanner{
		BaseURL:        cfg.NvdApiUrl,
		ApiKey:         cfg.NvdApiKey,
		ResultsPerPage: cfg.NvdResultsPerPage,
		HTTPClient:     &http.Client{Transport: tr},
		RateLimiter:    rl,
	}
}

// calculateBackoff determines the backoff duration with pure exponential increase
func calculateBackoff(retry int) time.Duration {
	// Calculate exponential backoff
	backoff := float64(initialBackoff) * math.Pow(backoffFactor, float64(retry))

	// Apply maximum backoff limit
	if backoff > float64(maxBackoff) {
		backoff = float64(maxBackoff)
	}

	return time.Duration(backoff)
}

// doWithRetry executes a request with exponential backoff retry using recursion
func (s *Scanner) doWithRetry(req *http.Request, retryCount int) ([]byte, error) {
	// Check if we've exceeded maximum retries
	if retryCount > maxRetries {
		return nil, fmt.Errorf("maximum retries exceeded")
	}

	// Clone the request to ensure a fresh request for each attempt
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("apiKey", s.ApiKey)

	// Log retry attempts (except first attempt)
	if retryCount > 0 {
		log.WithFields(log.Fields{
			"retry": retryCount,
			"url":   reqClone.URL.String(),
		}).Info("Retrying request after error")
	}

	// Execute the request through rate limiter
	resp, err := s.Do(reqClone)

	// Handle connection-level errors
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"retry": retryCount,
		}).Warn("HTTP request failed")

		// Calculate backoff and wait
		backoff := calculateBackoff(retryCount)
		log.WithFields(log.Fields{
			"backoff": backoff.String(),
		}).Debug("Backing off before retry")

		time.Sleep(backoff)

		// Recursively retry
		return s.doWithRetry(req, retryCount+1)
	}

	// Ensure response body is closed after we're done with it
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success! Read the body and return
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return body, nil
	}

	// Handle specific error status codes
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		// Read body before closing to get error details
		body, _ := io.ReadAll(resp.Body)

		log.WithFields(log.Fields{
			"statusCode": resp.StatusCode,
			"response":   string(body),
			"retry":      retryCount,
		}).Warn("Received error status code")

		// Get retry-after header if available
		var backoff time.Duration

		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			// Try to parse Retry-After header as seconds
			if seconds, err := time.ParseDuration(retryAfter + "s"); err == nil {
				backoff = seconds
				log.WithFields(log.Fields{
					"retryAfter": retryAfter,
					"backoff":    backoff.String(),
				}).Debug("Using Retry-After header for backoff")
			}
		}

		// If Retry-After header is not available or invalid, use exponential backoff
		if backoff == 0 {
			backoff = calculateBackoff(retryCount)
		}

		log.WithFields(log.Fields{
			"backoff": backoff.String(),
		}).Debug("Backing off before retry")

		time.Sleep(backoff)

		// Recursively retry
		return s.doWithRetry(req, retryCount+1)
	}

	// Client errors (4xx) other than 429 are not retried
	body, _ := io.ReadAll(resp.Body)
	return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
}

func (s *Scanner) GetCVEs(filter models.CveFilter) ([]models.CveItem, error) {
	index := 0
	totalResults := 1
	allCves := []models.CveItem{}

	for index < totalResults {
		url := s.createUrl(filter, index)

		// Create a new request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("couldn't create new request: %w", err)
		}

		// Execute the request with recursive retry logic (starting with retry count 0)
		body, err := s.doWithRetry(req, 0)
		if err != nil {
			return nil, fmt.Errorf("request failed after retries: %w", err)
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
