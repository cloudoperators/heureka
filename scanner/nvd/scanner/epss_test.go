// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudoperators/heureka/scanner/nvd/models"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/util/wait"
)

// testEpssClient returns an EpssClient pointed at baseURL with no rate-limiting
// and a single-attempt retry backoff (Steps=1) so failure tests run instantly.
func testEpssClient(baseURL string, batchSize int) *EpssClient {
	return &EpssClient{
		BaseURL:      baseURL,
		BatchSize:    batchSize,
		HTTPClient:   &http.Client{Timeout: 5 * time.Second},
		RateLimiter:  rate.NewLimiter(rate.Inf, 0),
		fetchBackoff: wait.Backoff{Steps: 1},
	}
}

func makeCVEItems(ids ...string) []models.CveItem {
	items := make([]models.CveItem, len(ids))
	for i, id := range ids {
		items[i] = models.CveItem{Cve: models.Cve{Id: id}}
	}
	return items
}

func epssResponse(data []models.EpssData) []byte {
	b, _ := json.Marshal(models.EpssResponse{Status: "OK", StatusCode: 200, Data: data})
	return b
}

// TestEnrichCVEsWithEPSS_BatchPartitioning verifies that a slice larger than
// BatchSize is split into the correct number of HTTP requests.
func TestEnrichCVEsWithEPSS_BatchPartitioning(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		ids := strings.Split(r.URL.Query().Get("cve"), ",")
		var data []models.EpssData
		for _, id := range ids {
			data = append(data, models.EpssData{Cve: id, Epss: "0.01000", Percentile: "0.50000"})
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(epssResponse(data))
	}))
	defer server.Close()

	// 5 CVEs, BatchSize=2 → ceil(5/2)=3 requests
	cves := makeCVEItems("CVE-2024-0001", "CVE-2024-0002", "CVE-2024-0003", "CVE-2024-0004", "CVE-2024-0005")
	testEpssClient(server.URL, 2).EnrichCVEsWithEPSS(context.Background(), cves)

	if got := int(requestCount.Load()); got != 3 {
		t.Errorf("expected 3 requests (2+2+1), got %d", got)
	}
	for _, item := range cves {
		if item.Cve.EpssScore == nil {
			t.Errorf("%s: expected EpssScore to be set, got nil", item.Cve.Id)
		}
		if item.Cve.EpssPercentile == nil {
			t.Errorf("%s: expected EpssPercentile to be set, got nil", item.Cve.Id)
		}
	}
}

// TestEnrichCVEsWithEPSS_MissingIDs verifies that CVEs absent from the API
// response retain nil EpssScore and EpssPercentile.
func TestEnrichCVEsWithEPSS_MissingIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only return data for CVE-2024-0001; CVE-2024-0002 is intentionally absent.
		w.Write(epssResponse([]models.EpssData{
			{Cve: "CVE-2024-0001", Epss: "0.04258", Percentile: "0.87612"},
		}))
	}))
	defer server.Close()

	cves := makeCVEItems("CVE-2024-0001", "CVE-2024-0002")
	testEpssClient(server.URL, 100).EnrichCVEsWithEPSS(context.Background(), cves)

	if cves[0].Cve.EpssScore == nil {
		t.Error("CVE-2024-0001: expected EpssScore, got nil")
	}
	if cves[1].Cve.EpssScore != nil {
		t.Errorf("CVE-2024-0002: expected nil EpssScore for missing CVE, got %v", *cves[1].Cve.EpssScore)
	}
	if cves[1].Cve.EpssPercentile != nil {
		t.Errorf("CVE-2024-0002: expected nil EpssPercentile for missing CVE, got %v", *cves[1].Cve.EpssPercentile)
	}
}

// TestEnrichCVEsWithEPSS_Non2xxResponse verifies that a non-2xx HTTP status
// causes the batch to be skipped and all CVEs in it keep nil scores.
func TestEnrichCVEsWithEPSS_Non2xxResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cves := makeCVEItems("CVE-2024-0001", "CVE-2024-0002")
	testEpssClient(server.URL, 100).EnrichCVEsWithEPSS(context.Background(), cves)

	for _, item := range cves {
		if item.Cve.EpssScore != nil {
			t.Errorf("%s: expected nil EpssScore after server error, got %v", item.Cve.Id, *item.Cve.EpssScore)
		}
	}
}

// TestEnrichCVEsWithEPSS_MalformedJSON verifies that a truncated/invalid JSON
// body causes the batch to be skipped and CVEs retain nil scores.
func TestEnrichCVEsWithEPSS_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK","data":[`)) // truncated
	}))
	defer server.Close()

	cves := makeCVEItems("CVE-2024-0001")
	testEpssClient(server.URL, 100).EnrichCVEsWithEPSS(context.Background(), cves)

	if cves[0].Cve.EpssScore != nil {
		t.Errorf("expected nil EpssScore after malformed JSON, got %v", *cves[0].Cve.EpssScore)
	}
}

// TestEnrichCVEsWithEPSS_PartialBatchFailure verifies that a failure in one
// batch does not prevent successfully-fetched batches from enriching their CVEs.
func TestEnrichCVEsWithEPSS_PartialBatchFailure(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callCount.Add(1) == 1 {
			// First batch succeeds.
			w.Write(epssResponse([]models.EpssData{
				{Cve: "CVE-2024-0001", Epss: "0.04258", Percentile: "0.87612"},
				{Cve: "CVE-2024-0002", Epss: "0.00100", Percentile: "0.30000"},
			}))
			return
		}
		// Second batch (and any retries) fails.
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	// BatchSize=2: batch 1 = [0001,0002], batch 2 = [0003,0004]
	cves := makeCVEItems("CVE-2024-0001", "CVE-2024-0002", "CVE-2024-0003", "CVE-2024-0004")
	testEpssClient(server.URL, 2).EnrichCVEsWithEPSS(context.Background(), cves)

	for _, id := range []string{"CVE-2024-0001", "CVE-2024-0002"} {
		for _, item := range cves {
			if item.Cve.Id == id && item.Cve.EpssScore == nil {
				t.Errorf("%s: expected EpssScore from successful batch, got nil", id)
			}
		}
	}
	for _, item := range cves[2:] {
		if item.Cve.EpssScore != nil {
			t.Errorf("%s: expected nil EpssScore after batch failure, got %v", item.Cve.Id, *item.Cve.EpssScore)
		}
	}
}

// TestEnrichCVEsWithEPSS_EmptyInput verifies that empty and nil CVE slices
// make no HTTP requests.
func TestEnrichCVEsWithEPSS_EmptyInput(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
	}))
	defer server.Close()

	c := testEpssClient(server.URL, 100)
	c.EnrichCVEsWithEPSS(context.Background(), nil)
	c.EnrichCVEsWithEPSS(context.Background(), []models.CveItem{})

	if n := int(requestCount.Load()); n != 0 {
		t.Errorf("expected no requests for empty input, got %d", n)
	}
}
