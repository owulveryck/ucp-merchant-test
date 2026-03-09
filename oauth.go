package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// OAuth2 server for UCP identity linking.
// Implements a minimal Authorization Code + PKCE flow for public clients.

// In-memory stores for OAuth state.
var (
	oauthMu       sync.Mutex
	authCodes     = map[string]*authCode{}     // code -> authCode
	accessTokens  = map[string]*tokenEntry{}   // token -> tokenEntry
	refreshTokens = map[string]*refreshEntry{} // token -> refreshEntry
)

type authCode struct {
	Code          string
	UserID        string
	Country       string
	CodeChallenge string
	RedirectURI   string
	ExpiresAt     time.Time
}

type tokenEntry struct {
	Token     string
	UserID    string
	Country   string
	ExpiresAt time.Time
}

type refreshEntry struct {
	Token   string
	UserID  string
	Country string
	Created time.Time
}

const (
	oauthClientID       = "agentflowui"
	accessTokenDuration = 1 * time.Hour
	authCodeDuration    = 5 * time.Minute
)

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// handleOAuthMetadata serves RFC 8414 OAuth Authorization Server Metadata.
func handleOAuthMetadata(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	base := fmt.Sprintf("%s://localhost:%d", scheme(), listenPort)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"issuer":                                base,
		"authorization_endpoint":                base + "/oauth2/authorize",
		"token_endpoint":                        base + "/oauth2/token",
		"revocation_endpoint":                   base + "/oauth2/revoke",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"token_endpoint_auth_methods_supported": []string{"none"},
		"code_challenge_methods_supported":      []string{"S256"},
	})
}

// handleOAuthAuthorize shows a minimal login/consent page.
func handleOAuthAuthorize(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")

	if clientID != oauthClientID {
		http.Error(w, "Invalid client_id", http.StatusBadRequest)
		return
	}
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		http.Error(w, "PKCE S256 required", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		// Show login page
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, oauthLoginHTML, merchantName, redirectURI, state, codeChallenge)
		return
	}

	// POST: process login form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad form data", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	country := strings.TrimSpace(r.FormValue("country"))
	redirectURIForm := r.FormValue("redirect_uri")
	stateForm := r.FormValue("state")
	codeChallengeForm := r.FormValue("code_challenge")

	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Generate authorization code
	code := randomHex(16)
	oauthMu.Lock()
	authCodes[code] = &authCode{
		Code:          code,
		UserID:        username,
		Country:       country,
		CodeChallenge: codeChallengeForm,
		RedirectURI:   redirectURIForm,
		ExpiresAt:     time.Now().Add(authCodeDuration),
	}
	oauthMu.Unlock()

	log.Printf("OAuth: user '%s' authorized, code issued", username)

	// Redirect back with code
	sep := "?"
	if strings.Contains(redirectURIForm, "?") {
		sep = "&"
	}
	http.Redirect(w, r, fmt.Sprintf("%s%scode=%s&state=%s", redirectURIForm, sep, code, stateForm), http.StatusFound)
}

// handleOAuthToken exchanges auth code for tokens, or refreshes tokens.
func handleOAuthToken(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeOAuthError(w, "invalid_request", "Bad form data", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		handleTokenExchange(w, r)
	case "refresh_token":
		handleTokenRefresh(w, r)
	default:
		writeOAuthError(w, "unsupported_grant_type", "Unsupported grant_type", http.StatusBadRequest)
	}
}

func handleTokenExchange(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	codeVerifier := r.FormValue("code_verifier")

	if code == "" || codeVerifier == "" {
		writeOAuthError(w, "invalid_request", "Missing code or code_verifier", http.StatusBadRequest)
		return
	}

	oauthMu.Lock()
	ac, ok := authCodes[code]
	if ok {
		delete(authCodes, code) // one-time use
	}
	oauthMu.Unlock()

	if !ok {
		writeOAuthError(w, "invalid_grant", "Invalid or expired authorization code", http.StatusBadRequest)
		return
	}
	if time.Now().After(ac.ExpiresAt) {
		writeOAuthError(w, "invalid_grant", "Authorization code expired", http.StatusBadRequest)
		return
	}

	// Verify PKCE: SHA256(code_verifier) base64url-encoded must match code_challenge
	if !verifyPKCE(codeVerifier, ac.CodeChallenge) {
		writeOAuthError(w, "invalid_grant", "PKCE verification failed", http.StatusBadRequest)
		return
	}

	// Issue tokens
	accessToken := randomHex(32)
	refreshToken := randomHex(32)

	oauthMu.Lock()
	accessTokens[accessToken] = &tokenEntry{
		Token:     accessToken,
		UserID:    ac.UserID,
		Country:   ac.Country,
		ExpiresAt: time.Now().Add(accessTokenDuration),
	}
	refreshTokens[refreshToken] = &refreshEntry{
		Token:   refreshToken,
		UserID:  ac.UserID,
		Country: ac.Country,
		Created: time.Now(),
	}
	oauthMu.Unlock()

	log.Printf("OAuth: tokens issued for user '%s' (country: %s)", ac.UserID, ac.Country)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(accessTokenDuration.Seconds()),
		"refresh_token": refreshToken,
	})
}

func handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	rt := r.FormValue("refresh_token")
	if rt == "" {
		writeOAuthError(w, "invalid_request", "Missing refresh_token", http.StatusBadRequest)
		return
	}

	oauthMu.Lock()
	entry, ok := refreshTokens[rt]
	if !ok {
		oauthMu.Unlock()
		writeOAuthError(w, "invalid_grant", "Invalid refresh token", http.StatusBadRequest)
		return
	}

	// Issue new access token (keep same refresh token)
	accessToken := randomHex(32)
	accessTokens[accessToken] = &tokenEntry{
		Token:     accessToken,
		UserID:    entry.UserID,
		Country:   entry.Country,
		ExpiresAt: time.Now().Add(accessTokenDuration),
	}
	oauthMu.Unlock()

	log.Printf("OAuth: access token refreshed for user '%s'", entry.UserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(accessTokenDuration.Seconds()),
		"refresh_token": rt,
	})
}

// handleOAuthRevoke revokes a token.
func handleOAuthRevoke(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusOK) // RFC 7009: always 200
		return
	}

	token := r.FormValue("token")
	if token == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	oauthMu.Lock()
	// Try as access token
	if entry, ok := accessTokens[token]; ok {
		log.Printf("OAuth: access token revoked for user '%s'", entry.UserID)
		delete(accessTokens, token)
	}
	// Try as refresh token
	if entry, ok := refreshTokens[token]; ok {
		log.Printf("OAuth: refresh token revoked for user '%s'", entry.UserID)
		delete(refreshTokens, token)
		// Also revoke all access tokens for this user
		for k, v := range accessTokens {
			if v.UserID == entry.UserID {
				delete(accessTokens, k)
			}
		}
	}
	oauthMu.Unlock()

	w.WriteHeader(http.StatusOK)
}

// extractUserFromToken extracts the user ID from a Bearer token in the request.
// Returns empty string if no valid token is present (guest mode).
func extractUserFromToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	oauthMu.Lock()
	defer oauthMu.Unlock()

	entry, ok := accessTokens[token]
	if !ok {
		return ""
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(accessTokens, token)
		return ""
	}
	return entry.UserID
}

// extractUserCountry extracts the user's country from a Bearer token in the request.
// Returns empty string if no valid token is present.
func extractUserCountry(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	oauthMu.Lock()
	defer oauthMu.Unlock()

	entry, ok := accessTokens[token]
	if !ok {
		return ""
	}
	if time.Now().After(entry.ExpiresAt) {
		return ""
	}
	return entry.Country
}

// isTokenExpired checks if the Authorization header has an expired token.
// Returns true only if a token is present but expired (not if absent).
func isTokenExpired(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return false
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	oauthMu.Lock()
	defer oauthMu.Unlock()

	entry, ok := accessTokens[token]
	if !ok {
		return true // unknown token = treat as expired
	}
	return time.Now().After(entry.ExpiresAt)
}

func writeOAuthError(w http.ResponseWriter, code, description string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             code,
		"error_description": description,
	})
}

// verifyPKCE checks that SHA256(codeVerifier) base64url-encoded matches codeChallenge.
func verifyPKCE(codeVerifier, codeChallenge string) bool {
	h := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])
	return computed == codeChallenge
}

const oauthLoginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Sign in - %s</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#0f1117;color:#e0e0e0;min-height:100vh;display:flex;align-items:center;justify-content:center}
.card{background:#1a1d27;border:1px solid #2a2d3a;border-radius:12px;padding:32px;max-width:400px;width:90%%}
h1{font-size:20px;margin-bottom:4px;color:#fff}
.subtitle{color:#888;font-size:14px;margin-bottom:24px}
label{display:block;font-size:13px;color:#aaa;margin-bottom:6px}
input[type=text],select{width:100%%;padding:10px 12px;border:1px solid #2a2d3a;border-radius:8px;background:#252836;color:#e0e0e0;font-size:14px;outline:none}
input[type=text]:focus,select:focus{border-color:#7c3aed}
select{margin-top:12px}
.btn{display:block;width:100%%;padding:12px;border:none;border-radius:8px;background:#7c3aed;color:#fff;font-size:14px;font-weight:600;cursor:pointer;margin-top:16px}
.btn:hover{background:#6d28d9}
.info{margin-top:16px;font-size:12px;color:#555;text-align:center}
</style>
</head>
<body>
<form class="card" method="POST">
  <h1>Sign in</h1>
  <p class="subtitle">Enter a username to link your account</p>
  <label for="username">Username</label>
  <input type="text" id="username" name="username" placeholder="e.g., alice" required autofocus>
  <label for="country" style="margin-top:12px">Country</label>
  <select id="country" name="country">
    <option value="">Select your country</option>
    <option value="US">United States</option>
    <option value="CA">Canada</option>
    <option value="GB">United Kingdom</option>
    <option value="FR">France</option>
    <option value="DE">Germany</option>
    <option value="JP">Japan</option>
    <option value="IT">Italy</option>
    <option value="ES">Spain</option>
    <option value="AU">Australia</option>
    <option value="BR">Brazil</option>
  </select>
  <input type="hidden" name="redirect_uri" value="%s">
  <input type="hidden" name="state" value="%s">
  <input type="hidden" name="code_challenge" value="%s">
  <button type="submit" class="btn">Sign in & Authorize</button>
  <p class="info">This is a sample merchant login for testing UCP identity linking.</p>
</form>
</body>
</html>`
