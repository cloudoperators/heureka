// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"

	"github.com/cloudoperators/heureka/internal/api/graphql/access/auth"
	authctx "github.com/cloudoperators/heureka/internal/api/graphql/access/context"
)

func NewAuth(cfg *util.Config, resolve auth.UserResolver, enableLog bool) *Auth {
	l := newLogger(enableLog)
	a := Auth{logger: l}

	a.appendInstance(auth.NewTokenAuthMethod(l, cfg, resolve))
	a.appendInstance(auth.NewOidcAuthMethod(l, cfg, resolve))

	return &a
}

// NewDbUserResolver returns a UserResolver that looks up users by unique_user_id in the given DB.
func NewDbUserResolver(db database.Database) auth.UserResolver {
	return func(ctx context.Context, uniqueUserId string) (int64, entity.UserType, error) {
		filter := &entity.UserFilter{UniqueUserID: []*string{&uniqueUserId}}
		users, err := db.GetUsers(ctx, filter, nil)
		if err != nil {
			return 0, entity.InvalidUserType, fmt.Errorf("db lookup failed: %w", err)
		}
		if len(users) == 0 {
			return 0, entity.InvalidUserType, fmt.Errorf("user not found")
		}
		return users[0].Id, users[0].Type, nil
	}
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

			for _, am := range a.chain {
				if err := am.Verify(authCtx); err == nil {
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
	}
}

func (a *Auth) appendInstance(am authMethod) {
	if am != nil &&
		(reflect.ValueOf(am).Kind() != reflect.Pointer || !reflect.ValueOf(am).IsNil()) {
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
