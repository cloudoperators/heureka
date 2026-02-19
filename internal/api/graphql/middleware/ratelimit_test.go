// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func TestIPRateLimiter_AllowsRequestsUnderLimit(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 5)
	defer limiter.Stop()

	ip := "192.168.1.1"
	l := limiter.getLimiter(ip)

	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Errorf("Request %d should be allowed within burst", i+1)
		}
	}
}

func TestIPRateLimiter_DeniesRequestsOverLimit(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 1)
	defer limiter.Stop()

	ip := "192.168.1.1"
	l := limiter.getLimiter(ip)

	if !l.Allow() {
		t.Fatal("First request should be allowed")
	}

	if l.Allow() {
		t.Error("Second request should be denied (burst exhausted)")
	}
}

func TestIPRateLimiter_IsolatesLimitersPerIP(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 1)
	defer limiter.Stop()

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	l1 := limiter.getLimiter(ip1)
	l2 := limiter.getLimiter(ip2)

	if !l1.Allow() {
		t.Fatal("IP1's first request should be allowed")
	}
	if l1.Allow() {
		t.Fatal("IP1's second request should be denied")
	}

	if !l2.Allow() {
		t.Error("IP2's first request should be allowed (independent quota)")
	}
}

func TestIPRateLimiter_CleanupRemovesStaleEntries(t *testing.T) {
	testStaleAfter := 100 * time.Millisecond
	limiter := NewIPRateLimiter(rate.Limit(10), 1)
	limiter.staleAfter = testStaleAfter
	defer limiter.Stop()

	for i := 0; i < 100; i++ {
		ip := string(rune(i))
		limiter.getLimiter(ip)
	}

	limiter.mu.RLock()
	initialCount := len(limiter.ips)
	limiter.mu.RUnlock()

	if initialCount != 100 {
		t.Fatalf("Expected 100 limiters initially, got %d", initialCount)
	}

	limiter.mu.Lock()
	for i := 0; i < 50; i++ {
		ip := string(rune(i))
		if client, ok := limiter.ips[ip]; ok {
			client.lastSeen = time.Now().Add(-testStaleAfter * 2)
		}
	}
	limiter.mu.Unlock()

	limiter.mu.Lock()
	for i := 50; i < 100; i++ {
		ip := string(rune(i))
		if client, ok := limiter.ips[ip]; ok {
			client.lastSeen = time.Now()
		}
	}
	limiter.mu.Unlock()

	limiter.cleanup()

	limiter.mu.RLock()
	afterCleanupCount := len(limiter.ips)
	limiter.mu.RUnlock()

	if afterCleanupCount != 50 {
		t.Errorf("Expected 50 limiters to remain after cleanup, but %d remain", afterCleanupCount)
	}

	limiter.mu.RLock()
	_, staleExists := limiter.ips[string(rune(0))]
	_, freshExists := limiter.ips[string(rune(50))]
	limiter.mu.RUnlock()

	if staleExists {
		t.Error("Expected stale IP (index 0) to be removed, but it still exists")
	}
	if !freshExists {
		t.Error("Expected fresh IP (index 50) to remain, but it was removed")
	}
}

func TestIPRateLimiter_MiddlewareReturns429(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(1), 1)
	defer limiter.Stop()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(limiter.Middleware())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "192.168.1.1")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request should succeed, got %d", w1.Code)
	}

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.1")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request should get 429, got %d", w2.Code)
	}

	if w2.Header().Get("Retry-After") == "" {
		t.Error("Expected Retry-After header to be set")
	}
}

func TestIPRateLimiter_BurstBehavior(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(10), 5)
	defer limiter.Stop()

	ip := "192.168.1.1"
	l := limiter.getLimiter(ip)

	for i := 0; i < 5; i++ {
		if !l.Allow() {
			t.Fatalf("Should allow up to burst size of 5")
		}
	}

	if l.Allow() {
		t.Error("Sixth request should fail (rate not satisfied)")
	}

	time.Sleep(150 * time.Millisecond)

	if !l.Allow() {
		t.Error("After refill, should allow one more request")
	}
}
