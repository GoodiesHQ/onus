package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/goodieshq/onus/internal/util"
	"golang.org/x/oauth2"
)

// ProviderOIDC implements the AuthProvider interface for OIDC authentication
type ProviderOIDC struct {
	issuserURL     string
	provider       *oidc.Provider
	verifier       *oidc.IDTokenVerifier
	config         *oauth2.Config
	sm             *scs.SessionManager
	enableUserInfo bool
}

// ConfigOIDC holds the configuration for OIDC provider
type ConfigOIDC struct {
	IssuerURL      string
	ClientID       string
	ClientSecret   string
	RedirectURL    string
	Scopes         []string
	EnableUserInfo bool
}

// Set of standard OIDC claims
type claimsOIDC struct {
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	UPN               string `json:"upn"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
}

// merge merges non-empty fields from other claims into c, does not overwrite existing non-empty fields
func (c *claimsOIDC) merge(other *claimsOIDC) {
	if c == nil || other == nil {
		return
	}

	if c.Email == "" {
		c.Email = other.Email
	}
	if c.PreferredUsername == "" {
		c.PreferredUsername = other.PreferredUsername
	}
	if c.UPN == "" {
		c.UPN = other.UPN
	}
	if c.Name == "" {
		c.Name = other.Name
	}
	if c.GivenName == "" {
		c.GivenName = other.GivenName
	}
	if c.FamilyName == "" {
		c.FamilyName = other.FamilyName
	}
}

var defaultScopesOIDC = []string{oidc.ScopeOpenID, "profile", "email"}

// NewProviderOIDC creates a new OIDC authentication provider
func NewProviderOIDC(ctx context.Context, cfg *ConfigOIDC, sm *scs.SessionManager) (*ProviderOIDC, error) {
	// Validate required config fields
	if cfg == nil {
		return nil, fmt.Errorf("oidc config cannot be nil")
	}

	if sm == nil {
		return nil, fmt.Errorf("session manager cannot be nil")
	}

	if cfg.IssuerURL == "" {
		return nil, fmt.Errorf("oidc issuer URL cannot be empty")
	}

	if cfg.ClientID == "" {
		return nil, fmt.Errorf("oidc client ID cannot be empty")
	}

	if cfg.ClientSecret == "" {
		return nil, fmt.Errorf("oidc client secret cannot be empty")
	}

	if cfg.RedirectURL == "" {
		return nil, fmt.Errorf("oidc redirect URL cannot be empty")
	}

	// Set default scopes if not provided
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = defaultScopesOIDC
	}

	// Create OIDC provider
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new oidc provider: %w", err)
	}

	// Create ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	// Create OAuth2 config
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURL,
		Scopes:       scopes,
	}

	// Return the OIDC provider instance
	return &ProviderOIDC{
		issuserURL:     cfg.IssuerURL,
		provider:       provider,
		verifier:       verifier,
		config:         oauthCfg,
		sm:             sm,
		enableUserInfo: cfg.EnableUserInfo,
	}, nil
}

// StartAuth initiates the OIDC authentication flow
func (idp *ProviderOIDC) StartAuth(w http.ResponseWriter, r *http.Request) {
	// Generate a random state and nonce

	state, err := util.GenerateCSRFToken()
	if err != nil {
		http.Error(w, "Failed to generate an OIDC state", http.StatusInternalServerError)
		return
	}

	nonce, err := util.GenerateCSRFToken()
	if err != nil {
		http.Error(w, "Failed to generate an OIDC nonce", http.StatusInternalServerError)
		return
	}

	// Store state and nonce in the session
	idp.sm.Put(r.Context(), "oidc_state", state)
	idp.sm.Put(r.Context(), "oidc_nonce", nonce)

	// Redirect to the OIDC provider's auth URL
	authURL := idp.config.AuthCodeURL(state, oidc.Nonce(nonce))
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleCallback processes the OIDC authentication callback
func (idp *ProviderOIDC) HandleCallback(w http.ResponseWriter, r *http.Request) (*Principal, error) {
	ctx := r.Context()

	// Verify state parameter
	state := r.URL.Query().Get("state")
	if state == "" {
		return nil, fmt.Errorf("missing oidc state")
	}

	// Retrieve expected state from session
	stateExpected := idp.sm.GetString(ctx, "oidc_state")
	if state != stateExpected {
		return nil, fmt.Errorf("invalid oidc state")
	}

	// Verify code parameter
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, fmt.Errorf("missing oidc code")
	}

	// Exchange the code for an OAuth2 token
	token, err := idp.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange oidc code for token: %w", err)
	}

	// Extract the ID token from OAuth2 token
	tokenRaw, ok := token.Extra("id_token").(string)
	if !ok || tokenRaw == "" {
		return nil, fmt.Errorf("missing id_token in oidc token response")
	}

	// Verify the ID token
	tokenID, err := idp.verifier.Verify(ctx, tokenRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to verify oidc id_token: %w", err)
	}

	// Verify nonce
	nonceExpected := idp.sm.GetString(ctx, "oidc_nonce")
	if nonceExpected == "" || tokenID.Nonce != nonceExpected {
		return nil, fmt.Errorf("invalid oidc nonce")
	}

	// Remove state and nonce from the session
	idp.sm.Remove(ctx, "oidc_state")
	idp.sm.Remove(ctx, "oidc_nonce")

	// Extract claims from the ID token
	var claims claimsOIDC
	if err := tokenID.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims from oidc id_token: %w", err)
	}

	// Optionally fetch userinfo claims
	if idp.enableUserInfo {
		userInfo, err := idp.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch oidc userinfo: %w", err)
		}

		// Extract claims from userinfo
		var claimsUserInfo claimsOIDC
		if err := userInfo.Claims(&claimsUserInfo); err != nil {
			return nil, fmt.Errorf("failed to extract claims from oidc userinfo: %w", err)
		}

		// Merge userinfo claims into ID token claims
		claims.merge(&claimsUserInfo)
	}

	// Determine email from claims
	email := util.FirstNonEmpty(
		claims.Email,
		claims.UPN,
		claims.PreferredUsername,
	)
	if email == "" || !util.IsValidEmail(email) {
		return nil, fmt.Errorf("oidc claims do not contain a valid email")
	}

	// Determine name from claims
	name := util.FirstNonEmpty(
		claims.Name,
		strings.TrimSpace(claims.GivenName+" "+claims.FamilyName),
	)
	if name == "" {
		name = email
	}

	return &Principal{
		Subject: tokenID.Subject,
		Issuer:  tokenID.Issuer,
		Email:   email,
		Name:    name,
	}, nil
}
