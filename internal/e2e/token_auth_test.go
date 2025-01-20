// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"time"

	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"

	. "github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Getting access via API", Label("e2e", "TokenAuthorization"), func() {
	var s *server.Server
	var cfg util.Config
	var queryUrl string

	BeforeEach(func() {
		var err error
		_ = dbm.NewTestSchema()
		Expect(err).To(BeNil(), "Database Seeder Setup should work")

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		cfg.AuthTokenSecret = "xxx"
		s = server.NewServer(cfg)

		queryUrl = fmt.Sprintf("http://localhost:%s/query", cfg.Port)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("trying to access API with valid token", func() {
		It("respond with 200", func() {
			token := GenerateJwt(cfg.AuthTokenSecret, 1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": token})
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
	When("trying to access API without 'X-Service-Authorization' header", func() {
		It("respond with 401", func() {
			resp := SendGetRequest(queryUrl, nil)
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(No authorization header)")
		})
	})
	When("trying to access API with invalid 'X-Service-Authorization' header", func() {
		It("respond with 401", func() {
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": "invalidHeader"})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access API with expired token", func() {
		It("respond with 401", func() {
			token := GenerateJwt(cfg.AuthTokenSecret, -1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access API with token created using invalid secret", func() {
		It("respond with 401", func() {
			token := GenerateJwt("invalidSecret", 1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access API with token created using invalid signing method", func() {
		It("respond with 401", func() {
			token := GenerateJwtWithInvalidSigningMethod(cfg.AuthTokenSecret, 1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access API with invalid token", func() {
		It("respond with 401", func() {
			token := GenerateInvalidJwt(cfg.AuthTokenSecret)
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(Missing ExpiresAt in token claims)")
		})
	})
})
