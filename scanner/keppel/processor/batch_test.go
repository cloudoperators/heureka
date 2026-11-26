// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aquasecurity/trivy/pkg/types"
	"github.com/cloudoperators/heureka/scanners/keppel/models"
)

// gqlHandler is a test helper that captures each incoming request body and
// responds with a pre-recorded JSON payload.
type gqlHandler struct {
	// responses is consumed in order; the last entry is reused once exhausted.
	responses []string
	mu        sync.Mutex
	bodies    []string
	idx       int32
}

func (h *gqlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	h.mu.Lock()
	h.bodies = append(h.bodies, string(body))
	h.mu.Unlock()

	idx := int(atomic.AddInt32(&h.idx, 1)) - 1
	if idx >= len(h.responses) {
		idx = len(h.responses) - 1
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(h.responses[idx]))
}

func newTestProcessor(t *testing.T, srv *httptest.Server, batchSize int) *Processor {
	t.Helper()
	cfg := Config{
		HeurekaUrl:          srv.URL,
		IssueRepositoryName: "nvd",
		IssueRepositoryUrl:  "https://nvd.nist.gov/",
		CveDetailsUrl:       "https://nvd.nist.gov/vuln/detail/",
		HeurekaRateLimit:    1000,
		HeurekaRateBurst:    1000,
		BatchSize:           batchSize,
	}
	p := NewProcessor(cfg, "test-tag")
	p.IssueRepositoryId = "repo-1"
	return p
}

func issueJSON(id, primaryName string) string {
	return `{"id":"` + id + `","primaryName":"` + primaryName + `","description":"desc","type":"Vulnerability"}`
}

func variantJSON(id, secondaryName, issueId string) string {
	return `{"id":"` + id + `","secondaryName":"` + secondaryName + `","issueId":"` + issueId + `"}`
}

func linkJSON(id string) string {
	return `{"id":"` + id + `"}`
}

func buildDataPayload(data map[string]string) string {
	b, _ := json.Marshal(map[string]interface{}{"data": jsonRawMap(data), "errors": []interface{}{}})
	return string(b)
}

func buildPartialPayload(data map[string]string, errAliases map[string]string) string {
	rawData := make(map[string]interface{}, len(data))
	for k, v := range data {
		rawData[k] = json.RawMessage(v)
	}
	for alias := range errAliases {
		rawData[alias] = json.RawMessage("null")
	}
	var errors []map[string]interface{}
	for alias, msg := range errAliases {
		errors = append(errors, map[string]interface{}{
			"message": msg,
			"path":    []string{alias},
		})
	}
	b, _ := json.Marshal(map[string]interface{}{"data": rawData, "errors": errors})
	return string(b)
}

// jsonRawMap embeds values verbatim rather than as double-encoded strings.
func jsonRawMap(m map[string]string) map[string]json.RawMessage {
	out := make(map[string]json.RawMessage, len(m))
	for k, v := range m {
		out[k] = json.RawMessage(v)
	}
	return out
}

// ---------------------------------------------------------------------------
// chunks
// ---------------------------------------------------------------------------

func TestChunks_EmptySlice(t *testing.T) {
	result := chunks([]int{}, 5)
	if result != nil {
		t.Fatalf("expected nil for empty input, got %v", result)
	}
}

func TestChunks_ZeroSize(t *testing.T) {
	result := chunks([]int{1, 2, 3}, 0)
	if result != nil {
		t.Fatalf("expected nil for n=0, got %v", result)
	}
}

func TestChunks_NegativeSize(t *testing.T) {
	result := chunks([]int{1, 2, 3}, -1)
	if result != nil {
		t.Fatalf("expected nil for n<0, got %v", result)
	}
}

func TestChunks_SingleItem(t *testing.T) {
	result := chunks([]int{42}, 10)
	if len(result) != 1 || len(result[0]) != 1 || result[0][0] != 42 {
		t.Fatalf("unexpected result: %v", result)
	}
}

func TestChunks_ExactMultiple(t *testing.T) {
	result := chunks([]int{1, 2, 3, 4}, 2)
	if len(result) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(result))
	}
	if result[0][0] != 1 || result[0][1] != 2 {
		t.Fatalf("unexpected first chunk: %v", result[0])
	}
	if result[1][0] != 3 || result[1][1] != 4 {
		t.Fatalf("unexpected second chunk: %v", result[1])
	}
}

func TestChunks_WithRemainder(t *testing.T) {
	result := chunks([]int{1, 2, 3, 4, 5}, 2)
	if len(result) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(result))
	}
	if len(result[2]) != 1 || result[2][0] != 5 {
		t.Fatalf("unexpected last chunk: %v", result[2])
	}
}

func TestChunks_SizeEqualsLength(t *testing.T) {
	result := chunks([]int{1, 2, 3}, 3)
	if len(result) != 1 || len(result[0]) != 3 {
		t.Fatalf("expected single chunk of 3, got %v", result)
	}
}

// ---------------------------------------------------------------------------
// batchCreateIssues
// ---------------------------------------------------------------------------

func TestBatchCreateIssues_AllSucceed(t *testing.T) {
	h := &gqlHandler{
		responses: []string{
			buildDataPayload(map[string]string{
				"i0": issueJSON("id-1", "CVE-2024-0001"),
				"i1": issueJSON("id-2", "CVE-2024-0002"),
			}),
		},
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	vulns := []newVuln{
		{primaryName: "CVE-2024-0001", description: "desc1"},
		{primaryName: "CVE-2024-0002", description: "desc2"},
	}

	results := p.batchCreateIssues(context.Background(), vulns)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results["CVE-2024-0001"].Id != "id-1" {
		t.Errorf("wrong id for CVE-2024-0001: %s", results["CVE-2024-0001"].Id)
	}
	if results["CVE-2024-0002"].Id != "id-2" {
		t.Errorf("wrong id for CVE-2024-0002: %s", results["CVE-2024-0002"].Id)
	}
}

func TestBatchCreateIssues_PartialAliasFail(t *testing.T) {
	h := &gqlHandler{
		responses: []string{
			buildPartialPayload(
				map[string]string{"i1": issueJSON("id-2", "CVE-2024-0002")},
				map[string]string{"i0": "already exists"},
			),
		},
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	vulns := []newVuln{
		{primaryName: "CVE-2024-0001", description: "desc1"},
		{primaryName: "CVE-2024-0002", description: "desc2"},
	}

	results := p.batchCreateIssues(context.Background(), vulns)

	if _, ok := results["CVE-2024-0001"]; ok {
		t.Error("CVE-2024-0001 should be absent (alias failed)")
	}
	if results["CVE-2024-0002"].Id != "id-2" {
		t.Errorf("CVE-2024-0002 should succeed, got %v", results["CVE-2024-0002"])
	}
}

func TestBatchCreateIssues_WholeChunkHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	vulns := []newVuln{
		{primaryName: "CVE-2024-0001"},
		{primaryName: "CVE-2024-0002"},
	}

	results := p.batchCreateIssues(context.Background(), vulns)

	if len(results) != 0 {
		t.Fatalf("expected empty results on HTTP error, got %v", results)
	}
}

func TestBatchCreateIssues_ChunkBoundary(t *testing.T) {
	singleChunkResp := func() string {
		return buildDataPayload(map[string]string{
			"i0": issueJSON("id-x", "CVE-X"),
			"i1": issueJSON("id-y", "CVE-Y"),
		})
	}

	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(singleChunkResp()))
	}))
	defer srv.Close()

	p := newTestProcessor(t, srv, 2)
	vulns := make([]newVuln, 5)
	for i := range vulns {
		vulns[i] = newVuln{primaryName: "CVE-placeholder"}
	}

	p.batchCreateIssues(context.Background(), vulns)

	if got := atomic.LoadInt32(&callCount); got != 3 {
		t.Fatalf("expected 3 HTTP calls for 5 vulns with batchSize=2, got %d", got)
	}
}

func TestBatchCreateIssues_ZeroBatchSize(t *testing.T) {
	h := &gqlHandler{
		responses: []string{
			buildDataPayload(map[string]string{
				"i0": issueJSON("id-1", "CVE-2024-0001"),
			}),
		},
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	// BatchSize=0 is clamped to 1; vulnerabilities must still be processed.
	p := newTestProcessor(t, srv, 0)
	if p.BatchSize != 1 {
		t.Fatalf("expected BatchSize clamped to 1, got %d", p.BatchSize)
	}

	vulns := []newVuln{{primaryName: "CVE-2024-0001"}}
	results := p.batchCreateIssues(context.Background(), vulns)

	if len(results) != 1 {
		t.Fatalf("expected 1 result after clamp, got %d", len(results))
	}
}

func TestBatchCreateIssues_RequestContainsAlias(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(buildDataPayload(map[string]string{
			"i0": issueJSON("id-1", "CVE-2024-0001"),
		})))
	}))
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	p.batchCreateIssues(context.Background(), []newVuln{{primaryName: "CVE-2024-0001", description: "desc"}})

	if !strings.Contains(captured, `i0: createIssue`) {
		t.Errorf("request body missing alias i0: createIssue, got: %s", captured)
	}
	if !strings.Contains(captured, "CVE-2024-0001") {
		t.Errorf("request body missing CVE name, got: %s", captured)
	}
}

// ---------------------------------------------------------------------------
// batchCreateIssueVariants
// ---------------------------------------------------------------------------

func TestBatchCreateIssueVariants_AllSucceed(t *testing.T) {
	h := &gqlHandler{
		responses: []string{
			buildDataPayload(map[string]string{
				"i0": variantJSON("v-1", "CVE-2024-0001", "id-1"),
				"i1": variantJSON("v-2", "CVE-2024-0002", "id-2"),
			}),
		},
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	vulns := []newVuln{
		{primaryName: "CVE-2024-0001", severity: "HIGH"},
		{primaryName: "CVE-2024-0002", severity: "LOW"},
	}
	issueIDByName := map[string]string{
		"CVE-2024-0001": "id-1",
		"CVE-2024-0002": "id-2",
	}

	failed := p.batchCreateIssueVariants(context.Background(), vulns, issueIDByName)

	if len(failed) != 0 {
		t.Fatalf("expected no failures, got %v", failed)
	}
}

func TestBatchCreateIssueVariants_VariantFailedSetPopulated(t *testing.T) {
	h := &gqlHandler{
		responses: []string{
			buildPartialPayload(
				map[string]string{"i1": variantJSON("v-2", "CVE-2024-0002", "id-2")},
				map[string]string{"i0": "constraint violation"},
			),
		},
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	vulns := []newVuln{
		{primaryName: "CVE-2024-0001", severity: "HIGH"},
		{primaryName: "CVE-2024-0002", severity: "LOW"},
	}
	issueIDByName := map[string]string{
		"CVE-2024-0001": "id-1",
		"CVE-2024-0002": "id-2",
	}

	failed := p.batchCreateIssueVariants(context.Background(), vulns, issueIDByName)

	if _, ok := failed["CVE-2024-0001"]; !ok {
		t.Error("CVE-2024-0001 should be in failed set")
	}
	if _, ok := failed["CVE-2024-0002"]; ok {
		t.Error("CVE-2024-0002 should not be in failed set")
	}
}

func TestBatchCreateIssueVariants_WholeChunkHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	vulns := []newVuln{
		{primaryName: "CVE-2024-0001", severity: "HIGH"},
		{primaryName: "CVE-2024-0002", severity: "LOW"},
	}
	issueIDByName := map[string]string{
		"CVE-2024-0001": "id-1",
		"CVE-2024-0002": "id-2",
	}

	failed := p.batchCreateIssueVariants(context.Background(), vulns, issueIDByName)

	if _, ok := failed["CVE-2024-0001"]; !ok {
		t.Error("CVE-2024-0001 should be in failed set on HTTP error")
	}
	if _, ok := failed["CVE-2024-0002"]; !ok {
		t.Error("CVE-2024-0002 should be in failed set on HTTP error")
	}
}

func TestBatchCreateIssueVariants_SeverityEnumNotQuoted(t *testing.T) {
	var captured string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(buildDataPayload(map[string]string{
			"i0": variantJSON("v-1", "CVE-2024-0001", "id-1"),
		})))
	}))
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	p.batchCreateIssueVariants(
		context.Background(),
		[]newVuln{{primaryName: "CVE-2024-0001", severity: "HIGH"}},
		map[string]string{"CVE-2024-0001": "id-1"},
	)

	// SeverityValues is a GQL enum — must appear without surrounding quotes.
	// e.g. "rating: High" not "rating: \"High\""
	if strings.Contains(captured, `rating: \"`) {
		t.Errorf("severity rating should not be quoted as a string in GQL, got: %s", captured)
	}
	if !strings.Contains(captured, "High") {
		t.Errorf("severity High missing from request, got: %s", captured)
	}
}

// ---------------------------------------------------------------------------
// batchAddComponentVersionToIssue
// ---------------------------------------------------------------------------

func TestBatchAddComponentVersionToIssue_AllSucceed(t *testing.T) {
	h := &gqlHandler{
		responses: []string{
			buildDataPayload(map[string]string{
				"i0": linkJSON("link-1"),
				"i1": linkJSON("link-2"),
			}),
		},
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	pairs := []linkPair{
		{primaryName: "CVE-2024-0001", issueId: "id-1", componentVersionId: "cv-1"},
		{primaryName: "CVE-2024-0002", issueId: "id-2", componentVersionId: "cv-1"},
	}

	var callCount int32
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(buildDataPayload(map[string]string{
			"i0": linkJSON("link-1"),
			"i1": linkJSON("link-2"),
		})))
	})

	p.batchAddComponentVersionToIssue(context.Background(), pairs)

	if got := atomic.LoadInt32(&callCount); got != 1 {
		t.Fatalf("expected 1 HTTP call, got %d", got)
	}
}

func TestBatchAddComponentVersionToIssue_PartialAliasFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(buildPartialPayload(
			map[string]string{"i1": linkJSON("link-2")},
			map[string]string{"i0": "not found"},
		)))
	}))
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	pairs := []linkPair{
		{primaryName: "CVE-2024-0001", issueId: "id-1", componentVersionId: "cv-1"},
		{primaryName: "CVE-2024-0002", issueId: "id-2", componentVersionId: "cv-1"},
	}

	p.batchAddComponentVersionToIssue(context.Background(), pairs)
}

func TestBatchAddComponentVersionToIssue_ChunkBoundary(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(buildDataPayload(map[string]string{
			"i0": linkJSON("link-x"),
			"i1": linkJSON("link-y"),
		})))
	}))
	defer srv.Close()

	p := newTestProcessor(t, srv, 2)
	pairs := make([]linkPair, 5)
	for i := range pairs {
		pairs[i] = linkPair{issueId: "id-x", componentVersionId: "cv-1"}
	}

	p.batchAddComponentVersionToIssue(context.Background(), pairs)

	if got := atomic.LoadInt32(&callCount); got != 3 {
		t.Fatalf("expected 3 HTTP calls, got %d", got)
	}
}

func TestBatchAddComponentVersionToIssue_EmptyPairs(t *testing.T) {
	var callCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
	}))
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)
	p.batchAddComponentVersionToIssue(context.Background(), nil)

	if got := atomic.LoadInt32(&callCount); got != 0 {
		t.Fatalf("expected 0 HTTP calls for empty pairs, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// ProcessReport — dedup across Results entries
// ---------------------------------------------------------------------------

// TestProcessReport_DuplicateCVEAcrossResults verifies that a CVE appearing in
// more than one Results entry is only created once.
func TestProcessReport_DuplicateCVEAcrossResults(t *testing.T) {
	var mu sync.Mutex
	var bodies []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(body))
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		// Determine which operation this is by inspecting the body.
		s := string(body)
		switch {
		case strings.Contains(s, "ListIssues"):
			_, _ = w.Write([]byte(`{"data":{"Issues":{"edges":[]}}}`))
		case strings.Contains(s, "createIssue"):
			_, _ = w.Write([]byte(buildDataPayload(map[string]string{
				"i0": issueJSON("id-1", "CVE-2024-0001"),
			})))
		case strings.Contains(s, "createIssueVariant"):
			_, _ = w.Write([]byte(buildDataPayload(map[string]string{
				"i0": variantJSON("v-1", "CVE-2024-0001", "id-1"),
			})))
		case strings.Contains(s, "addComponentVersionToIssue"):
			_, _ = w.Write([]byte(buildDataPayload(map[string]string{
				"i0": linkJSON("link-1"),
			})))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	p := newTestProcessor(t, srv, 25)

	report := models.TrivyReport{
		ArtifactName: "test-image",
		Results: []types.Result{
			{Vulnerabilities: []types.DetectedVulnerability{
				{VulnerabilityID: "CVE-2024-0001"},
			}},
			// Same CVE in a second result entry — must be deduplicated.
			{Vulnerabilities: []types.DetectedVulnerability{
				{VulnerabilityID: "CVE-2024-0001"},
			}},
		},
	}

	p.ProcessReport(report, "cv-1")

	createIssueCount := 0
	mu.Lock()
	for _, b := range bodies {
		if strings.Contains(b, "createIssue") && !strings.Contains(b, "createIssueVariant") {
			createIssueCount++
		}
	}
	mu.Unlock()

	if createIssueCount != 1 {
		t.Fatalf("expected exactly 1 createIssue batch request, got %d", createIssueCount)
	}
}
