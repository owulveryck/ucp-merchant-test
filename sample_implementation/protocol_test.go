package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestDiscoveryURLs(t *testing.T) {
	ts := newTestServer(t)

	resp, data := ts.doRequest("GET", "/.well-known/ucp", nil, ts.getHeaders(""))
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	// Extract and validate spec/schema URLs
	ucp := data["ucp"].(map[string]interface{})
	capabilities := ucp["capabilities"].([]interface{})
	for _, cap := range capabilities {
		capMap := cap.(map[string]interface{})
		if spec, ok := capMap["spec"].(string); ok && spec != "" {
			specResp, _ := ts.doRequest("GET", urlToPath(spec, ts.URL), nil, ts.getHeaders(""))
			if specResp.StatusCode != 200 {
				t.Fatalf("Spec URL %s returned %d", spec, specResp.StatusCode)
			}
		}
		if schema, ok := capMap["schema"].(string); ok && schema != "" {
			schemaResp, _ := ts.doRequest("GET", urlToPath(schema, ts.URL), nil, ts.getHeaders(""))
			if schemaResp.StatusCode != 200 {
				t.Fatalf("Schema URL %s returned %d", schema, schemaResp.StatusCode)
			}
		}
	}

	// Payment handler URLs
	payment := data["payment"].(map[string]interface{})
	handlers := payment["handlers"].([]interface{})
	for _, h := range handlers {
		hMap := h.(map[string]interface{})
		if spec, ok := hMap["spec"].(string); ok && spec != "" {
			specResp, _ := ts.doRequest("GET", urlToPath(spec, ts.URL), nil, ts.getHeaders(""))
			if specResp.StatusCode != 200 {
				t.Fatalf("Payment spec URL %s returned %d", spec, specResp.StatusCode)
			}
		}
	}
}

func TestDiscovery(t *testing.T) {
	ts := newTestServer(t)

	resp, data := ts.doRequest("GET", "/.well-known/ucp", nil, ts.getHeaders(""))
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	ucp := data["ucp"].(map[string]interface{})
	if ucp["version"] != "2026-01-11" {
		t.Fatalf("Unexpected UCP version: %v", ucp["version"])
	}

	// Verify capabilities
	capabilities := ucp["capabilities"].([]interface{})
	expectedCaps := map[string]bool{
		"dev.ucp.shopping.checkout":      false,
		"dev.ucp.shopping.order":         false,
		"dev.ucp.shopping.discount":      false,
		"dev.ucp.shopping.fulfillment":   false,
		"dev.ucp.shopping.buyer_consent": false,
	}
	for _, cap := range capabilities {
		capMap := cap.(map[string]interface{})
		name := capMap["name"].(string)
		if _, ok := expectedCaps[name]; ok {
			expectedCaps[name] = true
		}
	}
	for name, found := range expectedCaps {
		if !found {
			t.Fatalf("Missing capability: %s", name)
		}
	}

	// Verify payment handlers
	payment := data["payment"].(map[string]interface{})
	handlers := payment["handlers"].([]interface{})
	expectedHandlers := map[string]bool{
		"google_pay":           false,
		"mock_payment_handler": false,
		"shop_pay":             false,
	}
	for _, h := range handlers {
		hMap := h.(map[string]interface{})
		id := hMap["id"].(string)
		if _, ok := expectedHandlers[id]; ok {
			expectedHandlers[id] = true
		}
	}
	for id, found := range expectedHandlers {
		if !found {
			t.Fatalf("Missing payment handler: %s", id)
		}
	}

	// Verify shop_pay config
	for _, h := range handlers {
		hMap := h.(map[string]interface{})
		if hMap["id"] == "shop_pay" {
			if hMap["name"] != "com.shopify.shop_pay" {
				t.Fatalf("Shop Pay name mismatch: %v", hMap["name"])
			}
			config := hMap["config"].(map[string]interface{})
			if _, ok := config["shop_id"]; !ok {
				t.Fatal("Shop Pay config missing shop_id")
			}
		}
	}

	// Verify shopping service
	services := ucp["services"].(map[string]interface{})
	bindings, ok := services["dev.ucp.shopping"].([]interface{})
	if !ok || len(bindings) == 0 {
		t.Fatal("Shopping service not found")
	}
	// Find the REST binding
	var restBinding map[string]interface{}
	for _, b := range bindings {
		bm := b.(map[string]interface{})
		if bm["transport"] == "rest" {
			restBinding = bm
			break
		}
	}
	if restBinding == nil {
		t.Fatal("REST binding not found for shopping service")
	}
	if restBinding["version"] != "2026-01-11" {
		t.Fatalf("Shopping service version mismatch: %v", restBinding["version"])
	}
	if restBinding["endpoint"] == nil || restBinding["endpoint"] == "" {
		t.Fatal("Endpoint not found for shopping service")
	}
	// Verify MCP binding exists
	var mcpFound bool
	for _, b := range bindings {
		bm := b.(map[string]interface{})
		if bm["transport"] == "mcp" {
			mcpFound = true
			break
		}
	}
	if !mcpFound {
		t.Fatal("MCP binding not found for shopping service")
	}
}

func TestVersionNegotiation(t *testing.T) {
	ts := newTestServer(t)

	payload := ts.createCheckoutPayload("", 0)

	// Compatible version
	headers := ts.getHeaders("")
	headers["UCP-Agent"] = fmt.Sprintf(`profile="%s/.well-known/ucp"; version="2026-01-11"`, ts.URL)
	resp, _ := ts.doRequest("POST", "/shopping-api/checkout-sessions", payload, headers)
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		t.Fatalf("Expected 200/201, got %d", resp.StatusCode)
	}

	// Incompatible version
	headers2 := ts.getHeaders("")
	headers2["UCP-Agent"] = fmt.Sprintf(`profile="%s/.well-known/ucp"; version="2099-01-01"`, ts.URL)
	resp2, _ := ts.doRequest("POST", "/shopping-api/checkout-sessions", payload, headers2)
	if resp2.StatusCode != 400 {
		t.Fatalf("Expected 400 for incompatible version, got %d", resp2.StatusCode)
	}
}

func urlToPath(fullURL, baseURL string) string {
	if strings.HasPrefix(fullURL, baseURL) {
		return strings.TrimPrefix(fullURL, baseURL)
	}
	// For local URLs using localhost, extract path
	if idx := strings.Index(fullURL, "://"); idx >= 0 {
		rest := fullURL[idx+3:]
		if slashIdx := strings.Index(rest, "/"); slashIdx >= 0 {
			return rest[slashIdx:]
		}
	}
	return fullURL
}
