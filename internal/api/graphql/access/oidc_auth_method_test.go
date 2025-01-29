// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package access_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/access"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/cloudoperators/heureka/pkg/oidc"
)

const (
	clientId     = "mock-client-id"
	testUserName = "dummyUserName"
)

var _ = Describe("Pass OIDC token data via context when using OIDC auth middleware", Label("api", "OidcAuthorization"), func() {
	var oidcProvider *oidc.Provider
	var testServer *TestServer
	var oidcTokenStringHandler func(j *Jwt) string

	BeforeEach(func() {
		oidcProviderUrl := fmt.Sprintf("http://localhost:%s", util2.GetRandomFreePort())
		oidcProvider = oidc.NewProvider(oidcProviderUrl)
		oidcProvider.Start()

		auth := access.NewAuth(&util.Config{AuthOidcUrl: oidcProviderUrl, AuthOidcClientId: clientId})
		testServer = NewTestServer(auth)
		testServer.StartInBackground()

		oidcTokenStringHandler = CreateOidcTokenStringHandler(oidcProviderUrl, clientId, testUserName)
	})

	AfterEach(func() {
		testServer.Stop()
		oidcProvider.Stop()
	})

	When("User access api through OIDC token auth middleware with valid token", func() {
		BeforeEach(func() {
			token := GenerateJwtWithRsaSignature(oidcTokenStringHandler, oidcProvider.GetRsaPrivateKey(), 1*time.Hour)
			resp := SendGetRequest(testServer.EndpointUrl(), map[string]string{"Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(200))
		})
		It("Should be able to access user name from request context", func() {
			name, err := access.UserNameFromContext(testServer.Context())
			Expect(err).To(BeNil())
			Expect(name).To(BeEquivalentTo(testUserName))
		})
	})

	When("User access api through OIDC token auth middleware with invalid token", func() {
		BeforeEach(func() {
			token := GenerateJwtWithRsaSignature(InvalidTokenStringHandler, oidcProvider.GetRsaPrivateKey(), 1*time.Hour)
			resp := SendGetRequest(testServer.EndpointUrl(), map[string]string{"Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
		})
		It("Should not store gin context in request context", func() {
			_, err := access.UserNameFromContext(testServer.Context())
			Expect(err).ShouldNot(BeNil())
		})
	})
})
