// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package access

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/cloudoperators/heureka/internal/util"
)

const (
	tokenAuthHeader     string = "X-Service-Authorization"
	tokenAuthMethodName string = "TokenAuthMethod"
)

func NewTokenAuthMethod(l Logger, cfg *util.Config) authMethod {
	if cfg.AuthTokenSecret != "" {
		l.Info("Initializing Token auth")
		return &TokenAuthMethod{logger: l, secret: []byte(cfg.AuthTokenSecret)}
	}
	return nil
}

type TokenAuthMethod struct {
	logger Logger
	secret []byte
}

func (tam TokenAuthMethod) Verify(c *gin.Context) error {
	tokenString, err := getAuthTokenFromHeader(tokenAuthHeader, c)
	if err != nil {
		return verifyError(tokenAuthMethodName, err)
	}

	claims, err := tam.verifyTokenAndGetClaimsFromTokenString(tokenString)
	if err != nil {
		return err
	}

	err = tam.verifyTokenExpiration(claims)
	if err != nil {
		return err
	}

	ginContextSet(c, scannerNameKey, claims.RegisteredClaims.Subject)

	return nil
}

func (tam TokenAuthMethod) verifyTokenAndGetClaimsFromTokenString(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, tam.parse)
	if err != nil {
		tam.logger.Error("JWT parsing error: ", err)
		err = verifyError(tokenAuthMethodName, fmt.Errorf("Token parsing error"))
	} else if !token.Valid {
		tam.logger.Error("Invalid token")
		err = verifyError(tokenAuthMethodName, fmt.Errorf("Invalid token"))
	}
	return claims, err
}

func (tam TokenAuthMethod) verifyTokenExpiration(tc *TokenClaims) error {
	var err error
	if tc.ExpiresAt == nil {
		tam.logger.Error("Missing ExpiresAt in token claims")
		err = verifyError(tokenAuthMethodName, fmt.Errorf("Missing ExpiresAt in token claims"))
	} else if tc.ExpiresAt.Before(time.Now()) {
		tam.logger.Warn("Expired token")
		err = verifyError(tokenAuthMethodName, fmt.Errorf("Expired token"))
	}
	return err
}

func (tam *TokenAuthMethod) parse(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Invalid JWT parse method")
	}
	return tam.secret, nil
}
