package access

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type TokenAuth struct {
	logger           Logger
	scannerSecretMap map[string]string
}

func NewTokenAuth(l Logger) *TokenAuth {
	ta := TokenAuth{logger: l}

	env := os.Getenv("AUTH_TOKEN_MAP")
	if env != "" {
		err := json.Unmarshal([]byte(env), &ta.scannerSecretMap)
		if err != nil {
			l.Error("Error parsing JSON: ", err.Error())
		}
	}

	return &ta
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

		token, err := ta.parseFromString(tokenString)
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

		c.Next()
	}
}

func (ta *TokenAuth) parseFromString(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, ta.parse)
}

func (ta *TokenAuth) parse(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Invalid JWT parse method")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("Could not get claims from JWT")
	}
	scannerName, ok := claims["scanner_name"].(string)
	if !ok {
		return nil, fmt.Errorf("Could not claim scanner_name from JWT")
	}
	secret, ok := ta.scannerSecretMap[scannerName]
	if !ok {
		return nil, fmt.Errorf("Could not find secret for scanner: %s", scannerName)
	}
	return []byte(secret), nil
}
