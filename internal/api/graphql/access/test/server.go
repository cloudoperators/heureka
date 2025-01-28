package util

import (
	"context"
	"fmt"
	"net/http"
	"time"

	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/gin-gonic/gin"

	"github.com/cloudoperators/heureka/internal/api/graphql/access"
)

const (
	testEndpoint = "/testendpoint"
)

type TestServer struct {
	port           string
	auth           *access.Auth
	cancel         context.CancelFunc
	ctx            context.Context
	srv            *http.Server
	lastRequestCtx context.Context
}

func NewTestServer(auth *access.Auth) *TestServer {
	port := util2.GetRandomFreePort()
	return &TestServer{
		port: port,
		auth: auth,
	}
}

func (ts *TestServer) StartInBackground() {
	ts.lastRequestCtx = context.TODO()
	r := gin.Default()
	r.Use(ts.auth.GetMiddleware())
	r.GET(testEndpoint, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
		ts.lastRequestCtx = c.Request.Context()
	})

	ts.ctx, ts.cancel = context.WithCancel(context.Background())

	ts.srv = &http.Server{Addr: fmt.Sprintf(":%s", ts.port), Handler: r}
	util2.FirstListenThenServe(ts.srv)
}

func (ts *TestServer) Stop() {
	fmt.Println("Shuting down the server...")
	ts.cancel()

	ctxTimeout, cancelTimeout := context.WithTimeout(ts.ctx, 5*time.Second)
	defer cancelTimeout()

	if err := ts.srv.Shutdown(ctxTimeout); err != nil {
		fmt.Println("Server forced to shutdown: ", err)
	}

	fmt.Println("Server exiting")
}

func (ts *TestServer) Context() context.Context {
	return ts.lastRequestCtx
}

func (ts *TestServer) EndpointUrl() string {
	return fmt.Sprintf("http://localhost:%s", ts.port) + testEndpoint
}
