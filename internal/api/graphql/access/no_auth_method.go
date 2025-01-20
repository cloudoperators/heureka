package access

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	noAuthMethodName string = "NoAuthMethod"
)

type NoAuthMethod struct {
	logger         Logger
	authMethodName string
	msg            string
}

func NewNoAuthMethod(l Logger, authMethodName string, msg string) authMethod {
	return &NoAuthMethod{logger: l, authMethodName: authMethodName, msg: msg}
}

func (nam NoAuthMethod) AddWhitelistRoutes(*gin.Engine) {
}

func (nam NoAuthMethod) Verify(*gin.Context) error {
	return verifyError(nam.authMethodName, fmt.Errorf("Auth failed: %s", nam.msg))

}
