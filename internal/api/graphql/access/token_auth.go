package access

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.wdf.sap.corp/cc/heureka/internal/util"
)

const (
	ginContextKey ginContextKeyType = "GinContextKey"
	usernameKey   string            = "username"
)

type ginContextKeyType string

type TokenAuth struct {
	logger Logger
	secret []byte
}

func NewTokenAuth(l Logger, cfg *util.Config) *TokenAuth {
	return &TokenAuth{logger: l, secret: []byte(cfg.AuthTokenSecret)}
}

type TokenClaims struct {
	Version string `json:"version"`
	jwt.RegisteredClaims
}

func (ta *TokenAuth) GetMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			ta.logger.Error("Trying to use API without authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		token, claims, err := ta.parseFromString(tokenString)
		if err != nil {
			ta.logger.Error("JWT parsing error: ", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token parsing error"})
			c.Abort()
			return
		} else if !token.Valid || claims.ExpiresAt == nil {
			ta.logger.Error("Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		} else if claims.ExpiresAt.Before(time.Now()) {
			ta.logger.Warn("Expired token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			c.Abort()
			return
		}

		c.Set(usernameKey, claims.RegisteredClaims.Subject)
		ctx := context.WithValue(c.Request.Context(), ginContextKey, c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (ta *TokenAuth) parseFromString(tokenString string) (*jwt.Token, *TokenClaims, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, ta.parse)
	return token, claims, err
}

func (ta *TokenAuth) parse(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Invalid JWT parse method")
	}
	return ta.secret, nil
}

func UsernameFromContext(ctx context.Context) (string, error) {
	gc, err := ginContextFromContext(ctx)
	if err != nil {
		return "", err
	}

	u, ok := gc.Get(usernameKey)
	if !ok {
		return "", fmt.Errorf("could not find username in gin.Context")
	}
	us, ok := u.(string)
	if !ok {
		return "", fmt.Errorf("invalid username type")
	}
	return us, nil
}

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
