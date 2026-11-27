// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultExpiresInHours = 30 * 24
	jwtIssuer             = "heureka"
)

type tokenClaims struct {
	Version string `json:"version"`
	jwt.RegisteredClaims
}

func generateJWT(secret []byte, expireIn time.Duration, subject string) (string, error) {
	claims := tokenClaims{
		Version: "0.3.1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuer,
			Subject:   subject,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <subject>\n", os.Args[0])
		os.Exit(2)
	}
	subject := os.Args[1]

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable not set")
	}

	expiration := time.Hour * defaultExpiresInHours
	if hours, err := strconv.Atoi(os.Getenv("EXPIRES_IN_HOURS")); err == nil {
		expiration = time.Hour * time.Duration(hours)
	}

	token, err := generateJWT([]byte(secret), expiration, subject)
	if err != nil {
		panic(err)
	}
	fmt.Println(token)
}
