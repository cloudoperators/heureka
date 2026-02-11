// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package model

// SIEM models (kept separate to avoid touching generated models.go)

type SIEMAlertInput struct {
	Name         *string         `json:"name,omitempty"`
	Description  *string         `json:"description,omitempty"`
	Severity     *SeverityValues `json:"severity,omitempty"`
	URL          *string         `json:"url,omitempty"`
	Service      *string         `json:"service,omitempty"`
	SupportGroup *string         `json:"supportGroup,omitempty"`
	Region       *string         `json:"region,omitempty"`
	Cluster      *string         `json:"cluster,omitempty"`
	Namespace    *string         `json:"namespace,omitempty"`
	Pod          *string         `json:"pod,omitempty"`
	Container    *string         `json:"container,omitempty"`
}

type SIEMAlert struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Severity    *SeverityValues `json:"severity,omitempty"`
	URL         *string         `json:"url,omitempty"`
	// Optional fields — returned to clients for convenience
	Service      *string `json:"service,omitempty"`
	SupportGroup *string `json:"supportGroup,omitempty"`
	Region       *string `json:"region,omitempty"`
	Cluster      *string `json:"cluster,omitempty"`
	Namespace    *string `json:"namespace,omitempty"`
	Pod          *string `json:"pod,omitempty"`
	Container    *string `json:"container,omitempty"`
}
