package main

import (
	"fmt"
	"testing"
	"time"
)

func TestOrderRetrieval(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, true)
	checkoutID := data["id"].(string)
	_, completeData := ts.completeCheckout(checkoutID, nil)
	orderID := completeData["order"].(map[string]interface{})["id"].(string)

	resp, orderData := ts.doRequest("GET", "/orders/"+orderID, nil, ts.getHeaders(""))
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if orderData["id"] != orderID {
		t.Fatalf("Order ID mismatch: %v", orderData["id"])
	}
	if orderData["checkout_id"] != checkoutID {
		t.Fatalf("Checkout ID mismatch: %v", orderData["checkout_id"])
	}
}

func TestOrderFulfillmentRetrieval(t *testing.T) {
	ts := newTestServer(t)
	data := ts.createCheckoutSession("", 0, nil, false)
	checkoutID := data["id"].(string)

	// Update with address
	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type": "shipping",
				"destinations": []map[string]interface{}{
					{
						"id":               "dest_manual",
						"full_name":        "Jane Doe",
						"street_address":   "123 Main St",
						"address_locality": "Springfield",
						"address_region":   "IL",
						"postal_code":      "62704",
						"address_country":  "US",
					},
				},
				"selected_destination_id": "dest_manual",
			},
		},
	}

	_, updated := ts.updateCheckout(checkoutID, updatePayload, nil)
	options := getOptions(t, updated)
	if len(options) == 0 {
		t.Fatal("No options returned")
	}

	optionID := options[0]["id"].(string)
	optionTitle := options[0]["title"].(string)

	updatePayload["fulfillment"].(map[string]interface{})["methods"].([]map[string]interface{})[0]["groups"] = []map[string]interface{}{
		{"selected_option_id": optionID},
	}
	ts.updateCheckout(checkoutID, updatePayload, nil)

	// Complete
	_, completeData := ts.completeCheckout(checkoutID, nil)
	orderID := completeData["order"].(map[string]interface{})["id"].(string)

	// Get Order
	resp, orderData := ts.doRequest("GET", "/orders/"+orderID, nil, ts.getHeaders(""))
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	fulfillment := orderData["fulfillment"].(map[string]interface{})
	expectations := fulfillment["expectations"].([]interface{})
	if len(expectations) == 0 {
		t.Fatal("No expectations in order")
	}
	expect := expectations[0].(map[string]interface{})
	if expect["description"] != optionTitle {
		t.Fatalf("Expectation description mismatch: expected '%s', got '%v'", optionTitle, expect["description"])
	}
}

func TestOrderUpdate(t *testing.T) {
	ts := newTestServer(t)
	orderID := ts.createCompletedOrder()

	// Get order
	_, orderData := ts.doRequest("GET", "/orders/"+orderID, nil, ts.getHeaders(""))

	// Add shipment event
	lineItems := orderData["line_items"].([]interface{})
	var eventLineItems []map[string]interface{}
	for _, li := range lineItems {
		liMap := li.(map[string]interface{})
		qty := liMap["quantity"].(map[string]interface{})
		eventLineItems = append(eventLineItems, map[string]interface{}{
			"id":       liMap["id"],
			"quantity": qty["total"],
		})
	}

	newEvent := map[string]interface{}{
		"id":              fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		"occurred_at":     time.Now().UTC().Format(time.RFC3339),
		"type":            "shipped",
		"line_items":      eventLineItems,
		"tracking_number": "TRACK123",
		"tracking_url":    "http://track.me/123",
		"description":     "Shipped via FedEx",
	}

	fulfillment := orderData["fulfillment"].(map[string]interface{})
	events := []interface{}{newEvent}
	fulfillment["events"] = events
	orderData["fulfillment"] = fulfillment

	resp, updatedOrder := ts.doRequest("PUT", "/orders/"+orderID, orderData, ts.getHeaders(""))
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	updatedEvents := updatedOrder["fulfillment"].(map[string]interface{})["events"].([]interface{})
	if len(updatedEvents) == 0 {
		t.Fatal("No events returned")
	}
	firstEvent := updatedEvents[0].(map[string]interface{})
	if firstEvent["tracking_number"] != "TRACK123" {
		t.Fatalf("Order event not persisted: %v", firstEvent["tracking_number"])
	}
}

func TestOrderAdjustments(t *testing.T) {
	ts := newTestServer(t)
	orderID := ts.createCompletedOrder()

	_, orderData := ts.doRequest("GET", "/orders/"+orderID, nil, ts.getHeaders(""))

	adj := map[string]interface{}{
		"id":          fmt.Sprintf("adj_%d", time.Now().UnixNano()),
		"type":        "refund",
		"occurred_at": time.Now().UTC().Format(time.RFC3339),
		"status":      "pending",
		"amount":      500,
		"description": "Customer refund request",
	}

	orderData["adjustments"] = []interface{}{adj}

	resp, updatedOrder := ts.doRequest("PUT", "/orders/"+orderID, orderData, ts.getHeaders(""))
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	adjustments := updatedOrder["adjustments"].([]interface{})
	if len(adjustments) == 0 {
		t.Fatal("No adjustments returned")
	}
	firstAdj := adjustments[0].(map[string]interface{})
	if int(firstAdj["amount"].(float64)) != 500 {
		t.Fatalf("Expected amount 500, got %v", firstAdj["amount"])
	}
	if firstAdj["type"] != "refund" {
		t.Fatalf("Expected type 'refund', got '%v'", firstAdj["type"])
	}
}
