// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/api/graphql/access/auth"
	authcontext "github.com/cloudoperators/heureka/internal/api/graphql/access/context"
	"github.com/cloudoperators/heureka/internal/api/graphql/access/middleware"
	"github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
)

const (
	testTokenUserName        = "dummyTokenUserName"
	testTokenUserId          = int64(42)
	authTokenSecret          = "xxx"
	enableTokenMiddlewareLog = false
	enableTokenServerLog     = false
)

func technicalUserResolver(expectedSubject string) auth.UserResolver {
	return func(_ context.Context, uniqueUserId string) (int64, entity.UserType, error) {
		if uniqueUserId != expectedSubject {
			return 0, entity.InvalidUserType, fmt.Errorf("user not found")
		}
		return testTokenUserId, entity.TechnicalUserType, nil
	}
}

var _ = Describe(
	"Pass token data via context when using token auth middleware",
	Label("api", "TokenAuthorization"),
	func() {
		var testServer *test.TestServer

		BeforeEach(func() {
			a := middleware.NewAuth(
				&util.Config{AuthTokenSecret: authTokenSecret},
				technicalUserResolver(testTokenUserName),
				enableTokenMiddlewareLog,
			)
			testServer = test.NewTestServer(a, enableTokenServerLog)
			testServer.StartInBackground()
		})

		AfterEach(func() {
			testServer.Stop()
		})

		When("User access api through token auth middleware with valid token", func() {
			BeforeEach(func() {
				token := test.GenerateJwtWithName(
					test.TokenStringHandler,
					authTokenSecret,
					1*time.Hour,
					testTokenUserName,
				)
				resp := test.SendGetRequest(
					testServer.EndpointUrl(),
					map[string]string{"X-Service-Authorization": test.WithBearer(token)},
				)
				Expect(resp.StatusCode).To(Equal(200))
			})
			It("Should be able to access user name from request context", func() {
				name, err := authcontext.UserNameFromContext(testServer.Context())
				Expect(err).To(BeNil())
				Expect(name).To(BeEquivalentTo(testTokenUserName))
			})
			It("Should be able to access user id from request context", func() {
				id, err := authcontext.UserIdFromContext(testServer.Context())
				Expect(err).To(BeNil())
				Expect(id).To(Equal(testTokenUserId))
			})
		})

		When("User access api through token auth middleware with invalid token", func() {
			BeforeEach(func() {
				token := test.GenerateJwt(
					test.InvalidTokenStringHandler,
					authTokenSecret,
					1*time.Hour,
				)
				resp := test.SendGetRequest(
					testServer.EndpointUrl(),
					map[string]string{"X-Service-Authorization": test.WithBearer(token)},
				)
				Expect(resp.StatusCode).To(Equal(401))
			})
			It("Should not store gin context in request context", func() {
				_, err := authcontext.UserNameFromContext(testServer.Context())
				Expect(err).ShouldNot(BeNil())
			})
		})

		When("Subject is not registered as a user", func() {
			BeforeEach(func() {
				token := test.GenerateJwtWithName(
					test.TokenStringHandler,
					authTokenSecret,
					1*time.Hour,
					"unknown-subject",
				)
				resp := test.SendGetRequest(
					testServer.EndpointUrl(),
					map[string]string{"X-Service-Authorization": test.WithBearer(token)},
				)
				Expect(resp.StatusCode).To(Equal(401))
			})
			It("Should reject with 401", func() {})
		})

		When("Subject resolves to a non-technical user", func() {
			var nonTechServer *test.TestServer

			BeforeEach(func() {
				nonTechResolver := func(_ context.Context, uniqueUserId string) (int64, entity.UserType, error) {
					return testTokenUserId, entity.HumanUserType, nil
				}
				a := middleware.NewAuth(
					&util.Config{AuthTokenSecret: authTokenSecret},
					nonTechResolver,
					enableTokenMiddlewareLog,
				)
				nonTechServer = test.NewTestServer(a, enableTokenServerLog)
				nonTechServer.StartInBackground()

				token := test.GenerateJwtWithName(
					test.TokenStringHandler,
					authTokenSecret,
					1*time.Hour,
					testTokenUserName,
				)
				resp := test.SendGetRequest(
					nonTechServer.EndpointUrl(),
					map[string]string{"X-Service-Authorization": test.WithBearer(token)},
				)
				Expect(resp.StatusCode).To(Equal(401))
			})
			AfterEach(func() {
				nonTechServer.Stop()
			})
			It("Should reject with 401", func() {})
		})
	},
)
