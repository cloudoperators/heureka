// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package access

import (
	"github.com/gin-gonic/gin"
)

type NoAuth struct {
}

func NewNoAuth() *NoAuth {
	return &NoAuth{}
}

func (no *NoAuth) GetMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
