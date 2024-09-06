package access_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.wdf.sap.corp/cc/heureka/internal/api/graphql/access/test"
	util2 "github.wdf.sap.corp/cc/heureka/pkg/util"

	"github.com/gin-gonic/gin"

	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/access"
	"github.wdf.sap.corp/cc/heureka/internal/util"
)

const (
	testEndpoint    = "/testendpoint"
	testUsername    = "testAccessUser"
	authTokenSecret = "xxx"
)

type noLogLogger struct {
}

func (nll noLogLogger) Error(...interface{}) {
}

func (nll noLogLogger) Warn(...interface{}) {
}

type server struct {
	cancel         context.CancelFunc
	ctx            context.Context
	srv            *http.Server
	lastRequestCtx context.Context
}

func (s *server) startInBackground(port string) {
	s.lastRequestCtx = context.TODO()
	auth := access.NewTokenAuth(&noLogLogger{}, &util.Config{AuthTokenSecret: authTokenSecret})
	r := gin.Default()
	r.Use(auth.GetMiddleware())
	r.GET(testEndpoint, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
		s.lastRequestCtx = c.Request.Context()
	})

	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.srv = &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: r}
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Unexpected error when running test server")
		}
	}()
}

func (s *server) stop() {
	fmt.Println("Shuting down the server...")
	s.cancel()

	ctxTimeout, cancelTimeout := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancelTimeout()

	if err := s.srv.Shutdown(ctxTimeout); err != nil {
		fmt.Println("Server forced to shutdown: ", err)
	}

	fmt.Println("Server exiting")
}

func (s *server) context() context.Context {
	return s.lastRequestCtx
}

var _ = Describe("Pass token data via context when using token auth middleware", Label("api", "TokenAuthorization"), func() {
	var testServer server
	var url string

	BeforeEach(func() {
		port := util2.GetRandomFreePort()
		url = fmt.Sprintf("http://localhost:%s%s", port, testEndpoint)
		testServer.startInBackground(port)
	})

	AfterEach(func() {
		testServer.stop()
	})

	When("User access api through token auth middleware with valid token", func() {
		BeforeEach(func() {
			token := GenerateJwtWithUsername(authTokenSecret, 1*time.Hour, testUsername)
			resp := SendGetRequest(url, map[string]string{"Authorization": token})
			Expect(resp.StatusCode).To(Equal(200))
		})
		It("Should be able to access user name from request context", func() {
			username, err := access.UsernameFromContext(testServer.context())
			Expect(err).To(BeNil())
			Expect(username).To(BeEquivalentTo(testUsername))
		})
	})

	When("User access api through token auth middleware with invalid token", func() {
		BeforeEach(func() {
			token := GenerateInvalidJwt(authTokenSecret)
			resp := SendGetRequest(url, map[string]string{"Authorization": token})
			Expect(resp.StatusCode).To(Equal(401))
		})
		It("Should not store gin context in request context", func() {
			_, err := access.UsernameFromContext(testServer.context())
			Expect(err).ShouldNot(BeNil())
		})
	})
})

func TestTokenAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Token Auth Suite")
}
