// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package loaders

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ContextKey string

const LoadersKey = ContextKey("dataloaders")

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = withLoaders(c.Request)
		c.Next()
	}
}

func withLoaders(r *http.Request) *http.Request {
	loaders := newLoaders()
	ctx := context.WithValue(r.Context(), LoadersKey, loaders)
	return r.WithContext(ctx)
}

func getLoaders(ctx context.Context) (*Loaders, bool) {
	l, ok := ctx.Value(LoadersKey).(*Loaders)
	return l, ok
}

type Loaders struct {
	listSupportGroups *listSupportGroupsLoader
}

func newLoaders() *Loaders {
	return &Loaders{
		listSupportGroups: newListSupportGroupsLoader(),
	}
}
