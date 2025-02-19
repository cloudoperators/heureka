// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/access/context"
	"github.com/cloudoperators/heureka/internal/api/graphql/access/middleware"
	"github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/cloudoperators/heureka/pkg/oidc"
)

const (
	clientId     = "mock-client-id"
	testUserName = "dummyUserName"
)

var _ = Describe("Pass OIDC token data via context when using OIDC auth middleware", Label("api", "OidcAuthorization"), func() {
	var oidcProvider *oidc.Provider
	var testServer *test.TestServer
	var oidcTokenStringHandler func(j *test.Jwt) string

	BeforeEach(func() {
		oidcProviderUrl := fmt.Sprintf("http://localhost:%s", util2.GetRandomFreePort())
		oidcProvider = oidc.NewProvider(oidcProviderUrl)
		oidcProvider.Start()

		a := middleware.NewAuth(&util.Config{AuthOidcUrl: oidcProviderUrl, AuthOidcClientId: clientId})
		testServer = test.NewTestServer(a)
		testServer.StartInBackground()

		oidcTokenStringHandler = test.CreateOidcTokenStringHandler(oidcProviderUrl, clientId, testUserName)
	})

	AfterEach(func() {
		testServer.Stop()
		oidcProvider.Stop()
	})

	When("User access api through OIDC token auth middleware with valid token", func() {
		BeforeEach(func() {
			token := test.GenerateJwtWithRsaSignature(oidcTokenStringHandler, oidcProvider.GetRsaPrivateKey(), 1*time.Hour)
			resp := test.SendGetRequest(testServer.EndpointUrl(), map[string]string{"Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(200))
		})
		It("Should be able to access user name from request context", func() {
			name, err := context.UserNameFromContext(testServer.Context())
			Expect(err).To(BeNil())
			Expect(name).To(BeEquivalentTo(testUserName))
		})
	})

	When("User access api through OIDC token auth middleware with invalid token", func() {
		BeforeEach(func() {
			token := test.GenerateJwtWithRsaSignature(test.InvalidTokenStringHandler, oidcProvider.GetRsaPrivateKey(), 1*time.Hour)
			resp := test.SendGetRequest(testServer.EndpointUrl(), map[string]string{"Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
		})
		It("Should not store gin context in request context", func() {
			_, err := context.UserNameFromContext(testServer.Context())
			Expect(err).ShouldNot(BeNil())
		})
	})
})
