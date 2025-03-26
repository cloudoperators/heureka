// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/square/go-jose/v3"
)

type Provider struct {
	router           *gin.Engine
	server           *http.Server
	ctx              context.Context
	cancel           context.CancelFunc
	log              *logrus.Logger
	url              string
	rsaPrivateKey    *rsa.PrivateKey
	rsaPublicKey     jose.JSONWebKey
	clientId         string
	authCodeNonceMap map[string]string
}

func NewProvider(url string, enableLog bool) *Provider {
	l := logrus.New()
	if !enableLog {
		l.SetOutput(ioutil.Discard)
	}
	gin.DefaultWriter = l.Writer()
	oidcProvider := Provider{
		log:              l,
		url:              url,
		router:           gin.New(),
		clientId:         "mock-client-id",
		authCodeNonceMap: make(map[string]string),
	}
	oidcProvider.initRsa()
	oidcProvider.addRoutes()
	return &oidcProvider
}

func (p *Provider) Start() {
	p.ctx, p.cancel = context.WithCancel(context.Background())

	p.server = &http.Server{Handler: p.router.Handler()}

	serverAddr := "'default'"
	re := regexp.MustCompile(`:[0-9]+`)
	matches := re.FindAllString(p.url, -1)
	if len(matches) > 0 {
		p.server.Addr = matches[0]
		serverAddr = p.server.Addr
	}

	go func() {
		p.log.Printf("Starting OIDC provider on %s", serverAddr)
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.log.Fatalf("Server failed: %v", err)
		}
	}()
}

func (p *Provider) StartForeground() {
	p.server = &http.Server{Handler: p.router.Handler()}

	serverAddr := "'default'"
	re := regexp.MustCompile(`:[0-9]+`)
	matches := re.FindAllString(p.url, -1)
	if len(matches) > 0 {
		p.server.Addr = matches[0]
		serverAddr = p.server.Addr
	}

	p.log.Printf("Starting server on %s", serverAddr)
	if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		p.log.Fatalf("Server failed: %v", err)
	}
}

func (p *Provider) GetRsaPrivateKey() *rsa.PrivateKey {
	return p.rsaPrivateKey
}

func (p *Provider) Stop() {
	p.log.Println("Stopping server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer shutdownCancel()
	if err := p.server.Shutdown(shutdownCtx); err != nil {
		p.log.Fatalf("Failed to gracefully shut down server: %v", err)
	}
	p.cancel()
	p.log.Println("Server stopped successfully")
}

func (p *Provider) addRoutes() {
	p.router.GET("/.well-known/openid-configuration", p.openidConfigurationHandler)
	p.router.GET("/auth", p.authHandler)
	p.router.GET("/jwks", p.jwksHandler)
	p.router.POST("/token", p.tokenHandler)
}

func (p *Provider) initRsa() {
	var err error
	p.rsaPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		p.log.Fatalf("Failed to generate RSA private key: %v", err)
	}
	p.rsaPublicKey = jose.JSONWebKey{
		Key:       p.rsaPrivateKey.Public(),
		KeyID:     "mock-key-id",
		Algorithm: "RS256",
		Use:       "sig",
	}
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (p *Provider) openidConfigurationHandler(c *gin.Context) {
	p.log.Println("/.well-known/openid-configuration")
	c.Header("Content-Type", "application/json")
	metadata := map[string]interface{}{
		"issuer":                                p.url,
		"authorization_endpoint":                fmt.Sprintf("%s/auth", p.url),
		"token_endpoint":                        fmt.Sprintf("%s/token", p.url),
		"jwks_uri":                              fmt.Sprintf("%s/jwks", p.url),
		"id_token_signing_alg_values_supported": []string{"RS256"},
	}
	c.JSON(http.StatusOK, metadata)
}

func (p *Provider) authHandler(c *gin.Context) {
	p.log.Println("/auth")
	state := c.Query("state")
	p.log.Println("STATE: ", state)
	nonce := c.Query("nonce")
	p.log.Println("NONCE: ", nonce)
	if nonce == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nonce is required"})
		return
	}

	// Simulate generating a unique authorization code
	authCode, err := randString(16)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not generate auth code"})
		return
	}
	p.authCodeNonceMap[authCode] = nonce

	redirectUri := c.Query("redirect_uri")
	if redirectUri != "" {
		p.log.Println(redirectUri + "?code=" + authCode + "&state=" + state)
		c.Redirect(http.StatusFound, redirectUri+"?code="+authCode+"&state="+state)
	}
}

func (p *Provider) jwksHandler(c *gin.Context) {
	p.log.Println("/jwks")
	keySet := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{p.rsaPublicKey},
	}
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, keySet)
}

func (p *Provider) tokenHandler(c *gin.Context) {
	p.log.Println("/token")

	code := c.PostForm("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}
	nonce, ok := p.authCodeNonceMap[code]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code"})
		return
	}

	// Generate the ID Token with the correct nonce
	idToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":   "1234567890",
		"name":  "John Doe",
		"email": "johndoe@example.com",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
		"aud":   p.clientId,
		"iss":   p.url,
		"nonce": nonce,
	})

	tokenString, err := idToken.SignedString(p.rsaPrivateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	c.Header("Content-Type", "application/json")

	c.JSON(http.StatusOK, map[string]string{
		"access_token": "mock-access-token",
		"id_token":     tokenString,
		"token_type":   "Bearer",
	})
}
