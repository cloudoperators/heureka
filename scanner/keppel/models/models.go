// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"fmt"
	"strings"
	"time"

	ftypes "github.com/aquasecurity/trivy/pkg/fanal/types"
	"github.com/aquasecurity/trivy/pkg/types"
	"github.com/cloudoperators/heureka/scanners/keppel/client"
)

// For more attributes check
// https://github.com/sapcc/keppel/blob/master/internal/models/account.go
type Account struct {
	Name         string
	AuthTenantID string `json:"auth_tenant_id"`
}

type AccountResponse struct {
	Accounts []Account
}

type RepositoryResponse struct {
	Repositories []Repository
}

type Repository struct {
	Name          string `json:"name"`
	ManifestCount uint64 `json:"manifest_count"`
	TagCount      uint64 `json:"tag_count"`
	SizeBytes     uint64 `json:"size_bytes,omitempty"`
	PushedAt      int64  `json:"pushed_at,omitempty"`
}

type Manifest struct {
	Digest              string     `json:"digest"`
	MediaType           string     `json:"media_type"`
	SizeBytes           uint64     `json:"size_bytes"`
	PushedAt            uint64     `json:"pushed_at"`
	LastPulledAt        uint64     `json:"last_pulled_at"`
	VulnerabilityStatus string     `json:"vulnerability_status"`
	Labels              Labels
	Tags                []Tag      `json:"tags"`
	Children            []Manifest `json:"manifests"`
	MaxLayerCreatedAt   int64      `json:"max_layer_created_at"`
	MinLayerCreatedAt   int64      `json:"min_layer_created_at"`
	Platform            Platform   `json:"platform"`
}

type Platform struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
}

// ManifestResponse is the paginated response body of the Keppel
// GET /keppel/v1/accounts/{account}/repositories/{repo}/_manifests endpoint.
// When Truncated is true the caller must resend with marker set to the digest
// of the last manifest to retrieve the next page.
type ManifestResponse struct {
	Manifests []Manifest `json:"manifests"`
	Truncated bool       `json:"truncated"`
}

// Tag represents a single tag entry nested within a Keppel manifest listing
// response (GET /keppel/v1/accounts/{account}/repositories/{repo}/_manifests).
type Tag struct {
	Name         string `json:"name"`
	PushedAt     int64  `json:"pushed_at"`
	LastPulledAt int64  `json:"last_pulled_at"`
}

type Labels struct {
	Maintainer       string
	Maintainers      string
	SourceRepository string `json:"source_repository"`
}

type TrivyReport struct {
	SchemaVersion int            `json:",omitempty"`
	CreatedAt     time.Time      `json:",omitempty"`
	ArtifactName  string         `json:",omitempty"`
	ArtifactType  string         `json:",omitempty"` // generic replacement for original type `artifact.Type`
	Metadata      TrivyMetadata  `json:",omitempty"` // generic replacement for original type `types.Metadata`
	Results       []types.Result `json:",omitempty"`
}

type TrivyMetadata struct {
	Size int64      `json:",omitempty"`
	OS   *ftypes.OS `json:",omitempty"`

	// Container image
	ImageID     string         `json:",omitempty"`
	DiffIDs     []string       `json:",omitempty"`
	RepoTags    []string       `json:",omitempty"`
	RepoDigests []string       `json:",omitempty"`
	ImageConfig map[string]any `json:",omitempty"`
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error: status code %d, body: %s", e.StatusCode, e.Body)
}

func GetSeverityValue(severity string) client.SeverityValues {
	s := strings.ToLower(severity)
	switch s {
	case "critical":
		return client.SeverityValuesCritical
	case "high":
		return client.SeverityValuesHigh
	case "medium":
		return client.SeverityValuesMedium
	case "low":
		return client.SeverityValuesLow
	default:
		return client.SeverityValuesNone
	}
}
