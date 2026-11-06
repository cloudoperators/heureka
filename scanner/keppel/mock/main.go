// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

// Command mock is a local stand-in for a Keppel image registry plus its
// OpenStack Keystone identity endpoint, so the keppel scanner can be run
// end-to-end on a developer laptop without access to a real registry.
//
// It starts two listeners:
//
//   - Keystone on plain HTTP (default :8080). gophercloud verifies TLS on the
//     identity endpoint and the scanner's InsecureSkipVerify does NOT apply to
//     it, so this one must be HTTP.
//   - Keppel on HTTPS with a self-signed cert (default :8443). The scanner's
//     Keppel base URL is hardcoded to https:// and it skips TLS verification,
//     so a self-signed cert is fine.
//
// Run it (from this directory):
//
//	go run .
//
// then point the scanner at it:
//
//	KEPPEL_FQDN=localhost:8443
//	IDENTITY_ENDPOINT=http://localhost:8080/v3
//
// See ../../../docs/local-environment.md for the full workflow.
package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	keystoneAddr := flag.String("keystone-addr", envOr("MOCK_KEYSTONE_ADDR", ":8080"), "Keystone (HTTP) listen address")
	keppelAddr := flag.String("keppel-addr", envOr("MOCK_KEPPEL_ADDR", ":8443"), "Keppel (HTTPS) listen address")
	record := flag.Bool("record", os.Getenv("KEPPEL_RECORD") == "1", "record mode: proxy to the real upstreams (RECORD_IDENTITY_UPSTREAM/RECORD_KEPPEL_UPSTREAM) and write fixtures")
	flag.Parse()

	if *record {
		if err := runRecord(*keystoneAddr, *keppelAddr); err != nil {
			log.Fatalf("record mode failed: %v", err)
		}
		return
	}
	if err := runServe(*keystoneAddr, *keppelAddr); err != nil {
		log.Fatalf("serve mode failed: %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
