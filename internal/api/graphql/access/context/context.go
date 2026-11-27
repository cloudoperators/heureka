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
	ginContextKey             ginContextKeyType = "GinContextKey"
	UserNameKey               string            = "username"
	UserIdKey                 string            = "userId"
	authenticationRequiredKey string            = "authenticationRequired"
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

func ginContextSet[T any](c *gin.Context, key string, val T) {
	c.Set(key, val)
	ctx := context.WithValue(c.Request.Context(), ginContextKey, c)
	c.Request = c.Request.WithContext(ctx)
}

func ginContextGet(ctx context.Context, key string) (string, error) {
	var result string

	gc, err := ginContextFromContext(ctx)
	if err != nil {
		return result, err
	}

	v, ok := gc.Get(key)
	if !ok {
		return result, fmt.Errorf("could not find key: '%s' in gin.Context", key)
	}

	result, ok = v.(string)
	if !ok {
		return result, fmt.Errorf("invalid key type: '%T', should be string", v)
	}

	return result, nil
}

func ginContextGetBool(ctx context.Context, key string) bool {
	gc, err := ginContextFromContext(ctx)
	if err != nil {
		return false
	}

	return gc.GetBool(key)
}

func ginContextGetInt64(ctx context.Context, key string) (int64, error) {
	gc, err := ginContextFromContext(ctx)
	if err != nil {
		return 0, err
	}

	v, ok := gc.Get(key)
	if !ok {
		return 0, fmt.Errorf("could not find key: '%s' in gin.Context", key)
	}

	id, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("invalid key type: '%T', should be int64", v)
	}

	return id, nil
}

func UserNameToContext(c *gin.Context, username string) {
	ginContextSet(c, UserNameKey, username)
}

func UserNameFromContext(ctx context.Context) (string, error) {
	return ginContextGet(ctx, UserNameKey)
}

func UserIdToContext(c *gin.Context, id int64) {
	ginContextSet(c, UserIdKey, id)
}

func UserIdFromContext(ctx context.Context) (int64, error) {
	return ginContextGetInt64(ctx, UserIdKey)
}

func SetAuthenticationRequired(c *gin.Context) {
	ginContextSet(c, authenticationRequiredKey, true)
}

func IsAuthenticationRequired(ctx context.Context) bool {
	return ginContextGetBool(ctx, authenticationRequiredKey)
}
