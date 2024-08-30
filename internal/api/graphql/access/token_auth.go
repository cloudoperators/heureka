package access

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type TokenAuth struct {
	logger Logger
	secret []byte
}

func NewTokenAuth(l Logger) *TokenAuth {
	return &TokenAuth{logger: l, secret: []byte(os.Getenv("AUTH_TOKEN_SECRET"))}
}

type TokenClaims struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
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
		} else if !token.Valid {
			ta.logger.Error("Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("authinfo", claims)

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
