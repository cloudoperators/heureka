// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"github.com/cloudoperators/heureka/internal/database/querycounter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// QueryCounter is a Gin middleware that initializes a per-request DB query counter
// and logs the total count at request completion.
func QueryCounter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := querycounter.Init(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		count := querycounter.GetQueryCount(c.Request.Context())
		logrus.WithFields(logrus.Fields{
			"db_query_count": count,
			"method":         c.Request.Method,
			"path":           c.Request.URL.Path,
		}).Debug("Request completed")
	}
}
