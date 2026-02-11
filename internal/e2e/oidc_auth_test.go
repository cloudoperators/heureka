// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/database/mariadb"
	e2e_common "github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/cloudoperators/heureka/pkg/oidc"

	"github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const enableOidcProviderLog = false

var _ = Describe("Getting access via API", Label("e2e", "OidcAuthorization"), func() {
	var oidcProvider *oidc.Provider
	var s *server.Server
	var cfg util.Config
	var queryUrl string
	var oidcTokenStringHandler func(j *test.Jwt) string
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		db = dbm.NewTestSchemaWithoutMigration()
		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		cfg.AuthOidcClientId = "mock-client-id"
		cfg.AuthOidcUrl = fmt.Sprintf("http://localhost:%s", util2.GetRandomFreePort())
		oidcProvider = oidc.NewProvider(cfg.AuthOidcUrl, enableOidcProviderLog)
		oidcProvider.Start()

		queryUrl = fmt.Sprintf("http://localhost:%s/query", cfg.Port)
		s = e2e_common.NewRunningServer(cfg)

		oidcTokenStringHandler = test.CreateOidcTokenStringHandler(cfg.AuthOidcUrl, cfg.AuthOidcClientId, "dummyUserName")
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		oidcProvider.Stop()
		_ = dbm.TestTearDown(db)
	})

	When("trying to access query resource with valid oidc token", func() {
		It("respond with 200", func() {
			token := test.GenerateJwtWithRsaSignature(oidcTokenStringHandler, oidcProvider.GetRsaPrivateKey(), 1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
	When("trying to access query resource without 'Authorization' header", func() {
		It("respond with 401", func() {
			resp := test.SendGetRequest(queryUrl, nil)
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "OidcAuthMethod(No authorization header)")
		})
	})
	When("trying to access query resource with invalid 'Authorization' header", func() {
		It("respond with 401", func() {
			resp := test.SendGetRequest(queryUrl, map[string]string{"Authorization": "invalidHeader"})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "OidcAuthMethod(Invalid authorization header)")
		})
	})
	When("trying to access query resource with expired oidc token", func() {
		It("respond with 401", func() {
			token := test.GenerateJwtWithRsaSignature(oidcTokenStringHandler, oidcProvider.GetRsaPrivateKey(), -1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectRegexErrorMessage(resp, "OidcAuthMethod\\(oidc: token is expired \\(Token Expiry: .*\\)\\)")
		})
	})
	When("trying to access query resource with oidc token created using invalid Rsa", func() {
		It("respond with 401", func() {
			token := test.GenerateJwtWithRsaSignature(oidcTokenStringHandler, test.GenerateRsaPrivateKey(), 1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "OidcAuthMethod(failed to verify signature: failed to verify id token signature)")
		})
	})
	When("trying to access query resource with oidc token created using invalid signing method", func() {
		It("respond with 401", func() {
			token := test.GenerateJwt(oidcTokenStringHandler, "dummySecret", 1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "OidcAuthMethod(oidc: id token signed with unsupported algorithm, expected [\"RS256\"] got \"HS256\")")
		})
	})

	When("trying to access query resource with invalid token", func() {
		It("respond with 401", func() {
			token := test.GenerateJwtWithRsaSignature(test.InvalidTokenStringHandler, oidcProvider.GetRsaPrivateKey(), 1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectRegexErrorMessage(resp, "OidcAuthMethod\\(oidc: id token issued by a different provider, expected \".*\" got \"\"\\)")
		})
	})

	When("trying to access query resource with invalid audience", func() {
		It("respond with 401", func() {
			invalidAudienceOidcTokenStringHandler := test.CreateOidcTokenStringHandler(cfg.AuthOidcUrl, "invalidAudience", "dummyUserName")
			token := test.GenerateJwtWithRsaSignature(invalidAudienceOidcTokenStringHandler, oidcProvider.GetRsaPrivateKey(), 1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "OidcAuthMethod(oidc: expected audience \"mock-client-id\" got [\"invalidAudience\"])")
		})
	})
})
