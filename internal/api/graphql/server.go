// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package graphqlapi

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/cloudoperators/heureka/internal/api/graphql/access/middleware"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/resolver"
	"github.com/cloudoperators/heureka/internal/app"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/gin-gonic/gin"
)

type GraphQLAPI struct {
	Server *handler.Server
	App    app.Heureka

	auth *middleware.Auth
}

func NewGraphQLAPI(a app.Heureka, cfg util.Config) *GraphQLAPI {
	server := handler.NewDefaultServer(graph.NewExecutableSchema(resolver.NewResolver(a)))

	// Set our custom error presenter
	// Check out https://gqlgen.com/reference/errors/
	server.SetErrorPresenter(graph.ErrorPresenter)

	graphQLAPI := GraphQLAPI{
		Server: server,
		App:    a,
		auth:   middleware.NewAuth(&cfg, true),
	}
	return &graphQLAPI
}

func (g *GraphQLAPI) CreateEndpoints(router *gin.Engine) {
	router.Use(g.auth.Middleware())
	router.GET("/playground", g.playgroundHandler())
	router.POST("/query", g.graphqlHandler())
}

func (g *GraphQLAPI) graphqlHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		g.Server.ServeHTTP(c.Writer, c.Request)
	}
}

func (g *GraphQLAPI) playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
