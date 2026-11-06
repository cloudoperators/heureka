// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudoperators/heureka/pkg/util"
	"golang.org/x/time/rate"
)

// roundTripFunc lets tests inject a custom RoundTrip implementation.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func newRequest(t *testing.T) *http.Request {
	t.Helper()

	body := []byte(`{"query":"{}"}`)

	req, err := http.NewRequest(http.MethodPost, "http://example.com/graphql", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	// http.NewRequest sets GetBody for *bytes.Reader so the body can be replayed
	return req
}

// response429 returns a 429 response. Pass retryAfterSecs >= 1 to set the
// Retry-After header to that many seconds; pass 0 or negative to omit it
// (triggering exponential-backoff fallback).
func response429(retryAfterSecs int) *http.Response {
	h := http.Header{}

	if retryAfterSecs > 0 {
		h.Set("Retry-After", strconv.Itoa(retryAfterSecs))
	}

	return &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     h,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}
}

func response200() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}
}

func TestRetryAfterRoundTripper_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	rt := &util.RetryAfterRoundTripper{
		Base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++

			return response200(), nil
		}),
		MaxRetries: 3,
	}

	req := newRequest(t)

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetryAfterRoundTripper_HonorsRetryAfterHeader(t *testing.T) {
	var calls int32

	start := time.Now()
	rt := &util.RetryAfterRoundTripper{
		Base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			n := atomic.AddInt32(&calls, 1)
			if n == 1 {
				return response429(1), nil // server says wait 1s
			}

			return response200(), nil
		}),
		MaxRetries: 3,
	}

	req := newRequest(t)

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Fatalf("expected at least ~1s sleep from Retry-After header, got %v", elapsed)
	}

	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", atomic.LoadInt32(&calls))
	}
}

func TestRetryAfterRoundTripper_ExponentialFallbackWhenNoHeader(t *testing.T) {
	var calls int32

	rt := &util.RetryAfterRoundTripper{
		Base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			n := atomic.AddInt32(&calls, 1)
			if n == 1 {
				return response429(0), nil // no header → exponential fallback (sleeps 1s)
			}

			return response200(), nil
		}),
		MaxRetries: 3,
	}

	req := newRequest(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req = req.WithContext(ctx)

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", atomic.LoadInt32(&calls))
	}
}

func TestRetryAfterRoundTripper_StopsAfterMaxRetries(t *testing.T) {
	var calls int32

	rt := &util.RetryAfterRoundTripper{
		Base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&calls, 1)

			return response429(0), nil
		}),
		MaxRetries: 2,
	}

	req := newRequest(t)

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After MaxRetries exhausted the last 429 is returned as-is
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}

	// initial attempt + 2 retries = 3 calls
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", atomic.LoadInt32(&calls))
	}
}

func TestRetryAfterRoundTripper_ContextCancellationDuringSleep(t *testing.T) {
	var calls int32

	rt := &util.RetryAfterRoundTripper{
		Base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&calls, 1)

			return response429(60), nil // 60s sleep — will be cancelled
		}),
		MaxRetries: 3,
	}

	ctx, cancel := context.WithCancel(context.Background())

	req := newRequest(t)
	req = req.WithContext(ctx)

	// Cancel after the first 429 is received
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected an error from context cancellation, got nil")
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected exactly 1 call before cancellation, got %d", atomic.LoadInt32(&calls))
	}
}

func TestRetryAfterRoundTripper_SkipsRetryWhenBodyCannotBeReplayed(t *testing.T) {
	var calls int32

	rt := &util.RetryAfterRoundTripper{
		Base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&calls, 1)

			return response429(0), nil
		}),
		MaxRetries: 3,
	}

	// Manually construct a request with a body but no GetBody — simulates a
	// non-replayable stream.
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/graphql", io.NopCloser(bytes.NewReader([]byte("{}"))))
	req.GetBody = nil // explicitly clear it

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return the 429 without retrying
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", resp.StatusCode)
	}

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected exactly 1 call (no retry), got %d", atomic.LoadInt32(&calls))
	}
}

func TestNewRateLimitedHTTPClient_WrapsWithRetryAfter(t *testing.T) {
	// Verify that the client produced by the constructor correctly retries on 429.
	var calls int32

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			return response429(0), nil
		}

		return response200(), nil
	})

	client := util.NewRateLimitedHTTPClient(rate.Limit(1000), 1000, transport)

	req := newRequest(t)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", atomic.LoadInt32(&calls))
	}
}
