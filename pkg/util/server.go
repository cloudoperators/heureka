// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

// FirstListenThenServe is a utility function that ensures that first a listener is spin up
// then the http server is setup for serving asynchronously
// this is requried to ensure in tests that the server is spinned up before jumping to tests.
func FirstListenThenServe(srv *http.Server, log *logrus.Logger) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		log.Info("Starting Non Blocking HTTP Server...")
		ln, err := net.Listen("tcp", srv.Addr)
		if err != nil {
			log.WithError(err).Fatalf("Error while start listening...")
		}
		go func() {
			waitGroup.Done()
		}()
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatalf("Error while serving HTTP Server.")
		}
	}()
	waitGroup.Wait()
}
