// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	authctx "github.com/cloudoperators/heureka/internal/api/graphql/access/context"
	"github.com/cloudoperators/heureka/internal/util"
)

const (
	tokenAuthHeader     string = "X-Service-Authorization"
	tokenAuthMethodName string = "TokenAuthMethod"
)

func NewTokenAuthMethod(l *logrus.Logger, cfg *util.Config) authMethod {
	if cfg.AuthTokenSecret != "" {
		l.Info("Initializing Token auth")
		return &TokenAuthMethod{logger: l, secret: []byte(cfg.AuthTokenSecret)}
	}
	return nil
}

type TokenAuthMethod struct {
	logger *logrus.Logger
	secret []byte
}

func (tam TokenAuthMethod) Verify(c *gin.Context) error {
	tokenString, err := getAuthTokenFromHeader(tokenAuthHeader, c)
	if err != nil {
		return verifyError(tokenAuthMethodName, err)
	}

	claims, err := tam.parseTokenWithClaims(tokenString)
	if err != nil {
		return err
	}

	err = tam.verifyTokenExpiration(claims)
	if err != nil {
		return err
	}

	authctx.UserNameToContext(c, claims.Subject)

	return nil
}

func (tam TokenAuthMethod) parseTokenWithClaims(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, tam.parse)
	if err != nil {
		tam.logger.Error("JWT parsing error: ", err)
		err = verifyError(tokenAuthMethodName, fmt.Errorf("token parsing error"))
	} else if !token.Valid {
		tam.logger.Error("Invalid token")
		err = verifyError(tokenAuthMethodName, fmt.Errorf("invalid token"))
	}
	return claims, err
}

func (tam TokenAuthMethod) verifyTokenExpiration(tc *TokenClaims) error {
	var err error
	if tc.ExpiresAt == nil {
		tam.logger.Error("Missing ExpiresAt in token claims")
		err = verifyError(tokenAuthMethodName, fmt.Errorf("missing ExpiresAt in token claims"))
	} else if tc.ExpiresAt.Before(time.Now()) {
		tam.logger.Warn("Expired token")
		err = verifyError(tokenAuthMethodName, fmt.Errorf("expired token"))
	}
	return err
}

func (tam *TokenAuthMethod) parse(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("invalid JWT parse method")
	}
	return tam.secret, nil
}
