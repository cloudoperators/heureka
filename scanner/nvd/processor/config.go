// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor

type Config struct {
	HeurekaUrl          string `envconfig:"HEUREKA_URL" required:"true" json:"-"`
	HeurekaApiToken     string `envconfig:"HEUREKA_API_TOKEN" required:"true" json:"-"`
	IssueRepositoryName string `envconfig:"ISSUE_REPOSITORY_NAME" required:"true" default:"nvd" json:"-"`
	IssueRepositoryUrl  string `envconfig:"ISSUE_REPOSITORY_URL" required:"true" default:"https://nvd.nist.gov/" json:"-"`
	CveDetailsUrl       string `envconfig:"CVE_DETAILS_URL" required:"true" default:"https://nvd.nist.gov/vuln/detail/" json:"-"`
}
