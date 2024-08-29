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
