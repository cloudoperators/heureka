// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package graphqlapi

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/access"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph"
	"github.wdf.sap.corp/cc/heureka/internal/api/graphql/graph/resolver"
	"github.wdf.sap.corp/cc/heureka/internal/app"
)

type GraphQLAPI struct {
	Server *handler.Server
	App    app.Heureka

	auth access.Auth
}

func NewGraphQLAPI(a app.Heureka) *GraphQLAPI {
	graphQLAPI := GraphQLAPI{
		Server: handler.NewDefaultServer(graph.NewExecutableSchema(resolver.NewResolver(a))),
		App:    a,
		auth:   access.NewAuth(),
	}
	return &graphQLAPI
}

func (g *GraphQLAPI) CreateEndpoints(router *gin.Engine) {
	router.Use(g.auth.GetMiddleware())
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
