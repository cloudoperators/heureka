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

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	ips        map[string]*clientLimiter
	mu         *sync.RWMutex
	rate       rate.Limit
	burst      int
	stopCh     chan struct{}
	wg         sync.WaitGroup
	staleAfter time.Duration
}

const cleanupInterval = 1 * time.Minute
const staleAfter = 5 * time.Minute

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips:        make(map[string]*clientLimiter),
		mu:         &sync.RWMutex{},
		rate:       r,
		burst:      b,
		stopCh:     make(chan struct{}),
		staleAfter: staleAfter,
	}

	i.wg.Add(1)
	go i.cleanupRoutine()

	return i
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	client, exists := i.ips[ip]
	if !exists {
		client = &clientLimiter{
			limiter:  rate.NewLimiter(i.rate, i.burst),
			lastSeen: time.Now(),
		}
		i.ips[ip] = client
	} else {
		client.lastSeen = time.Now()
	}

	return client.limiter
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

	ticker := time.NewTicker(cleanupInterval)
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

	for ip, client := range i.ips {
		if time.Since(client.lastSeen) > i.staleAfter {
			delete(i.ips, ip)
		}
	}
}

func (i *IPRateLimiter) Stop() {
	close(i.stopCh)
	i.wg.Wait()
}

func (i *IPRateLimiter) SetRate(r rate.Limit) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.rate = r
	for _, client := range i.ips {
		client.limiter.SetLimit(r)
	}
}

func (i *IPRateLimiter) SetBurst(b int) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.burst = b
	for _, client := range i.ips {
		client.limiter.SetBurst(b)
	}
}
