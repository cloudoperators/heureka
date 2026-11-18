// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const defaultMaxRetries = 3

type RateLimitedRoundTripper struct {
	Base    http.RoundTripper
	Limiter *rate.Limiter
}

func (r *RateLimitedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	err := r.Limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}

	base := r.Base
	if base == nil {
		base = http.DefaultTransport
	}

	return base.RoundTrip(req)
}

// RetryAfterRoundTripper retries requests that receive HTTP 429 Too Many Requests.
// It reads the Retry-After response header to determine how long to wait before
// retrying; if the header is absent it falls back to exponential backoff.
// Retries are skipped when the request body cannot be replayed (GetBody is nil).
type RetryAfterRoundTripper struct {
	Base       http.RoundTripper
	MaxRetries int
}

func (r *RetryAfterRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	maxRetries := r.MaxRetries
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}

	for attempt := 0; ; attempt++ {
		resp, err := r.Base.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests || attempt >= maxRetries {
			return resp, nil
		}

		// if the body can't be replayed, return the 429 intact so the caller
		// can still read the error payload
		if req.Body != nil && req.GetBody == nil {
			return resp, nil
		}

		// drain and close the 429 body so the connection can be reused
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		delay := parseRetryAfter(resp.Header.Get("Retry-After"))
		if delay == 0 {
			// exponential fallback: 1s, 2s, 4s, …
			delay = time.Duration(1<<uint(attempt)) * time.Second
		}

		select {
		case <-req.Context().Done():
			return nil, fmt.Errorf("rate-limit retry cancelled: %w", req.Context().Err())
		case <-time.After(delay):
		}

		if req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("rate-limit retry: rebuild request body: %w", err)
			}

			req.Body = body
		}
	}
}

// parseRetryAfter parses an integer-seconds Retry-After header value (RFC 7231 §7.1.3).
func parseRetryAfter(s string) time.Duration {
	secs, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || secs <= 0 {
		return 0
	}

	return time.Duration(secs) * time.Second
}

func NewRateLimitedHTTPClient(limit rate.Limit, burst int, base http.RoundTripper) *http.Client {
	rateLimited := &RateLimitedRoundTripper{
		Base:    base,
		Limiter: rate.NewLimiter(limit, burst),
	}

	return &http.Client{
		Transport: &RetryAfterRoundTripper{
			Base:       rateLimited,
			MaxRetries: defaultMaxRetries,
		},
	}
}
