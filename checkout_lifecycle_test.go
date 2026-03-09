package main

import (
	"testing"
)

func TestCreateCheckout(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	if data["id"] == nil || data["id"] == "" {
		t.Fatal("Created checkout missing ID")
	}
}

func TestGetCheckout(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	resp, getData := ts.getCheckout(checkoutID)
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if getData["id"] != checkoutID {
		t.Fatalf("Get checkout returned wrong ID: %v", getData["id"])
	}
}

func TestUpdateCheckout(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["line_items"].([]map[string]interface{})[0]["quantity"] = 2

	resp, _ := ts.updateCheckout(checkoutID, updatePayload, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestCancelCheckout(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	resp, cancelData := ts.cancelCheckout(checkoutID)
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if cancelData["status"] != "canceled" {
		t.Fatalf("Expected status 'canceled', got '%v'", cancelData["status"])
	}
}

func TestCompleteCheckout(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	resp, completeData := ts.completeCheckout(checkoutID, nil)
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if completeData["status"] != "completed" {
		t.Fatalf("Expected status 'completed', got '%v'", completeData["status"])
	}
	order, ok := completeData["order"].(map[string]interface{})
	if !ok || order["id"] == nil {
		t.Fatal("order.id missing in completion response")
	}
	if order["permalink_url"] == nil || order["permalink_url"] == "" {
		t.Fatal("order.permalink_url missing")
	}
}

func TestCancelIsIdempotent(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	ts.cancelCheckout(checkoutID)

	resp, _ := ts.cancelCheckout(checkoutID)
	if resp.StatusCode == 200 {
		t.Fatal("Should not be able to cancel an already canceled checkout")
	}
}

func TestCannotUpdateCanceledCheckout(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	ts.cancelCheckout(checkoutID)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["line_items"].([]map[string]interface{})[0]["quantity"] = 2

	resp, _ := ts.updateCheckout(checkoutID, updatePayload, nil)
	if resp.StatusCode == 200 {
		t.Fatal("Should not be able to update a canceled checkout")
	}
}

func TestCannotCompleteCanceledCheckout(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	ts.cancelCheckout(checkoutID)

	resp, _ := ts.completeCheckout(checkoutID, nil)
	if resp.StatusCode == 200 {
		t.Fatal("Should not be able to complete a canceled checkout")
	}
}

func TestCompleteIsIdempotent(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	ts.completeCheckout(checkoutID, nil)

	resp, _ := ts.completeCheckout(checkoutID, nil)
	if resp.StatusCode == 200 {
		t.Fatal("Should not be able to complete an already completed checkout")
	}
}

func TestCannotUpdateCompletedCheckout(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	ts.completeCheckout(checkoutID, nil)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["line_items"].([]map[string]interface{})[0]["quantity"] = 2

	resp, _ := ts.updateCheckout(checkoutID, updatePayload, nil)
	if resp.StatusCode == 200 {
		t.Fatal("Should not be able to update a completed checkout")
	}
}

func TestCannotCancelCompletedCheckout(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	ts.completeCheckout(checkoutID, nil)

	resp, _ := ts.cancelCheckout(checkoutID)
	if resp.StatusCode == 200 {
		t.Fatal("Should not be able to cancel a completed checkout")
	}
}
