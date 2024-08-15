// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"time"

	ftypes "github.com/aquasecurity/trivy/pkg/fanal/types"
	stypes "github.com/aquasecurity/trivy/pkg/module/serialize"
)

// For more attributes check
// https://github.com/sapcc/keppel/blob/master/internal/models/account.go
type Account struct {
	Name         string
	AuthTenantId string `json:"auth_tenant_id"`
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
	Digest              string
	MediaType           string `json:"media_type"`
	SizeBytes           uint64 `json:"size_bytes"`
	PushedAt            uint64 `json:"pushed_at"`
	LastPulledAt        uint64 `json:"last_pulled_at"`
	VulnerabilityStatus string `json:"vulnerability_status"`
	Labels              Labels
	MaxLayerCreatedAt   int64 `json:"max_layer_created_at"`
	MinLayerCreatedAt   int64 `json:"min_layer_created_at"`
}

type ManifestResponse struct {
	Manifests []Manifest
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
	Results       stypes.Results `json:",omitempty"` // compatible replacement for original type `types.Results`
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

type Component struct {
	ID   string  `json:"id"`
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
}

type ComponentVersion struct {
	ID          string  `json:"id"`
	Version     *string `json:"version,omitempty"`
	ComponentID *string `json:"componentId,omitempty"`
}
