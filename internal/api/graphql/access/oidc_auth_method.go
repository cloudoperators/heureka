package access

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/cloudoperators/heureka/internal/util"
)

const (
	oidcAuthMethodName string = "OidcAuthMethod"
	oidcClientSecret   string = "xxx"
)

type OidcAuthMethod struct {
	logger    Logger
	provider  *oidc.Provider
	config    *oidc.Config
	verifier  *oidc.IDTokenVerifier
	oauth2Cfg *oauth2.Config
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func NewOidcAuthMethod(l Logger, cfg *util.Config) authMethod {
	if cfg.AuthOidcClientId != "" && cfg.AuthOidcUrl != "" {
		l.Info("Initializing OIDC auth")
		ctx := context.Background()
		oidcProvider, err := oidc.NewProvider(ctx, cfg.AuthOidcUrl)
		if err != nil {
			l.Error("Could not initialize OIDC provider: ", err, " (", cfg.AuthOidcUrl, ")")
			return &NoAuthMethod{logger: l, authMethodName: oidcAuthMethodName, msg: "Could not initialize OIDC provider: " + err.Error()}
		}
		oidcConfig := &oidc.Config{
			ClientID: cfg.AuthOidcClientId,
		}

		oidcVerifier := oidcProvider.Verifier(oidcConfig)

		config := oauth2.Config{
			ClientID:     cfg.AuthOidcClientId,
			ClientSecret: oidcClientSecret,
			Endpoint:     oidcProvider.Endpoint(),
			RedirectURL:  fmt.Sprintf("http://%s:%s/callback", GetHostname(), cfg.Port),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}
		fmt.Println(config.RedirectURL)

		return &OidcAuthMethod{logger: l, provider: oidcProvider, config: oidcConfig, verifier: oidcVerifier, oauth2Cfg: &config}
	}
	return nil
}

func (oam OidcAuthMethod) AddWhitelistRoutes(r *gin.Engine) {
	r.GET("/callback", oam.callbackHandler)
}

func (oam OidcAuthMethod) Verify(c *gin.Context) error {
	authenticated := false
	rawIdToken, err := c.Cookie("id_token")
	if err == nil {
		idToken, err := oam.verifier.Verify(context.Background(), rawIdToken)
		if err == nil {
			authenticated = true       // Mark as authenticated if verification is successful
			c.Set("id_token", idToken) // Store token in context for downstream handlers
		}
	}

	if !authenticated {
		state, err := randString(16)
		if err != nil {
			return verifyError(oidcAuthMethodName, fmt.Errorf("Internal error when trying to initialize state"))
		}
		nonce, err := randString(16)
		if err != nil {
			return verifyError(oidcAuthMethodName, fmt.Errorf("Internal error when trying to initialize nonce"))
		}
		c.SetCookie("state", state, int(time.Hour.Seconds()), "", "", c.Request.TLS != nil, true)
		c.SetCookie("nonce", nonce, int(time.Hour.Seconds()), "", "", c.Request.TLS != nil, true)

		authUrl := oam.oauth2Cfg.AuthCodeURL(state, oidc.Nonce(nonce))
		c.Redirect(http.StatusFound, authUrl)
		c.Abort()
		return nil
	}
	return nil
}

func (oam OidcAuthMethod) callbackHandler(c *gin.Context) {
	state, err := c.Cookie("state")
	if err != nil {
		oam.logger.Error("state cookie not found: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "state not found"})
		return
	}
	if c.Query("state") != state {
		oam.logger.Error("state cookie does not match")
		c.JSON(http.StatusBadRequest, gin.H{"error": "state did not match"})
		return
	}

	oauth2Token, err := oam.oauth2Cfg.Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		oam.logger.Error("Token exchange error: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		oam.logger.Error("No id_token field in oauth2 token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No id_token field in oauth2 token."})
		return
	}
	idToken, err := oam.verifier.Verify(c.Request.Context(), rawIDToken)
	if err != nil {
		oam.logger.Error("Failed to verify ID Token: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify ID Token: " + err.Error()})
		return
	}

	nonce, err := c.Cookie("nonce")
	if err != nil {
		oam.logger.Error("nonce cookie not found: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "nonce not found"})
		return
	}
	if idToken.Nonce != nonce {
		oam.logger.Error("nonce cookie does not match")
		c.JSON(http.StatusBadRequest, gin.H{"error": "nonce did not match"})
		return
	}

	//oauth2Token.AccessToken = "*REDACTED*"

	resp := struct {
		OAuth2Token   *oauth2.Token
		IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
	}{oauth2Token, new(json.RawMessage)}

	c.SetCookie("id_token", rawIDToken, int(24*time.Hour.Seconds()), "/", "", c.Request.TLS != nil, true)

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		oam.logger.Error("Could not marshal authentication data: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}
