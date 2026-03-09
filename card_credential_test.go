package main

import (
	"testing"
)

func TestCardCredentialPayment(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	paymentPayload := map[string]interface{}{
		"payment_data": map[string]interface{}{
			"id":           "instr_card",
			"handler_id":   "mock_payment_handler",
			"handler_name": "mock_payment_handler",
			"type":         "card",
			"brand":        "Visa",
			"last_digits":  "1111",
			"credential": map[string]interface{}{
				"type":             "card",
				"card_number_type": "fpan",
				"number":           "4242424242424242",
				"expiry_month":     12,
				"expiry_year":      2030,
				"cvc":              "123",
				"name":             "John Doe",
			},
		},
		"risk_signals": map[string]interface{}{},
	}

	resp, result := ts.doRequest("POST", "/shopping-api/checkout-sessions/"+checkoutID+"/complete", paymentPayload, ts.getHeaders(""))
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if result["status"] != "completed" {
		t.Fatalf("Expected status 'completed', got '%v'", result["status"])
	}
}
