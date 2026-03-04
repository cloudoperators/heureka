// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/cloudoperators/heureka/internal/util"

	. "github.com/cloudoperators/heureka/internal/api/graphql/access/auth"
	authctx "github.com/cloudoperators/heureka/internal/api/graphql/access/context"
)

func NewAuth(cfg *util.Config, enableLog bool) *Auth {
	l := newLogger(enableLog)
	auth := Auth{logger: l}
	auth.appendInstance(NewTokenAuthMethod(l, cfg))
	auth.appendInstance(NewOidcAuthMethod(l, cfg))
	return &auth
}

type Auth struct {
	chain  []authMethod
	logger *logrus.Logger
}

type authMethod interface {
	Verify(*gin.Context) error
}

func (a *Auth) Middleware() gin.HandlerFunc {
	return func(authCtx *gin.Context) {
		if len(a.chain) > 0 {
			authctx.SetAuthenticationRequired(authCtx)
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

func newLogger(enableLog bool) *logrus.Logger {
	l := logrus.New()
	if !enableLog {
		l.SetOutput(ioutil.Discard)
	}
	return l
}
