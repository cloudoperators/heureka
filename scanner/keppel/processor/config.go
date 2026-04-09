// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

type Config struct {
	HeurekaUrl          string  `envconfig:"HEUREKA_URL"           required:"true" json:"-"`
	IssueRepositoryName string  `envconfig:"ISSUE_REPOSITORY_NAME" required:"true" json:"-" default:"nvd"`
	IssueRepositoryUrl  string  `envconfig:"ISSUE_REPOSITORY_URL"  required:"true" json:"-" default:"https://nvd.nist.gov/"`
	CveDetailsUrl       string  `envconfig:"CVE_DETAILS_URL"       required:"true" json:"-" default:"https://nvd.nist.gov/vuln/detail/"`
	HeurekaRateLimit    float64 `envconfig:"HEUREKA_RATE_LIMIT"                    json:"-" default:"100.0"`
	HeurekaRateBurst    int     `envconfig:"HEUREKA_RATE_BURST"                    json:"-" default:"100"`
}
