// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package ai

import (
	"github.com/cloudoperators/heureka/internal/api/ai/llm"
	graphqlapi "github.com/cloudoperators/heureka/internal/api/graphql"
	gqlmiddleware "github.com/cloudoperators/heureka/internal/api/graphql/middleware"
	"github.com/cloudoperators/heureka/internal/util"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type AIAPI struct {
	handler     *Handler
	rateLimiter *gqlmiddleware.IPRateLimiter
}

func NewAIAPI(cfg util.Config, client llm.Client) *AIAPI {
	return &AIAPI{
		handler: newHandler(cfg, client, graphqlapi.Schema()),
		rateLimiter: gqlmiddleware.NewIPRateLimiter(
			rate.Limit(cfg.AIHttpRateLimit),
			cfg.AIHttpRateBurst,
		),
	}
}

func (a *AIAPI) CreateEndpoints(router *gin.Engine) {
	ai := router.Group("/ai")
	ai.Use(a.rateLimiter.Middleware())
	{
		ai.POST("/query", a.handler.QueryHandler)
	}
}
