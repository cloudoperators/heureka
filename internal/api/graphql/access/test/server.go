// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/cloudoperators/heureka/internal/api/graphql/access/middleware"
)

const (
	testEndpoint = "/testendpoint"
)

type TestServer struct {
	port           string
	auth           *middleware.Auth
	cancel         context.CancelFunc
	ctx            context.Context
	srv            *http.Server
	lastRequestCtx context.Context
	log            *logrus.Logger
}

func NewTestServer(auth *middleware.Auth, enableLog bool) *TestServer {
	port := util2.GetRandomFreePort()
	log := logrus.New()
	if !enableLog {
		log.SetOutput(ioutil.Discard)
	}
	return &TestServer{
		port: port,
		auth: auth,
		log:  log,
	}
}

func (ts *TestServer) StartInBackground() {
	ts.lastRequestCtx = context.TODO()
	gin.DefaultWriter = ts.log.Writer()
	r := gin.Default()
	r.Use(ts.auth.Middleware())
	r.GET(testEndpoint, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
		ts.lastRequestCtx = c.Request.Context()
	})

	ts.ctx, ts.cancel = context.WithCancel(context.Background())

	ts.srv = &http.Server{Addr: fmt.Sprintf(":%s", ts.port), Handler: r}
	util2.FirstListenThenServe(ts.srv, ts.log)
}

func (ts *TestServer) Stop() {
	ts.log.Info("Shuting down the server...")
	ts.cancel()

	ctxTimeout, cancelTimeout := context.WithTimeout(ts.ctx, 5*time.Second)
	defer cancelTimeout()

	if err := ts.srv.Shutdown(ctxTimeout); err != nil {
		ts.log.Error("Server forced to shutdown: ", err)
	}

	ts.log.Info("Server exiting")
}

func (ts *TestServer) Context() context.Context {
	return ts.lastRequestCtx
}

func (ts *TestServer) EndpointUrl() string {
	return fmt.Sprintf("http://localhost:%s", ts.port) + testEndpoint
}
