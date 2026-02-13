// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	authctx "github.com/cloudoperators/heureka/internal/api/graphql/access/context"
	"github.com/cloudoperators/heureka/internal/util"
)

const (
	oidcAuthHeader     string = "Authorization"
	oidcAuthMethodName string = "OidcAuthMethod"
)

type IDTokenClaims struct {
	Sub           string `json:"sub"`
	EmailVerified bool   `json:"email_verified"`
	Mail          string `json:"mail"`
	Iss           string `json:"iss"`
	LastName      string `json:"last_name"`
	GivenName     string `json:"access_token"`
	Aud           string `json:"aud"`
	UserUuid      string `json:"user_uuid"`
	Exp           int    `json:"exp"`
	IAT           int    `json:"iat"`
	FirstName     string `json:"first_name"`
	FamilyName    string `json:"family_name"`
	JTI           string `json:"jti"`
	Email         string `json:"email"`
}

type OidcAuthMethod struct {
	logger   *logrus.Logger
	provider *oidc.Provider
	config   *oidc.Config
	verifier *oidc.IDTokenVerifier
}

func NewOidcAuthMethod(l *logrus.Logger, cfg *util.Config) authMethod {
	if cfg.AuthOidcUrl != "" {
		oidcProvider, err := oidc.NewProvider(context.TODO(), cfg.AuthOidcUrl)
		if err != nil {
			l.Error("Could not initialize OIDC provider: ", err, " (", cfg.AuthOidcUrl, ")")
			return &NoAuthMethod{
				logger:         l,
				authMethodName: oidcAuthMethodName,
				msg:            "Could not initialize OIDC provider: " + err.Error(),
			}
		}

		oidcConfig := &oidc.Config{
			ClientID: cfg.AuthOidcClientId,
		}

		oidcVerifier := oidcProvider.Verifier(oidcConfig)

		l.Info("Initializing OIDC auth")
		return &OidcAuthMethod{logger: l, provider: oidcProvider, config: oidcConfig, verifier: oidcVerifier}
	}
	return nil
}

func (oam OidcAuthMethod) Verify(c *gin.Context) error {
	rawToken, err := getAuthTokenFromHeader(oidcAuthHeader, c)
	if err != nil {
		return verifyError(oidcAuthMethodName, err)
	}

	token, err := oam.verifier.Verify(c.Request.Context(), rawToken)
	if err != nil {
		return verifyError(oidcAuthMethodName, err)
	}

	var claims IDTokenClaims
	err = token.Claims(&claims)
	if err != nil {
		return verifyError(oidcAuthMethodName, fmt.Errorf("failed to parse token claims: %w", err))
	}

	authctx.UserNameToContext(c, claims.Sub)

	return nil
}
