package util

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/square/go-jose/v3"
)

type OidcProvider struct {
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

func NewOidcProvider(url string) *OidcProvider {
	oidcProvider := OidcProvider{
		log:              logrus.New(),
		url:              url,
		router:           gin.New(),
		clientId:         "mock-client-id",
		authCodeNonceMap: make(map[string]string),
	}
	oidcProvider.initRsa()
	oidcProvider.addRoutes()
	return &oidcProvider
}

func (op *OidcProvider) Start() {
	op.ctx, op.cancel = context.WithCancel(context.Background())
	op.server = &http.Server{Handler: op.router.Handler()}

	serverAddr := "'default"
	re := regexp.MustCompile(`:[0-9]+`)
	matches := re.FindAllString(op.url, -1)
	if len(matches) > 0 {
		op.server.Addr = matches[0]
		serverAddr = op.server.Addr
	}

	go func() {
		op.log.Printf("Starting OIDC provider on %s\n", serverAddr)
		if err := op.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			op.log.Fatalf("Server failed: %v", err)
		}
	}()
}

func (op *OidcProvider) StartForeground() {
	op.server = &http.Server{Handler: op.router.Handler()}

	serverAddr := "'default"
	re := regexp.MustCompile(`:[0-9]+`)
	matches := re.FindAllString(op.url, -1)
	if len(matches) > 0 {
		op.server.Addr = matches[0]
		serverAddr = op.server.Addr
	}

	op.log.Printf("Starting server on %s\n", serverAddr)
	if err := op.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		op.log.Fatalf("Server failed: %v", err)
	}
}

func (op *OidcProvider) GetRsaPrivateKey() *rsa.PrivateKey {
	return op.rsaPrivateKey
}

func (op *OidcProvider) Stop() {
	op.log.Println("Stopping server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(op.ctx, 5*time.Second)
	defer shutdownCancel()
	if err := op.server.Shutdown(shutdownCtx); err != nil {
		op.log.Fatalf("Failed to gracefully shut down server: %v", err)
	}
	op.cancel()
	op.log.Println("Server stopped successfully")
}

func (op *OidcProvider) addRoutes() {
	op.router.GET("/.well-known/openid-configuration", op.openidConfigurationHandler)
	op.router.GET("/auth", op.authHandler)
	op.router.GET("/jwks", op.jwksHandler)
	op.router.POST("/token", op.tokenHandler)
}

func (op *OidcProvider) initRsa() {
	var err error
	op.rsaPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		op.log.Fatalf("Failed to generate RSA private key: %v", err)
	}
	op.rsaPublicKey = jose.JSONWebKey{
		Key:       op.rsaPrivateKey.Public(),
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

func (op *OidcProvider) openidConfigurationHandler(c *gin.Context) {
	op.log.Println("/.well-known/openid-configuration")
	c.Header("Content-Type", "application/json")
	metadata := map[string]interface{}{
		"issuer":                                op.url,
		"authorization_endpoint":                fmt.Sprintf("%s/auth", op.url),
		"token_endpoint":                        fmt.Sprintf("%s/token", op.url),
		"jwks_uri":                              fmt.Sprintf("%s/jwks", op.url),
		"id_token_signing_alg_values_supported": []string{"RS256"},
	}
	c.JSON(http.StatusOK, metadata)
}

func (op *OidcProvider) authHandler(c *gin.Context) {
	op.log.Println("/auth")
	state := c.Query("state")
	op.log.Println("STATE: ", state)
	nonce := c.Query("nonce")
	op.log.Println("NONCE: ", nonce)
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
	op.authCodeNonceMap[authCode] = nonce

	redirectUri := c.Query("redirect_uri")
	if redirectUri != "" {
		op.log.Println(redirectUri + "?code=" + authCode + "&state=" + state)
		c.Redirect(http.StatusFound, redirectUri+"?code="+authCode+"&state="+state)
	}
}

func (op *OidcProvider) jwksHandler(c *gin.Context) {
	op.log.Println("/jwks")
	keySet := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{op.rsaPublicKey},
	}
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, keySet)
}

func (op *OidcProvider) tokenHandler(c *gin.Context) {
	op.log.Println("/token")

	code := c.PostForm("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}
	nonce, ok := op.authCodeNonceMap[code]
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
		"aud":   op.clientId,
		"iss":   op.url,
		"nonce": nonce,
	})

	tokenString, err := idToken.SignedString(op.rsaPrivateKey)
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
