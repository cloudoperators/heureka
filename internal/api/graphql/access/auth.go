// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package access

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"github.com/cloudoperators/heureka/internal/util"
)

type Logger interface {
	Error(...interface{})
	Warn(...interface{})
	Info(...interface{})
}

func NewAuth(cfg *util.Config) *Auth {
	l := newLogger()
	auth := Auth{logger: l}
	auth.appendInstance(NewTokenAuthMethod(l, cfg))
	auth.appendInstance(NewOidcAuthMethod(l, cfg))
	return &auth
}

type Auth struct {
	chain  []authMethod
	logger Logger
}

type authMethod interface {
	AddWhitelistRoutes(*gin.Engine)
	Verify(*gin.Context) error
}

func (a *Auth) AddWhitelistRoutes(router *gin.Engine) {
	for _, auth := range a.chain {
		auth.AddWhitelistRoutes(router)
	}
}

func (a *Auth) GetMiddleware() gin.HandlerFunc {
	return func(authCtx *gin.Context) {
		if len(a.chain) > 0 {
			var retMsg string
			for _, auth := range a.chain {
				if err := auth.Verify(authCtx); err == nil {
					authCtx.Next()
					return
				} else {
					if retMsg != "" {
						retMsg = fmt.Sprintf("%s, ", retMsg)
					}
					retMsg = fmt.Sprintf("%s%s", retMsg, err)
				}
			}
			a.logger.Error("Unauthorized access: ", retMsg)
			authCtx.JSON(http.StatusUnauthorized, gin.H{"error": retMsg})
			authCtx.Abort()
			return
		}
		authCtx.Next()
		return
	}
}

func (a *Auth) appendInstance(am authMethod) {
	if am != nil && !(reflect.ValueOf(am).Kind() == reflect.Ptr && reflect.ValueOf(am).IsNil()) {
		a.chain = append(a.chain, am)
	}
}

func newLogger() Logger {
	return logrus.New()
}

//TODO: move to auth_util:

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
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	return tokenString, nil
}

func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}
