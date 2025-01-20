// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"fmt"
	"io"
	"os"

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
	})

	AfterEach(func() {
		s.BlockingStop()
		oidcProvider.Stop()
	})
	When("trying to access query resource", func() {
		It("respond with 200", func() {
			jar := CreateCookieJar()
			resp := SendGetRequestWithCookieJar(queryUrl, nil, jar)
			fmt.Println(jar)
			Expect(resp.StatusCode).To(Equal(200))
			defer resp.Body.Close()
			// Read and print the response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading body:", err)
				return
			}
			fmt.Println("Body:", string(body))

			fmt.Println("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
			resp = SendGetRequestWithCookieJar(queryUrl, nil, jar)
			fmt.Println(jar)
			Expect(resp.StatusCode).To(Equal(200))
			fmt.Println("YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY")
			defer resp.Body.Close()
			// Read and print the response body
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading body:", err)
				return
			}
			fmt.Println("Body:", string(body))
			fmt.Println("YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY")

		})
	})
})

/* WHEN Trying to access API for the first time
IT Respond with 200
AND Rsponse contains auth data
AND Auth token is stored in cookie sotrage

WHEN Trying to access API with properly generated Auth token in cookie storage
IT Respond with 200
AND Response contains API data;

WHEN Trying to access API with expired Auth token in cookie storage
IT Respond with ...
AND Response contains error message

WHEN Trying to access API with invalid Auth token in cookie storage
IT Respond with ...
AND Response contains error message

WHEN Trying to access API with invalid clientSecret
IT respond with ...
AND Response contains error message

WHEN Trying to access API with invalid clientId
IT respond with ...
AND Response contains error message

WHEN Trying to access API without Auth token cookie in the storage
IT respond with ...
AND Response contains error message
*/
/*
...
	When("trying to access query resource with valid token", func() {
		It("respond with 200", func() {
			token := GenerateJwt(cfg.AuthTokenSecret, 1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": token})
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
	When("trying to access query resource without 'X-Service-Authorization' header", func() {
		It("respond with 401", func() {
			resp := SendGetRequest(queryUrl, nil)
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(No authorization header)")
		})
	})
	When("trying to access query resource with invalid 'X-Service-Authorization' header", func() {
		It("respond with 401", func() {
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": "invalidHeader"})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access query resource with expired token", func() {
		It("respond with 401", func() {
			token := GenerateJwt(cfg.AuthTokenSecret, -1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access query resource with token created using invalid secret", func() {
		It("respond with 401", func() {
			token := GenerateJwt("invalidSecret", 1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access query resource with token created using invalid signing method", func() {
		It("respond with 401", func() {
			token := GenerateJwtWithInvalidSigningMethod(cfg.AuthTokenSecret, 1*time.Hour)
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(Token parsing error)")
		})
	})
	When("trying to access query resource with invalid token", func() {
		It("respond with 401", func() {
			token := GenerateInvalidJwt(cfg.AuthTokenSecret)
			resp := SendGetRequest(queryUrl, map[string]string{"X-Service-Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))
			ExpectErrorMessage(resp, "TokenAuthMethod(Missing ExpiresAt in token claims)")
		})
	})
*/
