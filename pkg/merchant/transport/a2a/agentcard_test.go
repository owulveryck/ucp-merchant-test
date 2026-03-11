package a2a

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	a2alib "github.com/a2aproject/a2a-go/a2a"

	"github.com/owulveryck/ucp-merchant-test/pkg/auth"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/merchanttest"
)

func TestHandleAgentCard_Fields(t *testing.T) {
	mock := merchanttest.NewMock()
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	s := New(mock, authSrv,
		WithMerchantName("Test Merchant"),
		WithListenPort(func() int { return 8080 }),
	)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent-card.json", nil)
	w := httptest.NewRecorder()

	s.HandleAgentCard(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var card a2alib.AgentCard
	if err := json.Unmarshal(w.Body.Bytes(), &card); err != nil {
		t.Fatalf("failed to parse agent card: %v", err)
	}

	if card.Name != "Test Merchant" {
		t.Errorf("expected name=Test Merchant, got %s", card.Name)
	}

	if card.URL != "http://localhost:8080/a2a" {
		t.Errorf("expected URL=http://localhost:8080/a2a, got %s", card.URL)
	}

	if len(card.Skills) != 4 {
		t.Errorf("expected 4 skills, got %d", len(card.Skills))
	}

	if len(card.DefaultInputModes) != 1 || card.DefaultInputModes[0] != "application/json" {
		t.Errorf("expected defaultInputModes=[application/json], got %v", card.DefaultInputModes)
	}
}

func TestHandleAgentCard_UCPExtension(t *testing.T) {
	mock := merchanttest.NewMock()
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	s := New(mock, authSrv)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent-card.json", nil)
	w := httptest.NewRecorder()

	s.HandleAgentCard(w, req)

	var card a2alib.AgentCard
	json.Unmarshal(w.Body.Bytes(), &card)

	if len(card.Capabilities.Extensions) == 0 {
		t.Fatal("expected at least one extension")
	}

	ext := card.Capabilities.Extensions[0]
	if ext.URI != ucpExtensionURI {
		t.Errorf("expected URI=%s, got %s", ucpExtensionURI, ext.URI)
	}

	params, ok := ext.Params["capabilities"]
	if !ok {
		t.Fatal("expected capabilities in extension params")
	}
	caps, ok := params.(map[string]any)
	if !ok {
		t.Fatal("expected capabilities to be a map")
	}
	if _, ok := caps["dev.ucp.shopping.checkout"]; !ok {
		t.Error("expected dev.ucp.shopping.checkout capability")
	}
}

func TestHandleAgentCard_Skills(t *testing.T) {
	mock := merchanttest.NewMock()
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	s := New(mock, authSrv)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent-card.json", nil)
	w := httptest.NewRecorder()

	s.HandleAgentCard(w, req)

	var card a2alib.AgentCard
	json.Unmarshal(w.Body.Bytes(), &card)

	skillIDs := make(map[string]bool)
	for _, skill := range card.Skills {
		skillIDs[skill.ID] = true
	}

	expected := []string{"catalog", "cart", "checkout", "orders"}
	for _, id := range expected {
		if !skillIDs[id] {
			t.Errorf("expected skill %s in agent card", id)
		}
	}
}

func TestHandleAgentCard_OPTIONS(t *testing.T) {
	mock := merchanttest.NewMock()
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	s := New(mock, authSrv)

	req := httptest.NewRequest(http.MethodOptions, "/.well-known/agent-card.json", nil)
	w := httptest.NewRecorder()

	s.HandleAgentCard(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}
