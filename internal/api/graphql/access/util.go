// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package access

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ginContextKeyType string

const (
	ginContextKey  ginContextKeyType = "GinContextKey"
	scannerNameKey string            = "scannername"
	userNameKey    string            = "username"
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

func ginContextFromContext(ctx context.Context) (*gin.Context, error) {
	ginContext := ctx.Value(ginContextKey)
	if ginContext == nil {
		return nil, fmt.Errorf("could not retrieve gin.Context")
	}

	gc, ok := ginContext.(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("gin.Context has wrong type")
	}
	return gc, nil
}

func ginContextSet(c *gin.Context, key string, val string) {
	c.Set(scannerNameKey, val)
	ctx := context.WithValue(c.Request.Context(), ginContextKey, c)
	c.Request = c.Request.WithContext(ctx)
}

func ginContextGet(ctx context.Context, key string) (string, error) {
	gc, err := ginContextFromContext(ctx)
	if err != nil {
		return "", err
	}

	s, ok := gc.Get(key)
	if !ok {
		return "", fmt.Errorf("could not find key: '%s' in gin.Context", key)
	}
	ss, ok := s.(string)
	if !ok {
		return "", fmt.Errorf("invalid key type: '%T', should be string", s)
	}
	return ss, nil
}

func UserNameFromContext(ctx context.Context) (string, error) {
	return ginContextGet(ctx, userNameKey)
}

func ScannerNameFromContext(ctx context.Context) (string, error) {
	return ginContextGet(ctx, scannerNameKey)
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
