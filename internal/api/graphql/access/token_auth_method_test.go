// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package access_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudoperators/heureka/internal/api/graphql/access/test"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/gin-gonic/gin"

	"github.com/cloudoperators/heureka/internal/api/graphql/access"
	"github.com/cloudoperators/heureka/internal/util"
)

const (
	testEndpoint    = "/testendpoint"
	testScannerName = "testAccessScanner"
	authTokenSecret = "xxx"
)

type server struct {
	cancel         context.CancelFunc
	ctx            context.Context
	srv            *http.Server
	lastRequestCtx context.Context
}

func (s *server) startInBackground(port string) {
	s.lastRequestCtx = context.TODO()
	auth := access.NewAuth(&util.Config{AuthTokenSecret: authTokenSecret})
	r := gin.Default()
	r.Use(auth.GetMiddleware())
	r.GET(testEndpoint, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
		s.lastRequestCtx = c.Request.Context()
	})

	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.srv = &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: r}
	util2.FirstListenThenServe(s.srv)
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

	When("Scanner access api through token auth middleware with valid token", func() {
		BeforeEach(func() {
			token := GenerateJwtWithName(TokenStringHandler, authTokenSecret, 1*time.Hour, testScannerName)
			resp := SendGetRequest(url, map[string]string{"X-Service-Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(200))
		})
		It("Should be able to access scanner name from request context", func() {
			name, err := access.ScannerNameFromContext(testServer.context())
			Expect(err).To(BeNil())
			Expect(name).To(BeEquivalentTo(testScannerName))
		})
	})

	When("Scanner access api through token auth middleware with invalid token", func() {
		BeforeEach(func() {
			token := GenerateJwt(InvalidTokenStringHandler, authTokenSecret, 1*time.Hour)
			resp := SendGetRequest(url, map[string]string{"X-Service-Authorization": WithBearer(token)})
			Expect(resp.StatusCode).To(Equal(401))
		})
		It("Should not store gin context in request context", func() {
			_, err := access.ScannerNameFromContext(testServer.context())
			Expect(err).ShouldNot(BeNil())
		})
	})
})

func TestTokenAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Token Auth Suite")
}
