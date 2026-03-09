package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveWebhookURLEmpty(t *testing.T) {
	url := ResolveWebhookURL("")
	if url != "" {
		t.Errorf("expected empty URL, got %s", url)
	}
}

func TestResolveWebhookURLNoProfile(t *testing.T) {
	url := ResolveWebhookURL("agent_name=test")
	if url != "" {
		t.Errorf("expected empty URL when no profile, got %s", url)
	}
}

func TestResolveWebhookURLWithProfile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ucp": map[string]interface{}{
				"capabilities": []interface{}{
					map[string]interface{}{
						"name": "webhook",
						"config": map[string]interface{}{
							"webhook_url": "https://example.com/webhook",
						},
					},
				},
			},
		})
	}))
	defer ts.Close()

	url := ResolveWebhookURL("agent_name=test; profile=" + ts.URL)
	if url != "https://example.com/webhook" {
		t.Errorf("expected https://example.com/webhook, got %s", url)
	}
}

func TestSendWebhookEventEmpty(t *testing.T) {
	// Should not panic with empty URL
	SendWebhookEvent("", map[string]interface{}{"event_type": "test"})
}
