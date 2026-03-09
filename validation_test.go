package main

import (
	"strings"
	"testing"
)

func TestOutOfStock(t *testing.T) {
	ts := newTestServer(t)
	payload := ts.createCheckoutPayload("gardenias", 1)

	resp, _ := ts.createCheckout(payload)
	if resp.StatusCode != 400 {
		t.Fatalf("Expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateInventoryValidation(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["line_items"].([]map[string]interface{})[0]["quantity"] = 10001

	resp, _ := ts.updateCheckout(checkoutID, updatePayload, nil)
	if resp.StatusCode != 400 {
		t.Fatalf("Expected 400, got %d", resp.StatusCode)
	}
}

func TestProductNotFound(t *testing.T) {
	ts := newTestServer(t)
	payload := ts.createCheckoutPayload("pink_wumpus", 1)

	resp, data := ts.createCheckout(payload)
	if resp.StatusCode != 400 {
		t.Fatalf("Expected 400, got %d", resp.StatusCode)
	}
	detail, _ := data["detail"].(string)
	if !strings.Contains(strings.ToLower(detail), "not found") {
		t.Fatalf("Expected 'not found' message, got '%s'", detail)
	}
}

func TestPaymentFailure(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	resp, _ := ts.completeCheckout(checkoutID, ts.getFailPaymentPayload())
	if resp.StatusCode != 402 {
		t.Fatalf("Expected 402, got %d", resp.StatusCode)
	}
}

func TestCompleteWithoutFulfillment(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	resp, result := ts.completeCheckout(checkoutID, nil)
	if resp.StatusCode != 400 {
		t.Fatalf("Expected 400, got %d", resp.StatusCode)
	}
	detail, _ := result["detail"].(string)
	if !strings.Contains(detail, "Fulfillment address and option must be selected") {
		t.Fatalf("Expected fulfillment error message, got '%s'", detail)
	}
}

func TestStructuredErrorMessages(t *testing.T) {
	ts := newTestServer(t)
	payload := ts.createCheckoutPayload("gardenias", 1)

	resp, data := ts.createCheckout(payload)
	if resp.StatusCode != 400 {
		t.Fatalf("Expected 400, got %d", resp.StatusCode)
	}
	detail, ok := data["detail"].(string)
	if !ok || detail == "" {
		t.Fatal("Error response missing 'detail' field")
	}
	if !strings.Contains(detail, "Insufficient stock") {
		t.Fatalf("Expected 'Insufficient stock' in detail, got '%s'", detail)
	}
}
