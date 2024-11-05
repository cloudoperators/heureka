// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package access

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/cloudoperators/heureka/internal/util"
)

type ginContextKeyType string

const (
	ginContextKey   ginContextKeyType = "GinContextKey"
	scannerNameKey  string            = "scannername"
	tokenAuthHeader string            = "X-Service-Authorization"
)

func NewTokenAuthMethod(l Logger, cfg *util.Config) *TokenAuthMethod {
	if cfg.AuthTokenSecret != "" {
		return &TokenAuthMethod{logger: l, secret: []byte(cfg.AuthTokenSecret)}
	}
	return nil
}

type TokenClaims struct {
	Version string `json:"version"`
	jwt.RegisteredClaims
}

type TokenAuthMethod struct {
	logger Logger
	secret []byte
}

func (tam TokenAuthMethod) Verify(c *gin.Context) error {
	verifyError := func(s string) error {
		return fmt.Errorf("TokenAuthMethod(%s)", s)
	}

	tokenString := c.GetHeader(tokenAuthHeader)
	if tokenString == "" {
		return verifyError("No authorization header")
	}
	token, claims, err := tam.parseFromString(tokenString)
	if err != nil {
		tam.logger.Error("JWT parsing error: ", err)
		return verifyError("Token parsing error")
	} else if !token.Valid || claims.ExpiresAt == nil {
		tam.logger.Error("Invalid token")
		return verifyError("Invalid token")
	} else if claims.ExpiresAt.Before(time.Now()) {
		tam.logger.Warn("Expired token")
		return verifyError("Token expired")
	}
	c.Set(scannerNameKey, claims.RegisteredClaims.Subject)
	ctx := context.WithValue(c.Request.Context(), ginContextKey, c)
	c.Request = c.Request.WithContext(ctx)
	return nil
}

func (tam TokenAuthMethod) parseFromString(tokenString string) (*jwt.Token, *TokenClaims, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, tam.parse)
	return token, claims, err
}

func (tam *TokenAuthMethod) parse(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Invalid JWT parse method")
	}
	return tam.secret, nil
}

func ScannerNameFromContext(ctx context.Context) (string, error) {
	gc, err := ginContextFromContext(ctx)
	if err != nil {
		return "", err
	}

	s, ok := gc.Get(scannerNameKey)
	if !ok {
		return "", fmt.Errorf("could not find scanner name in gin.Context")
	}
	ss, ok := s.(string)
	if !ok {
		return "", fmt.Errorf("invalid scanner name type")
	}
	return ss, nil
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
