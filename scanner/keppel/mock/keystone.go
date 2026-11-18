// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// keystoneHandler emulates the small slice of the OpenStack Keystone v3 API
// that gophercloud's AuthenticateV3 exercises.
//
// gophercloud (v1.14.1) POSTs to {IDENTITY_ENDPOINT}/auth/tokens and then:
//   - reads the issued token id from the X-Subject-Token response header
//     (this is what scanner.CreateAuthToken returns via iClient.Token()), and
//   - unmarshals the body, requiring a well-formed token.expires_at and a
//     token.catalog (which may be empty — the scanner never uses it).
func keystoneHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v3/auth/tokens", handleToken)
	// Version discovery is defensive: v1.14.1 posts directly, but some code
	// paths probe the unversioned/base endpoint first.
	mux.HandleFunc("GET /v3", handleVersion)
	mux.HandleFunc("GET /", handleVersion)
	return mux
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	body := map[string]any{
		"token": map[string]any{
			"issued_at":  now.Format(time.RFC3339),
			"expires_at": now.Add(24 * time.Hour).Format(time.RFC3339),
			"methods":    []string{"password"},
			"user": map[string]any{
				"id":   "mock-user",
				"name": "mock",
			},
			// Empty catalog is accepted: the scanner builds the Keppel URL from
			// KEPPEL_FQDN and never consults the service catalog.
			"catalog": []any{},
		},
	}
	w.Header().Set("X-Subject-Token", "mock-subject-token")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("keystone: encode token response: %v", err)
	}
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	body := map[string]any{
		"version": map[string]any{
			"id":     "v3.14",
			"status": "stable",
			"links": []map[string]string{
				{"rel": "self", "href": "http://" + r.Host + "/v3"},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("keystone: encode version response: %v", err)
	}
}
