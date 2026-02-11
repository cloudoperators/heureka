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

	"github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Getting access via API", Label("e2e", "TokenAuthorization"), func() {
	var s *server.Server
	var cfg util.Config
	var queryUrl string
	var db *mariadb.SqlDatabase

	BeforeEach(func() {
		db = dbm.NewTestSchemaWithoutMigration()
		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		cfg.AuthTokenSecret = "xxx"
		s = e2e_common.NewRunningServer(cfg)
		queryUrl = fmt.Sprintf("http://localhost:%s/query", cfg.Port)
	})

	AfterEach(func() {
		e2e_common.ServerTeardown(s)
		_ = dbm.TestTearDown(db)
	})

	When("trying to access query resource with valid token", func() {
		It("respond with 200", func() {
			token := test.GenerateJwt(test.TokenStringHandler, cfg.AuthTokenSecret, 1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
	When("trying to access query resource without 'X-Service-Authorization' header", func() {
		It("respond with 401", func() {
			resp := test.SendGetRequest(queryUrl, nil)
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "TokenAuthMethod(No authorization header)")
		})
	})
	When("trying to access query resource with invalid 'X-Service-Authorization' header", func() {
		It("respond with 401", func() {
			resp := test.SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": "invalidHeader"})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "TokenAuthMethod(Invalid authorization header)")
		})
	})
	When("trying to access query resource with expired token", func() {
		It("respond with 401", func() {
			token := test.GenerateJwt(test.TokenStringHandler, cfg.AuthTokenSecret, -1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access query resource with token created using invalid secret", func() {
		It("respond with 401", func() {
			token := test.GenerateJwt(test.TokenStringHandler, "invalidSecret", 1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access query resource with token created using invalid signing method", func() {
		It("respond with 401", func() {
			token := test.GenerateJwtWithRsaSignature(test.TokenStringHandler, test.GenerateRsaPrivateKey(), 1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access query resource with invalid token", func() {
		It("respond with 401", func() {
			token := test.GenerateJwt(test.InvalidTokenStringHandler, cfg.AuthTokenSecret, 1*time.Hour)
			resp := test.SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": test.WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
			test.ExpectErrorMessage(resp, "TokenAuthMethod(Missing ExpiresAt in token claims)")
		})
	})
})
