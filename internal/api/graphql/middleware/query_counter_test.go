// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudoperators/heureka/internal/database/querycounter"
	"github.com/gin-gonic/gin"
)

func TestQueryCounter_InitializesCounter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedCount int

	router := gin.New()
	router.Use(QueryCounter())
	router.POST("/query", func(c *gin.Context) {
		// Simulate DB queries
		querycounter.Increment(c.Request.Context())
		querycounter.Increment(c.Request.Context())
		querycounter.Increment(c.Request.Context())
		capturedCount = querycounter.GetQueryCount(c.Request.Context())
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedCount != 3 {
		t.Errorf("expected query count 3, got %d", capturedCount)
	}
}

func TestQueryCounter_IsolatesRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	counts := make([]int, 2)
	callIndex := 0

	router := gin.New()
	router.Use(QueryCounter())
	router.POST("/query", func(c *gin.Context) {
		idx := callIndex
		callIndex++
		// First request: 5 queries, second: 2 queries
		n := 5
		if idx == 1 {
			n = 2
		}

		for range n {
			querycounter.Increment(c.Request.Context())
		}

		counts[idx] = querycounter.GetQueryCount(c.Request.Context())
		c.Status(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodPost, "/query", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/query", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if counts[0] != 5 {
		t.Errorf("first request: expected 5, got %d", counts[0])
	}

	if counts[1] != 2 {
		t.Errorf("second request: expected 2, got %d", counts[1])
	}
}
