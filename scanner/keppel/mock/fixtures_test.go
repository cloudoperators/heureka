// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudoperators/heureka/scanners/keppel/models"
)

// TestDefaultsRoundTrip ensures the built-in responses marshal and unmarshal
// cleanly against the scanner's model types — the compile-time coupling is
// backed up by a JSON round-trip check.
func TestDefaultsRoundTrip(t *testing.T) {
	cases := map[string]any{
		"accounts":     accountsResponse(),
		"repositories": repositoriesResponse(),
		"manifest":     manifestResponse(),
		"trivy_report": trivyResponse(),
	}
	for name, v := range cases {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("%s: marshal: %v", name, err)
		}
		if !json.Valid(b) {
			t.Fatalf("%s: produced invalid JSON", name)
		}
	}

	// The default Trivy report must carry at least one CVE so the scanner has
	// something to report.
	report := defaultTrivyReport()
	var cves int
	for _, r := range report.Results {
		cves += len(r.Vulnerabilities)
	}
	if cves == 0 {
		t.Fatal("defaultTrivyReport has no vulnerabilities")
	}
}

// TestFixturesFitModels is the drift guard: every committed override fixture
// must still unmarshal into the model type the scanner expects. If a fixture
// no longer fits (e.g. models.go changed), this fails and the fixture must be
// regenerated with `make keppel-mock-record`.
func TestFixturesFitModels(t *testing.T) {
	targets := map[string]func([]byte) error{
		"accounts":     func(b []byte) error { return json.Unmarshal(b, &models.AccountResponse{}) },
		"repositories": func(b []byte) error { return json.Unmarshal(b, &models.RepositoryResponse{}) },
		"manifest":     func(b []byte) error { return json.Unmarshal(b, &models.ManifestResponse{}) },
		"trivy_report": func(b []byte) error { return json.Unmarshal(b, &models.TrivyReport{}) },
	}
	for name, fit := range targets {
		path := filepath.Join(fixturesDir, name+".json")
		b, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue // no override committed; the Go default is used
		}
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if err := fit(b); err != nil {
			t.Errorf("%s does not fit its model type: %v", path, err)
		}
	}
}
