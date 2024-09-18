// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/cloudoperators/heureka/internal/api/graphql/access"

	"github.com/golang-jwt/jwt/v5"

	. "github.com/onsi/gomega"
)

const (
	testUsername = "testUser"
)

func SendGetRequest(url string, headers map[string]string) *http.Response {
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

func ExpectErrorMessage(resp *http.Response, expectedMsg string) {
	var respData struct {
		Error string `json:"error" required:true`
	}
	unmarshalResponseData(resp, &respData)
	Expect(respData.Error).To(Equal(expectedMsg))
}

type Jwt struct {
	signingMethod jwt.SigningMethod
	signKey       interface{}
	expiresAt     *jwt.NumericDate
	username      string
}

func NewJwt(secret string) *Jwt {
	return &Jwt{signKey: []byte(secret), signingMethod: jwt.SigningMethodHS256}
}

func NewRsaJwt(privKey *rsa.PrivateKey) *Jwt {
	return &Jwt{signKey: privKey, signingMethod: jwt.SigningMethodRS256}
}

func (j *Jwt) WithUsername(username string) *Jwt {
	j.username = username
	return j
}

func (j *Jwt) WithExpiresAt(t time.Time) *Jwt {
	j.expiresAt = jwt.NewNumericDate(t)
	return j
}

func (j *Jwt) String() string {
	claims := access.TokenClaims{
		Version: "0.3.1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: j.expiresAt,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "heureka",
			Subject:   j.username,
		},
	}
	token := jwt.NewWithClaims(j.signingMethod, claims)

	tokenString, err := token.SignedString(j.signKey)
	Expect(err).To(BeNil())
	return tokenString
}

func GenerateJwt(jwtSecret string, expiresIn time.Duration) string {
	return NewJwt(jwtSecret).WithExpiresAt(time.Now().Add(expiresIn)).WithUsername(testUsername).String()
}

func GenerateJwtWithUsername(jwtSecret string, expiresIn time.Duration, username string) string {
	return NewJwt(jwtSecret).WithExpiresAt(time.Now().Add(expiresIn)).WithUsername(username).String()
}

func GenerateInvalidJwt(jwtSecret string) string {
	return NewJwt(jwtSecret).WithUsername(testUsername).String()
}

func GenerateRsaPrivateKey() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	Expect(err).To(BeNil())
	return privateKey
}

func GenerateJwtWithInvalidSigningMethod(jwtSecret string, expiresIn time.Duration) string {
	return NewRsaJwt(GenerateRsaPrivateKey()).WithExpiresAt(time.Now().Add(expiresIn)).WithUsername(testUsername).String()
}
