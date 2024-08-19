// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

type Config struct {
	KeppelBaseUrl    string `envconfig:"KEPPEL_BASE_URL" required:"true" json:"-"`
	KeppelUsername   string `envconfig:"KEPPEL_USERNAME" required:"true" json:"-"`
	KeppelPassword   string `envconfig:"KEPPEL_PASSWORD" required:"true" json:"-"`
	Domain           string `envconfig:"KEPPEL_DOMAIN" required:"true" json:"-"`
	Project          string `envconfig:"KEPPEL_PROJECT" required:"true" json:"-"`
	Region           string `envconfig:"KEPPEL_REGION" required:"true" json:"-"`
	IdentityEndpoint string `envconfig:"IDENTITY_ENDPOINT" required:"true" json:"-"`
}
