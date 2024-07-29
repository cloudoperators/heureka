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

	"github.wdf.sap.corp/cc/heureka/scanners/cveScanner/models"
	"golang.org/x/time/rate"
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

func NewScanner(baseURL string, apiKey string, resultsPerPage string, rl *rate.Limiter) *Scanner {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Scanner{
		BaseURL:        baseURL,
		ApiKey:         apiKey,
		ResultsPerPage: resultsPerPage,
		HTTPClient:     &http.Client{Transport: tr},
		RateLimiter:    rl,
	}
}

func (s *Scanner) GetCVEs(filter models.CveFilter) ([]models.CveItem, error) {
	index := 0
	totalResults := 1

	allCves := []models.CveItem{}

	for index < totalResults {
		cveResponse := models.CveResponse{}
		url := s.createUrl(filter, index)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		req.Header.Add("apiKey", s.ApiKey)

		resp, err := s.HTTPClient.Do(req)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if err = json.Unmarshal(body, &cveResponse); err != nil {
			fmt.Println(err)
			return nil, err
		}

		allCves = append(allCves, cveResponse.Vulnerabilities...)
		index += cveResponse.ResultsPerPage
		totalResults = cveResponse.TotalResults
	}

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
	return url
}
