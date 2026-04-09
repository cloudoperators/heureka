// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import "fmt"

type Config struct {
	KeppelFQDN       string  `envconfig:"KEPPEL_FQDN"        required:"true" json:"-"`
	KeppelUsername   string  `envconfig:"KEPPEL_USERNAME"    required:"true" json:"-"`
	KeppelUserDomain string  `envconfig:"KEPPEL_USER_DOMAIN" required:"true" json:"-"`
	KeppelPassword   string  `envconfig:"KEPPEL_PASSWORD"    required:"true" json:"-"`
	Domain           string  `envconfig:"KEPPEL_DOMAIN"      required:"true" json:"-"`
	Project          string  `envconfig:"KEPPEL_PROJECT"     required:"true" json:"-"`
	IdentityEndpoint string  `envconfig:"IDENTITY_ENDPOINT"  required:"true" json:"-"`
	KeppelRateLimit  float64 `envconfig:"KEPPEL_RATE_LIMIT"                  json:"-" default:"1.0"`
	KeppelRateBurst  int     `envconfig:"KEPPEL_RATE_BURST"                  json:"-" default:"10"`
	TrivyRateLimit   float64 `envconfig:"TRIVY_RATE_LIMIT"                   json:"-" default:"0.0833"` // 5 requests per minute
	TrivyRateBurst   int     `envconfig:"TRIVY_RATE_BURST"                   json:"-" default:"1"`
}

func (c *Config) KeppelBaseUrl() string {
	return fmt.Sprintf("https://%s", c.KeppelFQDN)
}
