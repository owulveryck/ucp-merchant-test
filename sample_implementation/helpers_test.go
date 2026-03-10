package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var loadOnce sync.Once

type testServer struct {
	*httptest.Server
	t *testing.T
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	resetStores()

	loadOnce.Do(func() {
		if err := loadFlowerShopData("testdata/flower_shop", "csv"); err != nil {
			t.Fatalf("Failed to load test data: %v", err)
		}
	})

	// Reset dynamic state after loading (loadOnce won't re-load, but we need clean dynamic data)
	shopData.ResetDynamicAddresses()

	// Capture server logs: suppress during pass, dump on failure
	var logBuf bytes.Buffer
	origOutput := log.Writer()
	log.SetOutput(&logBuf)

	srv := httptest.NewServer(newMux())
	u, _ := url.Parse(srv.URL)
	port, _ := strconv.Atoi(u.Port())
	listenPort = port
	tlsEnabled = false
	simulationSecret = "test-secret"
	merchantName = "Test Merchant"

	t.Cleanup(func() {
		srv.Close()
		log.SetOutput(origOutput)
		if t.Failed() && logBuf.Len() > 0 {
			t.Log("Server logs:\n" + logBuf.String())
		}
	})

	return &testServer{Server: srv, t: t}
}

func (ts *testServer) getHeaders(idempotencyKey string) map[string]string {
	if idempotencyKey == "" {
		idempotencyKey = fmt.Sprintf("idem-%d", time.Now().UnixNano())
	}
	return map[string]string{
		"Content-Type":      "application/json",
		"idempotency-key":   idempotencyKey,
		"request-signature": "test",
	}
}

func (ts *testServer) getHeadersWithAgent(idempotencyKey, agentProfileURL string) map[string]string {
	h := ts.getHeaders(idempotencyKey)
	h["UCP-Agent"] = fmt.Sprintf(`profile="%s"`, agentProfileURL)
	return h
}

func (ts *testServer) doRequest(method, path string, body interface{}, headers map[string]string) (*http.Response, map[string]interface{}) {
	ts.t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, ts.URL+path, bodyReader)
	if err != nil {
		ts.t.Fatalf("Failed to create request: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ts.t.Fatalf("Request failed: %v", err)
	}

	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	// Re-create body for re-reading
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	return resp, result
}

func (ts *testServer) createCheckoutPayload(itemID string, qty int) map[string]interface{} {
	if itemID == "" {
		itemID = "bouquet_roses"
	}
	if qty == 0 {
		qty = 1
	}

	return map[string]interface{}{
		"currency": "USD",
		"line_items": []map[string]interface{}{
			{
				"item":     map[string]interface{}{"id": itemID, "title": "Test Item"},
				"quantity": qty,
			},
		},
		"payment": map[string]interface{}{
			"selected_instrument_id": "instr_1",
			"instruments":            []interface{}{},
			"handlers": []map[string]interface{}{
				{
					"id":                 "google_pay",
					"name":               "google.pay",
					"version":            "2026-01-11",
					"spec":               "https://example.com/spec",
					"config_schema":      "https://example.com/schema",
					"instrument_schemas": []string{"https://example.com/instrument_schema"},
					"config":             map[string]interface{}{},
				},
			},
		},
		"fulfillment": map[string]interface{}{
			"methods": []map[string]interface{}{
				{
					"type": "shipping",
					"destinations": []map[string]interface{}{
						{"id": "dest_1", "address_country": "US"},
					},
					"selected_destination_id": "dest_1",
					"groups": []map[string]interface{}{
						{"selected_option_id": "std-ship"},
					},
				},
			},
		},
	}
}

func (ts *testServer) createCheckoutPayloadNoFulfillment(itemID string, qty int) map[string]interface{} {
	p := ts.createCheckoutPayload(itemID, qty)
	delete(p, "fulfillment")
	return p
}

func (ts *testServer) createCheckout(payload map[string]interface{}) (*http.Response, map[string]interface{}) {
	ts.t.Helper()
	return ts.doRequest("POST", "/shopping-api/checkout-sessions", payload, ts.getHeaders(""))
}

func (ts *testServer) getCheckout(checkoutID string) (*http.Response, map[string]interface{}) {
	ts.t.Helper()
	return ts.doRequest("GET", "/shopping-api/checkout-sessions/"+checkoutID, nil, ts.getHeaders(""))
}

func (ts *testServer) updateCheckout(checkoutID string, payload map[string]interface{}, headers map[string]string) (*http.Response, map[string]interface{}) {
	ts.t.Helper()
	if headers == nil {
		headers = ts.getHeaders("")
	}
	return ts.doRequest("PUT", "/shopping-api/checkout-sessions/"+checkoutID, payload, headers)
}

func (ts *testServer) completeCheckout(checkoutID string, paymentPayload map[string]interface{}) (*http.Response, map[string]interface{}) {
	ts.t.Helper()
	if paymentPayload == nil {
		paymentPayload = ts.getValidPaymentPayload()
	}
	return ts.doRequest("POST", "/shopping-api/checkout-sessions/"+checkoutID+"/complete", paymentPayload, ts.getHeaders(""))
}

func (ts *testServer) cancelCheckout(checkoutID string) (*http.Response, map[string]interface{}) {
	ts.t.Helper()
	return ts.doRequest("POST", "/shopping-api/checkout-sessions/"+checkoutID+"/cancel", nil, ts.getHeaders(""))
}

func (ts *testServer) getValidPaymentPayload() map[string]interface{} {
	return map[string]interface{}{
		"payment_data": map[string]interface{}{
			"id":           "instr_1",
			"handler_id":   "mock_payment_handler",
			"handler_name": "mock_payment_handler",
			"type":         "card",
			"brand":        "Visa",
			"last_digits":  "1234",
			"credential": map[string]interface{}{
				"type":  "token",
				"token": "success_token",
			},
			"billing_address": map[string]interface{}{
				"street_address":   "123 Main St",
				"address_locality": "Springfield",
				"address_region":   "IL",
				"address_country":  "US",
				"postal_code":      "62704",
			},
		},
		"risk_signals": map[string]interface{}{},
	}
}

func (ts *testServer) getFailPaymentPayload() map[string]interface{} {
	return map[string]interface{}{
		"payment_data": map[string]interface{}{
			"id":           "instr_fail",
			"handler_id":   "mock_payment_handler",
			"handler_name": "mock_payment_handler",
			"type":         "card",
			"brand":        "Visa",
			"last_digits":  "9999",
			"credential": map[string]interface{}{
				"type":  "token",
				"token": "fail_token",
			},
		},
		"risk_signals": map[string]interface{}{},
	}
}

func (ts *testServer) ensureFulfillmentReady(checkoutID string) map[string]interface{} {
	ts.t.Helper()
	_, data := ts.getCheckout(checkoutID)

	if isReady(data) {
		return data
	}

	// 1. Inject a default address if needed
	hasDests := false
	if f, ok := data["fulfillment"].(map[string]interface{}); ok {
		if methods, ok := f["methods"].([]interface{}); ok && len(methods) > 0 {
			m := methods[0].(map[string]interface{})
			if dests, ok := m["destinations"].([]interface{}); ok && len(dests) > 0 {
				hasDests = true
			}
		}
	}

	if !hasDests {
		updatePayload := buildUpdateFromCheckout(data)
		updatePayload["fulfillment"] = map[string]interface{}{
			"methods": []map[string]interface{}{
				{
					"type": "shipping",
					"destinations": []map[string]interface{}{
						{
							"id":               "dest_default",
							"street_address":   "123 Default St",
							"address_locality": "City",
							"address_region":   "State",
							"postal_code":      "12345",
							"address_country":  "US",
						},
					},
					"selected_destination_id": "dest_default",
				},
			},
		}
		_, data = ts.updateCheckout(checkoutID, updatePayload, nil)
	}

	// 2. Select destination if not selected
	f := data["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})

	if method["selected_destination_id"] == nil || method["selected_destination_id"] == "" {
		if dests, ok := method["destinations"].([]interface{}); ok && len(dests) > 0 {
			destID := dests[0].(map[string]interface{})["id"].(string)
			updatePayload := buildUpdateFromCheckout(data)
			updatePayload["fulfillment"] = map[string]interface{}{
				"methods": []map[string]interface{}{
					{
						"type":                    "shipping",
						"selected_destination_id": destID,
					},
				},
			}
			_, data = ts.updateCheckout(checkoutID, updatePayload, nil)
		}
	}

	// 3. Select first option if not selected
	f = data["fulfillment"].(map[string]interface{})
	methods = f["methods"].([]interface{})
	method = methods[0].(map[string]interface{})

	hasSelection := false
	if groups, ok := method["groups"].([]interface{}); ok {
		for _, g := range groups {
			gm := g.(map[string]interface{})
			if gm["selected_option_id"] != nil && gm["selected_option_id"] != "" {
				hasSelection = true
				break
			}
		}
	}

	if !hasSelection {
		if groups, ok := method["groups"].([]interface{}); ok && len(groups) > 0 {
			g := groups[0].(map[string]interface{})
			if options, ok := g["options"].([]interface{}); ok && len(options) > 0 {
				optionID := options[0].(map[string]interface{})["id"].(string)
				updatePayload := buildUpdateFromCheckout(data)
				updatePayload["fulfillment"] = map[string]interface{}{
					"methods": []map[string]interface{}{
						{
							"type":   "shipping",
							"groups": []map[string]interface{}{{"selected_option_id": optionID}},
						},
					},
				}
				_, data = ts.updateCheckout(checkoutID, updatePayload, nil)
			}
		}
	}

	return data
}

func (ts *testServer) createCheckoutSession(itemID string, qty int, buyer map[string]interface{}, selectFulfillment bool) map[string]interface{} {
	ts.t.Helper()
	var payload map[string]interface{}
	if selectFulfillment {
		payload = ts.createCheckoutPayload(itemID, qty)
	} else {
		payload = ts.createCheckoutPayloadNoFulfillment(itemID, qty)
	}
	if buyer != nil {
		payload["buyer"] = buyer
	}
	resp, data := ts.createCheckout(payload)
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		ts.t.Fatalf("createCheckoutSession failed: status=%d", resp.StatusCode)
	}
	if selectFulfillment {
		data = ts.ensureFulfillmentReady(data["id"].(string))
	}
	return data
}

func (ts *testServer) createCompletedOrder() string {
	ts.t.Helper()
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)
	resp, completeData := ts.completeCheckout(checkoutID, nil)
	if resp.StatusCode != 200 {
		ts.t.Fatalf("createCompletedOrder: complete failed status=%d", resp.StatusCode)
	}
	order := completeData["order"].(map[string]interface{})
	return order["id"].(string)
}

func isReady(data map[string]interface{}) bool {
	f, ok := data["fulfillment"].(map[string]interface{})
	if !ok {
		return false
	}
	methods, ok := f["methods"].([]interface{})
	if !ok || len(methods) == 0 {
		return false
	}
	m := methods[0].(map[string]interface{})
	if m["selected_destination_id"] == nil || m["selected_destination_id"] == "" {
		return false
	}
	groups, ok := m["groups"].([]interface{})
	if !ok || len(groups) == 0 {
		return false
	}
	g := groups[0].(map[string]interface{})
	return g["selected_option_id"] != nil && g["selected_option_id"] != ""
}

func buildUpdateFromCheckout(data map[string]interface{}) map[string]interface{} {
	lineItems := data["line_items"].([]interface{})
	var updatedItems []map[string]interface{}
	for _, li := range lineItems {
		liMap := li.(map[string]interface{})
		item := liMap["item"].(map[string]interface{})
		updatedItems = append(updatedItems, map[string]interface{}{
			"id":       liMap["id"],
			"item":     map[string]interface{}{"id": item["id"], "title": item["title"]},
			"quantity": liMap["quantity"],
		})
	}
	result := map[string]interface{}{
		"id":         data["id"],
		"currency":   data["currency"],
		"line_items": updatedItems,
	}
	if p, ok := data["payment"]; ok {
		result["payment"] = p
	}
	return result
}

// webhookRecorder is an httptest.Server that records POSTed webhook events.
type webhookRecorder struct {
	server *httptest.Server
	mu     sync.Mutex
	events []map[string]interface{}
}

func newWebhookRecorder(t *testing.T) *webhookRecorder {
	wr := &webhookRecorder{}
	wr.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		var event map[string]interface{}
		json.Unmarshal(body, &event)
		wr.mu.Lock()
		wr.events = append(wr.events, event)
		wr.mu.Unlock()
		w.WriteHeader(200)
		fmt.Fprintf(w, `{"status":"ok"}`)
	}))
	t.Cleanup(func() { wr.server.Close() })
	return wr
}

func (wr *webhookRecorder) waitForEvents(n int, timeout time.Duration) []map[string]interface{} {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		wr.mu.Lock()
		count := len(wr.events)
		wr.mu.Unlock()
		if count >= n {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	wr.mu.Lock()
	defer wr.mu.Unlock()
	result := make([]map[string]interface{}, len(wr.events))
	copy(result, wr.events)
	return result
}

// agentProfileServer serves the shopping-agent-test.json template with the webhook URL injected.
func newAgentProfileServer(t *testing.T, webhookURL string) *httptest.Server {
	template := `{
  "ucp": {
    "version": "2026-01-11",
    "capabilities": [
      {
        "name": "dev.ucp.shopping.order",
        "version": "2026-01-11",
        "spec": "https://ucp.dev/specs/shopping/order",
        "schema": "https://ucp.dev/schemas/shopping/order.json",
        "config": {
          "webhook_url": "WEBHOOK_URL_PLACEHOLDER"
        }
      }
    ]
  }
}`
	content := strings.Replace(template, "WEBHOOK_URL_PLACEHOLDER", webhookURL, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprint(w, content)
	}))
	t.Cleanup(func() { srv.Close() })
	return srv
}
