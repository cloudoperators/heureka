// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"github.com/gin-gonic/gin"
)

type authMethod interface {
	Verify(*gin.Context) error
}
