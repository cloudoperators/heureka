// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"os"
	"time"

	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"

	. "github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}

var _ = Describe("Getting access via API", Label("e2e", "OidcAuthorization"), func() {
	var oidcProvider *util2.OidcProvider
	var s *server.Server
	var cfg util.Config
	var queryUrl string
	var oidcTokenStringHandler func(j *Jwt) string

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		cfg.AuthOidcClientId = "mock-client-id"
		cfg.AuthOidcUrl = fmt.Sprintf("http://localhost:%s", util2.GetRandomFreePort())
		oidcProvider = util2.NewOidcProvider(cfg.AuthOidcUrl)
		oidcProvider.Start()

		queryUrl = fmt.Sprintf("http://%s:%s/query", getHostname(), cfg.Port)
		s = server.NewServer(cfg)
		s.NonBlockingStart()

		oidcTokenStringHandler = CreateOidcTokenStringHandler(cfg.AuthOidcUrl, cfg.AuthOidcClientId)
	})

	AfterEach(func() {
		s.BlockingStop()
		oidcProvider.Stop()
	})

	When("trying to access query resource with valid oidc token", func() {
		It("respond with 200", func() {
			token := GenerateJwtWithRsaSignature(oidcTokenStringHandler, oidcProvider.GetRsaPrivateKey(), 1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
	When("trying to access query resource without 'Authorization' header", func() {
		It("respond with 401", func() {
			resp := SendGetRequest(queryUrl, nil)
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "OidcAuthMethod(No authorization header)")
		})
	})
	When("trying to access query resource with invalid 'Authorization' header", func() {
		It("respond with 401", func() {
			resp := SendGetRequest(queryUrl, map[string]string{"Authorization": "invalidHeader"})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "OidcAuthMethod(Invalid authorization header)")
		})
	})
	When("trying to access query resource with expired oidc token", func() {
		It("respond with 401", func() {
			token := GenerateJwtWithRsaSignature(oidcTokenStringHandler, oidcProvider.GetRsaPrivateKey(), -1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectRegexErrorMessage(resp, "OidcAuthMethod\\(oidc: token is expired \\(Token Expiry: .*\\)\\)")
		})
	})
	When("trying to access query resource with oidc token created using invalid Rsa", func() {
		It("respond with 401", func() {
			token := GenerateJwtWithRsaSignature(oidcTokenStringHandler, GenerateRsaPrivateKey(), 1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "OidcAuthMethod(failed to verify signature: failed to verify id token signature)")
		})
	})
	When("trying to access query resource with oidc token created using invalid signing method", func() {
		It("respond with 401", func() {
			token := GenerateJwt(oidcTokenStringHandler, "dummySecret", 1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "OidcAuthMethod(oidc: id token signed with unsupported algorithm, expected [\"RS256\"] got \"HS256\")")
		})
	})

	When("trying to access query resource with invalid token", func() {
		It("respond with 401", func() {
			token := GenerateJwtWithRsaSignature(InvalidTokenStringHandler, oidcProvider.GetRsaPrivateKey(), 1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectRegexErrorMessage(resp, "OidcAuthMethod\\(oidc: id token issued by a different provider, expected \".*\" got \"\"\\)")
		})
	})
})
