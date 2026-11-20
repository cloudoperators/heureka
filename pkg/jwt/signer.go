package jwt

import (
	"sync"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

const (
	skewBuffer = -30
	jwtIssuer  = "heureka"
)

type TokenClaims struct {
	Version string `json:"version"`
	gojwt.RegisteredClaims
}

func GenerateJWT(secret []byte, expireIn time.Duration, subject string) (string, error) {
	claims := TokenClaims{
		Version: "0.3.1",
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(expireIn)),
			IssuedAt:  gojwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuer,
			Subject:   subject,
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// Signer caches the current token and re-mints when it's close to expiry.
type Signer struct {
	secret  []byte
	ttl     time.Duration
	subject string

	mu      sync.Mutex
	token   string
	expires time.Time
}

func NewSigner(secret string, ttl time.Duration, subject string) *Signer {
	return &Signer{secret: []byte(secret), ttl: ttl, subject: subject}
}

func (s *Signer) Token() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if time.Now().Before(s.expires.Add(skewBuffer * time.Second)) {
		return s.token, nil
	}
	tok, err := GenerateJWT(s.secret, s.ttl, s.subject)
	if err != nil {
		return "", err
	}
	s.token = tok
	s.expires = time.Now().Add(s.ttl)
	return tok, nil
}
