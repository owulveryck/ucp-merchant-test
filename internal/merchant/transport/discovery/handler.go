package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

// Server handles UCP discovery, specs, and schemas endpoints.
type Server struct {
	baseURL func() string
}

// New creates a discovery Server. baseURL is called on each request
// to obtain the current base URL (e.g. "http://localhost:8182").
func New(baseURL func() string) *Server {
	return &Server{baseURL: baseURL}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")
}

// HandleDiscovery serves the /.well-known/ucp endpoint.
func (s *Server) HandleDiscovery(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	base := s.baseURL()
	json.NewEncoder(w).Encode(model.UCPDiscovery{
		UCP: model.UCPDiscoveryProfile{
			Version: "2026-01-11",
			Services: map[ucp.UCPService]model.UCPServiceEntry{
				ucp.ServiceShopping: {
					Version: "2026-01-11",
					Spec:    base + "/specs/shopping",
					Rest: &model.UCPTransportBinding{
						Schema:   base + "/schemas/shopping/rest.json",
						Endpoint: base + "/shopping-api",
					},
					MCP: &model.UCPTransportBinding{
						Schema:   base + "/schemas/shopping/mcp.openrpc.json",
						Endpoint: base + "/mcp",
					},
					A2A: &model.UCPTransportBinding{
						Endpoint: base + "/.well-known/agent-card.json",
					},
				},
			},
			Capabilities: []model.UCPCapabilityEntry{
				{Name: "dev.ucp.shopping.checkout", Version: "2026-01-11", Spec: base + "/specs/shopping/checkout", Schema: base + "/schemas/shopping/checkout.json"},
				{Name: "dev.ucp.shopping.order", Version: "2026-01-11", Spec: base + "/specs/shopping/order", Schema: base + "/schemas/shopping/order.json"},
				{Name: "dev.ucp.shopping.discount", Version: "2026-01-11", Spec: base + "/specs/shopping/discount", Schema: base + "/schemas/shopping/discount.json"},
				{Name: "dev.ucp.shopping.fulfillment", Version: "2026-01-11", Spec: base + "/specs/shopping/fulfillment", Schema: base + "/schemas/shopping/fulfillment.json"},
				{Name: "dev.ucp.shopping.buyer_consent", Version: "2026-01-11", Spec: base + "/specs/shopping/buyer_consent", Schema: base + "/schemas/shopping/buyer_consent.json"},
				{Name: "dev.ucp.common.identity_linking", Version: "2026-01-11", Spec: base + "/specs/identity_linking", Schema: base + "/schemas/identity_linking.json"},
			},
		},
		Payment: model.UCPPaymentProfile{
			Handlers: []map[string]any{
				{
					"id":                 "google_pay",
					"name":               "google.pay",
					"version":            "2026-01-11",
					"spec":               base + "/specs/payment/google_pay",
					"config_schema":      base + "/schemas/payment/google_pay.json",
					"instrument_schemas": []string{base + "/schemas/payment/google_pay_instrument.json"},
					"config":             map[string]any{},
				},
				{
					"id":                 "mock_payment_handler",
					"name":               "mock_payment_handler",
					"version":            "2026-01-11",
					"spec":               base + "/specs/payment/mock",
					"config_schema":      base + "/schemas/payment/mock.json",
					"instrument_schemas": []string{base + "/schemas/payment/mock_instrument.json"},
					"config":             map[string]any{},
				},
				{
					"id":                 "shop_pay",
					"name":               "com.shopify.shop_pay",
					"version":            "2026-01-11",
					"spec":               base + "/specs/payment/shop_pay",
					"config_schema":      base + "/schemas/payment/shop_pay.json",
					"instrument_schemas": []string{base + "/schemas/payment/shop_pay_instrument.json"},
					"config":             map[string]any{"shop_id": "merchant_1"},
				},
			},
		},
	})
}

// HandleSpecsAndSchemas serves /specs/ and /schemas/ endpoints.
func (s *Server) HandleSpecsAndSchemas(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if strings.HasPrefix(r.URL.Path, "/schemas/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object"}`)
	} else {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<!DOCTYPE html><html><head><title>UCP Spec</title></head><body><h1>%s</h1></body></html>`, r.URL.Path)
	}
}
