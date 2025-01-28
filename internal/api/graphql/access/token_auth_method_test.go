// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package access_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudoperators/heureka/internal/api/graphql/access/test"

	"github.com/cloudoperators/heureka/internal/api/graphql/access"
	"github.com/cloudoperators/heureka/internal/util"
)

const (
	testScannerName = "testAccessScanner"
	authTokenSecret = "xxx"
)

var _ = Describe("Pass token data via context when using token auth middleware", Label("api", "TokenAuthorization"), func() {
	var testServer *TestServer

	BeforeEach(func() {
		auth := access.NewAuth(&util.Config{AuthTokenSecret: authTokenSecret})
		testServer = NewTestServer(auth)
		testServer.StartInBackground()
	})

	AfterEach(func() {
		testServer.Stop()
	})

	When("Scanner access api through token auth middleware with valid token", func() {
		BeforeEach(func() {
			token := GenerateJwtWithName(TokenStringHandler, authTokenSecret, 1*time.Hour, testScannerName)
			resp := SendGetRequest(testServer.EndpointUrl(), map[string]string{"X-Service-Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(200))
		})
		It("Should be able to access scanner name from request context", func() {
			name, err := access.ScannerNameFromContext(testServer.Context())
			Expect(err).To(BeNil())
			Expect(name).To(BeEquivalentTo(testScannerName))
		})
	})

	When("Scanner access api through token auth middleware with invalid token", func() {
		BeforeEach(func() {
			token := GenerateJwt(InvalidTokenStringHandler, authTokenSecret, 1*time.Hour)
			resp := SendGetRequest(testServer.EndpointUrl(), map[string]string{"X-Service-Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
		})
		It("Should not store gin context in request context", func() {
			_, err := access.ScannerNameFromContext(testServer.Context())
			Expect(err).ShouldNot(BeNil())
		})
	})
})
