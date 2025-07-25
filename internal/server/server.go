// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	graphqlapi "github.com/cloudoperators/heureka/internal/api/graphql"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/database/mariadb"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"
	"github.com/onuryilmaz/ginprom"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router     *gin.Engine
	graphQLAPI *graphqlapi.GraphQLAPI
	config     util.Config

	// Use this context if you want your software
	// unit to be notified about heureka shutdown
	shutdownCtx context.Context

	// Heureka cancel function used by Heureka shutdown
	shutdownFunc context.CancelFunc

	// Use this workgroup if you want Heureka to
	// block shutdown until important job is done
	wg *sync.WaitGroup

	nonBlockingSrv *http.Server

	app *app.HeurekaApp
}

func NewServer(cfg util.Config) *Server {
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	wg := sync.WaitGroup{}

	db, err := mariadb.NewSqlDatabase(cfg)
	if err != nil {
		logrus.WithError(err).Fatalln("Error while Creating Db")
	}

	err = db.RunMigrations()
	if err != nil {
		logrus.WithError(err).Fatalln("Error while Migrating Db")
	}

	db, err = mariadb.NewSqlDatabase(cfg)
	if err != nil {
		logrus.WithError(err).Fatalln("Error while Creating Db")
	}

	application := app.NewHeurekaApp(ctx, &wg, db, cfg)

	s := Server{
		router:       &gin.Engine{},
		graphQLAPI:   graphqlapi.NewGraphQLAPI(application, cfg),
		config:       cfg,
		app:          application,
		shutdownCtx:  ctx,
		shutdownFunc: cancel,
		wg:           &wg,
	}

	if logrus.GetLevel() == logrus.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s.router = gin.New()
	s.router.ForwardedByClientIP = true
	s.router.Use(cors.New(cors.Config{
		AllowHeaders:     []string{"Authorization", "Content-Type", "Accept"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT"},
		AllowCredentials: true,
		MaxAge:           600,
		AllowAllOrigins:  true,
	}))

	s.initLogging()
	s.createEndpoints()

	return &s
}

func (s *Server) initLogging() {
	logFormatter := func(param gin.LogFormatterParams) string {
		var statusColor, methodColor, resetColor string
		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			methodColor = param.MethodColor()
			resetColor = param.ResetColor()
		}

		return fmt.Sprintf("[HEUREKA] %v |%s %3d %s| %13v | %15s | %10s |%s %-7s %s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusColor, param.StatusCode, resetColor,
			param.Latency,
			param.ClientIP,
			param.Request.Header.Get("x-remote-user"),
			methodColor, param.Method, resetColor,
			param.Path,
			param.ErrorMessage,
		)
	}

	logConfig := gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: logFormatter,
		SkipPaths: []string{"/"},
	})

	s.router.Use(logConfig, gin.Recovery(), ginprom.PromMiddleware(nil))
}

func (s *Server) createEndpoints() {
	s.router.GET("/", s.homeHandler)
	s.router.NoRoute(s.homeHandler)
	s.router.GET("/metrics", ginprom.PromHandler(promhttp.Handler()))
	s.router.GET("/status", s.statusHandler)
	s.router.GET("/ready", s.readyHandler)
	s.router.GET("info", s.infoHandler)

	s.graphQLAPI.CreateEndpoints(s.router)
}

func (s *Server) homeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"msg": "heureka"})
}

func (s *Server) infoHandler(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{"configuration": s.config})
}

func (s *Server) statusHandler(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"msg": "alive"})
}

func (s *Server) readyHandler(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"msg": "ready"})
}

func (s *Server) Start() {
	s.router.Run(s.config.Port)
}

func (s *Server) NonBlockingStart() {
	s.nonBlockingSrv = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.Port),
		Handler: s.router.Handler(),
	}

	util2.FirstListenThenServe(s.nonBlockingSrv, logrus.New())
}

func (s *Server) BlockingStop() {
	s.shutdownFunc()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.nonBlockingSrv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}
	if err := s.graphQLAPI.App.Shutdown(); err != nil {
		log.Fatalf("Error while shuting down Heureka App: %s", err)
	}
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		//log.Println("All goroutines exited cleanly.")
	case <-ctx.Done():
		log.Fatalf("Timeout: some goroutines did not exit in time.")
	}
}

func (s Server) App() *app.HeurekaApp {
	return s.app
}
