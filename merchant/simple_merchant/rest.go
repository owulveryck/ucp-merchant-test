package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/owulveryck/ucp-merchant-test/internal/idempotency"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/discount"
	mfulfillment "github.com/owulveryck/ucp-merchant-test/internal/merchant/fulfillment"
	mpayment "github.com/owulveryck/ucp-merchant-test/internal/merchant/payment"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/pricing"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
	"github.com/owulveryck/ucp-merchant-test/internal/webhook"
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
func handleIdempotency(w http.ResponseWriter, r *http.Request, body []byte) bool {
	key := r.Header.Get("idempotency-key")
	if key == "" {
		return false
	}
	payloadHash := idempotency.HashPayload(body)
	entry, exists := idempotencyStoreInstance.Check(key)
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
	idempotencyStoreInstance.Store(key, idempotency.HashPayload(body), statusCode, responseBody)
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

	var req model.CheckoutRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	lineItems, err := pricing.BuildLineItems(req.LineItems, catalogInstance)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	checkoutSeq++
	coID := fmt.Sprintf("co_%04d", checkoutSeq)

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	co := &model.Checkout{
		ID:        coID,
		Status:    "incomplete",
		UCP:       model.UCPEnvelope{Version: "2026-01-11", Capabilities: []model.Capability{}},
		Links:     []model.Link{{Type: "application/json", URL: fmt.Sprintf("%s://localhost:%d/shopping-api/checkout-sessions/%s", scheme(), listenPort, coID)}},
		Currency:  currency,
		LineItems: lineItems,
	}

	co.Totals = pricing.CalculateTotals(lineItems, 0, nil)
	co.Payment = mpayment.ParsePayment(req.Payment)
	co.Fulfillment = mfulfillment.ParseFulfillment(req.Fulfillment, nil, co, shopData, checkoutDestinations, checkoutOptionTitles, &addrSeqCounter, &addrSeqMu)
	co.Buyer = mpayment.ParseBuyer(req.Buyer)

	// Store webhook URL from UCP-Agent header
	webhookURL := webhook.ResolveWebhookURL(r.Header.Get("UCP-Agent"))
	if webhookURL != "" {
		checkoutWebhooks[coID] = webhookURL
	}

	checkouts[coID] = co

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
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	co, ok := checkouts[id]
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

	var req model.CheckoutRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	co, ok := checkouts[id]
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Checkout not found: %s", id))
		return
	}

	if co.Status == "completed" || co.Status == "canceled" {
		writeError(w, http.StatusConflict, fmt.Sprintf("Cannot update checkout in %s status", co.Status))
		return
	}

	// Update line items if provided
	if len(req.LineItems) > 0 {
		lineItems, err := pricing.BuildLineItems(req.LineItems, catalogInstance)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		co.LineItems = lineItems
	}

	// Update buyer if provided
	if req.Buyer != nil {
		co.Buyer = mpayment.ParseBuyer(req.Buyer)
	}

	// Update payment if provided
	if req.Payment != nil {
		co.Payment = mpayment.ParsePayment(req.Payment)
	}

	// Handle discounts
	shippingCost := mfulfillment.GetCurrentShippingCost(co)
	if req.Discounts != nil {
		co.Discounts = discount.ApplyDiscounts(req.Discounts, co.LineItems, shopData)
	}

	// Handle fulfillment
	if req.Fulfillment != nil {
		co.Fulfillment = mfulfillment.ParseFulfillment(req.Fulfillment, co.Buyer, co, shopData, checkoutDestinations, checkoutOptionTitles, &addrSeqCounter, &addrSeqMu)
		shippingCost = mfulfillment.GetCurrentShippingCost(co)
	}

	// Recalculate totals
	co.Totals = pricing.CalculateTotals(co.LineItems, shippingCost, co.Discounts)

	processAndRespond(w, r, body, http.StatusOK, co)
}

// restCompleteCheckout handles POST /shopping-api/checkout-sessions/{id}/complete
func restCompleteCheckout(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

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

	var req model.CheckoutRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	co, ok := checkouts[id]
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
	if !mfulfillment.IsFulfillmentComplete(co) {
		writeError(w, http.StatusBadRequest, "Fulfillment address and option must be selected")
		return
	}

	// Process payment
	if req.PaymentData != nil {
		if req.PaymentData.Credential != nil && req.PaymentData.Credential.Token == "fail_token" {
			writeError(w, http.StatusPaymentRequired, "Payment failed")
			return
		}
		if req.PaymentData.HandlerID != "" {
			validHandlers := map[string]bool{
				"google_pay":           true,
				"mock_payment_handler": true,
				"shop_pay":             true,
			}
			if !validHandlers[req.PaymentData.HandlerID] {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("Unknown payment handler: %s", req.PaymentData.HandlerID))
				return
			}
		}
	}

	// Create order
	orderSeq++
	orderID := fmt.Sprintf("order_%04d", orderSeq)

	var orderLineItems []model.OrderLineItem
	var expectationLineItems []model.EventLineItem
	for _, li := range co.LineItems {
		orderLineItems = append(orderLineItems, model.OrderLineItem{
			ID:       li.ID,
			Item:     li.Item,
			Quantity: model.OrderQuantity{Total: li.Quantity, Fulfilled: 0},
			Totals:   li.Totals,
			Status:   "processing",
		})
		expectationLineItems = append(expectationLineItems, model.EventLineItem{
			ID:       li.ID,
			Quantity: li.Quantity,
		})
	}

	// Build fulfillment expectations from selected option
	var expectations []model.Expectation
	optionTitle := checkoutOptionTitles[id]
	if optionTitle == "" {
		optionTitle = "Standard Shipping"
	}
	dest := checkoutDestinations[id]
	destVal := model.FulfillmentDestination{}
	if dest != nil {
		destVal = *dest
	}
	expectations = append(expectations, model.Expectation{
		ID:          "expect_1",
		LineItems:   expectationLineItems,
		MethodType:  "shipping",
		Description: optionTitle,
		Destination: destVal,
	})

	order := &model.Order{
		ID:           orderID,
		UCP:          model.UCPEnvelope{Version: "2026-01-11", Capabilities: []model.Capability{}},
		CheckoutID:   id,
		PermalinkURL: fmt.Sprintf("%s://localhost:%d/orders/%s", scheme(), listenPort, orderID),
		LineItems:    orderLineItems,
		Fulfillment: model.OrderFulfillment{
			Expectations: expectations,
		},
		Currency: co.Currency,
		Totals:   co.Totals,
	}

	orders[orderID] = order

	co.Status = "completed"
	co.Order = &model.OrderRef{
		ID:           orderID,
		PermalinkURL: order.PermalinkURL,
	}

	// Send webhook
	if webhookURL, ok := checkoutWebhooks[id]; ok {
		orderJSON, _ := json.Marshal(order)
		var orderMap map[string]interface{}
		json.Unmarshal(orderJSON, &orderMap)
		webhook.SendWebhookEvent(webhookURL, map[string]interface{}{
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

	storeMu.Lock()
	defer storeMu.Unlock()

	co, ok := checkouts[id]
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

	storeMu.Lock()
	defer storeMu.Unlock()

	order, ok := orders[id]
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

	storeMu.Lock()
	defer storeMu.Unlock()

	order, ok := orders[id]
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
				var events []model.FulfillmentEvent
				json.Unmarshal(eventsJSON, &events)
				order.Fulfillment.Events = events
			}
			if expectRaw, ok := fMap["expectations"]; ok {
				expectJSON, _ := json.Marshal(expectRaw)
				var expectations []model.Expectation
				json.Unmarshal(expectJSON, &expectations)
				order.Fulfillment.Expectations = expectations
			}
		}
	}

	// Update adjustments
	if adjRaw, ok := reqMap["adjustments"]; ok {
		adjJSON, _ := json.Marshal(adjRaw)
		var adjustments []model.Adjustment
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

	secret := r.Header.Get("Simulation-Secret")
	if secret != simulationSecret {
		writeError(w, http.StatusForbidden, "Invalid or missing simulation secret")
		return
	}

	id := extractPathParam(r.URL.Path, "/testing/simulate-shipping/")

	storeMu.Lock()
	defer storeMu.Unlock()

	order, ok := orders[id]
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Order not found: %s", id))
		return
	}

	event := model.FulfillmentEvent{
		ID:             fmt.Sprintf("evt_ship_%s", id),
		OccurredAt:     time.Now().UTC().Format(time.RFC3339),
		Type:           "shipped",
		TrackingNumber: fmt.Sprintf("TRK-%s", id),
		Description:    "Order shipped",
	}

	for _, li := range order.LineItems {
		event.LineItems = append(event.LineItems, model.EventLineItem{
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
		webhook.SendWebhookEvent(webhookURL, map[string]interface{}{
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
	s = strings.TrimSuffix(s, "/")
	return s
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
