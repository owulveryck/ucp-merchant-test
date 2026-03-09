package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	icatalog "github.com/owulveryck/ucp-merchant-test/internal/catalog"
	mpayment "github.com/owulveryck/ucp-merchant-test/internal/merchant/payment"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/pricing"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// In-memory MCP-specific state
var (
	mcpCheckoutStates = map[string]*model.MCPCheckoutState{}
	mcpOrderShipments = map[string]*model.Shipment{}
	mcpOrderOwners    = map[string]string{} // orderID -> ownerID
)

// extractUserID gets the _user_id injected by the MCP handler. Empty string = guest.
func extractUserID(args map[string]interface{}) string {
	uid, _ := args["_user_id"].(string)
	return uid
}

// extractUserCountryFromArgs gets the _user_country injected by the MCP handler.
func extractUserCountryFromArgs(args map[string]interface{}) string {
	c, _ := args["_user_country"].(string)
	return c
}

// canAccessEntity checks if a user can access an entity.
func canAccessEntity(userID, ownerID string) bool {
	if userID == "" {
		return ownerID == ""
	}
	return ownerID == userID
}

// Tool handlers

func handleListProducts(args map[string]interface{}) (interface{}, error) {
	category, _ := args["category"].(string)
	brand, _ := args["brand"].(string)
	query, _ := args["query"].(string)
	usageType, _ := args["usage_type"].(string)
	userCountry := extractUserCountryFromArgs(args)

	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 50 {
		limit = 50
	}

	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}
	if offset < 0 {
		offset = 0
	}

	filtered := catalogInstance.Filter(category, brand, query, usageType, userCountry, "", "")

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Rank != filtered[j].Rank {
			return filtered[i].Rank < filtered[j].Rank
		}
		return filtered[i].Title < filtered[j].Title
	})

	categories := catalogInstance.CategoryCount()

	total := len(filtered)

	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page := filtered[offset:end]

	type productInfo struct {
		ID                 string   `json:"id"`
		Title              string   `json:"title"`
		Category           string   `json:"category"`
		Brand              string   `json:"brand"`
		Price              string   `json:"price"`
		Rank               int      `json:"rank"`
		InStock            bool     `json:"in_stock"`
		ImageURL           string   `json:"image_url"`
		UsageType          string   `json:"usage_type,omitempty"`
		AvailableCountries []string `json:"available_countries,omitempty"`
	}
	products := make([]productInfo, 0, len(page))
	for _, p := range page {
		products = append(products, productInfo{
			ID:                 p.ID,
			Title:              p.Title,
			Category:           p.Category,
			Brand:              p.Brand,
			Price:              fmt.Sprintf("$%.2f", float64(p.Price)/100),
			Rank:               p.Rank,
			InStock:            p.Quantity > 0,
			ImageURL:           p.ImageURL,
			UsageType:          p.UsageType,
			AvailableCountries: p.AvailableCountries,
		})
	}

	hub.Publish(model.DashboardEvent{Type: "products_listed", Summary: fmt.Sprintf("Product catalog queried (showing %d of %d)", len(products), total), Timestamp: time.Now()})
	return map[string]interface{}{
		"products": products,
		"pagination": map[string]interface{}{
			"total":    total,
			"offset":   offset,
			"limit":    limit,
			"has_more": end < total,
		},
		"categories": categories,
	}, nil
}

func handleGetProductDetails(args map[string]interface{}) (interface{}, error) {
	id, _ := args["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	userCountry := extractUserCountryFromArgs(args)

	storeMu.Lock()
	p := catalogInstance.Find(id)
	if p == nil {
		storeMu.Unlock()
		return nil, fmt.Errorf("product not found: %s", id)
	}
	result := map[string]interface{}{
		"id":          p.ID,
		"title":       p.Title,
		"category":    p.Category,
		"brand":       p.Brand,
		"price":       fmt.Sprintf("$%.2f", float64(p.Price)/100),
		"price_cents": p.Price,
		"in_stock":    p.Quantity > 0,
		"rank":        p.Rank,
		"image_url":   p.ImageURL,
		"description": p.Description,
		"usage_type":  p.UsageType,
	}
	if len(p.AvailableCountries) > 0 {
		result["available_countries"] = p.AvailableCountries
		if userCountry != "" {
			result["available_in_your_country"] = icatalog.ContainsCountry(p.AvailableCountries, userCountry)
		}
	}
	storeMu.Unlock()

	hub.Publish(model.DashboardEvent{Type: "product_viewed", ID: p.ID, Summary: fmt.Sprintf("Product details viewed: %s", p.Title), Timestamp: time.Now()})
	return result, nil
}

func handleCreateCart(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	cartData, _ := args["cart"].(map[string]interface{})
	if cartData == nil {
		return nil, fmt.Errorf("missing cart parameter")
	}

	cartLineItems := parseLineItemRequests(cartData)
	if len(cartLineItems) == 0 {
		return nil, fmt.Errorf("cart must have at least one line item")
	}

	lineItems, err := pricing.BuildLineItems(cartLineItems, catalogInstance)
	if err != nil {
		return nil, err
	}

	cartSeq++
	cart := &model.Cart{
		ID:        fmt.Sprintf("cart-%04d", cartSeq),
		OwnerID:   extractUserID(args),
		LineItems: lineItems,
		Currency:  "USD",
		Totals:    pricing.CalculateTotals(lineItems, 0, nil),
	}
	carts[cart.ID] = cart
	hub.Publish(model.DashboardEvent{Type: "cart_created", ID: cart.ID, Summary: fmt.Sprintf("Cart %s created with %d items, total %s", cart.ID, len(cart.LineItems), cart.Totals[len(cart.Totals)-1].DisplayText), Timestamp: time.Now(), Data: *cart})
	return cart, nil
}

func handleGetCart(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	cart, ok := carts[id]
	if !ok || !canAccessEntity(extractUserID(args), cart.OwnerID) {
		return nil, fmt.Errorf("cart not found: %s", id)
	}
	return cart, nil
}

func handleUpdateCart(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	cart, ok := carts[id]
	if !ok || !canAccessEntity(extractUserID(args), cart.OwnerID) {
		return nil, fmt.Errorf("cart not found: %s", id)
	}

	cartData, _ := args["cart"].(map[string]interface{})
	if cartData == nil {
		return nil, fmt.Errorf("missing cart parameter")
	}
	cartLineItems := parseLineItemRequests(cartData)
	if len(cartLineItems) > 0 {
		lineItems, err := pricing.BuildLineItems(cartLineItems, catalogInstance)
		if err != nil {
			return nil, err
		}
		cart.LineItems = lineItems
		cart.Totals = pricing.CalculateTotals(lineItems, 0, nil)
	}
	hub.Publish(model.DashboardEvent{Type: "cart_updated", ID: cart.ID, Summary: fmt.Sprintf("Cart %s updated, total %s", cart.ID, cart.Totals[len(cart.Totals)-1].DisplayText), Timestamp: time.Now(), Data: *cart})
	return cart, nil
}

func handleCancelCart(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	cart, ok := carts[id]
	if !ok || !canAccessEntity(extractUserID(args), cart.OwnerID) {
		return nil, fmt.Errorf("cart not found: %s", id)
	}
	delete(carts, id)
	cart.Messages = append(cart.Messages, model.Message{Type: "info", Text: "Cart has been canceled"})
	hub.Publish(model.DashboardEvent{Type: "cart_canceled", ID: id, Summary: fmt.Sprintf("Cart %s canceled", id), Timestamp: time.Now()})
	return cart, nil
}

func handleCreateCheckout(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	checkoutData, _ := args["checkout"].(map[string]interface{})
	if checkoutData == nil {
		return nil, fmt.Errorf("missing checkout parameter")
	}

	var lineItems []model.LineItem

	// Check if creating from a cart
	if cartID, ok := checkoutData["cart_id"].(string); ok && cartID != "" {
		cart, exists := carts[cartID]
		if !exists {
			return nil, fmt.Errorf("cart not found: %s", cartID)
		}
		lineItems = cart.LineItems
	} else {
		coLineItems := parseLineItemRequests(checkoutData)
		if len(coLineItems) == 0 {
			return nil, fmt.Errorf("checkout must have line_items or cart_id")
		}
		var err error
		lineItems, err = pricing.BuildLineItems(coLineItems, catalogInstance)
		if err != nil {
			return nil, err
		}
	}

	// Validate country availability
	userCountry := extractUserCountryFromArgs(args)
	if userCountry != "" {
		for _, li := range lineItems {
			p := catalogInstance.Find(li.Item.ID)
			if p != nil && len(p.AvailableCountries) > 0 && !icatalog.ContainsCountry(p.AvailableCountries, userCountry) {
				return nil, fmt.Errorf("product %s (%s) is not available in %s", p.ID, p.Title, userCountry)
			}
		}
	}

	checkoutSeq++
	coID := fmt.Sprintf("checkout-%04d", checkoutSeq)

	co := &model.Checkout{
		ID:        coID,
		Status:    "incomplete",
		UCP:       model.UCPEnvelope{Version: "2026-01-11", Capabilities: []model.Capability{}},
		Links:     []model.Link{{Type: "application/json", URL: fmt.Sprintf("http://localhost:%d/checkout/%s", listenPort, coID)}},
		Currency:  "USD",
		LineItems: lineItems,
		Totals:    pricing.CalculateTotals(lineItems, 0, nil),
		Payment:   mpayment.DefaultPayment(),
	}

	ownerID := extractUserID(args)

	// Check for buyer info
	if buyerData, ok := checkoutData["buyer"].(map[string]interface{}); ok {
		co.Buyer = mpayment.ParseBuyer(parseBuyerRequest(buyerData))
	}

	state := &model.MCPCheckoutState{
		Checkout: co,
		OwnerID:  ownerID,
	}
	state.CheckoutHash = computeCheckoutHash(co, state.Shipping)

	checkouts[co.ID] = co
	mcpCheckoutStates[co.ID] = state

	hub.Publish(model.DashboardEvent{Type: "checkout_created", ID: co.ID, Summary: fmt.Sprintf("Checkout %s created, total %s", co.ID, co.Totals[len(co.Totals)-1].DisplayText), Timestamp: time.Now()})

	return mcpCheckoutResponse(co, state), nil
}

func handleGetCheckout(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	co, ok := checkouts[id]
	if !ok {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}
	state := mcpCheckoutStates[id]
	if state == nil || !canAccessEntity(extractUserID(args), state.OwnerID) {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}
	state.CheckoutHash = computeCheckoutHash(co, state.Shipping)
	return mcpCheckoutResponse(co, state), nil
}

func handleUpdateCheckout(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	co, ok := checkouts[id]
	if !ok {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}
	state := mcpCheckoutStates[id]
	if state == nil || !canAccessEntity(extractUserID(args), state.OwnerID) {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}
	if co.Status == "completed" || co.Status == "canceled" {
		return nil, fmt.Errorf("cannot update checkout in %s status", co.Status)
	}

	checkoutData, _ := args["checkout"].(map[string]interface{})
	if checkoutData == nil {
		return nil, fmt.Errorf("missing checkout parameter")
	}

	// Update line items if provided
	coLineItems := parseLineItemRequests(checkoutData)
	if len(coLineItems) > 0 {
		lineItems, err := pricing.BuildLineItems(coLineItems, catalogInstance)
		if err != nil {
			return nil, err
		}
		co.LineItems = lineItems
	}

	// Update buyer if provided
	if buyerData, ok := checkoutData["buyer"].(map[string]interface{}); ok {
		co.Buyer = mpayment.ParseBuyer(parseBuyerRequest(buyerData))
	}

	// Update shipping option if provided (MCP-specific)
	if shippingID, ok := checkoutData["shipping_option_id"].(string); ok {
		opt := findShippingOption(shippingID)
		if opt == nil {
			return nil, fmt.Errorf("unknown shipping option: %s — use get_shipping_options to see available options", shippingID)
		}
		state.Shipping = opt
	}

	// Recalculate totals
	shippingCost := 0
	if state.Shipping != nil {
		shippingCost = state.Shipping.Price
	}
	co.Totals = pricing.CalculateTotals(co.LineItems, shippingCost, co.Discounts)

	state.CheckoutHash = computeCheckoutHash(co, state.Shipping)
	hub.Publish(model.DashboardEvent{Type: "checkout_updated", ID: co.ID, Summary: fmt.Sprintf("Checkout %s updated, status: %s", co.ID, co.Status), Timestamp: time.Now()})
	return mcpCheckoutResponse(co, state), nil
}

func handleCompleteCheckout(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	co, ok := checkouts[id]
	if !ok {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}
	state := mcpCheckoutStates[id]
	if state == nil || !canAccessEntity(extractUserID(args), state.OwnerID) {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}
	if co.Status == "completed" {
		return nil, fmt.Errorf("checkout already completed")
	}
	if co.Status == "canceled" {
		return nil, fmt.Errorf("checkout has been canceled")
	}

	// Re-validate country availability at completion time
	userCountry := extractUserCountryFromArgs(args)
	country := userCountry
	if co.Buyer != nil && co.Buyer.FullName != "" {
		// Check if buyer has address info via fulfillment destinations
	}
	if country != "" {
		for _, li := range co.LineItems {
			p := catalogInstance.Find(li.Item.ID)
			if p != nil && len(p.AvailableCountries) > 0 && !icatalog.ContainsCountry(p.AvailableCountries, country) {
				return nil, fmt.Errorf("product %s (%s) is not available for delivery to %s", p.ID, p.Title, country)
			}
		}
	}

	// Verify approval hash
	approval, _ := args["approval"].(map[string]interface{})
	if approval == nil {
		return nil, fmt.Errorf("approval is required: the platform must present the checkout to the user for approval before completing")
	}
	submittedHash, _ := approval["checkout_hash"].(string)
	if submittedHash == "" {
		return nil, fmt.Errorf("approval.checkout_hash is required")
	}
	expectedHash := computeCheckoutHash(co, state.Shipping)
	if submittedHash != expectedHash {
		return nil, fmt.Errorf("checkout state changed since approval — re-fetch the checkout and request user approval again")
	}

	// Recalculate totals with shipping if selected
	if state.Shipping != nil {
		co.Totals = pricing.CalculateTotals(co.LineItems, state.Shipping.Price, co.Discounts)
	}

	orderSeq++
	orderID := fmt.Sprintf("order-%04d", orderSeq)

	// Build order line items
	var orderLineItems []model.OrderLineItem
	for _, li := range co.LineItems {
		orderLineItems = append(orderLineItems, model.OrderLineItem{
			ID:       li.ID,
			Item:     li.Item,
			Quantity: model.OrderQuantity{Total: li.Quantity, Fulfilled: 0},
			Totals:   li.Totals,
			Status:   "processing",
		})
	}

	order := &model.Order{
		ID:           orderID,
		UCP:          model.UCPEnvelope{Version: "2026-01-11", Capabilities: []model.Capability{}},
		CheckoutID:   co.ID,
		PermalinkURL: fmt.Sprintf("http://localhost:%d/orders/%s", listenPort, orderID),
		LineItems:    orderLineItems,
		Fulfillment:  model.OrderFulfillment{},
		Currency:     co.Currency,
		Totals:       co.Totals,
	}
	orders[orderID] = order
	mcpOrderOwners[orderID] = state.OwnerID

	co.Status = "completed"
	co.Order = &model.OrderRef{
		ID:           orderID,
		PermalinkURL: order.PermalinkURL,
	}
	state.CheckoutHash = computeCheckoutHash(co, state.Shipping)

	hub.Publish(model.DashboardEvent{Type: "checkout_completed", ID: co.ID, Summary: fmt.Sprintf("Order %s placed, total %s", orderID, co.Totals[len(co.Totals)-1].DisplayText), Timestamp: time.Now()})
	hub.Publish(model.DashboardEvent{Type: "order_confirmed", ID: orderID, Summary: fmt.Sprintf("Order %s confirmed", orderID), Timestamp: time.Now()})

	cancelCh := make(chan struct{})
	orderCancelChsMu.Lock()
	orderCancelChs[orderID] = cancelCh
	orderCancelChsMu.Unlock()

	go startOrderProgression(orderID, state.Shipping, cancelCh)

	return mcpCheckoutResponse(co, state), nil
}

func handleCancelCheckout(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	co, ok := checkouts[id]
	if !ok {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}
	state := mcpCheckoutStates[id]
	if state == nil || !canAccessEntity(extractUserID(args), state.OwnerID) {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}
	if co.Status == "completed" {
		return nil, fmt.Errorf("cannot cancel completed checkout")
	}
	co.Status = "canceled"
	state.CheckoutHash = computeCheckoutHash(co, state.Shipping)
	hub.Publish(model.DashboardEvent{Type: "checkout_canceled", ID: co.ID, Summary: fmt.Sprintf("Checkout %s canceled", co.ID), Timestamp: time.Now()})
	return mcpCheckoutResponse(co, state), nil
}

// mcpCheckoutResponse builds the MCP response for a checkout, adding MCP-specific fields.
func mcpCheckoutResponse(co *model.Checkout, state *model.MCPCheckoutState) interface{} {
	resp := map[string]interface{}{
		"id":         co.ID,
		"status":     co.Status,
		"currency":   co.Currency,
		"line_items": co.LineItems,
		"totals":     co.Totals,
		"links":      co.Links,
	}
	if state != nil && state.CheckoutHash != "" {
		resp["checkout_hash"] = state.CheckoutHash
	}
	if co.Buyer != nil {
		resp["buyer"] = co.Buyer
	}
	if state != nil && state.Shipping != nil {
		resp["selected_shipping"] = state.Shipping
	}
	if co.Order != nil {
		resp["order"] = map[string]interface{}{
			"id":            co.Order.ID,
			"permalink_url": co.Order.PermalinkURL,
		}
	}
	if co.Fulfillment != nil {
		resp["fulfillment"] = co.Fulfillment
	}
	if co.Discounts != nil {
		resp["discounts"] = co.Discounts
	}
	return resp
}

// extractImageURLs walks a result structure and returns all image_url values found.
func extractImageURLs(result interface{}) []string {
	data, err := json.Marshal(result)
	if err != nil {
		return nil
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	seen := map[string]bool{}
	var urls []string
	var walk func(v interface{})
	walk = func(v interface{}) {
		switch val := v.(type) {
		case map[string]interface{}:
			if u, ok := val["image_url"].(string); ok && u != "" && !seen[u] {
				seen[u] = true
				urls = append(urls, u)
			}
			for _, child := range val {
				walk(child)
			}
		case []interface{}:
			for _, child := range val {
				walk(child)
			}
		}
	}
	walk(raw)
	return urls
}

// Order management handlers

func handleGetOrder(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	ord, ok := orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	ownerID := mcpOrderOwners[id]
	if !canAccessEntity(extractUserID(args), ownerID) {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	return mcpOrderResponse(ord), nil
}

func handleListOrders(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	userID := extractUserID(args)

	type orderSummary struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		CheckoutID   string `json:"checkout_id"`
		PermalinkURL string `json:"permalink_url"`
		Total        string `json:"total"`
	}
	var summaries []orderSummary
	for _, ord := range orders {
		ownerID := mcpOrderOwners[ord.ID]
		if !canAccessEntity(userID, ownerID) {
			continue
		}
		totalText := ""
		for _, t := range ord.Totals {
			if t.Type == "total" {
				totalText = t.DisplayText
			}
		}
		summaries = append(summaries, orderSummary{
			ID:           ord.ID,
			Status:       "confirmed", // MCP orders start as confirmed
			CheckoutID:   ord.CheckoutID,
			PermalinkURL: ord.PermalinkURL,
			Total:        totalText,
		})
	}
	hub.Publish(model.DashboardEvent{Type: "orders_listed", Summary: "Order list queried", Timestamp: time.Now()})
	return map[string]interface{}{"orders": summaries}, nil
}

func handleCancelOrder(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	ord, ok := orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	ownerID := mcpOrderOwners[id]
	if !canAccessEntity(extractUserID(args), ownerID) {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	// Try to close cancel channel
	orderCancelChsMu.Lock()
	if ch, ok := orderCancelChs[id]; ok {
		close(ch)
		delete(orderCancelChs, id)
	}
	orderCancelChsMu.Unlock()

	hub.Publish(model.DashboardEvent{Type: "order_canceled", ID: ord.ID, Summary: fmt.Sprintf("Order %s canceled", ord.ID), Timestamp: time.Now()})
	return map[string]interface{}{"id": ord.ID, "status": "canceled", "message": "Order has been canceled"}, nil
}

func handleTrackOrder(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	ord, ok := orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	ownerID := mcpOrderOwners[id]
	if !canAccessEntity(extractUserID(args), ownerID) {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	result := map[string]interface{}{
		"order_id": ord.ID,
		"status":   "confirmed",
	}
	if shipment, ok := mcpOrderShipments[id]; ok {
		result["shipment"] = shipment
		result["status"] = "shipped"
	}
	return result, nil
}

// startOrderProgression simulates order fulfillment with timed status transitions.
func startOrderProgression(orderID string, shipping *model.ShippingOption, cancelCh chan struct{}) {
	type transition struct {
		delay  time.Duration
		status string
	}
	steps := []transition{
		{30 * time.Second, "processing"},
		{30 * time.Second, "shipped"},
		{30 * time.Second, "in_transit"},
		{30 * time.Second, "out_for_delivery"},
		{30 * time.Second, "delivered"},
	}

	for _, step := range steps {
		select {
		case <-cancelCh:
			return
		case <-time.After(step.delay):
		}

		storeMu.Lock()
		_, ok := orders[orderID]
		if !ok {
			storeMu.Unlock()
			return
		}

		if step.status == "shipped" {
			carrier := "FastShip Express"
			estimatedDays := 3
			if shipping != nil {
				carrier = shipping.Carrier
				estimatedDays = shipping.EstimatedDays
			}
			mcpOrderShipments[orderID] = &model.Shipment{
				TrackingNumber: fmt.Sprintf("TRK-%s-%06d", orderID, time.Now().UnixNano()%1000000),
				Carrier:        carrier,
				EstimatedDate:  time.Now().Add(time.Duration(estimatedDays) * 24 * time.Hour).Format("2006-01-02"),
				ShippedAt:      time.Now(),
			}
		}
		if step.status == "delivered" {
			if s, ok := mcpOrderShipments[orderID]; ok {
				s.DeliveredAt = time.Now()
			}
		}

		storeMu.Unlock()

		hub.Publish(model.DashboardEvent{
			Type:      "order_" + step.status,
			ID:        orderID,
			Summary:   fmt.Sprintf("Order %s → %s", orderID, step.status),
			Timestamp: time.Now(),
		})
	}
}

func handleGetShippingOptions(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	checkoutID, _ := args["checkout_id"].(string)
	if checkoutID == "" {
		return nil, fmt.Errorf("checkout_id is required")
	}
	co, ok := checkouts[checkoutID]
	if !ok {
		return nil, fmt.Errorf("checkout not found: %s", checkoutID)
	}
	state := mcpCheckoutStates[checkoutID]
	if state == nil || !canAccessEntity(extractUserID(args), state.OwnerID) {
		return nil, fmt.Errorf("checkout not found: %s", checkoutID)
	}

	options := getShippingOptions()

	result := map[string]interface{}{
		"checkout_id": checkoutID,
		"options":     options,
	}
	if state.Shipping != nil {
		result["selected_shipping"] = state.Shipping
	}
	// Check fulfillment for destination info to show relevant options
	if co.Fulfillment != nil {
		for _, m := range co.Fulfillment.Methods {
			for _, d := range m.Destinations {
				if d.AddressCountry != "" && d.AddressCountry != "US" && d.AddressCountry != "USA" {
					for i := range options {
						options[i].EstimatedDays += 2
					}
					break
				}
			}
		}
	}
	return result, nil
}

func handleTrackShipment(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["order_id"].(string)
	if id == "" {
		id, _ = args["id"].(string)
	}
	ord, ok := orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	ownerID := mcpOrderOwners[id]
	if !canAccessEntity(extractUserID(args), ownerID) {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	result := map[string]interface{}{
		"order_id": ord.ID,
		"status":   "confirmed",
	}
	if shipment, ok := mcpOrderShipments[id]; ok {
		result["shipment"] = shipment
		result["status"] = "shipped"
	}
	return result, nil
}

func getShippingOptions() []model.ShippingOption {
	return []model.ShippingOption{
		{ID: "standard", Method: "Standard Shipping", Carrier: "PostalService", EstimatedDays: 7, Price: 0, DisplayText: "Free — 5-7 business days"},
		{ID: "express", Method: "Express Shipping", Carrier: "FastShip Express", EstimatedDays: 3, Price: 999, DisplayText: "$9.99 — 2-3 business days"},
		{ID: "next_day", Method: "Next Day Delivery", Carrier: "FastShip Priority", EstimatedDays: 1, Price: 1999, DisplayText: "$19.99 — next business day"},
	}
}

func findShippingOption(id string) *model.ShippingOption {
	for _, opt := range getShippingOptions() {
		if opt.ID == id {
			return &opt
		}
	}
	return nil
}

// mcpOrderResponse builds an MCP-friendly response from a canonical Order.
func mcpOrderResponse(ord *model.Order) interface{} {
	return ord
}

// computeCheckoutHash produces a SHA-256 hash of the material checkout fields.
func computeCheckoutHash(co *model.Checkout, shipping *model.ShippingOption) string {
	type hashLineItem struct {
		ItemID   string `json:"item_id"`
		Title    string `json:"title"`
		Price    int    `json:"price"`
		Quantity int    `json:"quantity"`
	}
	type hashData struct {
		ID        string                `json:"id"`
		LineItems []hashLineItem        `json:"line_items"`
		Currency  string                `json:"currency"`
		Totals    []model.Total         `json:"totals"`
		Shipping  *model.ShippingOption `json:"shipping,omitempty"`
	}
	var items []hashLineItem
	for _, li := range co.LineItems {
		items = append(items, hashLineItem{
			ItemID:   li.Item.ID,
			Title:    li.Item.Title,
			Price:    li.Item.Price,
			Quantity: li.Quantity,
		})
	}
	data := hashData{
		ID:        co.ID,
		LineItems: items,
		Currency:  co.Currency,
		Totals:    co.Totals,
		Shipping:  shipping,
	}
	b, _ := json.Marshal(data)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// parseLineItemRequests converts a raw map's "line_items" field to typed requests.
func parseLineItemRequests(data map[string]interface{}) []model.LineItemRequest {
	rawItems, _ := data["line_items"].([]interface{})
	if len(rawItems) == 0 {
		return nil
	}
	items := make([]model.LineItemRequest, 0, len(rawItems))
	for _, raw := range rawItems {
		rawMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		li := model.LineItemRequest{}
		if id, ok := rawMap["id"].(string); ok {
			li.ID = id
		}
		if itemMap, ok := rawMap["item"].(map[string]interface{}); ok {
			if id, ok := itemMap["id"].(string); ok {
				li.Item = &model.ItemRef{ID: id}
			}
		}
		if pid, ok := rawMap["product_id"].(string); ok {
			li.ProductID = pid
		}
		if q, ok := rawMap["quantity"].(float64); ok {
			li.Quantity = int(q)
		}
		items = append(items, li)
	}
	return items
}

// parseBuyerRequest converts a raw buyer map to a typed BuyerRequest.
func parseBuyerRequest(data map[string]interface{}) *model.BuyerRequest {
	if data == nil {
		return nil
	}
	b := &model.BuyerRequest{}
	if v, ok := data["first_name"].(string); ok {
		b.FirstName = v
	}
	if v, ok := data["last_name"].(string); ok {
		b.LastName = v
	}
	if v, ok := data["fullName"].(string); ok {
		b.FullName = v
	}
	if v, ok := data["name"].(string); ok {
		b.Name = v
	}
	if v, ok := data["email"].(string); ok {
		b.Email = v
	}
	return b
}
