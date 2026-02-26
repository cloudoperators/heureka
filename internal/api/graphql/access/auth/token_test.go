// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/api/graphql/access/context"
	"github.com/cloudoperators/heureka/internal/api/graphql/access/middleware"
	"github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	"github.com/cloudoperators/heureka/internal/util"
)

const (
	testTokenUserName        = "dummyTokenUserName"
	authTokenSecret          = "xxx"
	enableTokenMiddlewareLog = false
	enableTokenServerLog     = false
)

var _ = Describe("Pass token data via context when using token auth middleware", Label("api", "TokenAuthorization"), func() {
	var testServer *test.TestServer

	BeforeEach(func() {
		a := middleware.NewAuth(&util.Config{AuthTokenSecret: authTokenSecret}, enableTokenMiddlewareLog)
		testServer = test.NewTestServer(a, enableTokenServerLog)
		testServer.StartInBackground()
	})

	AfterEach(func() {
		testServer.Stop()
	})

	When("User access api through token auth middleware with valid token", func() {
		BeforeEach(func() {
			token := test.GenerateJwtWithName(test.TokenStringHandler, authTokenSecret, 1*time.Hour, testTokenUserName)
			resp := test.SendGetRequest(testServer.EndpointUrl(), map[string]string{"X-Service-Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(200))
		})
		It("Should be able to access user name from request context", func() {
			name, err := context.UserNameFromContext(testServer.Context())
			Expect(err).To(BeNil())
			Expect(name).To(BeEquivalentTo(testTokenUserName))
		})
	})

	When("User access api through token auth middleware with invalid token", func() {
		BeforeEach(func() {
			token := test.GenerateJwt(test.InvalidTokenStringHandler, authTokenSecret, 1*time.Hour)
			resp := test.SendGetRequest(testServer.EndpointUrl(), map[string]string{"X-Service-Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
		})
		It("Should not store gin context in request context", func() {
			_, err := context.UserNameFromContext(testServer.Context())
			Expect(err).ShouldNot(BeNil())
		})
	})
})
