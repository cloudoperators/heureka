// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

type Config struct {
	// KeppelFQDN       string `envconfig:"KEPPEL_FQDN" required:"true" json:"-"`
	OpenstackUsername string `envconfig:"OS_USERNAME" required:"true" json:"-"`
	OpenstackPassword string `envconfig:"OS_PASSWORD" required:"true" json:"-"`
	Domain            string `envconfig:"OS_DOMAIN" required:"true" json:"-"`
	Project           string `envconfig:"OS_PROJECT" required:"true" json:"-"`
	IdentityEndpoint  string `envconfig:"IDENTITY_ENDPOINT" required:"true" json:"-"`
}
