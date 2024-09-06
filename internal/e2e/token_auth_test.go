// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/access"
	"github.wdf.sap.corp/cc/heureka/internal/util"
	util2 "github.wdf.sap.corp/cc/heureka/pkg/util"

	"github.wdf.sap.corp/cc/heureka/internal/server"

	"github.com/golang-jwt/jwt/v5"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func sendGetRequest(url string, headers map[string]string) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	Expect(err).To(BeNil())
	return resp
}

func unmarshalResponseData(resp *http.Response, respData interface{}) {
	body, err := io.ReadAll(resp.Body)
	Expect(err).To(BeNil())
	err = json.Unmarshal(body, respData)
	Expect(err).To(BeNil())
}

func generateJwtSpec(signingMethod jwt.SigningMethod, signKey interface{}, expiresAt *jwt.NumericDate) string {
	claims := access.TokenClaims{
		Version: "0.3.1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expiresAt,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "heureka",
			Subject:   "testUser",
		},
	}
	token := jwt.NewWithClaims(signingMethod, claims)

	tokenString, err := token.SignedString(signKey)
	Expect(err).To(BeNil())
	return tokenString
}

func generateJwt(jwtSecret string, expiresIn time.Duration) string {
	expiresAt := jwt.NewNumericDate(time.Now().Add(expiresIn))
	return generateJwtSpec(jwt.SigningMethodHS256, []byte(jwtSecret), expiresAt)
}

func generateInvalidJwt(jwtSecret string) string {
	return generateJwtSpec(jwt.SigningMethodHS256, []byte(jwtSecret), nil)
}

func generateRsaPrivateKey() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	Expect(err).To(BeNil())
	return privateKey
}

func generateJwtWithInvalidSigningMethod(jwtSecret string, expiresIn time.Duration) string {
	privateKey := generateRsaPrivateKey()
	expiresAt := jwt.NewNumericDate(time.Now().Add(expiresIn))
	return generateJwtSpec(jwt.SigningMethodRS256, privateKey, expiresAt)
}

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
		cfg.AuthType = "token"
		cfg.AuthTokenSecret = "xxx"
		s = server.NewServer(cfg)

		queryUrl = fmt.Sprintf("http://localhost:%s/query", cfg.Port)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("trying to access query resource with valid token", func() {
		It("respond with 200", func() {
			token := generateJwt(cfg.AuthTokenSecret, 1*time.Hour)
			resp := sendGetRequest(queryUrl, map[string]string{"Authorization": token})
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
	When("trying to access query resource without 'Authorization' header", func() {
		It("respond with 401", func() {
			resp := sendGetRequest(queryUrl, nil)
			Expect(resp.StatusCode).To(Equal(401))

			var respData struct {
				Error string `json:"error" required:true`
			}
			unmarshalResponseData(resp, &respData)
			Expect(respData.Error).To(Equal("Authorization header is required"))
		})
	})
	When("trying to access query resource with invalid 'Authorization' header", func() {
		It("respond with 401", func() {
			resp := sendGetRequest(queryUrl, map[string]string{"Authorization": "invalidHeader"})
			Expect(resp.StatusCode).To(Equal(401))

			var respData struct {
				Error string `json:"error" required:true`
			}
			unmarshalResponseData(resp, &respData)
			Expect(respData.Error).To(Equal("Token parsing error"))
		})
	})
	When("trying to access query resource with expired token", func() {
		It("respond with 401", func() {
			token := generateJwt(cfg.AuthTokenSecret, -1*time.Hour)
			resp := sendGetRequest(queryUrl, map[string]string{"Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))

			var respData struct {
				Error string `json:"error" required:true`
			}
			unmarshalResponseData(resp, &respData)
			Expect(respData.Error).To(Equal("Token parsing error"))
		})
	})
	When("trying to access query resource with token created using invalid secret", func() {
		It("respond with 401", func() {
			token := generateJwt("invalidSecret", 1*time.Hour)
			resp := sendGetRequest(queryUrl, map[string]string{"Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))

			var respData struct {
				Error string `json:"error" required:true`
			}
			unmarshalResponseData(resp, &respData)
			Expect(respData.Error).To(Equal("Token parsing error"))
		})
	})
	When("trying to access query resource with token created using invalid signing method", func() {
		It("respond with 401", func() {
			token := generateJwtWithInvalidSigningMethod(cfg.AuthTokenSecret, 1*time.Hour)
			resp := sendGetRequest(queryUrl, map[string]string{"Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))

			var respData struct {
				Error string `json:"error" required:true`
			}
			unmarshalResponseData(resp, &respData)
			Expect(respData.Error).To(Equal("Token parsing error"))
		})
	})
	When("trying to access query resource with invalid token", func() {
		It("respond with 401", func() {
			token := generateInvalidJwt(cfg.AuthTokenSecret)
			resp := sendGetRequest(queryUrl, map[string]string{"Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))

			var respData struct {
				Error string `json:"error" required:true`
			}
			unmarshalResponseData(resp, &respData)
			Expect(respData.Error).To(Equal("Invalid token"))
		})
	})
})
