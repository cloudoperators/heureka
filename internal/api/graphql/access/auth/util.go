// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	Version string `json:"version"`
	jwt.RegisteredClaims
}

type OidcTokenClaims struct {
	Version       string `json:"version"`
	Sub           string `json:"sub"`
	EmailVerified bool   `json:"email_verified"`
	Mail          string `json:"mail"`
	LastName      string `json:"last_name"`
	GivenName     string `json:"access_token"`
	Aud           string `json:"aud"`
	UserUuid      string `json:"user_uuid"`
	FirstName     string `json:"first_name"`
	FamilyName    string `json:"family_name"`
	JTI           string `json:"jti"`
	Email         string `json:"email"`
	jwt.RegisteredClaims
}

func verifyError(methodName string, e error) error {
	return fmt.Errorf("%s(%w)", methodName, e)
}

func getAuthTokenFromHeader(header string, c *gin.Context) (string, error) {
	authHeader := c.GetHeader(header)
	if authHeader == "" {
		return "", fmt.Errorf("No authorization header")
	}
	if len(authHeader) < (len("Bearer ")+1) || !strings.Contains(authHeader, "Bearer ") {
		return "", fmt.Errorf("Invalid authorization header")
	}
	return strings.Split(authHeader, "Bearer ")[1], nil
}

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}
