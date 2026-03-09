package webhook_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/owulveryck/ucp-merchant-test/internal/webhook"
)

func ExampleResolveWebhookURL() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ucp":{"capabilities":[{"name":"webhooks","config":{"webhook_url":"https://agent.example.com/hooks"}}]}}`)
	}))
	defer ts.Close()

	header := fmt.Sprintf("agent_name=test; profile=%s", ts.URL)
	url := webhook.ResolveWebhookURL(header)
	fmt.Println(url)
	// Output:
	// https://agent.example.com/hooks
}
