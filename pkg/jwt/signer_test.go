// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package jwt_test

import (
	"testing"
	"time"

	heurekajwt "github.com/cloudoperators/heureka/pkg/jwt"
	gojwt "github.com/golang-jwt/jwt/v5"
)

const (
	testSecret  = "test-secret"
	testSubject = "scanner-test"
)

func parseToken(t *testing.T, tok, secret string) *heurekajwt.TokenClaims {
	t.Helper()

	claims := &heurekajwt.TokenClaims{}

	parsed, err := gojwt.ParseWithClaims(tok, claims, func(*gojwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if !parsed.Valid {
		t.Fatal("token not valid")
	}

	return claims
}

func TestGenerateJWT_ProducesValidHS256Token(t *testing.T) {
	tok, err := heurekajwt.GenerateJWT([]byte(testSecret), time.Minute, testSubject)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	claims := parseToken(t, tok, testSecret)

	if claims.Subject != testSubject {
		t.Errorf("Subject = %q, want %q", claims.Subject, testSubject)
	}

	if claims.Issuer != "heureka" {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, "heureka")
	}

	if claims.Version == "" {
		t.Error("Version claim missing")
	}

	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt claim missing")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt in the past")
	}
}

func TestGenerateJWT_RejectedByWrongSecret(t *testing.T) {
	tok, err := heurekajwt.GenerateJWT([]byte(testSecret), time.Minute, testSubject)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	_, err = gojwt.ParseWithClaims(tok, &heurekajwt.TokenClaims{}, func(*gojwt.Token) (any, error) {
		return []byte("wrong-secret"), nil
	})
	if err == nil {
		t.Fatal("expected error when parsing with wrong secret, got nil")
	}
}

func TestSigner_CachesWithinTTL(t *testing.T) {
	s := heurekajwt.NewSigner(testSecret, time.Hour, testSubject)

	first, err := s.Token()
	if err != nil {
		t.Fatalf("first Token: %v", err)
	}

	second, err := s.Token()
	if err != nil {
		t.Fatalf("second Token: %v", err)
	}

	if first != second {
		t.Error("expected cached token to be reused within TTL")
	}
}

func TestSigner_ReMintsWhenCacheIsStale(t *testing.T) {
	// TTL shorter than the 30s skew buffer forces re-mint on every call.
	s := heurekajwt.NewSigner(testSecret, 10*time.Second, testSubject)

	first, err := s.Token()
	if err != nil {
		t.Fatalf("first Token: %v", err)
	}

	// Small sleep so iat changes and tokens differ byte-wise.
	time.Sleep(1100 * time.Millisecond)

	second, err := s.Token()
	if err != nil {
		t.Fatalf("second Token: %v", err)
	}

	if first == second {
		t.Error("expected fresh mint when TTL is below skew buffer")
	}
}

func TestSigner_TokenValidatesAgainstVerifier(t *testing.T) {
	s := heurekajwt.NewSigner(testSecret, time.Minute, testSubject)

	tok, err := s.Token()
	if err != nil {
		t.Fatalf("Token: %v", err)
	}

	claims := parseToken(t, tok, testSecret)
	if claims.Subject != testSubject {
		t.Errorf("Subject = %q, want %q", claims.Subject, testSubject)
	}
}

func TestSigner_ConcurrentCallersShareCache(t *testing.T) {
	s := heurekajwt.NewSigner(testSecret, time.Hour, testSubject)

	const n = 50

	tokens := make([]string, n)
	errs := make([]error, n)
	done := make(chan int, n)

	for i := 0; i < n; i++ {
		go func(i int) {
			tokens[i], errs[i] = s.Token()
			done <- i
		}(i)
	}

	for i := 0; i < n; i++ {
		<-done
	}

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: %v", i, err)
		}
	}

	for i := 1; i < n; i++ {
		if tokens[i] != tokens[0] {
			t.Fatalf("goroutine %d produced a different token than goroutine 0", i)
		}
	}
}
