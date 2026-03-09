package main

import (
	"testing"
)

func TestTokenBindingCompletion(t *testing.T) {
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
				"type":  "stripe_token",
				"token": "success_token",
				"binding": map[string]interface{}{
					"checkout_id": checkoutID,
					"identity": map[string]interface{}{
						"access_token": "user_access_token",
					},
				},
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
