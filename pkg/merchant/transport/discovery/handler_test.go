package discovery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testServer() *Server {
	return New(func() string { return "http://test.example.com" })
}

func TestHandleDiscovery_JSONStructure(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/ucp", nil)
	w := httptest.NewRecorder()

	s.HandleDiscovery(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	var ucpProfile struct {
		Version      string                     `json:"version"`
		Services     map[string]json.RawMessage `json:"services"`
		Capabilities []json.RawMessage          `json:"capabilities"`
	}
	if err := json.Unmarshal(body["ucp"], &ucpProfile); err != nil {
		t.Fatalf("failed to parse ucp: %v", err)
	}
	if ucpProfile.Version != "2026-01-11" {
		t.Errorf("expected version 2026-01-11, got %s", ucpProfile.Version)
	}
	if _, ok := ucpProfile.Services["dev.ucp.shopping"]; !ok {
		t.Error("missing dev.ucp.shopping service")
	}

	var svc struct {
		Rest *json.RawMessage `json:"rest"`
		MCP  *json.RawMessage `json:"mcp"`
	}
	if err := json.Unmarshal(ucpProfile.Services["dev.ucp.shopping"], &svc); err != nil {
		t.Fatalf("failed to parse service: %v", err)
	}
	if svc.Rest == nil {
		t.Error("missing rest transport")
	}
	if svc.MCP == nil {
		t.Error("missing mcp transport")
	}

	if len(ucpProfile.Capabilities) != 6 {
		t.Errorf("expected 6 capabilities, got %d", len(ucpProfile.Capabilities))
	}

	var payment struct {
		Handlers []json.RawMessage `json:"handlers"`
	}
	if err := json.Unmarshal(body["payment"], &payment); err != nil {
		t.Fatalf("failed to parse payment: %v", err)
	}
	if len(payment.Handlers) != 3 {
		t.Errorf("expected 3 payment handlers, got %d", len(payment.Handlers))
	}
}

func TestHandleDiscovery_BaseURLInEndpoints(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/ucp", nil)
	w := httptest.NewRecorder()

	s.HandleDiscovery(w, req)

	body := w.Body.String()
	base := "http://test.example.com"

	endpoints := []string{
		base + "/specs/shopping",
		base + "/schemas/shopping/rest.json",
		base + "/shopping-api",
		base + "/schemas/shopping/mcp.openrpc.json",
		base + "/mcp",
	}
	for _, ep := range endpoints {
		if !strings.Contains(body, ep) {
			t.Errorf("response missing endpoint %s", ep)
		}
	}
}

func TestHandleDiscovery_CORS(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/ucp", nil)
	w := httptest.NewRecorder()

	s.HandleDiscovery(w, req)

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

func TestHandleDiscovery_OPTIONS(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodOptions, "/.well-known/ucp", nil)
	w := httptest.NewRecorder()

	s.HandleDiscovery(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("missing CORS header on OPTIONS")
	}
}

func TestHandleSpecsAndSchemas_Schema(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/schemas/shopping/rest.json", nil)
	w := httptest.NewRecorder()

	s.HandleSpecsAndSchemas(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
	var schema map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &schema); err != nil {
		t.Errorf("invalid JSON schema response: %v", err)
	}
}

func TestHandleSpecsAndSchemas_Spec(t *testing.T) {
	s := testServer()
	req := httptest.NewRequest(http.MethodGet, "/specs/shopping/checkout", nil)
	w := httptest.NewRecorder()

	s.HandleSpecsAndSchemas(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html" {
		t.Errorf("expected text/html, got %s", ct)
	}
	if !strings.Contains(w.Body.String(), "/specs/shopping/checkout") {
		t.Error("spec HTML should contain the request path")
	}
}
