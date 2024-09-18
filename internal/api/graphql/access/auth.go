// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package access

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/cloudoperators/heureka/internal/util"
)

type Logger interface {
	Error(...interface{})
	Warn(...interface{})
}

type Auth interface {
	GetMiddleware() gin.HandlerFunc
}

func NewAuth(cfg *util.Config) Auth {
	l := newLogger()

	authType := strings.ToLower(cfg.AuthType)
	if authType == "token" {
		return NewTokenAuth(l, cfg)
	} else if authType == "none" {
		return NewNoAuth()
	}

	l.Warn("AUTH_TYPE is not set, assuming 'none' authorization method")

	return NewNoAuth()
}

func newLogger() Logger {
	return logrus.New()
}
