// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	dbtypes "github.com/aquasecurity/trivy-db/pkg/types"
	"github.com/aquasecurity/trivy/pkg/types"
	"github.com/cloudoperators/heureka/scanners/keppel/models"
)

// keppelHandler serves the four Keppel endpoints the scanner calls. Repository
// names can contain slashes (organization/repo), which net/http path wildcards
// don't match, so we route by inspecting the path instead of registering
// patterns.
//
// All response bodies are built from the scanner's own structs in
// scanner/keppel/models (and the upstream trivy types), so a change to those
// types is a compile error here — that is the primary drift guard. A committed
// fixture under testdata/fixtures/<name>.json, if present, overrides the
// built-in default (see fixtures.go).
func keppelHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/keppel/v1/accounts":
			writeJSON(w, http.StatusOK, accountsResponse())
		case strings.HasPrefix(path, "/keppel/v1/accounts/") && strings.HasSuffix(path, "/repositories"):
			writeJSON(w, http.StatusOK, repositoriesResponse())
		case strings.HasPrefix(path, "/keppel/v1/accounts/") && strings.HasSuffix(path, "/trivy_report"):
			writeJSON(w, http.StatusOK, trivyResponse())
		case strings.HasPrefix(path, "/keppel/v1/accounts/") && strings.HasSuffix(path, "/_manifests"):
			// Tag listing used by the scanner's ListTags (available-version-update
			// feature). Returns one non-truncated page of manifests carrying tags.
			writeJSON(w, http.StatusOK, manifestListResponse())
		case strings.HasPrefix(path, "/v2/") && strings.Contains(path, "/manifests/"):
			digest := path[strings.LastIndex(path, "/manifests/")+len("/manifests/"):]
			writeManifest(w, digest)
		default:
			http.NotFound(w, r)
		}
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("keppel: encode response: %v", err)
	}
}

// writeManifest returns a single-arch manifest (no children) so the scanner
// proceeds to fetch a Trivy report. The scanner reads the vulnerability status
// and digest from response headers, not the body.
func writeManifest(w http.ResponseWriter, digest string) {
	w.Header().Set("Content-Type", "application/json")
	// "Low" is a valid (non-skipped) status, so the scanner fetches the report.
	w.Header().Set("X-Keppel-Vulnerability-Status", "Low")
	if digest != "" {
		w.Header().Set("Docker-Content-Digest", digest)
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(manifestResponse()); err != nil {
		log.Printf("keppel: encode manifest: %v", err)
	}
}

func accountsResponse() models.AccountResponse {
	v := models.AccountResponse{
		Accounts: []models.Account{{Name: "mock-account", AuthTenantID: "mock-tenant"}},
	}
	mustOverride("accounts", &v)
	return v
}

func repositoriesResponse() models.RepositoryResponse {
	v := models.RepositoryResponse{
		Repositories: []models.Repository{{Name: "mock-repo", ManifestCount: 1, TagCount: 1}},
	}
	mustOverride("repositories", &v)
	return v
}

// manifestResponse carries the children list only (empty = single-arch image).
func manifestResponse() models.ManifestResponse {
	v := models.ManifestResponse{Manifests: []models.Manifest{}}
	mustOverride("manifest", &v)
	return v
}

// manifestListResponse answers the paginated _manifests listing used by
// ListTags. It returns one non-truncated page of manifests whose Tags include
// several semantic versions (plus a floating "latest" that HighestSemverTag
// ignores), so the scanner records 2.3.1 as the available latest version.
func manifestListResponse() models.ManifestResponse {
	v := models.ManifestResponse{
		Truncated: false,
		Manifests: []models.Manifest{
			{
				Digest: "sha256:aaaa000000000000000000000000000000000000000000000000000000000001",
				Tags: []models.Tag{
					{Name: "1.0.0"},
					{Name: "2.3.1"},
					{Name: "2.0.5"},
					{Name: "latest"},
				},
			},
		},
	}
	mustOverride("manifest_list", &v)
	return v
}

func trivyResponse() models.TrivyReport {
	v := defaultTrivyReport()
	mustOverride("trivy_report", &v)
	return v
}

// defaultTrivyReport is built from the upstream trivy structs so that any
// breaking change to those types fails to compile here.
func defaultTrivyReport() models.TrivyReport {
	return models.TrivyReport{
		SchemaVersion: 2,
		CreatedAt:     time.Now().UTC(),
		ArtifactName:  "mock-account/mock-repo",
		ArtifactType:  "container_image",
		Results: []types.Result{
			{
				Target: "mock-account/mock-repo (debian 12)",
				Vulnerabilities: []types.DetectedVulnerability{
					{
						VulnerabilityID:  "CVE-2021-44228",
						PkgName:          "log4j",
						InstalledVersion: "2.14.1",
						FixedVersion:     "2.17.0",
						PrimaryURL:       "https://nvd.nist.gov/vuln/detail/CVE-2021-44228",
						Vulnerability: dbtypes.Vulnerability{
							Title:       "Log4Shell",
							Description: "Mock critical vulnerability (Log4Shell) for local scanner testing.",
							Severity:    "CRITICAL",
							CVSS: dbtypes.VendorCVSS{
								"nvd": dbtypes.CVSS{
									V3Vector: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
									V3Score:  10.0,
								},
							},
						},
					},
					{
						VulnerabilityID:  "CVE-2023-38545",
						PkgName:          "curl",
						InstalledVersion: "7.88.0",
						FixedVersion:     "8.4.0",
						PrimaryURL:       "https://nvd.nist.gov/vuln/detail/CVE-2023-38545",
						Vulnerability: dbtypes.Vulnerability{
							Title:       "curl SOCKS5 heap buffer overflow",
							Description: "Mock high severity vulnerability for local scanner testing.",
							Severity:    "HIGH",
							CVSS: dbtypes.VendorCVSS{
								"nvd": dbtypes.CVSS{
									V3Vector: "CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:U/C:H/I:H/A:H",
									V3Score:  8.1,
								},
							},
						},
					},
				},
			},
		},
	}
}
