// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import "net/http"

type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("X-Service-Authorization", "Bearer "+t.token)

	return t.base.RoundTrip(req)
}

func WithAuth(client *http.Client, token string) *http.Client {
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}

	wrapped := *client
	wrapped.Transport = &authTransport{base: base, token: token}

	return &wrapped
}
