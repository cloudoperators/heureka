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
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
)

const (
	tokenAuthHeader     string = "X-Service-Authorization"
	tokenAuthMethodName string = "TokenAuthMethod"
)

func NewTokenAuthMethod(l *logrus.Logger, cfg *util.Config, resolve UserResolver) authMethod {
	if cfg.AuthTokenSecret != "" {
		if resolve == nil {
			l.Fatal("NewTokenAuthMethod: resolve is nil")
		}
		l.Info("Initializing Token auth")
		return &TokenAuthMethod{logger: l, secret: []byte(cfg.AuthTokenSecret), resolve: resolve}
	}

	return nil
}

type TokenAuthMethod struct {
	logger  *logrus.Logger
	secret  []byte
	resolve UserResolver
}

func (tam *TokenAuthMethod) Verify(c *gin.Context) error {
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

	userId, userType, err := tam.resolve(c.Request.Context(), claims.Subject)
	if err != nil {
		tam.logger.WithError(err).WithField("subject", claims.Subject).Warn("user associated with token not found")
		return verifyError(tokenAuthMethodName, fmt.Errorf("invalid token"))
	}

	if userType != entity.TechnicalUserType {
		tam.logger.WithField("subject", claims.Subject).Warn("user associated with token is not a technical user")
		return verifyError(tokenAuthMethodName, fmt.Errorf("invalid token"))
	}

	authctx.UserNameToContext(c, claims.Subject)
	authctx.UserIdToContext(c, userId)

	return nil
}

func (tam *TokenAuthMethod) parseTokenWithClaims(tokenString string) (*TokenClaims, error) {
	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, tam.parse)
	if err != nil {
		tam.logger.WithError(err).Warn("JWT parsing error")
		err = verifyError(tokenAuthMethodName, fmt.Errorf("invalid token"))
	} else if !token.Valid {
		tam.logger.Warn("JWT signature invalid")

		err = verifyError(tokenAuthMethodName, fmt.Errorf("invalid token"))
	}

	return claims, err
}

func (tam *TokenAuthMethod) verifyTokenExpiration(tc *TokenClaims) error {
	var err error

	if tc.ExpiresAt == nil {
		tam.logger.Warn("missing ExpiresAt in token claims")

		err = verifyError(tokenAuthMethodName, fmt.Errorf("invalid token"))
	} else if tc.ExpiresAt.Before(time.Now()) {
		tam.logger.Warn("expired token")

		err = verifyError(tokenAuthMethodName, fmt.Errorf("invalid token"))
	}

	return err
}

func (tam *TokenAuthMethod) parse(token *jwt.Token) (any, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("invalid JWT parse method")
	}

	return tam.secret, nil
}
