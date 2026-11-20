// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	heurekajwt "github.com/cloudoperators/heureka/pkg/jwt"
	"github.com/cloudoperators/heureka/pkg/util"
)

func okResponse() *http.Response {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
}

func TestWithAuth_SetsServiceAuthorizationHeader(t *testing.T) {
	var seenHeader string
	base := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		seenHeader = r.Header.Get("X-Service-Authorization")
		return okResponse(), nil
	})
	signer := heurekajwt.NewSigner("secret", time.Minute, "test-caller")
	client := util.WithAuth(&http.Client{Transport: base}, signer)

	req, _ := http.NewRequest(http.MethodPost, "http://example.test/q", nil)
	if _, err := client.Do(req); err != nil {
		t.Fatalf("Do: %v", err)
	}

	if !strings.HasPrefix(seenHeader, "Bearer ey") {
		t.Errorf("X-Service-Authorization = %q, want prefix %q", seenHeader, "Bearer ey")
	}
}

func TestWithAuth_DoesNotMutateOriginalRequest(t *testing.T) {
	base := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return okResponse(), nil
	})
	signer := heurekajwt.NewSigner("secret", time.Minute, "test-caller")
	client := util.WithAuth(&http.Client{Transport: base}, signer)

	req, _ := http.NewRequest(http.MethodPost, "http://example.test/q", nil)
	if _, err := client.Do(req); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got := req.Header.Get("X-Service-Authorization"); got != "" {
		t.Errorf("original request mutated: X-Service-Authorization = %q", got)
	}
}

func TestWithAuth_PropagatesTransportErrors(t *testing.T) {
	wantErr := errors.New("error")
	base := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, wantErr
	})
	signer := heurekajwt.NewSigner("secret", time.Minute, "test-caller")
	client := util.WithAuth(&http.Client{Transport: base}, signer)

	req, _ := http.NewRequest(http.MethodPost, "http://example.test/q", nil)
	_, err := client.Do(req)
	if err == nil || !strings.Contains(err.Error(), "error") {
		t.Errorf("err = %v, want containing %q", err, "error")
	}
}

func TestWithAuth_PreservesExistingClientFields(t *testing.T) {
	base := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return okResponse(), nil
	})
	original := &http.Client{
		Transport: base,
		Timeout:   7 * time.Second,
	}
	signer := heurekajwt.NewSigner("secret", time.Minute, "test-caller")
	wrapped := util.WithAuth(original, signer)

	if wrapped.Timeout != original.Timeout {
		t.Errorf("Timeout lost: got %v, want %v", wrapped.Timeout, original.Timeout)
	}
	if wrapped == original {
		t.Error("WithAuth returned the same pointer; should be a copy to avoid caller mutation")
	}
	if wrapped.Transport == original.Transport {
		t.Error("Transport not wrapped; got original back")
	}
}

func TestWithAuth_UsesDefaultTransportWhenNil(t *testing.T) {
	signer := heurekajwt.NewSigner("secret", time.Minute, "test-caller")
	// Client with no explicit Transport — WithAuth should fall back to http.DefaultTransport.
	client := util.WithAuth(&http.Client{}, signer)

	if client.Transport == nil {
		t.Fatal("wrapped client has nil Transport")
	}
}
