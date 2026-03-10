package auth

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

	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

const (
	// OAuthClientID is the expected OAuth 2.0 client identifier for UCP platforms.
	OAuthClientID = "agentflowui"
	// AccessTokenDuration is the validity period for issued access tokens (1 hour).
	AccessTokenDuration = 1 * time.Hour
	// AuthCodeDuration is the validity period for authorization codes (5 minutes).
	AuthCodeDuration = 5 * time.Minute
)

type authCode struct {
	Code          string
	UserID        string
	Country       ucp.Country
	CodeChallenge string
	RedirectURI   string
	ExpiresAt     time.Time
}

type tokenEntry struct {
	Token     string
	UserID    string
	Country   ucp.Country
	ExpiresAt time.Time
}

type refreshEntry struct {
	Token   string
	UserID  string
	Country ucp.Country
	Created time.Time
}

// OAuthServer manages OAuth2 state and handles.
type OAuthServer struct {
	mu            sync.Mutex
	authCodes     map[string]*authCode
	accessTokens  map[string]*tokenEntry
	refreshTokens map[string]*refreshEntry

	MerchantName string
	Scheme       func() string
	ListenPort   func() int
}

// NewOAuthServer creates a new OAuth server.
func NewOAuthServer(merchantName string, scheme func() string, listenPort func() int) *OAuthServer {
	return &OAuthServer{
		authCodes:     map[string]*authCode{},
		accessTokens:  map[string]*tokenEntry{},
		refreshTokens: map[string]*refreshEntry{},
		MerchantName:  merchantName,
		Scheme:        scheme,
		ListenPort:    listenPort,
	}
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Reset clears all OAuth state.
func (s *OAuthServer) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessTokens = map[string]*tokenEntry{}
	s.refreshTokens = map[string]*refreshEntry{}
	s.authCodes = map[string]*authCode{}
}

// HandleMetadata serves RFC 8414 OAuth Authorization Server Metadata.
func (s *OAuthServer) HandleMetadata(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	base := fmt.Sprintf("%s://localhost:%d", s.Scheme(), s.ListenPort())
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

// HandleAuthorize shows a minimal login/consent page.
func (s *OAuthServer) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")

	if clientID != OAuthClientID {
		http.Error(w, "Invalid client_id", http.StatusBadRequest)
		return
	}
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		http.Error(w, "PKCE S256 required", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, oauthLoginHTML, s.MerchantName, redirectURI, state, codeChallenge)
		return
	}

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

	code := randomHex(16)
	s.mu.Lock()
	s.authCodes[code] = &authCode{
		Code:          code,
		UserID:        username,
		Country:       ucp.NewCountry(country),
		CodeChallenge: codeChallengeForm,
		RedirectURI:   redirectURIForm,
		ExpiresAt:     time.Now().Add(AuthCodeDuration),
	}
	s.mu.Unlock()

	log.Printf("OAuth: user '%s' authorized, code issued", username)

	sep := "?"
	if strings.Contains(redirectURIForm, "?") {
		sep = "&"
	}
	http.Redirect(w, r, fmt.Sprintf("%s%scode=%s&state=%s", redirectURIForm, sep, code, stateForm), http.StatusFound)
}

// HandleToken exchanges auth code for tokens, or refreshes tokens.
func (s *OAuthServer) HandleToken(w http.ResponseWriter, r *http.Request) {
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

	switch r.FormValue("grant_type") {
	case "authorization_code":
		s.handleTokenExchange(w, r)
	case "refresh_token":
		s.handleTokenRefresh(w, r)
	default:
		writeOAuthError(w, "unsupported_grant_type", "Unsupported grant_type", http.StatusBadRequest)
	}
}

func (s *OAuthServer) handleTokenExchange(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	codeVerifier := r.FormValue("code_verifier")

	if code == "" || codeVerifier == "" {
		writeOAuthError(w, "invalid_request", "Missing code or code_verifier", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	ac, ok := s.authCodes[code]
	if ok {
		delete(s.authCodes, code)
	}
	s.mu.Unlock()

	if !ok {
		writeOAuthError(w, "invalid_grant", "Invalid or expired authorization code", http.StatusBadRequest)
		return
	}
	if time.Now().After(ac.ExpiresAt) {
		writeOAuthError(w, "invalid_grant", "Authorization code expired", http.StatusBadRequest)
		return
	}

	if !verifyPKCE(codeVerifier, ac.CodeChallenge) {
		writeOAuthError(w, "invalid_grant", "PKCE verification failed", http.StatusBadRequest)
		return
	}

	accessToken := randomHex(32)
	refreshToken := randomHex(32)

	s.mu.Lock()
	s.accessTokens[accessToken] = &tokenEntry{
		Token:     accessToken,
		UserID:    ac.UserID,
		Country:   ac.Country,
		ExpiresAt: time.Now().Add(AccessTokenDuration),
	}
	s.refreshTokens[refreshToken] = &refreshEntry{
		Token:   refreshToken,
		UserID:  ac.UserID,
		Country: ac.Country,
		Created: time.Now(),
	}
	s.mu.Unlock()

	log.Printf("OAuth: tokens issued for user '%s' (country: %s)", ac.UserID, ac.Country)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(AccessTokenDuration.Seconds()),
		"refresh_token": refreshToken,
	})
}

func (s *OAuthServer) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	rt := r.FormValue("refresh_token")
	if rt == "" {
		writeOAuthError(w, "invalid_request", "Missing refresh_token", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	entry, ok := s.refreshTokens[rt]
	if !ok {
		s.mu.Unlock()
		writeOAuthError(w, "invalid_grant", "Invalid refresh token", http.StatusBadRequest)
		return
	}

	accessToken := randomHex(32)
	s.accessTokens[accessToken] = &tokenEntry{
		Token:     accessToken,
		UserID:    entry.UserID,
		Country:   entry.Country,
		ExpiresAt: time.Now().Add(AccessTokenDuration),
	}
	s.mu.Unlock()

	log.Printf("OAuth: access token refreshed for user '%s'", entry.UserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(AccessTokenDuration.Seconds()),
		"refresh_token": rt,
	})
}

// HandleRevoke revokes a token.
func (s *OAuthServer) HandleRevoke(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusOK)
		return
	}

	token := r.FormValue("token")
	if token == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mu.Lock()
	if entry, ok := s.accessTokens[token]; ok {
		log.Printf("OAuth: access token revoked for user '%s'", entry.UserID)
		delete(s.accessTokens, token)
	}
	if entry, ok := s.refreshTokens[token]; ok {
		log.Printf("OAuth: refresh token revoked for user '%s'", entry.UserID)
		delete(s.refreshTokens, token)
		for k, v := range s.accessTokens {
			if v.UserID == entry.UserID {
				delete(s.accessTokens, k)
			}
		}
	}
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

// ExtractUserFromToken extracts the user ID from a Bearer token.
func (s *OAuthServer) ExtractUserFromToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.accessTokens[token]
	if !ok {
		return ""
	}
	if time.Now().After(entry.ExpiresAt) {
		delete(s.accessTokens, token)
		return ""
	}
	return entry.UserID
}

// ExtractUserCountry extracts the user's country from a Bearer token.
func (s *OAuthServer) ExtractUserCountry(r *http.Request) ucp.Country {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.accessTokens[token]
	if !ok {
		return ""
	}
	if time.Now().After(entry.ExpiresAt) {
		return ""
	}
	return entry.Country
}

// IsTokenExpired checks if the Authorization header has an expired token.
func (s *OAuthServer) IsTokenExpired(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return false
	}
	token := strings.TrimPrefix(auth, "Bearer ")

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.accessTokens[token]
	if !ok {
		return true
	}
	return time.Now().After(entry.ExpiresAt)
}

// InjectToken creates a token directly (for testing).
func (s *OAuthServer) InjectToken(userID string, country ucp.Country, expiresAt time.Time) string {
	token := randomHex(16)
	s.mu.Lock()
	s.accessTokens[token] = &tokenEntry{
		Token:     token,
		UserID:    userID,
		Country:   country,
		ExpiresAt: expiresAt,
	}
	s.mu.Unlock()
	return token
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")
}

func writeOAuthError(w http.ResponseWriter, code, description string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             code,
		"error_description": description,
	})
}

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
