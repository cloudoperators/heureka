// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter manages rate limiters for each client IP address.
type IPRateLimiter struct {
	ips    map[string]*rate.Limiter
	mu     *sync.RWMutex
	rate   rate.Limit
	burst  int
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewIPRateLimiter creates and initializes a new rate limiter.
// Parameters:
//   - r: rate.Limit - requests per second (e.g., rate.Limit(50) for 50 RPS)
//   - b: int - burst size allowing short spikes (e.g., 100)
//
// Returns a *IPRateLimiter that automatically cleans up stale entries every minute.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips:    make(map[string]*rate.Limiter),
		mu:     &sync.RWMutex{},
		rate:   r,
		burst:  b,
		stopCh: make(chan struct{}),
	}

	i.wg.Add(1)
	go i.cleanupRoutine()

	return i
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.ips[ip] = limiter
	}

	return limiter
}

func (i *IPRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := i.getLimiter(ip)

		if !limiter.Allow() {
			c.Header("Retry-After", "1")
			c.Header("X-RateLimit-Limit", "")
			c.Header("X-RateLimit-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			return
		}

		c.Next()
	}
}

func (i *IPRateLimiter) cleanupRoutine() {
	defer i.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.cleanup()
		case <-i.stopCh:
			return
		}
	}
}

func (i *IPRateLimiter) cleanup() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.ips = make(map[string]*rate.Limiter)
}

func (i *IPRateLimiter) Stop() {
	close(i.stopCh)
	i.wg.Wait()
}
