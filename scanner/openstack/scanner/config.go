// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

type Config struct {
	OpenstackUsername string `envconfig:"OS_USERNAME" required:"true" json:"-"`
	OpenstackPassword string `envconfig:"OS_PASSWORD" required:"true" json:"-"`
	ProjectDomain     string `envconfig:"OS_PROJECT_DOMAIN_NAME" required:"true" json:"-"`
	Domain            string `envconfig:"OS_DOMAIN_NAME" required:"true" json:"-"`
	Project           string `envconfig:"OS_PROJECT_NAME" required:"true" json:"-"`
	Region            string `envconfig:"OS_REGION_NAME" required:"true" json:"-"`
	IdentityEndpoint  string `envconfig:"OS_AUTH_URL" required:"true" json:"-"`
	ScannerTimeout    string `envconfig:"SCANNER_TIMEOUT" default:"30m"`
}
