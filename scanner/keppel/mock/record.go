// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// runRecord runs the mock as a transparent recording proxy in front of a real
// Keppel + Keystone. Point the scanner at the mock as usual (KEPPEL_FQDN and
// IDENTITY_ENDPOINT), set the real upstreams via RECORD_KEPPEL_UPSTREAM and
// RECORD_IDENTITY_UPSTREAM, provide real KEPPEL_* credentials, and run the
// scanner once. Every Keppel response the scanner fetches is written to
// testdata/fixtures/<name>.json, so the committed fixtures can be refreshed
// from reality when the real API changes.
func runRecord(keystoneAddr, keppelAddr string) error {
	idUp := mustURL("RECORD_IDENTITY_UPSTREAM")
	kpUp := mustURL("RECORD_KEPPEL_UPSTREAM")

	// Keystone is a stable protocol; proxy it (so gophercloud can authenticate)
	// but don't record it.
	keystone := &http.Server{Addr: keystoneAddr, Handler: recordingProxy(idUp, false)}
	go func() {
		log.Printf("[record] keystone proxy %s -> %s", keystoneAddr, idUp)
		if err := keystone.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("keystone proxy: %v", err)
		}
	}()

	cert, err := selfSignedCert()
	if err != nil {
		return err
	}
	keppel := &http.Server{
		Addr:      keppelAddr,
		Handler:   recordingProxy(kpUp, true),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{*cert}},
	}
	log.Printf("[record] keppel proxy %s -> %s (writing fixtures to %s)", keppelAddr, kpUp, fixturesDir)
	if err := keppel.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func mustURL(key string) *url.URL {
	raw := envOr(key, "")
	if raw == "" {
		log.Fatalf("record mode requires %s", key)
	}
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("invalid %s=%q: %v", key, raw, err)
	}
	return u
}

// recordingProxy forwards requests to upstream (preserving the incoming path),
// skipping TLS verification. When record is true, each JSON response body is
// classified and written to a fixture file.
func recordingProxy(upstream *url.URL, record bool) http.Handler {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = upstream.Scheme
			req.URL.Host = upstream.Host
			req.Host = upstream.Host
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // recording against staging only
		},
	}
	if record {
		proxy.ModifyResponse = func(resp *http.Response) error {
			name := fixtureName(resp.Request.URL.Path)
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return err
			}
			resp.Body = io.NopCloser(bytes.NewReader(body))
			if name != "" && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				var pretty bytes.Buffer
				if json.Indent(&pretty, body, "", "  ") == nil {
					if err := writeFixture(name, pretty.Bytes()); err != nil {
						log.Printf("[record] write %s: %v", name, err)
					} else {
						log.Printf("[record] saved %s (%d bytes) from %s", name, pretty.Len(), resp.Request.URL.Path)
						if name == "manifest" {
							log.Printf("[record] manifest headers: X-Keppel-Vulnerability-Status=%q Docker-Content-Digest=%q",
								resp.Header.Get("X-Keppel-Vulnerability-Status"), resp.Header.Get("Docker-Content-Digest"))
						}
					}
				}
			}
			return nil
		}
	}
	return proxy
}

// fixtureName maps a Keppel request path to the fixture it should update, or ""
// to skip recording.
func fixtureName(path string) string {
	switch {
	case path == "/keppel/v1/accounts":
		return "accounts"
	case strings.HasSuffix(path, "/repositories"):
		return "repositories"
	case strings.HasSuffix(path, "/trivy_report"):
		return "trivy_report"
	case strings.HasPrefix(path, "/v2/") && strings.Contains(path, "/manifests/"):
		return "manifest"
	default:
		return ""
	}
}
