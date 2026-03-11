package a2a_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	a2alib "github.com/a2aproject/a2a-go/a2a"

	"github.com/owulveryck/ucp-merchant-test/internal/auth"
	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/merchanttest"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/transport/a2a"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

// ExampleServer_HandleAgentCard demonstrates serving the A2A Agent Card
// which advertises the merchant's UCP capabilities.
func ExampleServer_HandleAgentCard() {
	mock := merchanttest.NewMock()
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	server := a2a.New(mock, authSrv, a2a.WithMerchantName("Flower Shop"))

	req := httptest.NewRequest(http.MethodGet, "/.well-known/agent-card.json", nil)
	w := httptest.NewRecorder()
	server.HandleAgentCard(w, req)

	var card a2alib.AgentCard
	json.Unmarshal(w.Body.Bytes(), &card)

	fmt.Println("Name:", card.Name)
	fmt.Println("Skills:", len(card.Skills))
	fmt.Println("Extensions:", len(card.Capabilities.Extensions))
	// Output:
	// Name: Flower Shop
	// Skills: 4
	// Extensions: 1
}

// ExampleNew demonstrates creating an A2A transport server and listing
// products via a simulated A2A message/send request.
func ExampleNew() {
	mock := merchanttest.NewMock()
	mock.FilterFunc = func(ucp.Category, string, string, ucp.Country, ucp.Currency, ucp.Language) []catalog.Product {
		return []catalog.Product{
			{ID: "SKU-001", Title: "Red Roses"},
			{ID: "SKU-002", Title: "White Tulips"},
		}
	}
	mock.CategoryCountFunc = func() []catalog.CategoryStat {
		return nil
	}

	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	server := a2a.New(mock, authSrv, a2a.WithMerchantName("Flower Shop"))

	// Simulate a message/send JSON-RPC request with a list_products action.
	body := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "message/send",
		"params": {
			"message": {
				"role": "user",
				"parts": [{"kind": "data", "data": {"action": "list_products"}}],
				"messageId": "msg-1",
				"kind": "message"
			}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/a2a", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = http.NoBody // will be replaced below
	req = httptest.NewRequest(http.MethodPost, "/a2a", stringReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	// Output:
	// Status: 200
}

type stringReaderType struct{ s string }

func (r *stringReaderType) Read(p []byte) (int, error) {
	n := copy(p, r.s)
	r.s = r.s[n:]
	if len(r.s) == 0 {
		return n, fmt.Errorf("EOF")
	}
	return n, nil
}

func stringReader(s string) *stringReaderType {
	return &stringReaderType{s: s}
}
