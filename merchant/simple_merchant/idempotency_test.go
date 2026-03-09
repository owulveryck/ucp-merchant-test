package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestIdempotencyCreate(t *testing.T) {
	ts := newTestServer(t)
	idemKey := fmt.Sprintf("idem-create-%d", time.Now().UnixNano())
	payload := ts.createCheckoutPayload("", 0)

	headers := ts.getHeaders(idemKey)
	resp1, data1 := ts.doRequest("POST", "/shopping-api/checkout-sessions", payload, headers)
	if resp1.StatusCode != 201 && resp1.StatusCode != 200 {
		t.Fatalf("Expected 200/201, got %d", resp1.StatusCode)
	}

	// Duplicate
	resp2, data2 := ts.doRequest("POST", "/shopping-api/checkout-sessions", payload, headers)
	if resp2.StatusCode != resp1.StatusCode {
		t.Fatalf("Idempotency Failed: status code mismatch %d vs %d", resp1.StatusCode, resp2.StatusCode)
	}
	if !jsonEqual(data1, data2) {
		t.Fatal("Idempotency Failed: response body mismatch")
	}

	// Conflict
	conflictPayload := ts.createCheckoutPayload("", 0)
	conflictPayload["currency"] = "EUR"
	resp3, _ := ts.doRequest("POST", "/shopping-api/checkout-sessions", conflictPayload, headers)
	if resp3.StatusCode != 409 {
		t.Fatalf("Expected 409 conflict, got %d", resp3.StatusCode)
	}
}

func TestIdempotencyUpdate(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	idemKey := fmt.Sprintf("idem-update-%d", time.Now().UnixNano())
	headers := ts.getHeaders(idemKey)

	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["line_items"].([]map[string]interface{})[0]["quantity"] = 2

	resp1, data1 := ts.doRequest("PUT", "/shopping-api/checkout-sessions/"+checkoutID, updatePayload, headers)
	if resp1.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp1.StatusCode)
	}

	// Duplicate
	resp2, data2 := ts.doRequest("PUT", "/shopping-api/checkout-sessions/"+checkoutID, updatePayload, headers)
	if resp2.StatusCode != 200 {
		t.Fatalf("Idempotency Update Failed: status code mismatch %d", resp2.StatusCode)
	}
	if !jsonEqual(data1, data2) {
		t.Fatal("Idempotency Update Failed: response body mismatch")
	}

	// Conflict
	conflictPayload := buildUpdateFromCheckout(data)
	conflictPayload["line_items"].([]map[string]interface{})[0]["quantity"] = 3
	resp3, _ := ts.doRequest("PUT", "/shopping-api/checkout-sessions/"+checkoutID, conflictPayload, headers)
	if resp3.StatusCode != 409 {
		t.Fatalf("Expected 409 conflict, got %d", resp3.StatusCode)
	}
}

func TestIdempotencyComplete(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	idemKey := fmt.Sprintf("idem-complete-%d", time.Now().UnixNano())
	headers := ts.getHeaders(idemKey)
	completePayload := ts.getValidPaymentPayload()

	resp1, data1 := ts.doRequest("POST", "/shopping-api/checkout-sessions/"+checkoutID+"/complete", completePayload, headers)
	if resp1.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp1.StatusCode)
	}

	// Duplicate
	resp2, data2 := ts.doRequest("POST", "/shopping-api/checkout-sessions/"+checkoutID+"/complete", completePayload, headers)
	if resp2.StatusCode != 200 {
		t.Fatalf("Idempotency Complete Failed: status code mismatch %d", resp2.StatusCode)
	}
	if !jsonEqual(data1, data2) {
		t.Fatal("Idempotency Complete Failed: response body mismatch")
	}

	// Conflict
	conflictPayload := ts.getValidPaymentPayload()
	conflictPayload["payment_data"].(map[string]interface{})["credential"].(map[string]interface{})["token"] = "different_token"
	resp3, _ := ts.doRequest("POST", "/shopping-api/checkout-sessions/"+checkoutID+"/complete", conflictPayload, headers)
	if resp3.StatusCode != 409 {
		t.Fatalf("Expected 409 conflict, got %d", resp3.StatusCode)
	}
}

func TestIdempotencyCancel(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)

	idemKey := fmt.Sprintf("idem-cancel-%d", time.Now().UnixNano())
	headers := ts.getHeaders(idemKey)

	resp1, data1 := ts.doRequest("POST", "/shopping-api/checkout-sessions/"+checkoutID+"/cancel", nil, headers)
	if resp1.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp1.StatusCode)
	}

	// Duplicate
	resp2, data2 := ts.doRequest("POST", "/shopping-api/checkout-sessions/"+checkoutID+"/cancel", nil, headers)
	if resp2.StatusCode != 200 {
		t.Fatalf("Idempotency Cancel Failed: status code mismatch %d", resp2.StatusCode)
	}
	if !jsonEqual(data1, data2) {
		t.Fatal("Idempotency Cancel Failed: response body mismatch")
	}
}

func jsonEqual(a, b map[string]interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	var aVal, bVal interface{}
	json.Unmarshal(aJSON, &aVal)
	json.Unmarshal(bJSON, &bVal)
	return reflect.DeepEqual(aVal, bVal)
}
