// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

// mock_oidc_provider/main.go
package main

import (
	"log"
	"os"

	"github.com/cloudoperators/heureka/pkg/oidc"
)

func main() {
	url := os.Getenv("OIDC_PROVIDER_URL")
	if url == "" {
		log.Fatal("Could not start OIDC provider. OIDC_PROVIDER_URL is not set.")
	}
    oidcProvider := oidc.NewProvider(url)
    oidcProvider.StartForeground()
}
