// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type NoAuthMethod struct {
	logger         *logrus.Logger
	authMethodName string
	msg            string
}

func NewNoAuthMethod(l *logrus.Logger, authMethodName string, msg string) authMethod {
	return &NoAuthMethod{logger: l, authMethodName: authMethodName, msg: msg}
}

func (nam NoAuthMethod) Verify(*gin.Context) error {
	return verifyError(nam.authMethodName, fmt.Errorf("auth failed: %s", nam.msg))
}
