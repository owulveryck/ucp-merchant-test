package a2aclient

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenCache caches OAuth tokens per base URL.
type TokenCache struct {
	mu      sync.Mutex
	entries map[string]*tokenEntry
}

type tokenEntry struct {
	accessToken string
	expiresAt   time.Time
}

// NewTokenCache creates a new token cache.
func NewTokenCache() *TokenCache {
	return &TokenCache{entries: make(map[string]*tokenEntry)}
}

// Get returns a valid cached token for the given base URL, or empty string.
func (tc *TokenCache) Get(baseURL string) string {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	e, ok := tc.entries[baseURL]
	if !ok || time.Now().After(e.expiresAt) {
		return ""
	}
	return e.accessToken
}

// Set stores a token for a base URL.
func (tc *TokenCache) Set(baseURL, token string, expiresIn int) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.entries[baseURL] = &tokenEntry{
		accessToken: token,
		expiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
}

// ObtainToken performs the OAuth2 + PKCE flow against the merchant's OAuth server.
func ObtainToken(httpClient *http.Client, baseURL, username, country string) (string, int, error) {
	// Generate PKCE code verifier and challenge
	verifierBytes := make([]byte, 32)
	if _, err := rand.Read(verifierBytes); err != nil {
		return "", 0, fmt.Errorf("generate verifier: %w", err)
	}
	codeVerifier := hex.EncodeToString(verifierBytes)
	h := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(h[:])

	// Step 1: Authorize
	authData := url.Values{
		"response_type":         {"code"},
		"client_id":             {"agentflowui"},
		"redirect_uri":          {"http://localhost:0/callback"},
		"state":                 {"demo-state"},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
		"username":              {username},
		"country":               {country},
	}

	noRedirect := &http.Client{
		Transport: httpClient.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	authResp, err := noRedirect.PostForm(baseURL+"/oauth2/authorize", authData)
	if err != nil {
		return "", 0, fmt.Errorf("authorize: %w", err)
	}
	defer authResp.Body.Close()

	if authResp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(authResp.Body)
		return "", 0, fmt.Errorf("authorize: expected 302, got %d: %s", authResp.StatusCode, body)
	}

	loc := authResp.Header.Get("Location")
	locURL, err := url.Parse(loc)
	if err != nil {
		return "", 0, fmt.Errorf("parse redirect: %w", err)
	}
	code := locURL.Query().Get("code")
	if code == "" {
		return "", 0, fmt.Errorf("no code in redirect: %s", loc)
	}

	// Step 2: Token exchange
	tokenData := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {"http://localhost:0/callback"},
		"code_verifier": {codeVerifier},
	}

	tokenReq, err := http.NewRequest("POST", baseURL+"/oauth2/token",
		strings.NewReader(tokenData.Encode()))
	if err != nil {
		return "", 0, fmt.Errorf("token request: %w", err)
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.SetBasicAuth("agentflowui", "super-secret-client-key")

	tokenResp, err := httpClient.Do(tokenReq)
	if err != nil {
		return "", 0, fmt.Errorf("token exchange: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(tokenResp.Body)
		return "", 0, fmt.Errorf("token: status %d: %s", tokenResp.StatusCode, body)
	}

	var tokenResult struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
		return "", 0, fmt.Errorf("decode token: %w", err)
	}
	return tokenResult.AccessToken, tokenResult.ExpiresIn, nil
}
