package main

import (
	"fmt"
	"testing"
	"time"
)

func TestInvalidAdjustmentStatus(t *testing.T) {
	ts := newTestServer(t)
	orderID := ts.createCompletedOrder()

	_, orderData := ts.doRequest("GET", "/orders/"+orderID, nil, ts.getHeaders(""))

	adj := map[string]interface{}{
		"id":          fmt.Sprintf("adj_%d", time.Now().UnixNano()),
		"type":        "refund",
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
		"status":      "INVALID_STATUS",
		"amount":      500,
	}

	orderData["adjustments"] = []interface{}{adj}

	resp, _ := ts.doRequest("PUT", "/orders/"+orderID, orderData, ts.getHeaders(""))
	if resp.StatusCode != 422 {
		t.Fatalf("Expected 422, got %d", resp.StatusCode)
	}
}

func TestUnknownDiscountCode(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["discounts"] = map[string]interface{}{"codes": []string{"INVALID_CODE_123"}}

	resp, updated := ts.updateCheckout(checkoutID, updatePayload, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	totals := updated["totals"].([]interface{})
	for _, tot := range totals {
		totMap := tot.(map[string]interface{})
		if totMap["type"] == "discount" {
			t.Fatal("Unknown discount code should not apply discount")
		}
	}
}

func TestMalformedAdjustmentPayload(t *testing.T) {
	ts := newTestServer(t)
	orderID := ts.createCompletedOrder()

	_, orderData := ts.doRequest("GET", "/orders/"+orderID, nil, ts.getHeaders(""))

	// Malform: dict instead of list
	orderData["adjustments"] = map[string]interface{}{"id": "adj_1", "amount": 100}

	resp, _ := ts.doRequest("PUT", "/orders/"+orderID, orderData, ts.getHeaders(""))
	if resp.StatusCode != 422 {
		t.Fatalf("Expected 422, got %d", resp.StatusCode)
	}
}
