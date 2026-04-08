// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net/http"

	"golang.org/x/time/rate"
)

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

func NewRateLimitedHTTPClient(limit rate.Limit, burst int, base http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: &RateLimitedRoundTripper{
			Base:    base,
			Limiter: rate.NewLimiter(limit, burst),
		},
	}
}
