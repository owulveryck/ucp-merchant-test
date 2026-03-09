package main

import (
	"testing"
)

func TestAp2MandateCompletion(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	paymentPayload := map[string]interface{}{
		"payment_data": map[string]interface{}{
			"id":           "instr_1",
			"handler_id":   "mock_payment_handler",
			"handler_name": "mock_payment_handler",
			"type":         "card",
			"brand":        "visa",
			"last_digits":  "4242",
			"credential": map[string]interface{}{
				"type":  "token",
				"token": "success_token",
			},
		},
		"risk_signals": map[string]interface{}{},
		"ap2": map[string]interface{}{
			"checkout_mandate": "header.payload.signature~kb_signature",
		},
	}

	resp, result := ts.doRequest("POST", "/shopping-api/checkout-sessions/"+checkoutID+"/complete", paymentPayload, ts.getHeaders(""))
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if result["status"] != "completed" {
		t.Fatalf("Expected status 'completed', got '%v'", result["status"])
	}
}
