// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/access/auth"
	authcontext "github.com/cloudoperators/heureka/internal/api/graphql/access/context"
	"github.com/cloudoperators/heureka/internal/api/graphql/access/middleware"
	"github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/cloudoperators/heureka/pkg/oidc"
)

const (
	clientId                = "mock-client-id"
	testOidcUserName        = "dummyOidcUserName"
	testOidcUserId          = int64(7)
	enableOidcProviderLog   = false
	enableOidcMiddlewareLog = false
	enableOidcServerLog     = false
)

func humanUserResolver(expectedSubject string) auth.UserResolver {
	return func(_ context.Context, uniqueUserId string) (int64, entity.UserType, error) {
		if uniqueUserId != expectedSubject {
			return 0, entity.InvalidUserType, fmt.Errorf("user not found")
		}
		return testOidcUserId, entity.HumanUserType, nil
	}
}

var _ = Describe(
	"Pass OIDC token data via context when using OIDC auth middleware",
	Label("api", "OidcAuthorization"),
	func() {
		var (
			oidcProvider           *oidc.Provider
			testServer             *test.TestServer
			oidcTokenStringHandler func(j *test.Jwt) string
		)

		BeforeEach(func() {
			oidcProviderUrl := fmt.Sprintf("http://localhost:%s", util2.GetRandomFreePort())
			oidcProvider = oidc.NewProvider(oidcProviderUrl, enableOidcProviderLog)
			oidcProvider.Start()

			a := middleware.NewAuth(
				&util.Config{AuthOidcUrl: oidcProviderUrl, AuthOidcClientId: clientId},
				humanUserResolver(testOidcUserName),
				enableOidcMiddlewareLog,
			)
			testServer = test.NewTestServer(a, enableOidcServerLog)
			testServer.StartInBackground()

			oidcTokenStringHandler = test.CreateOidcTokenStringHandler(
				oidcProviderUrl,
				clientId,
				testOidcUserName,
			)
		})

		AfterEach(func() {
			testServer.Stop()
			oidcProvider.Stop()
		})

		When("User access api through OIDC token auth middleware with valid token", func() {
			BeforeEach(func() {
				token := test.GenerateJwtWithRsaSignature(
					oidcTokenStringHandler,
					oidcProvider.GetRsaPrivateKey(),
					1*time.Hour,
				)
				resp := test.SendGetRequest(
					testServer.EndpointUrl(),
					map[string]string{"Authorization": test.WithBearer(token)},
				)
				Expect(resp.StatusCode).To(Equal(200))
			})
			It("Should be able to access user name from request context", func() {
				name, err := authcontext.UserNameFromContext(testServer.Context())
				Expect(err).To(BeNil())
				Expect(name).To(BeEquivalentTo(testOidcUserName))
			})
			It("Should be able to access user id from request context", func() {
				id, err := authcontext.UserIdFromContext(testServer.Context())
				Expect(err).To(BeNil())
				Expect(id).To(Equal(testOidcUserId))
			})
		})

		When("User access api through OIDC token auth middleware with invalid token", func() {
			BeforeEach(func() {
				token := test.GenerateJwtWithRsaSignature(
					test.InvalidTokenStringHandler,
					oidcProvider.GetRsaPrivateKey(),
					1*time.Hour,
				)
				resp := test.SendGetRequest(
					testServer.EndpointUrl(),
					map[string]string{"Authorization": test.WithBearer(token)},
				)
				Expect(resp.StatusCode).To(Equal(401))
			})
			It("Should not store gin context in request context", func() {
				_, err := authcontext.UserNameFromContext(testServer.Context())
				Expect(err).ShouldNot(BeNil())
			})
		})
	},
)
