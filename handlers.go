package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Shipment struct {
	TrackingNumber string    `json:"tracking_number"`
	Carrier        string    `json:"carrier"`
	EstimatedDate  string    `json:"estimated_delivery,omitempty"`
	ShippedAt      time.Time `json:"shipped_at,omitempty"`
	DeliveredAt    time.Time `json:"delivered_at,omitempty"`
}

// UCP data types

type Item struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Price    int    `json:"price"`
	ImageURL string `json:"image_url,omitempty"`
}

type Total struct {
	Type        string `json:"type"`
	DisplayText string `json:"display_text,omitempty"`
	Amount      int    `json:"amount"`
}

type LineItem struct {
	ID       string  `json:"id"`
	Item     Item    `json:"item"`
	Quantity int     `json:"quantity"`
	Totals   []Total `json:"totals"`
}

type Address struct {
	Street  string `json:"street,omitempty"`
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	Zip     string `json:"zip,omitempty"`
	Country string `json:"country,omitempty"`
}

type Buyer struct {
	Name    string   `json:"name,omitempty"`
	Email   string   `json:"email,omitempty"`
	Address *Address `json:"address,omitempty"`
}

type Link struct {
	Rel string `json:"rel"`
	URL string `json:"url"`
}

type Order struct {
	ID            string     `json:"id"`
	OwnerID       string     `json:"owner_id,omitempty"`
	Status        string     `json:"status"`
	ConfirmationN string     `json:"confirmation_number"`
	LineItems     []LineItem `json:"line_items,omitempty"`
	Currency      string     `json:"currency,omitempty"`
	Totals        []Total    `json:"totals,omitempty"`
	Buyer         *Buyer     `json:"buyer,omitempty"`
	Shipment      *Shipment  `json:"shipment,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	cancelCh      chan struct{}
}

type Cart struct {
	ID        string     `json:"id"`
	OwnerID   string     `json:"owner_id,omitempty"`
	LineItems []LineItem `json:"line_items"`
	Currency  string     `json:"currency"`
	Totals    []Total    `json:"totals"`
	Messages  []Message  `json:"messages,omitempty"`
}

type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ShippingOption struct {
	ID            string `json:"id"`
	Method        string `json:"method"`
	Carrier       string `json:"carrier"`
	EstimatedDays int    `json:"estimated_days"`
	Price         int    `json:"price"`
	DisplayText   string `json:"display_text"`
}

type Checkout struct {
	ID               string          `json:"id"`
	OwnerID          string          `json:"owner_id,omitempty"`
	LineItems        []LineItem      `json:"line_items"`
	Status           string          `json:"status"`
	Currency         string          `json:"currency"`
	Totals           []Total         `json:"totals"`
	CheckoutHash     string          `json:"checkout_hash,omitempty"`
	Links            []Link          `json:"links,omitempty"`
	Buyer            *Buyer          `json:"buyer,omitempty"`
	SelectedShipping *ShippingOption `json:"selected_shipping,omitempty"`
	ContinueURL      string          `json:"continue_url,omitempty"`
	Order            *Order          `json:"order,omitempty"`
}

// In-memory stores
var (
	carts       = map[string]*Cart{}
	checkouts   = map[string]*Checkout{}
	orders      = map[string]*Order{}
	cartSeq     int
	checkoutSeq int
	orderSeq    int
	storeMu     sync.Mutex
)

const taxRate = 0.20 // 20% tax

func getShippingOptions() []ShippingOption {
	return []ShippingOption{
		{ID: "standard", Method: "Standard Shipping", Carrier: "PostalService", EstimatedDays: 7, Price: 0, DisplayText: "Free — 5-7 business days"},
		{ID: "express", Method: "Express Shipping", Carrier: "FastShip Express", EstimatedDays: 3, Price: 999, DisplayText: "$9.99 — 2-3 business days"},
		{ID: "next_day", Method: "Next Day Delivery", Carrier: "FastShip Priority", EstimatedDays: 1, Price: 1999, DisplayText: "$19.99 — next business day"},
	}
}

func findShippingOption(id string) *ShippingOption {
	for _, opt := range getShippingOptions() {
		if opt.ID == id {
			return &opt
		}
	}
	return nil
}

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
// Authenticated users can only access their own entities.
// Guests (empty userID) can access guest entities (empty ownerID).
func canAccessEntity(userID, ownerID string) bool {
	if userID == "" {
		return ownerID == ""
	}
	return ownerID == userID
}

func buildLineItems(rawItems []map[string]interface{}) ([]LineItem, error) {
	var items []LineItem
	for i, raw := range rawItems {
		productID, _ := raw["product_id"].(string)
		if productID == "" {
			// Try item.id as fallback
			if itemMap, ok := raw["item"].(map[string]interface{}); ok {
				productID, _ = itemMap["id"].(string)
			}
		}
		if productID == "" {
			return nil, fmt.Errorf("line item %d: missing product_id", i)
		}
		product := findProduct(productID)
		if product == nil {
			return nil, fmt.Errorf("product not found: %s", productID)
		}
		qty := 1
		if q, ok := raw["quantity"].(float64); ok {
			qty = int(q)
		}
		if qty < 1 {
			qty = 1
		}
		lineTotal := product.Price * qty
		li := LineItem{
			ID:       fmt.Sprintf("LI-%03d", i+1),
			Item:     Item{ID: product.ID, Title: product.Title, Price: product.Price, ImageURL: product.ImageURL},
			Quantity: qty,
			Totals:   []Total{{Type: "subtotal", Amount: lineTotal}},
		}
		items = append(items, li)
	}
	return items, nil
}

func calculateTotals(items []LineItem) []Total {
	return calculateTotalsWithShipping(items, nil)
}

func calculateTotalsWithShipping(items []LineItem, shipping *ShippingOption) []Total {
	subtotal := 0
	for _, li := range items {
		for _, t := range li.Totals {
			if t.Type == "subtotal" {
				subtotal += t.Amount
			}
		}
	}
	tax := int(float64(subtotal) * taxRate)
	total := subtotal + tax

	totals := []Total{
		{Type: "subtotal", DisplayText: fmt.Sprintf("$%.2f", float64(subtotal)/100), Amount: subtotal},
		{Type: "tax", DisplayText: fmt.Sprintf("$%.2f (20%%)", float64(tax)/100), Amount: tax},
	}

	if shipping != nil {
		totals = append(totals, Total{
			Type:        "shipping",
			DisplayText: fmt.Sprintf("$%.2f (%s)", float64(shipping.Price)/100, shipping.Method),
			Amount:      shipping.Price,
		})
		total += shipping.Price
	}

	totals = append(totals, Total{
		Type:        "total",
		DisplayText: fmt.Sprintf("$%.2f", float64(total)/100),
		Amount:      total,
	})
	return totals
}

// Tool handlers

func handleListProducts(args map[string]interface{}) (interface{}, error) {
	// Parse optional filters
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

	// Filter
	filtered := filterCatalog(category, brand, query, usageType, userCountry)

	// Sort by Rank ascending, then Title for stability
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Rank != filtered[j].Rank {
			return filtered[i].Rank < filtered[j].Rank
		}
		return filtered[i].Title < filtered[j].Title
	})

	// Categories (always computed from full catalog for discovery)
	categories := categoryCount(catalog)

	total := len(filtered)

	// Apply pagination
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page := filtered[offset:end]

	// Build response
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

	hub.Publish(DashboardEvent{Type: "products_listed", Summary: fmt.Sprintf("Product catalog queried (showing %d of %d)", len(products), total), Timestamp: time.Now()})
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
	p := findProduct(id)
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
			result["available_in_your_country"] = containsCountry(p.AvailableCountries, userCountry)
		}
	}
	storeMu.Unlock()

	hub.Publish(DashboardEvent{Type: "product_viewed", ID: p.ID, Summary: fmt.Sprintf("Product details viewed: %s", p.Title), Timestamp: time.Now()})
	return result, nil
}

func handleCreateCart(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	cartData, _ := args["cart"].(map[string]interface{})
	if cartData == nil {
		return nil, fmt.Errorf("missing cart parameter")
	}
	rawItems, _ := cartData["line_items"].([]interface{})
	if len(rawItems) == 0 {
		return nil, fmt.Errorf("cart must have at least one line item")
	}

	var itemMaps []map[string]interface{}
	for _, ri := range rawItems {
		if m, ok := ri.(map[string]interface{}); ok {
			itemMaps = append(itemMaps, m)
		}
	}

	lineItems, err := buildLineItems(itemMaps)
	if err != nil {
		return nil, err
	}

	cartSeq++
	cart := &Cart{
		ID:        fmt.Sprintf("cart-%04d", cartSeq),
		OwnerID:   extractUserID(args),
		LineItems: lineItems,
		Currency:  "USD",
		Totals:    calculateTotals(lineItems),
	}
	carts[cart.ID] = cart
	hub.Publish(DashboardEvent{Type: "cart_created", ID: cart.ID, Summary: fmt.Sprintf("Cart %s created with %d items, total %s", cart.ID, len(cart.LineItems), cart.Totals[len(cart.Totals)-1].DisplayText), Timestamp: time.Now(), Data: *cart})
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
	rawItems, _ := cartData["line_items"].([]interface{})
	if len(rawItems) > 0 {
		var itemMaps []map[string]interface{}
		for _, ri := range rawItems {
			if m, ok := ri.(map[string]interface{}); ok {
				itemMaps = append(itemMaps, m)
			}
		}
		lineItems, err := buildLineItems(itemMaps)
		if err != nil {
			return nil, err
		}
		cart.LineItems = lineItems
		cart.Totals = calculateTotals(lineItems)
	}
	hub.Publish(DashboardEvent{Type: "cart_updated", ID: cart.ID, Summary: fmt.Sprintf("Cart %s updated, total %s", cart.ID, cart.Totals[len(cart.Totals)-1].DisplayText), Timestamp: time.Now(), Data: *cart})
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
	cart.Messages = append(cart.Messages, Message{Type: "info", Text: "Cart has been canceled"})
	hub.Publish(DashboardEvent{Type: "cart_canceled", ID: id, Summary: fmt.Sprintf("Cart %s canceled", id), Timestamp: time.Now()})
	return cart, nil
}

func handleCreateCheckout(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	checkoutData, _ := args["checkout"].(map[string]interface{})
	if checkoutData == nil {
		return nil, fmt.Errorf("missing checkout parameter")
	}

	var lineItems []LineItem

	// Check if creating from a cart
	if cartID, ok := checkoutData["cart_id"].(string); ok && cartID != "" {
		cart, exists := carts[cartID]
		if !exists {
			return nil, fmt.Errorf("cart not found: %s", cartID)
		}
		lineItems = cart.LineItems
	} else {
		rawItems, _ := checkoutData["line_items"].([]interface{})
		if len(rawItems) == 0 {
			return nil, fmt.Errorf("checkout must have line_items or cart_id")
		}
		var itemMaps []map[string]interface{}
		for _, ri := range rawItems {
			if m, ok := ri.(map[string]interface{}); ok {
				itemMaps = append(itemMaps, m)
			}
		}
		var err error
		lineItems, err = buildLineItems(itemMaps)
		if err != nil {
			return nil, err
		}
	}

	// Validate country availability
	userCountry := extractUserCountryFromArgs(args)
	if userCountry != "" {
		for _, li := range lineItems {
			p := findProduct(li.Item.ID)
			if p != nil && len(p.AvailableCountries) > 0 && !containsCountry(p.AvailableCountries, userCountry) {
				return nil, fmt.Errorf("product %s (%s) is not available in %s", p.ID, p.Title, userCountry)
			}
		}
	}

	checkoutSeq++
	co := &Checkout{
		ID:        fmt.Sprintf("checkout-%04d", checkoutSeq),
		OwnerID:   extractUserID(args),
		LineItems: lineItems,
		Status:    "incomplete",
		Currency:  "USD",
		Totals:    calculateTotals(lineItems),
		Links:     []Link{{Rel: "self", URL: fmt.Sprintf("http://localhost:%d/checkout/checkout-%04d", listenPort, checkoutSeq)}},
	}

	// Check for buyer info
	if buyerData, ok := checkoutData["buyer"].(map[string]interface{}); ok {
		co.Buyer = parseBuyer(buyerData)
		if co.Buyer.Address != nil {
			co.Status = "ready_for_complete"
		}
	}

	co.CheckoutHash = computeCheckoutHash(co)
	checkouts[co.ID] = co
	hub.Publish(DashboardEvent{Type: "checkout_created", ID: co.ID, Summary: fmt.Sprintf("Checkout %s created, total %s", co.ID, co.Totals[len(co.Totals)-1].DisplayText), Timestamp: time.Now(), Data: *co})
	return co, nil
}

func handleGetCheckout(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	co, ok := checkouts[id]
	if !ok || !canAccessEntity(extractUserID(args), co.OwnerID) {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}
	co.CheckoutHash = computeCheckoutHash(co)
	return co, nil
}

func handleUpdateCheckout(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	co, ok := checkouts[id]
	if !ok || !canAccessEntity(extractUserID(args), co.OwnerID) {
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
	if rawItems, ok := checkoutData["line_items"].([]interface{}); ok && len(rawItems) > 0 {
		var itemMaps []map[string]interface{}
		for _, ri := range rawItems {
			if m, ok := ri.(map[string]interface{}); ok {
				itemMaps = append(itemMaps, m)
			}
		}
		lineItems, err := buildLineItems(itemMaps)
		if err != nil {
			return nil, err
		}
		co.LineItems = lineItems
		co.Totals = calculateTotals(lineItems)
	}

	// Update buyer if provided
	if buyerData, ok := checkoutData["buyer"].(map[string]interface{}); ok {
		co.Buyer = parseBuyer(buyerData)
		if co.Buyer.Address != nil && co.Status == "incomplete" {
			co.Status = "ready_for_complete"
		}
	}

	// Update shipping option if provided
	if shippingID, ok := checkoutData["shipping_option_id"].(string); ok {
		opt := findShippingOption(shippingID)
		if opt == nil {
			return nil, fmt.Errorf("unknown shipping option: %s — use get_shipping_options to see available options", shippingID)
		}
		co.SelectedShipping = opt
		co.Totals = calculateTotalsWithShipping(co.LineItems, co.SelectedShipping)
	}

	co.CheckoutHash = computeCheckoutHash(co)
	hub.Publish(DashboardEvent{Type: "checkout_updated", ID: co.ID, Summary: fmt.Sprintf("Checkout %s updated, status: %s", co.ID, co.Status), Timestamp: time.Now(), Data: *co})
	return co, nil
}

func handleCompleteCheckout(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	co, ok := checkouts[id]
	if !ok || !canAccessEntity(extractUserID(args), co.OwnerID) {
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
	if co.Buyer != nil && co.Buyer.Address != nil && co.Buyer.Address.Country != "" {
		country = co.Buyer.Address.Country
	}
	if country != "" {
		for _, li := range co.LineItems {
			p := findProduct(li.Item.ID)
			if p != nil && len(p.AvailableCountries) > 0 && !containsCountry(p.AvailableCountries, country) {
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
	expectedHash := computeCheckoutHash(co)
	if submittedHash != expectedHash {
		return nil, fmt.Errorf("checkout state changed since approval — re-fetch the checkout and request user approval again")
	}

	// Recalculate totals with shipping if selected
	if co.SelectedShipping != nil {
		co.Totals = calculateTotalsWithShipping(co.LineItems, co.SelectedShipping)
	}

	orderSeq++
	now := time.Now()
	cancelCh := make(chan struct{})
	ord := &Order{
		ID:            fmt.Sprintf("order-%04d", orderSeq),
		OwnerID:       co.OwnerID,
		Status:        "confirmed",
		ConfirmationN: fmt.Sprintf("CONF-%06d", orderSeq*1000+1),
		LineItems:     co.LineItems,
		Currency:      co.Currency,
		Totals:        co.Totals,
		Buyer:         co.Buyer,
		CreatedAt:     now,
		UpdatedAt:     now,
		cancelCh:      cancelCh,
	}
	orders[ord.ID] = ord

	co.Status = "completed"
	co.Order = &Order{
		ID:            ord.ID,
		Status:        ord.Status,
		ConfirmationN: ord.ConfirmationN,
	}
	co.CheckoutHash = computeCheckoutHash(co)

	// Copy order for event to avoid race
	ordCopy := *ord
	ordCopy.cancelCh = nil

	hub.Publish(DashboardEvent{Type: "checkout_completed", ID: co.ID, Summary: fmt.Sprintf("Order %s placed (%s), total %s", co.Order.ID, co.Order.ConfirmationN, co.Totals[len(co.Totals)-1].DisplayText), Timestamp: time.Now(), Data: *co})
	hub.Publish(DashboardEvent{Type: "order_confirmed", ID: ord.ID, Summary: fmt.Sprintf("Order %s confirmed", ord.ID), Timestamp: time.Now(), Data: ordCopy})

	go startOrderProgression(ord.ID, co.SelectedShipping, cancelCh)

	return co, nil
}

func handleCancelCheckout(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	co, ok := checkouts[id]
	if !ok || !canAccessEntity(extractUserID(args), co.OwnerID) {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}
	if co.Status == "completed" {
		return nil, fmt.Errorf("cannot cancel completed checkout")
	}
	co.Status = "canceled"
	co.CheckoutHash = computeCheckoutHash(co)
	hub.Publish(DashboardEvent{Type: "checkout_canceled", ID: co.ID, Summary: fmt.Sprintf("Checkout %s canceled", co.ID), Timestamp: time.Now(), Data: *co})
	return co, nil
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

func parseBuyer(data map[string]interface{}) *Buyer {
	b := &Buyer{}
	if name, ok := data["name"].(string); ok {
		b.Name = name
	}
	if email, ok := data["email"].(string); ok {
		b.Email = email
	}
	if addrData, ok := data["address"].(map[string]interface{}); ok {
		b.Address = &Address{}
		if s, ok := addrData["street"].(string); ok {
			b.Address.Street = s
		}
		if s, ok := addrData["city"].(string); ok {
			b.Address.City = s
		}
		if s, ok := addrData["state"].(string); ok {
			b.Address.State = s
		}
		if s, ok := addrData["zip"].(string); ok {
			b.Address.Zip = s
		}
		if s, ok := addrData["country"].(string); ok {
			b.Address.Country = s
		}
	}
	return b
}

// Order management handlers

func handleGetOrder(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	ord, ok := orders[id]
	if !ok || !canAccessEntity(extractUserID(args), ord.OwnerID) {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	return ord, nil
}

func handleListOrders(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	userID := extractUserID(args)

	type orderSummary struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		ConfirmationN string `json:"confirmation_number"`
		Total         string `json:"total"`
		CreatedAt     string `json:"created_at"`
	}
	var summaries []orderSummary
	for _, ord := range orders {
		if !canAccessEntity(userID, ord.OwnerID) {
			continue
		}
		totalText := ""
		for _, t := range ord.Totals {
			if t.Type == "total" {
				totalText = t.DisplayText
			}
		}
		summaries = append(summaries, orderSummary{
			ID:            ord.ID,
			Status:        ord.Status,
			ConfirmationN: ord.ConfirmationN,
			Total:         totalText,
			CreatedAt:     ord.CreatedAt.Format(time.RFC3339),
		})
	}
	hub.Publish(DashboardEvent{Type: "orders_listed", Summary: "Order list queried", Timestamp: time.Now()})
	return map[string]interface{}{"orders": summaries}, nil
}

func handleCancelOrder(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	ord, ok := orders[id]
	if !ok || !canAccessEntity(extractUserID(args), ord.OwnerID) {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	if ord.Status != "confirmed" && ord.Status != "processing" {
		return nil, fmt.Errorf("cannot cancel order in %s status — only cancellable before shipping", ord.Status)
	}

	close(ord.cancelCh)
	ord.Status = "canceled"
	ord.UpdatedAt = time.Now()

	ordCopy := *ord
	ordCopy.cancelCh = nil

	hub.Publish(DashboardEvent{Type: "order_canceled", ID: ord.ID, Summary: fmt.Sprintf("Order %s canceled", ord.ID), Timestamp: time.Now(), Data: ordCopy})
	return map[string]interface{}{"id": ord.ID, "status": ord.Status, "message": "Order has been canceled"}, nil
}

func handleTrackOrder(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["id"].(string)
	ord, ok := orders[id]
	if !ok || !canAccessEntity(extractUserID(args), ord.OwnerID) {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	statusMessages := map[string]string{
		"confirmed":        "Order has been confirmed and is awaiting processing",
		"processing":       "Order is being prepared for shipment",
		"shipped":          "Package has been shipped",
		"in_transit":       "Package is in transit",
		"out_for_delivery": "Package is out for delivery",
		"delivered":        "Package has been delivered",
		"canceled":         "Order has been canceled",
	}

	result := map[string]interface{}{
		"order_id":            ord.ID,
		"status":              ord.Status,
		"confirmation_number": ord.ConfirmationN,
		"status_message":      statusMessages[ord.Status],
	}
	if ord.Shipment != nil {
		result["shipment"] = ord.Shipment
	}
	return result, nil
}

// startOrderProgression simulates order fulfillment with timed status transitions.
func startOrderProgression(orderID string, shipping *ShippingOption, cancelCh chan struct{}) {
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
		ord, ok := orders[orderID]
		if !ok || ord.Status == "canceled" {
			storeMu.Unlock()
			return
		}

		ord.Status = step.status
		ord.UpdatedAt = time.Now()

		if step.status == "shipped" {
			carrier := "FastShip Express"
			estimatedDays := 3
			if shipping != nil {
				carrier = shipping.Carrier
				estimatedDays = shipping.EstimatedDays
			}
			ord.Shipment = &Shipment{
				TrackingNumber: fmt.Sprintf("TRK-%s-%06d", orderID, time.Now().UnixNano()%1000000),
				Carrier:        carrier,
				EstimatedDate:  time.Now().Add(time.Duration(estimatedDays) * 24 * time.Hour).Format("2006-01-02"),
				ShippedAt:      time.Now(),
			}
		}
		if step.status == "delivered" && ord.Shipment != nil {
			ord.Shipment.DeliveredAt = time.Now()
		}

		ordCopy := *ord
		ordCopy.cancelCh = nil

		storeMu.Unlock()

		hub.Publish(DashboardEvent{
			Type:      "order_" + step.status,
			ID:        orderID,
			Summary:   fmt.Sprintf("Order %s → %s", orderID, step.status),
			Timestamp: time.Now(),
			Data:      ordCopy,
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
	if !ok || !canAccessEntity(extractUserID(args), co.OwnerID) {
		return nil, fmt.Errorf("checkout not found: %s", checkoutID)
	}

	options := getShippingOptions()

	// Adjust estimated days for non-US countries
	if co.Buyer != nil && co.Buyer.Address != nil && co.Buyer.Address.Country != "" {
		country := co.Buyer.Address.Country
		if country != "US" && country != "USA" && country != "United States" {
			for i := range options {
				options[i].EstimatedDays += 2
			}
		}
	}

	result := map[string]interface{}{
		"checkout_id": checkoutID,
		"options":     options,
	}
	if co.SelectedShipping != nil {
		result["selected_shipping"] = co.SelectedShipping
	}
	return result, nil
}

func handleTrackShipment(args map[string]interface{}) (interface{}, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	id, _ := args["order_id"].(string)
	if id == "" {
		// Fallback to "id" for compat
		id, _ = args["id"].(string)
	}
	ord, ok := orders[id]
	if !ok || !canAccessEntity(extractUserID(args), ord.OwnerID) {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	statusMessages := map[string]string{
		"confirmed":        "Order has been confirmed and is awaiting processing",
		"processing":       "Order is being prepared for shipment",
		"shipped":          "Package has been shipped",
		"in_transit":       "Package is in transit",
		"out_for_delivery": "Package is out for delivery",
		"delivered":        "Package has been delivered",
		"canceled":         "Order has been canceled",
	}

	result := map[string]interface{}{
		"order_id":            ord.ID,
		"status":              ord.Status,
		"confirmation_number": ord.ConfirmationN,
		"status_message":      statusMessages[ord.Status],
	}
	if ord.Shipment != nil {
		result["shipment"] = ord.Shipment
	}
	return result, nil
}

// computeCheckoutHash produces a SHA-256 hash of the material checkout fields.
// Only server-side Go code computes this, so there are no cross-language canonicalization issues.
func computeCheckoutHash(co *Checkout) string {
	type hashLineItem struct {
		ItemID   string `json:"item_id"`
		Title    string `json:"title"`
		Price    int    `json:"price"`
		Quantity int    `json:"quantity"`
	}
	type hashData struct {
		ID        string          `json:"id"`
		LineItems []hashLineItem  `json:"line_items"`
		Currency  string          `json:"currency"`
		Totals    []Total         `json:"totals"`
		Shipping  *ShippingOption `json:"shipping,omitempty"`
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
		Shipping:  co.SelectedShipping,
	}
	b, _ := json.Marshal(data)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
