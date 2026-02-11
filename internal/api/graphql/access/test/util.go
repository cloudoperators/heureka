// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudoperators/heureka/internal/api/graphql/access/auth"

	"github.com/golang-jwt/jwt/v5"

	// nolint due to importing all functions from gomega package
	//nolint: staticcheck
	. "github.com/onsi/gomega"
)

const (
	testClientName = "testClientName"
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
		Error string `json:"error"`
	}
	unmarshalResponseData(resp, &respData)
	Expect(respData.Error).To(Equal(expectedMsg))
}

func ExpectRegexErrorMessage(resp *http.Response, expectedRegexMsg string, args ...interface{}) {
	var respData struct {
		Error string `json:"error"`
	}
	unmarshalResponseData(resp, &respData)
	Expect(respData.Error).Should(MatchRegexp(expectedRegexMsg, args...))
}

type Jwt struct {
	signingMethod jwt.SigningMethod
	signKey       interface{}
	expiresAt     *jwt.NumericDate
	name          string
}

func NewJwt(secret string) *Jwt {
	return &Jwt{signKey: []byte(secret), signingMethod: jwt.SigningMethodHS256}
}

func NewRsaJwt(privKey *rsa.PrivateKey) *Jwt {
	return &Jwt{signKey: privKey, signingMethod: jwt.SigningMethodRS256}
}

func (j *Jwt) WithName(name string) *Jwt {
	j.name = name
	return j
}

func (j *Jwt) WithExpiresAt(t time.Time) *Jwt {
	j.expiresAt = jwt.NewNumericDate(t)
	return j
}

func GenerateJwtWithRsaSignature(strGen func(*Jwt) string, rsaPrivateKey *rsa.PrivateKey, expiresIn time.Duration) string {
	return strGen(NewRsaJwt(rsaPrivateKey).WithExpiresAt(time.Now().Add(expiresIn)).WithName(testClientName))
}

func GenerateJwt(strGen func(*Jwt) string, jwtSecret string, expiresIn time.Duration) string {
	return strGen(NewJwt(jwtSecret).WithExpiresAt(time.Now().Add(expiresIn)).WithName(testClientName))
}

func GenerateJwtWithName(strGen func(*Jwt) string, jwtSecret string, expiresIn time.Duration, name string) string {
	return strGen(NewJwt(jwtSecret).WithExpiresAt(time.Now().Add(expiresIn)).WithName(name))
}

func GenerateRsaPrivateKey() *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	Expect(err).To(BeNil())
	return privateKey
}

func TokenStringHandler(j *Jwt) string {
	claims := auth.TokenClaims{
		Version: "0.3.1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: j.expiresAt,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "heureka",
			Subject:   j.name,
		},
	}
	token := jwt.NewWithClaims(j.signingMethod, claims)

	tokenString, err := token.SignedString(j.signKey)
	Expect(err).To(BeNil())
	return tokenString
}

func InvalidTokenStringHandler(j *Jwt) string {
	type InvalidTokenClaims struct{ jwt.RegisteredClaims }
	claims := InvalidTokenClaims{}
	token := jwt.NewWithClaims(j.signingMethod, claims)

	tokenString, err := token.SignedString(j.signKey)
	Expect(err).To(BeNil())
	return tokenString
}

func CreateOidcTokenStringHandler(issuer string, clientId string, userName string) func(j *Jwt) string {
	return func(j *Jwt) string {
		claims := auth.OidcTokenClaims{
			Version:       "0.0.1",
			Sub:           userName,
			EmailVerified: false,
			Mail:          "dummy.mail@heureka.com",
			LastName:      "dummyLastName",
			GivenName:     "dummyGivenName",
			Aud:           clientId,
			UserUuid:      "dummyUuid",
			FirstName:     "dummyFirstName",
			FamilyName:    "dummyFamilyName",
			JTI:           "dummyJTI",
			Email:         "dummyMail",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: j.expiresAt,
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    issuer,
				Subject:   j.name,
			},
		}
		token := jwt.NewWithClaims(j.signingMethod, claims)

		tokenString, err := token.SignedString(j.signKey)
		Expect(err).To(BeNil())
		return tokenString
	}
}

func WithBearer(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}
