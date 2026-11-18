// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"log"
	"net/http"
)

// runServe starts the canned Keystone (HTTP) and Keppel (HTTPS) listeners.
func runServe(keystoneAddr, keppelAddr string) error {
	keystone := &http.Server{Addr: keystoneAddr, Handler: keystoneHandler()}
	go func() {
		log.Printf("keystone (http) listening on %s", keystoneAddr)
		if err := keystone.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("keystone listener: %v", err)
		}
	}()

	cert, err := selfSignedCert()
	if err != nil {
		return err
	}
	keppel := &http.Server{
		Addr:      keppelAddr,
		Handler:   keppelHandler(),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{*cert}},
	}
	log.Printf("keppel (https) listening on %s", keppelAddr)
	// ListenAndServeTLS with empty cert/key files uses TLSConfig.Certificates.
	if err := keppel.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
