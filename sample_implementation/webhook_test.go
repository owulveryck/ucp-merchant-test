package main

import (
	"testing"
	"time"
)

func TestWebhookEventStream(t *testing.T) {
	ts := newTestServer(t)
	wr := newWebhookRecorder(t)
	agentSrv := newAgentProfileServer(t, wr.server.URL+"/webhooks/partners/test_partner/events/order")

	// Create checkout with webhook-enabled agent profile
	payload := ts.createCheckoutPayload("", 0)
	headers := ts.getHeadersWithAgent("", agentSrv.URL+"/profiles/shopping-agent.json")
	resp, data := ts.doRequest("POST", "/shopping-api/checkout-sessions", payload, headers)
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		t.Fatalf("Expected 200/201, got %d", resp.StatusCode)
	}
	checkoutID := data["id"].(string)

	// Ensure fulfillment ready
	ts.ensureFulfillmentReady(checkoutID)

	// Complete checkout
	_, completeData := ts.completeCheckout(checkoutID, nil)
	orderID := completeData["order"].(map[string]interface{})["id"].(string)

	// Trigger shipping
	shipHeaders := ts.getHeaders("")
	shipHeaders["Simulation-Secret"] = simulationSecret
	shipResp, _ := ts.doRequest("POST", "/testing/simulate-shipping/"+orderID, nil, shipHeaders)
	if shipResp.StatusCode != 200 {
		t.Fatalf("Expected 200 from simulate-shipping, got %d", shipResp.StatusCode)
	}

	// Wait for webhook events
	events := wr.waitForEvents(2, 3*time.Second)
	if len(events) < 2 {
		t.Fatalf("Expected at least 2 webhook events, got %d", len(events))
	}

	// Find order_placed event
	var placedEvent map[string]interface{}
	for _, e := range events {
		if e["event_type"] == "order_placed" {
			placedEvent = e
			break
		}
	}
	if placedEvent == nil {
		t.Fatal("Missing order_placed event")
	}
	if placedEvent["checkout_id"] != checkoutID {
		t.Fatalf("order_placed checkout_id mismatch: %v", placedEvent["checkout_id"])
	}
	placedOrder := placedEvent["order"].(map[string]interface{})
	if placedOrder["id"] != orderID {
		t.Fatalf("order_placed order.id mismatch: %v", placedOrder["id"])
	}

	// Find order_shipped event
	var shippedEvent map[string]interface{}
	for _, e := range events {
		if e["event_type"] == "order_shipped" {
			shippedEvent = e
			break
		}
	}
	if shippedEvent == nil {
		t.Fatal("Missing order_shipped event")
	}
	if shippedEvent["checkout_id"] != checkoutID {
		t.Fatalf("order_shipped checkout_id mismatch: %v", shippedEvent["checkout_id"])
	}
	shippedOrder := shippedEvent["order"].(map[string]interface{})
	if shippedOrder["id"] != orderID {
		t.Fatalf("order_shipped order.id mismatch: %v", shippedOrder["id"])
	}
	fulfillment := shippedOrder["fulfillment"].(map[string]interface{})
	fulfillmentEvents, ok := fulfillment["events"].([]interface{})
	if !ok || len(fulfillmentEvents) == 0 {
		t.Fatal("order_shipped event missing fulfillment events")
	}
	hasShipped := false
	for _, fe := range fulfillmentEvents {
		feMap := fe.(map[string]interface{})
		if feMap["type"] == "shipped" {
			hasShipped = true
			break
		}
	}
	if !hasShipped {
		t.Fatal("order_shipped event did not contain shipped fulfillment event")
	}
}

func TestWebhookOrderAddressKnownCustomer(t *testing.T) {
	ts := newTestServer(t)
	wr := newWebhookRecorder(t)
	agentSrv := newAgentProfileServer(t, wr.server.URL+"/webhooks/partners/test_partner/events/order")

	buyer := map[string]interface{}{
		"full_name": "John Doe",
		"email":     "john.doe@example.com",
	}
	payload := ts.createCheckoutPayloadNoFulfillment("", 0)
	payload["buyer"] = buyer
	headers := ts.getHeadersWithAgent("", agentSrv.URL+"/profiles/shopping-agent.json")

	resp, data := ts.doRequest("POST", "/shopping-api/checkout-sessions", payload, headers)
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		t.Fatalf("Expected 200/201, got %d", resp.StatusCode)
	}
	checkoutID := data["id"].(string)

	// Trigger fulfillment update to inject address
	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{{"type": "shipping"}},
	}
	ts.updateCheckout(checkoutID, updatePayload, nil)

	// Fetch to get injected destinations
	_, checkoutData := ts.getCheckout(checkoutID)

	// Select destination
	f := checkoutData["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	if dests, ok := method["destinations"].([]interface{}); ok && len(dests) > 0 {
		destID := dests[0].(map[string]interface{})["id"].(string)
		updatePayload2 := buildUpdateFromCheckout(checkoutData)
		updatePayload2["fulfillment"] = map[string]interface{}{
			"methods": []map[string]interface{}{
				{"type": "shipping", "selected_destination_id": destID},
			},
		}
		ts.updateCheckout(checkoutID, updatePayload2, nil)

		// Fetch to get options
		_, checkoutData = ts.getCheckout(checkoutID)
		f = checkoutData["fulfillment"].(map[string]interface{})
		methods = f["methods"].([]interface{})
		method = methods[0].(map[string]interface{})
		if groups, ok := method["groups"].([]interface{}); ok && len(groups) > 0 {
			g := groups[0].(map[string]interface{})
			if options, ok := g["options"].([]interface{}); ok && len(options) > 0 {
				optionID := options[0].(map[string]interface{})["id"].(string)
				updatePayload3 := buildUpdateFromCheckout(checkoutData)
				updatePayload3["fulfillment"] = map[string]interface{}{
					"methods": []map[string]interface{}{
						{
							"type":                    "shipping",
							"selected_destination_id": destID,
							"groups":                  []map[string]interface{}{{"selected_option_id": optionID}},
						},
					},
				}
				ts.updateCheckout(checkoutID, updatePayload3, nil)
			}
		}
	}

	// Complete checkout
	_, completeData := ts.completeCheckout(checkoutID, nil)
	orderID := completeData["order"].(map[string]interface{})["id"].(string)

	events := wr.waitForEvents(1, 3*time.Second)
	var orderEvent map[string]interface{}
	for _, e := range events {
		if o, ok := e["order"].(map[string]interface{}); ok && o["id"] == orderID {
			orderEvent = e
			break
		}
	}
	if orderEvent == nil {
		t.Fatal("No webhook event for order")
	}
	order := orderEvent["order"].(map[string]interface{})
	fulfillment := order["fulfillment"].(map[string]interface{})
	expectations := fulfillment["expectations"].([]interface{})
	if len(expectations) == 0 {
		t.Fatal("No expectations in webhook order")
	}
	dest := expectations[0].(map[string]interface{})["destination"].(map[string]interface{})
	if dest["address_country"] != "US" {
		t.Fatalf("Expected address_country US, got %v", dest["address_country"])
	}
}

func TestWebhookOrderAddressNewAddress(t *testing.T) {
	ts := newTestServer(t)
	wr := newWebhookRecorder(t)
	agentSrv := newAgentProfileServer(t, wr.server.URL+"/webhooks/partners/test_partner/events/order")

	buyer := map[string]interface{}{
		"full_name": "John Doe",
		"email":     "john.doe@example.com",
	}
	payload := ts.createCheckoutPayloadNoFulfillment("", 0)
	payload["buyer"] = buyer
	headers := ts.getHeadersWithAgent("", agentSrv.URL+"/profiles/shopping-agent.json")

	resp, data := ts.doRequest("POST", "/shopping-api/checkout-sessions", payload, headers)
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		t.Fatalf("Expected 200/201, got %d", resp.StatusCode)
	}
	checkoutID := data["id"].(string)

	// Send new address
	newAddress := map[string]interface{}{
		"id":              "dest_new_webhook",
		"address_country": "CA",
		"postal_code":     "M5V 2H1",
		"street_address":  "Webhook St",
	}
	updatePayload := buildUpdateFromCheckout(data)
	updatePayload["fulfillment"] = map[string]interface{}{
		"methods": []map[string]interface{}{
			{
				"type":                    "shipping",
				"destinations":            []map[string]interface{}{newAddress},
				"selected_destination_id": "dest_new_webhook",
			},
		},
	}
	ts.updateCheckout(checkoutID, updatePayload, nil)

	// Fetch to get options
	_, checkoutData := ts.getCheckout(checkoutID)
	f := checkoutData["fulfillment"].(map[string]interface{})
	methods := f["methods"].([]interface{})
	method := methods[0].(map[string]interface{})
	if groups, ok := method["groups"].([]interface{}); ok && len(groups) > 0 {
		g := groups[0].(map[string]interface{})
		if options, ok := g["options"].([]interface{}); ok && len(options) > 0 {
			optionID := options[0].(map[string]interface{})["id"].(string)
			updatePayload2 := buildUpdateFromCheckout(checkoutData)
			updatePayload2["fulfillment"] = map[string]interface{}{
				"methods": []map[string]interface{}{
					{
						"type":                    "shipping",
						"selected_destination_id": "dest_new_webhook",
						"groups":                  []map[string]interface{}{{"selected_option_id": optionID}},
					},
				},
			}
			ts.updateCheckout(checkoutID, updatePayload2, nil)
		}
	}

	// Complete checkout
	_, completeData := ts.completeCheckout(checkoutID, nil)
	orderID := completeData["order"].(map[string]interface{})["id"].(string)

	events := wr.waitForEvents(1, 3*time.Second)
	var orderEvent map[string]interface{}
	for _, e := range events {
		if o, ok := e["order"].(map[string]interface{}); ok && o["id"] == orderID {
			orderEvent = e
			break
		}
	}
	if orderEvent == nil {
		t.Fatal("No webhook event for order")
	}
	order := orderEvent["order"].(map[string]interface{})
	fulfillment := order["fulfillment"].(map[string]interface{})
	expectations := fulfillment["expectations"].([]interface{})
	if len(expectations) == 0 {
		t.Fatal("No expectations in webhook order")
	}
	dest := expectations[0].(map[string]interface{})["destination"].(map[string]interface{})
	if dest["address_country"] != "CA" {
		t.Fatalf("Expected address_country CA, got %v", dest["address_country"])
	}
	if dest["street_address"] != "Webhook St" {
		t.Fatalf("Expected street_address 'Webhook St', got %v", dest["street_address"])
	}
}
