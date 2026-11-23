// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net/http"

	"github.com/cloudoperators/heureka/pkg/jwt"
)

type authTransport struct {
	base   http.RoundTripper
	signer *jwt.Signer
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.signer.Token()
	if err != nil {
		return nil, err
	}

	req = req.Clone(req.Context())
	req.Header.Set("X-Service-Authorization", "Bearer "+token)

	return t.base.RoundTrip(req)
}

func WithAuth(client *http.Client, signer *jwt.Signer) *http.Client {
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}

	wrapped := *client
	wrapped.Transport = &authTransport{base: base, signer: signer}

	return &wrapped
}
