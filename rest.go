package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// REST API stores (separate from MCP stores to avoid conflicts).
var (
	restCheckouts   = map[string]*RestCheckout{}
	restOrders      = map[string]*RestOrder{}
	restCheckoutSeq int
	restOrderSeq    int
	restStoreMu     sync.Mutex
	// Map checkout ID -> webhook URL for sending events.
	checkoutWebhooks = map[string]string{}
	// Map checkout ID -> selected fulfillment destination (for building order expectations).
	checkoutDestinations = map[string]*RestFulfillmentDestination{}
	// Map checkout ID -> selected fulfillment option title.
	checkoutOptionTitles = map[string]string{}
)

func writeJSONResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, detail string) {
	writeJSONResponse(w, status, map[string]string{"detail": detail})
}

// checkVersionNegotiation checks the UCP-Agent header for version compatibility.
// Returns true if the request should be rejected.
func checkVersionNegotiation(w http.ResponseWriter, r *http.Request) bool {
	ucpAgent := r.Header.Get("UCP-Agent")
	if ucpAgent == "" {
		return false
	}
	// Parse version="YYYY-MM-DD" from the header
	for _, part := range strings.Split(ucpAgent, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "version=") {
			version := strings.Trim(strings.TrimPrefix(part, "version="), "\"")
			if version != "" && version != "2026-01-11" {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("Incompatible UCP version: %s. Expected 2026-01-11", version))
				return true
			}
		}
	}
	return false
}

// handleIdempotency handles idempotency key checking.
// Returns (shouldReturn, cachedBody, cachedStatus).
func handleIdempotency(w http.ResponseWriter, r *http.Request, body []byte) bool {
	key := r.Header.Get("idempotency-key")
	if key == "" {
		return false
	}
	payloadHash := hashPayload(body)
	entry, exists := checkIdempotency(key)
	if !exists {
		return false
	}
	if entry.PayloadHash == payloadHash {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(entry.StatusCode)
		w.Write(entry.ResponseBody)
		return true
	}
	writeError(w, http.StatusConflict, "Idempotency key conflict: payload differs from original request")
	return true
}

func storeIdempotentResponse(r *http.Request, body []byte, statusCode int, responseBody []byte) {
	key := r.Header.Get("idempotency-key")
	if key == "" {
		return
	}
	storeIdempotency(key, hashPayload(body), statusCode, responseBody)
}

func processAndRespond(w http.ResponseWriter, r *http.Request, reqBody []byte, status int, result interface{}) {
	respBody, _ := json.Marshal(result)
	storeIdempotentResponse(r, reqBody, status, respBody)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(respBody)
}

// restCreateCheckout handles POST /shopping-api/checkout-sessions
func restCreateCheckout(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if checkVersionNegotiation(w, r) {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	if handleIdempotency(w, r, body) {
		return
	}

	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	restStoreMu.Lock()
	defer restStoreMu.Unlock()

	// Parse line items
	lineItems, err := restBuildLineItems(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	restCheckoutSeq++
	coID := fmt.Sprintf("co_%04d", restCheckoutSeq)

	co := &RestCheckout{
		ID:        coID,
		Status:    "incomplete",
		UCP:       RestUCP{Version: "2026-01-11", Capabilities: []RestCapability{}},
		Links:     []RestLink{{Type: "application/json", URL: fmt.Sprintf("%s://localhost:%d/shopping-api/checkout-sessions/%s", scheme(), listenPort, coID)}},
		Currency:  stringOr(req, "currency", "USD"),
		LineItems: lineItems,
	}

	// Calculate totals
	co.Totals = restCalculateTotals(lineItems, 0, nil)

	// Parse payment from request
	co.Payment = restParsePayment(req)

	// Parse fulfillment from request
	co.Fulfillment = restParseFulfillment(req, nil, co)

	// Parse buyer
	co.Buyer = restParseBuyer(req)

	// Store webhook URL from UCP-Agent header
	webhookURL := resolveWebhookURL(r.Header.Get("UCP-Agent"))
	if webhookURL != "" {
		checkoutWebhooks[coID] = webhookURL
	}

	restCheckouts[coID] = co

	processAndRespond(w, r, body, http.StatusCreated, co)
}

// restGetCheckout handles GET /shopping-api/checkout-sessions/{id}
func restGetCheckout(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	id := extractPathParam(r.URL.Path, "/shopping-api/checkout-sessions/")
	// Strip trailing path parts like /complete or /cancel
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}

	restStoreMu.Lock()
	defer restStoreMu.Unlock()

	co, ok := restCheckouts[id]
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Checkout not found: %s", id))
		return
	}

	writeJSONResponse(w, http.StatusOK, co)
}

// restUpdateCheckout handles PUT /shopping-api/checkout-sessions/{id}
func restUpdateCheckout(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if checkVersionNegotiation(w, r) {
		return
	}

	id := extractPathParam(r.URL.Path, "/shopping-api/checkout-sessions/")
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	if handleIdempotency(w, r, body) {
		return
	}

	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	restStoreMu.Lock()
	defer restStoreMu.Unlock()

	co, ok := restCheckouts[id]
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Checkout not found: %s", id))
		return
	}

	if co.Status == "completed" || co.Status == "canceled" {
		writeError(w, http.StatusConflict, fmt.Sprintf("Cannot update checkout in %s status", co.Status))
		return
	}

	// Update line items if provided
	if rawItems, ok := req["line_items"]; ok && rawItems != nil {
		lineItems, err := restBuildLineItems(req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		co.LineItems = lineItems
	}

	// Update buyer if provided
	if _, ok := req["buyer"]; ok {
		co.Buyer = restParseBuyer(req)
	}

	// Update payment if provided
	if _, ok := req["payment"]; ok {
		co.Payment = restParsePayment(req)
	}

	// Handle discounts
	shippingCost := restGetCurrentShippingCost(co)
	if discountsRaw, ok := req["discounts"]; ok && discountsRaw != nil {
		co.Discounts = restApplyDiscounts(discountsRaw, co.LineItems)
	}

	// Handle fulfillment
	if fulfillmentRaw, ok := req["fulfillment"]; ok && fulfillmentRaw != nil {
		co.Fulfillment = restParseFulfillment(req, co.Buyer, co)
		shippingCost = restGetCurrentShippingCost(co)
	}

	// Recalculate totals
	co.Totals = restCalculateTotals(co.LineItems, shippingCost, co.Discounts)

	processAndRespond(w, r, body, http.StatusOK, co)
}

// restCompleteCheckout handles POST /shopping-api/checkout-sessions/{id}/complete
func restCompleteCheckout(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Extract ID: /shopping-api/checkout-sessions/{id}/complete
	path := r.URL.Path
	prefix := "/shopping-api/checkout-sessions/"
	suffix := "/complete"
	id := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	if handleIdempotency(w, r, body) {
		return
	}

	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	restStoreMu.Lock()
	defer restStoreMu.Unlock()

	co, ok := restCheckouts[id]
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Checkout not found: %s", id))
		return
	}

	if co.Status == "completed" {
		writeError(w, http.StatusConflict, "Checkout already completed")
		return
	}
	if co.Status == "canceled" {
		writeError(w, http.StatusConflict, "Checkout has been canceled")
		return
	}

	// Validate fulfillment is complete
	if !restIsFulfillmentComplete(co) {
		writeError(w, http.StatusBadRequest, "Fulfillment address and option must be selected")
		return
	}

	// Process payment
	paymentData, _ := req["payment_data"].(map[string]interface{})
	if paymentData != nil {
		// Check for token-based payment
		credential, _ := paymentData["credential"].(map[string]interface{})
		if credential != nil {
			token, _ := credential["token"].(string)
			if token == "fail_token" {
				writeError(w, http.StatusPaymentRequired, "Payment failed")
				return
			}
		}
		// Check handler_id exists
		handlerID, _ := paymentData["handler_id"].(string)
		if handlerID != "" {
			validHandlers := map[string]bool{
				"google_pay":           true,
				"mock_payment_handler": true,
				"shop_pay":             true,
			}
			if !validHandlers[handlerID] {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("Unknown payment handler: %s", handlerID))
				return
			}
		}
	}

	// Create order
	restOrderSeq++
	orderID := fmt.Sprintf("order_%04d", restOrderSeq)

	// Build order line items
	var orderLineItems []RestOrderLineItem
	var expectationLineItems []RestEventLineItem
	for _, li := range co.LineItems {
		orderLineItems = append(orderLineItems, RestOrderLineItem{
			ID:       li.ID,
			Item:     li.Item,
			Quantity: RestOrderQuantity{Total: li.Quantity, Fulfilled: 0},
			Totals:   li.Totals,
			Status:   "processing",
		})
		expectationLineItems = append(expectationLineItems, RestEventLineItem{
			ID:       li.ID,
			Quantity: li.Quantity,
		})
	}

	// Build fulfillment expectations from selected option
	var expectations []RestExpectation
	optionTitle := checkoutOptionTitles[id]
	if optionTitle == "" {
		optionTitle = "Standard Shipping"
	}
	dest := checkoutDestinations[id]
	destVal := RestFulfillmentDestination{}
	if dest != nil {
		destVal = *dest
	}
	expectations = append(expectations, RestExpectation{
		ID:          "expect_1",
		LineItems:   expectationLineItems,
		MethodType:  "shipping",
		Description: optionTitle,
		Destination: destVal,
	})

	order := &RestOrder{
		ID:           orderID,
		UCP:          RestUCP{Version: "2026-01-11", Capabilities: []RestCapability{}},
		CheckoutID:   id,
		PermalinkURL: fmt.Sprintf("%s://localhost:%d/orders/%s", scheme(), listenPort, orderID),
		LineItems:    orderLineItems,
		Fulfillment: RestOrderFulfillment{
			Expectations: expectations,
		},
		Currency: co.Currency,
		Totals:   co.Totals,
	}

	restOrders[orderID] = order

	co.Status = "completed"
	co.Order = &RestOrderRef{
		ID:           orderID,
		PermalinkURL: order.PermalinkURL,
	}

	// Send webhook
	if webhookURL, ok := checkoutWebhooks[id]; ok {
		orderJSON, _ := json.Marshal(order)
		var orderMap map[string]interface{}
		json.Unmarshal(orderJSON, &orderMap)
		sendWebhookEvent(webhookURL, map[string]interface{}{
			"event_type":  "order_placed",
			"checkout_id": id,
			"order":       orderMap,
		})
	}

	processAndRespond(w, r, body, http.StatusOK, co)
}

// restCancelCheckout handles POST /shopping-api/checkout-sessions/{id}/cancel
func restCancelCheckout(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := r.URL.Path
	prefix := "/shopping-api/checkout-sessions/"
	suffix := "/cancel"
	id := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)

	body, _ := io.ReadAll(r.Body)

	if handleIdempotency(w, r, body) {
		return
	}

	restStoreMu.Lock()
	defer restStoreMu.Unlock()

	co, ok := restCheckouts[id]
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Checkout not found: %s", id))
		return
	}

	if co.Status == "canceled" {
		writeError(w, http.StatusConflict, "Checkout already canceled")
		return
	}
	if co.Status == "completed" {
		writeError(w, http.StatusConflict, "Cannot cancel completed checkout")
		return
	}

	co.Status = "canceled"

	processAndRespond(w, r, body, http.StatusOK, co)
}

// restGetOrder handles GET /orders/{id}
func restGetOrder(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	id := extractPathParam(r.URL.Path, "/orders/")

	restStoreMu.Lock()
	defer restStoreMu.Unlock()

	order, ok := restOrders[id]
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Order not found: %s", id))
		return
	}

	writeJSONResponse(w, http.StatusOK, order)
}

// restUpdateOrder handles PUT /orders/{id}
func restUpdateOrder(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	id := extractPathParam(r.URL.Path, "/orders/")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	// Validate the body can parse into our order model
	var reqMap map[string]interface{}
	if err := json.Unmarshal(body, &reqMap); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Invalid JSON")
		return
	}

	// Validate adjustments
	if adjRaw, ok := reqMap["adjustments"]; ok && adjRaw != nil {
		adjList, ok := adjRaw.([]interface{})
		if !ok {
			writeError(w, http.StatusUnprocessableEntity, "adjustments must be a list")
			return
		}
		validStatuses := map[string]bool{"pending": true, "approved": true, "rejected": true, "completed": true}
		for _, a := range adjList {
			adjMap, ok := a.(map[string]interface{})
			if !ok {
				writeError(w, http.StatusUnprocessableEntity, "Invalid adjustment format")
				return
			}
			status, _ := adjMap["status"].(string)
			if status != "" && !validStatuses[status] {
				writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("Invalid adjustment status: %s", status))
				return
			}
		}
	}

	restStoreMu.Lock()
	defer restStoreMu.Unlock()

	order, ok := restOrders[id]
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Order not found: %s", id))
		return
	}

	// Update fulfillment events
	if fulfillmentRaw, ok := reqMap["fulfillment"]; ok {
		fMap, _ := fulfillmentRaw.(map[string]interface{})
		if fMap != nil {
			if eventsRaw, ok := fMap["events"]; ok {
				eventsJSON, _ := json.Marshal(eventsRaw)
				var events []RestFulfillmentEvent
				json.Unmarshal(eventsJSON, &events)
				order.Fulfillment.Events = events
			}
			if expectRaw, ok := fMap["expectations"]; ok {
				expectJSON, _ := json.Marshal(expectRaw)
				var expectations []RestExpectation
				json.Unmarshal(expectJSON, &expectations)
				order.Fulfillment.Expectations = expectations
			}
		}
	}

	// Update adjustments
	if adjRaw, ok := reqMap["adjustments"]; ok {
		adjJSON, _ := json.Marshal(adjRaw)
		var adjustments []RestAdjustment
		json.Unmarshal(adjJSON, &adjustments)
		order.Adjustments = adjustments
	}

	writeJSONResponse(w, http.StatusOK, order)
}

// restSimulateShipping handles POST /testing/simulate-shipping/{id}
func restSimulateShipping(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Check simulation secret
	secret := r.Header.Get("Simulation-Secret")
	if secret != simulationSecret {
		writeError(w, http.StatusForbidden, "Invalid or missing simulation secret")
		return
	}

	id := extractPathParam(r.URL.Path, "/testing/simulate-shipping/")

	restStoreMu.Lock()
	defer restStoreMu.Unlock()

	order, ok := restOrders[id]
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Order not found: %s", id))
		return
	}

	// Add shipped event
	event := RestFulfillmentEvent{
		ID:             fmt.Sprintf("evt_ship_%s", id),
		OccurredAt:     time.Now().UTC().Format(time.RFC3339),
		Type:           "shipped",
		TrackingNumber: fmt.Sprintf("TRK-%s", id),
		Description:    "Order shipped",
	}

	// Add line items to event
	for _, li := range order.LineItems {
		event.LineItems = append(event.LineItems, RestEventLineItem{
			ID:       li.ID,
			Quantity: li.Quantity.Total,
		})
	}

	order.Fulfillment.Events = append(order.Fulfillment.Events, event)

	// Send webhook
	if webhookURL, ok := checkoutWebhooks[order.CheckoutID]; ok {
		orderJSON, _ := json.Marshal(order)
		var orderMap map[string]interface{}
		json.Unmarshal(orderJSON, &orderMap)
		sendWebhookEvent(webhookURL, map[string]interface{}{
			"event_type":  "order_shipped",
			"checkout_id": order.CheckoutID,
			"order":       orderMap,
		})
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "shipped"})
}

// Helper functions

func extractPathParam(path, prefix string) string {
	s := strings.TrimPrefix(path, prefix)
	// Remove trailing slash
	s = strings.TrimSuffix(s, "/")
	return s
}

func stringOr(m map[string]interface{}, key, def string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return def
}

func restBuildLineItems(req map[string]interface{}) ([]RestLineItem, error) {
	rawItems, _ := req["line_items"].([]interface{})
	if len(rawItems) == 0 {
		return nil, fmt.Errorf("line_items is required")
	}

	var items []RestLineItem
	for i, raw := range rawItems {
		rawMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract item ID
		itemID := ""
		if itemMap, ok := rawMap["item"].(map[string]interface{}); ok {
			itemID, _ = itemMap["id"].(string)
		}
		if itemID == "" {
			return nil, fmt.Errorf("line item %d: missing item.id", i)
		}

		product := findProduct(itemID)
		if product == nil {
			return nil, fmt.Errorf("Product not found: %s", itemID)
		}

		qty := 1
		if q, ok := rawMap["quantity"].(float64); ok {
			qty = int(q)
		}
		if qty < 1 {
			qty = 1
		}

		// Check stock
		if product.Quantity <= 0 {
			return nil, fmt.Errorf("Insufficient stock for product %s", itemID)
		}
		if qty > product.Quantity {
			return nil, fmt.Errorf("Insufficient stock for product %s: requested %d, available %d", itemID, qty, product.Quantity)
		}

		lineTotal := product.Price * qty

		liID := fmt.Sprintf("li_%03d", i+1)
		if existingID, ok := rawMap["id"].(string); ok && existingID != "" {
			liID = existingID
		}

		items = append(items, RestLineItem{
			ID: liID,
			Item: RestItem{
				ID:       product.ID,
				Title:    product.Title,
				Price:    product.Price,
				ImageURL: product.ImageURL,
			},
			Quantity: qty,
			Totals: []Total{
				{Type: "subtotal", Amount: lineTotal},
				{Type: "total", Amount: lineTotal},
			},
		})
	}
	return items, nil
}

func restCalculateTotals(items []RestLineItem, shippingCost int, discounts *RestDiscounts) []Total {
	subtotal := 0
	for _, li := range items {
		for _, t := range li.Totals {
			if t.Type == "subtotal" {
				subtotal += t.Amount
			}
		}
	}

	total := subtotal

	var totals []Total
	totals = append(totals, Total{
		Type:        "subtotal",
		DisplayText: fmt.Sprintf("$%.2f", float64(subtotal)/100),
		Amount:      subtotal,
	})

	// Apply discounts
	if discounts != nil {
		discountAmount := 0
		for _, d := range discounts.Applied {
			discountAmount += d.Amount
		}
		if discountAmount > 0 {
			total -= discountAmount
			totals = append(totals, Total{
				Type:        "discount",
				DisplayText: fmt.Sprintf("-$%.2f", float64(discountAmount)/100),
				Amount:      discountAmount,
			})
		}
	}

	if shippingCost > 0 {
		total += shippingCost
		totals = append(totals, Total{
			Type:        "fulfillment",
			DisplayText: fmt.Sprintf("$%.2f", float64(shippingCost)/100),
			Amount:      shippingCost,
		})
	}

	totals = append(totals, Total{
		Type:        "total",
		DisplayText: fmt.Sprintf("$%.2f", float64(total)/100),
		Amount:      total,
	})

	return totals
}

func restParsePayment(req map[string]interface{}) RestPayment {
	paymentRaw, ok := req["payment"]
	if !ok || paymentRaw == nil {
		return restDefaultPayment()
	}

	paymentMap, ok := paymentRaw.(map[string]interface{})
	if !ok {
		return restDefaultPayment()
	}

	p := &RestPayment{}

	if sid, ok := paymentMap["selected_instrument_id"].(string); ok {
		p.SelectedInstrumentID = sid
	}

	// Parse instruments
	if instRaw, ok := paymentMap["instruments"].([]interface{}); ok {
		for _, inst := range instRaw {
			if m, ok := inst.(map[string]interface{}); ok {
				p.Instruments = append(p.Instruments, m)
			}
		}
	}
	if p.Instruments == nil {
		p.Instruments = []map[string]interface{}{}
	}

	// Parse handlers
	if hRaw, ok := paymentMap["handlers"].([]interface{}); ok {
		for _, h := range hRaw {
			if m, ok := h.(map[string]interface{}); ok {
				p.Handlers = append(p.Handlers, m)
			}
		}
	}
	if p.Handlers == nil {
		p.Handlers = restDefaultHandlers()
	}

	return *p
}

func restDefaultPayment() RestPayment {
	return RestPayment{
		SelectedInstrumentID: "instr_1",
		Instruments:          []map[string]interface{}{},
		Handlers:             restDefaultHandlers(),
	}
}

func restDefaultHandlers() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":                 "google_pay",
			"name":               "google.pay",
			"version":            "2026-01-11",
			"spec":               "https://ucp.dev/specs/payment/google_pay",
			"config_schema":      "https://ucp.dev/schemas/payment/google_pay.json",
			"instrument_schemas": []string{"https://ucp.dev/schemas/payment/google_pay_instrument.json"},
			"config":             map[string]interface{}{},
		},
		{
			"id":                 "mock_payment_handler",
			"name":               "mock_payment_handler",
			"version":            "2026-01-11",
			"spec":               "https://ucp.dev/specs/payment/mock",
			"config_schema":      "https://ucp.dev/schemas/payment/mock.json",
			"instrument_schemas": []string{"https://ucp.dev/schemas/payment/mock_instrument.json"},
			"config":             map[string]interface{}{},
		},
		{
			"id":                 "shop_pay",
			"name":               "com.shopify.shop_pay",
			"version":            "2026-01-11",
			"spec":               "https://ucp.dev/specs/payment/shop_pay",
			"config_schema":      "https://ucp.dev/schemas/payment/shop_pay.json",
			"instrument_schemas": []string{"https://ucp.dev/schemas/payment/shop_pay_instrument.json"},
			"config":             map[string]interface{}{"shop_id": "merchant_1"},
		},
	}
}

func restParseBuyer(req map[string]interface{}) *RestBuyer {
	buyerRaw, ok := req["buyer"]
	if !ok || buyerRaw == nil {
		return nil
	}
	buyerMap, ok := buyerRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	b := &RestBuyer{}
	if v, ok := buyerMap["first_name"].(string); ok {
		b.FirstName = v
	}
	if v, ok := buyerMap["last_name"].(string); ok {
		b.LastName = v
	}
	if v, ok := buyerMap["fullName"].(string); ok {
		b.FullName = v
	}
	if v, ok := buyerMap["email"].(string); ok {
		b.Email = v
	}

	// Parse consent
	if consentRaw, ok := buyerMap["consent"].(map[string]interface{}); ok {
		c := &RestConsent{}
		if v, ok := consentRaw["marketing"].(bool); ok {
			c.Marketing = &v
		}
		if v, ok := consentRaw["analytics"].(bool); ok {
			c.Analytics = &v
		}
		if v, ok := consentRaw["sale_of_data"].(bool); ok {
			c.SaleOfData = &v
		}
		b.Consent = c
	}

	return b
}

func restParseFulfillment(req map[string]interface{}, buyer *RestBuyer, co *RestCheckout) *RestFulfillment {
	fulfillmentRaw, ok := req["fulfillment"]
	if !ok || fulfillmentRaw == nil {
		return nil
	}
	fMap, ok := fulfillmentRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	methodsRaw, _ := fMap["methods"].([]interface{})
	if len(methodsRaw) == 0 {
		return nil
	}

	f := &RestFulfillment{}
	for _, mRaw := range methodsRaw {
		mData, ok := mRaw.(map[string]interface{})
		if !ok {
			continue
		}
		method := RestFulfillmentMethod{}
		if v, ok := mData["id"].(string); ok {
			method.ID = v
		} else {
			method.ID = "method_shipping"
		}
		if v, ok := mData["type"].(string); ok {
			method.Type = v
		}
		// Collect line item IDs
		if co != nil {
			for _, li := range co.LineItems {
				method.LineItemIDs = append(method.LineItemIDs, li.ID)
			}
		}
		if method.LineItemIDs == nil {
			method.LineItemIDs = []string{}
		}

		// Parse destinations
		destsRaw, _ := mData["destinations"].([]interface{})
		if len(destsRaw) > 0 {
			for _, dRaw := range destsRaw {
				dMap, ok := dRaw.(map[string]interface{})
				if !ok {
					continue
				}
				dest := parseDestination(dMap, buyer)
				method.Destinations = append(method.Destinations, dest)
			}
		} else if method.Type == "shipping" {
			// Address injection: look up known customer addresses
			email := ""
			if buyer != nil {
				email = buyer.Email
			} else if co != nil && co.Buyer != nil {
				email = co.Buyer.Email
			}
			if email != "" {
				addresses := findAddressesForEmail(email)
				for _, addr := range addresses {
					method.Destinations = append(method.Destinations, RestFulfillmentDestination{
						ID:              addr.ID,
						StreetAddress:   addr.StreetAddress,
						AddressLocality: addr.City,
						AddressRegion:   addr.State,
						PostalCode:      addr.PostalCode,
						AddressCountry:  addr.Country,
					})
				}
			}
			if len(method.Destinations) == 0 {
				// No destinations to inject
				method.Destinations = nil
			}
		}

		// Selected destination
		if v, ok := mData["selected_destination_id"].(string); ok {
			method.SelectedDestinationID = v

			// Preserve existing destinations if we have them from a previous update
			if len(method.Destinations) == 0 && co != nil && co.Fulfillment != nil && len(co.Fulfillment.Methods) > 0 {
				method.Destinations = co.Fulfillment.Methods[0].Destinations
			}

			// Store the selected destination for order creation
			if co != nil {
				for _, d := range method.Destinations {
					if d.ID == v {
						dest := d
						checkoutDestinations[co.ID] = &dest
						break
					}
				}
			}

			// Generate shipping options based on destination country
			destCountry := ""
			for _, d := range method.Destinations {
				if d.ID == v {
					destCountry = d.AddressCountry
					break
				}
			}
			if destCountry != "" {
				options := restGenerateShippingOptions(destCountry, co)
				groupLineItemIDs := method.LineItemIDs
				if groupLineItemIDs == nil {
					groupLineItemIDs = []string{}
				}
				method.Groups = []RestFulfillmentGroup{
					{ID: "group_1", LineItemIDs: groupLineItemIDs, Options: options},
				}
			}
		}

		// Parse groups (for selected_option_id)
		if groupsRaw, ok := mData["groups"].([]interface{}); ok && len(groupsRaw) > 0 {
			// Preserve existing options if we have them
			existingOptions := []RestFulfillmentOption{}
			if co != nil && co.Fulfillment != nil && len(co.Fulfillment.Methods) > 0 {
				existingMethod := co.Fulfillment.Methods[0]
				if len(existingMethod.Groups) > 0 {
					existingOptions = existingMethod.Groups[0].Options
				}
				// Also preserve destinations and selection
				if len(method.Destinations) == 0 {
					method.Destinations = existingMethod.Destinations
				}
				if method.SelectedDestinationID == "" {
					method.SelectedDestinationID = existingMethod.SelectedDestinationID
				}
			}

			for gi, gRaw := range groupsRaw {
				gMap, ok := gRaw.(map[string]interface{})
				if !ok {
					continue
				}
				groupLineItemIDs := method.LineItemIDs
				if groupLineItemIDs == nil {
					groupLineItemIDs = []string{}
				}
				group := RestFulfillmentGroup{
					ID:          fmt.Sprintf("group_%d", gi+1),
					LineItemIDs: groupLineItemIDs,
				}
				if v, ok := gMap["selected_option_id"].(string); ok {
					group.SelectedOptionID = v
					// Store the selected option title for order expectations
					if co != nil {
						for _, opt := range existingOptions {
							if opt.ID == v {
								checkoutOptionTitles[co.ID] = opt.Title
								break
							}
						}
					}
				}
				group.Options = existingOptions
				method.Groups = append(method.Groups, group)
			}
		}

		f.Methods = append(f.Methods, method)
	}

	return f
}

func parseDestination(dMap map[string]interface{}, buyer *RestBuyer) RestFulfillmentDestination {
	dest := RestFulfillmentDestination{}
	if v, ok := dMap["id"].(string); ok {
		dest.ID = v
	}
	if v, ok := dMap["full_name"].(string); ok {
		dest.FullName = v
	}
	if v, ok := dMap["street_address"].(string); ok {
		dest.StreetAddress = v
	}
	if v, ok := dMap["address_locality"].(string); ok {
		dest.AddressLocality = v
	}
	if v, ok := dMap["address_region"].(string); ok {
		dest.AddressRegion = v
	}
	if v, ok := dMap["postal_code"].(string); ok {
		dest.PostalCode = v
	}
	if v, ok := dMap["address_country"].(string); ok {
		dest.AddressCountry = v
	}

	// If no ID provided, try to match existing or generate one
	if dest.ID == "" {
		email := ""
		if buyer != nil {
			email = buyer.Email
		}
		if email != "" {
			existingAddrs := findAddressesForEmail(email)
			matched := matchExistingAddress(existingAddrs, dest.StreetAddress, dest.AddressLocality, dest.AddressRegion, dest.PostalCode, dest.AddressCountry)
			if matched != nil {
				dest.ID = matched.ID
			} else {
				// Generate new ID and save
				addrSeqMu.Lock()
				addrSeqCounter++
				dest.ID = fmt.Sprintf("addr_dyn_%d", addrSeqCounter)
				addrSeqMu.Unlock()
				saveDynamicAddress(email, CSVAddress{
					ID:            dest.ID,
					StreetAddress: dest.StreetAddress,
					City:          dest.AddressLocality,
					State:         dest.AddressRegion,
					PostalCode:    dest.PostalCode,
					Country:       dest.AddressCountry,
				})
			}
		} else {
			addrSeqMu.Lock()
			addrSeqCounter++
			dest.ID = fmt.Sprintf("addr_dyn_%d", addrSeqCounter)
			addrSeqMu.Unlock()
		}
	}

	return dest
}

var (
	addrSeqCounter int
	addrSeqMu      sync.Mutex
)

func restGenerateShippingOptions(country string, co *RestCheckout) []RestFulfillmentOption {
	rates := getShippingRatesForCountry(country)
	var options []RestFulfillmentOption

	// Check promotions for free shipping
	freeShipping := false
	if co != nil {
		subtotal := 0
		var itemIDs []string
		for _, li := range co.LineItems {
			for _, t := range li.Totals {
				if t.Type == "subtotal" {
					subtotal += t.Amount
				}
			}
			itemIDs = append(itemIDs, li.Item.ID)
		}
		for _, promo := range shopData.Promotions {
			if promo.Type == "free_shipping" {
				if promo.MinSubtotal > 0 && subtotal >= promo.MinSubtotal {
					freeShipping = true
					break
				}
				if len(promo.EligibleItemIDs) > 0 {
					for _, eligible := range promo.EligibleItemIDs {
						for _, itemID := range itemIDs {
							if eligible == itemID {
								freeShipping = true
								break
							}
						}
						if freeShipping {
							break
						}
					}
				}
			}
			if freeShipping {
				break
			}
		}
	}

	for _, rate := range rates {
		price := rate.Price
		title := rate.Title
		if freeShipping && rate.ServiceLevel == "standard" {
			price = 0
			title = "Free Standard Shipping"
		}
		options = append(options, RestFulfillmentOption{
			ID:    rate.ID,
			Title: title,
			Totals: []Total{
				{Type: "fulfillment", Amount: price},
				{Type: "total", Amount: price},
			},
		})
	}
	return options
}

func restGetCurrentShippingCost(co *RestCheckout) int {
	if co.Fulfillment == nil {
		return 0
	}
	for _, m := range co.Fulfillment.Methods {
		for _, g := range m.Groups {
			if g.SelectedOptionID != "" {
				for _, opt := range g.Options {
					if opt.ID == g.SelectedOptionID {
						for _, t := range opt.Totals {
							if t.Type == "total" {
								return t.Amount
							}
						}
					}
				}
			}
		}
	}
	return 0
}

func restIsFulfillmentComplete(co *RestCheckout) bool {
	if co.Fulfillment == nil {
		return false
	}
	for _, m := range co.Fulfillment.Methods {
		if m.SelectedDestinationID == "" {
			return false
		}
		hasOption := false
		for _, g := range m.Groups {
			if g.SelectedOptionID != "" {
				hasOption = true
				break
			}
		}
		if !hasOption {
			return false
		}
	}
	return len(co.Fulfillment.Methods) > 0
}

func restApplyDiscounts(discountsRaw interface{}, lineItems []RestLineItem) *RestDiscounts {
	dMap, ok := discountsRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	codesRaw, _ := dMap["codes"].([]interface{})
	if len(codesRaw) == 0 {
		return nil
	}

	// Calculate subtotal
	subtotal := 0
	for _, li := range lineItems {
		for _, t := range li.Totals {
			if t.Type == "subtotal" {
				subtotal += t.Amount
			}
		}
	}

	result := &RestDiscounts{}
	for _, cRaw := range codesRaw {
		code, _ := cRaw.(string)
		if code == "" {
			continue
		}
		result.Codes = append(result.Codes, code)

		discount := findDiscountByCode(code)
		if discount == nil {
			continue // Unknown codes are silently ignored
		}

		var amount int
		switch discount.Type {
		case "percentage":
			amount = subtotal * discount.Value / 100
			subtotal -= amount // Apply sequentially for multiple discounts
		case "fixed_amount":
			amount = discount.Value
			subtotal -= amount
		}

		result.Applied = append(result.Applied, RestAppliedDiscount{
			Code:   discount.Code,
			Title:  discount.Description,
			Amount: amount,
		})
	}

	return result
}

// restHandleCheckoutSessions is the main router for /shopping-api/checkout-sessions
func restHandleCheckoutSessions(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/shopping-api/checkout-sessions")
	path = strings.TrimSuffix(path, "/")

	switch {
	case path == "" || path == "/":
		if r.Method == http.MethodPost {
			restCreateCheckout(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	case strings.HasSuffix(path, "/complete"):
		if r.Method == http.MethodPost {
			restCompleteCheckout(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	case strings.HasSuffix(path, "/cancel"):
		if r.Method == http.MethodPost {
			restCancelCheckout(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	default:
		switch r.Method {
		case http.MethodGet:
			restGetCheckout(w, r)
		case http.MethodPut:
			restUpdateCheckout(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// restHandleOrders is the main router for /orders
func restHandleOrders(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch r.Method {
	case http.MethodGet:
		restGetOrder(w, r)
	case http.MethodPut:
		restUpdateOrder(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
