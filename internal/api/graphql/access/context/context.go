// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package context

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ginContextKeyType string

const (
	ginContextKey  ginContextKeyType = "GinContextKey"
	scannerNameKey string            = "scannername"
	userNameKey    string            = "username"
)

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
	c.Set(key, val)
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

func UserNameToContext(c *gin.Context, username string) {
	ginContextSet(c, userNameKey, username)
}

func UserNameFromContext(ctx context.Context) (string, error) {
	return ginContextGet(ctx, userNameKey)
}

func ScannerNameToContext(c *gin.Context, scannername string) {
	ginContextSet(c, scannerNameKey, scannername)
}

func ScannerNameFromContext(ctx context.Context) (string, error) {
	return ginContextGet(ctx, scannerNameKey)
}
