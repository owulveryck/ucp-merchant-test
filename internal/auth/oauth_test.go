package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

func newTestServer() *OAuthServer {
	return NewOAuthServer("TestMerchant", func() string { return "http" }, func() int { return 8080 })
}

func TestInjectAndExtractToken(t *testing.T) {
	s := newTestServer()
	token := s.InjectToken("alice", ucp.Country("US"), time.Now().Add(time.Hour))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	user := s.ExtractUserFromToken(req)
	if user != "alice" {
		t.Errorf("expected alice, got %s", user)
	}

	country := s.ExtractUserCountry(req)
	if country != "US" {
		t.Errorf("expected US, got %s", country)
	}
}

func TestExpiredToken(t *testing.T) {
	s := newTestServer()
	token := s.InjectToken("alice", ucp.Country("US"), time.Now().Add(-time.Hour))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	user := s.ExtractUserFromToken(req)
	if user != "" {
		t.Error("expected empty user for expired token")
	}

	if !s.IsTokenExpired(req) {
		t.Error("expected token to be expired")
	}
}

func TestReset(t *testing.T) {
	s := newTestServer()
	s.InjectToken("alice", ucp.Country("US"), time.Now().Add(time.Hour))
	s.Reset()

	if len(s.accessTokens) != 0 {
		t.Error("expected empty access tokens after reset")
	}
}

func TestHandleMetadata(t *testing.T) {
	s := newTestServer()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/.well-known/oauth-authorization-server", nil)
	s.HandleMetadata(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "authorization_endpoint") {
		t.Error("expected metadata to contain authorization_endpoint")
	}
}

func TestHandleRevoke(t *testing.T) {
	s := newTestServer()
	token := s.InjectToken("alice", ucp.Country("US"), time.Now().Add(time.Hour))

	form := url.Values{"token": {token}}
	req := httptest.NewRequest("POST", "/oauth2/revoke", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	s.HandleRevoke(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Token should be gone
	checkReq := httptest.NewRequest("GET", "/", nil)
	checkReq.Header.Set("Authorization", "Bearer "+token)
	if s.ExtractUserFromToken(checkReq) != "" {
		t.Error("expected token to be revoked")
	}
}

func TestNoAuthHeader(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("GET", "/", nil)
	if s.ExtractUserFromToken(req) != "" {
		t.Error("expected empty user with no auth header")
	}
	if s.IsTokenExpired(req) {
		t.Error("expected not expired with no auth header")
	}
}

func TestTokenExchangeFlow(t *testing.T) {
	s := newTestServer()

	// Generate PKCE
	verifier := "test-verifier-string-at-least-43-chars-long-for-pkce"
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	// GET authorize page
	req := httptest.NewRequest("GET", "/oauth2/authorize?client_id=agentflowui&redirect_uri=http://localhost/cb&state=xyz&code_challenge="+challenge+"&code_challenge_method=S256", nil)
	w := httptest.NewRecorder()
	s.HandleAuthorize(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for authorize page, got %d", w.Code)
	}

	// POST authorize (simulate login)
	form := url.Values{
		"username":       {"testuser"},
		"country":        {"US"},
		"redirect_uri":   {"http://localhost/cb"},
		"state":          {"xyz"},
		"code_challenge": {challenge},
	}
	req = httptest.NewRequest("POST", "/oauth2/authorize?client_id=agentflowui&redirect_uri=http://localhost/cb&state=xyz&code_challenge="+challenge+"&code_challenge_method=S256", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	s.HandleAuthorize(w, req)
	if w.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", w.Code)
	}

	// Extract code from redirect
	loc := w.Header().Get("Location")
	u, _ := url.Parse(loc)
	code := u.Query().Get("code")
	if code == "" {
		t.Fatal("expected code in redirect")
	}

	// Exchange code for tokens
	tokenForm := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {verifier},
	}
	req = httptest.NewRequest("POST", "/oauth2/token", strings.NewReader(tokenForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	s.HandleToken(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for token exchange, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "access_token") {
		t.Error("expected access_token in response")
	}
}
