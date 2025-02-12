// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package access

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/cloudoperators/heureka/internal/util"
)

type Logger interface {
	Error(...interface{})
	Warn(...interface{})
}

func NewAuth(cfg *util.Config) *Auth {
	l := newLogger()
	auth := Auth{logger: l}
	auth.AppendInstance(NewTokenAuthMethod(l, cfg))
	//TODO: auth.AppendInstance(NewOidcAuthMethod(l, cfg))
	return &auth
}

type Auth struct {
	chain  []AuthMethod
	logger Logger
}

type AuthMethod interface {
	Verify(*gin.Context) error
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
			a.logger.Error("Unauthorized access: %s", retMsg)
			authCtx.JSON(http.StatusUnauthorized, gin.H{"error": retMsg})
			authCtx.Abort()
			return
		}
		authCtx.Next()
		return
	}
}

func (a *Auth) AppendInstance(am AuthMethod) {
	if !reflect.ValueOf(am).IsNil() {
		a.chain = append(a.chain, am)
	}
}

func newLogger() Logger {
	return logrus.New()
}
