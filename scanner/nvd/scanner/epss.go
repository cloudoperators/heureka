// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/cloudoperators/heureka/scanner/nvd/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

// EpssClient fetches EPSS scores from the FIRST.org EPSS API.
type EpssClient struct {
	BaseURL      string
	BatchSize    int
	HTTPClient   *http.Client
	RateLimiter  *rate.Limiter
	fetchBackoff wait.Backoff
}

// NewEpssClient creates an EpssClient using the provided configuration.
func NewEpssClient(cfg Config) *EpssClient {
	if cfg.EpssBatchSize <= 0 {
		log.Fatalf("HEUREKA_EPSS_BATCH_SIZE must be a positive integer, got %d", cfg.EpssBatchSize)
	}

	rl := rate.NewLimiter(rate.Limit(cfg.EpssRateLimit), cfg.EpssRateBurst)

	return &EpssClient{
		BaseURL:     cfg.EpssApiUrl,
		BatchSize:   cfg.EpssBatchSize,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
		RateLimiter: rl,
		fetchBackoff: wait.Backoff{
			Steps:    3,
			Duration: 2 * time.Second,
			Factor:   2.0,
			Cap:      30 * time.Second,
		},
	}
}

// fetchBatch requests EPSS scores for a slice of CVE IDs from the FIRST.org API.
func (e *EpssClient) fetchBatch(ctx context.Context, cveIDs []string) (map[string]models.EpssData, error) {
	url := fmt.Sprintf("%s?cve=%s", e.BaseURL, strings.Join(cveIDs, ","))

	var result map[string]models.EpssData

	err := retry.OnError(
		e.fetchBackoff,
		func(err error) bool { return true },
		func() error {
			if waitErr := e.RateLimiter.Wait(ctx); waitErr != nil {
				return waitErr
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("failed to create EPSS request: %w", err)
			}

			resp, err := e.HTTPClient.Do(req)
			if err != nil {
				log.WithFields(log.Fields{"url": url, "error": err}).Warn("EPSS request failed")
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("EPSS API returned HTTP %d: %s", resp.StatusCode, string(body))
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read EPSS response body: %w", err)
			}

			var epssResp models.EpssResponse
			if err := json.Unmarshal(body, &epssResp); err != nil {
				return fmt.Errorf("failed to unmarshal EPSS response: %w", err)
			}

			result = make(map[string]models.EpssData, len(epssResp.Data))
			for _, d := range epssResp.Data {
				result[d.Cve] = d
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("EPSS batch fetch failed after retries: %w", err)
	}

	return result, nil
}

// EnrichCVEsWithEPSS fetches EPSS scores for all CVEs in the slice and writes
// the score and percentile back into each Cve struct in-place.
// CVEs with no EPSS entry are left with nil EpssScore/EpssPercentile.
func (e *EpssClient) EnrichCVEsWithEPSS(ctx context.Context, cves []models.CveItem) {
	if len(cves) == 0 {
		return
	}

	// Collect all CVE IDs.
	ids := make([]string, len(cves))
	for i, item := range cves {
		ids[i] = item.Cve.Id
	}

	// Batch-query FIRST.org — merge all batch results into one map.
	scores := make(map[string]models.EpssData, len(ids))

	for start := 0; start < len(ids); start += e.BatchSize {
		end := min(start+e.BatchSize, len(ids))

		batch, err := e.fetchBatch(ctx, ids[start:end])
		if err != nil {
			log.WithFields(log.Fields{
				"batchStart": start,
				"batchEnd":   end,
				"error":      err,
			}).Warn("Failed to fetch EPSS batch; scores for this batch will be missing")

			continue
		}

		maps.Copy(scores, batch)
	}

	log.WithFields(log.Fields{
		"totalCVEs":   len(cves),
		"epssMatches": len(scores),
	}).Info("EPSS enrichment complete")

	// Write scores back into the CVE items.
	for i := range cves {
		cveID := cves[i].Cve.Id
		if d, ok := scores[cveID]; ok {
			cves[i].Cve.EpssScore = models.ParseEpssValue(d.Epss)
			cves[i].Cve.EpssPercentile = models.ParseEpssValue(d.Percentile)
		}
	}
}
