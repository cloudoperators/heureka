package access

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Error(...interface{})
	Warn(...interface{})
}

type Auth interface {
	GetMiddleware() gin.HandlerFunc
}

func NewAuth() Auth {
	l := newLogger()

	authType := strings.ToLower(os.Getenv("AUTH_TYPE"))
	if authType == "token" {
		return NewTokenAuth(l)
	} else if authType == "none" {
		return NewNoAuth()
	}

	l.Warn("AUTH_TYPE is not set, assuming 'none' authorization method")

	return NewNoAuth()
}

func newLogger() Logger {
	return logrus.New()
}
