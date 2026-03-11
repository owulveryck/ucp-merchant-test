package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/owulveryck/ucp-merchant-test/internal/auth"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/merchanttest"
)

func TestServeHTTP_CORS(t *testing.T) {
	mock := merchanttest.NewMock()
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	s := New(mock, authSrv)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	headers := map[string]string{
		"Access-Control-Allow-Origin":   "*",
		"Access-Control-Allow-Methods":  "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers":  "Content-Type, Mcp-Session-Id, Authorization",
		"Access-Control-Expose-Headers": "Mcp-Session-Id",
	}
	for k, v := range headers {
		if got := w.Header().Get(k); got != v {
			t.Errorf("header %s: expected %q, got %q", k, v, got)
		}
	}
}

func TestServeHTTP_OPTIONS(t *testing.T) {
	mock := merchanttest.NewMock()
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	s := New(mock, authSrv)

	req := httptest.NewRequest(http.MethodOptions, "/mcp", nil)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestServeHTTP_ExpiredToken(t *testing.T) {
	mock := merchanttest.NewMock()
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	s := New(mock, authSrv)

	// Inject an already-expired token
	token := authSrv.InjectToken("user1", "US", time.Now().Add(-time.Hour))

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	s.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "token_expired" {
		t.Errorf("expected error=token_expired, got %s", body["error"])
	}
}
