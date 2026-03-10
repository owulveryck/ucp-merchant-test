// Package main provides a concrete merchant.Merchant implementation for the
// UCP conformance test server. The simpleMerchant struct holds all business
// logic and in-memory state; transport packages (REST, MCP) delegate to it.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	icatalog "github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/discount"
	mfulfillment "github.com/owulveryck/ucp-merchant-test/internal/merchant/fulfillment"
	mpayment "github.com/owulveryck/ucp-merchant-test/internal/merchant/payment"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/pricing"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

// simpleMerchant is the concrete merchant.Merchant implementation used by
// the UCP conformance test server. It stores all shopping state in memory
// and delegates to internal business-logic packages for pricing, fulfillment,
// discounts, and payment.
type simpleMerchant struct {
	mu       sync.Mutex
	catalog  *catalogStore
	shopData shopDataSource

	// In-memory state
	checkouts   map[string]*model.Checkout
	orders      map[string]*model.Order
	carts       map[string]*model.Cart
	checkoutSeq int
	orderSeq    int
	cartSeq     int

	// Checkout metadata
	orderOwners          map[string]string // orderID -> ownerID
	checkoutOwners       map[string]string // checkoutID -> ownerID
	checkoutDestinations map[string]*model.FulfillmentDestination
	checkoutOptionTitles map[string]string
	addrSeqCounter       int

	// Separate mutex for address sequence counter (avoids deadlock with m.mu)
	addrSeqMu sync.Mutex

	// Port/scheme for link generation
	listenPort func() int
	scheme     func() string
}

// newSimpleMerchant creates a new simpleMerchant with the given catalog and
// data source. The listenPort and scheme functions are used for generating
// resource URLs in checkout and order responses.
func newSimpleMerchant(cat *catalogStore, data shopDataSource, listenPort func() int, schemeFn func() string) *simpleMerchant {
	return &simpleMerchant{
		catalog:              cat,
		shopData:             data,
		checkouts:            map[string]*model.Checkout{},
		orders:               map[string]*model.Order{},
		carts:                map[string]*model.Cart{},
		orderOwners:          map[string]string{},
		checkoutOwners:       map[string]string{},
		checkoutDestinations: map[string]*model.FulfillmentDestination{},
		checkoutOptionTitles: map[string]string{},
		listenPort:           listenPort,
		scheme:               schemeFn,
	}
}

// --- Cataloger (delegates to catalog store) ---

func (m *simpleMerchant) Find(id string) *icatalog.Product {
	return m.catalog.Find(id)
}

func (m *simpleMerchant) Filter(category ucp.Category, brand, query string, country ucp.Country, currency ucp.Currency, language ucp.Language) []icatalog.Product {
	return m.catalog.Filter(category, brand, query, country, currency, language)
}

func (m *simpleMerchant) CategoryCount() []icatalog.CategoryStat {
	return m.catalog.CategoryCount()
}

func (m *simpleMerchant) Lookup(id string, shipsTo ucp.Country) *icatalog.Product {
	return m.catalog.Lookup(id, shipsTo)
}

func (m *simpleMerchant) Search(params icatalog.SearchParams) []icatalog.SearchResult {
	return m.catalog.Search(params)
}

// --- Carter ---

// CreateCart creates a new shopping cart with the given line items.
func (m *simpleMerchant) CreateCart(ownerID string, items []model.LineItemRequest) (*model.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(items) == 0 {
		return nil, fmt.Errorf("cart must have at least one line item: %w", merchant.ErrBadRequest)
	}

	lineItems, err := pricing.BuildLineItems(items, m.catalog)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", err.Error(), merchant.ErrBadRequest)
	}

	m.cartSeq++
	cart := &model.Cart{
		ID:        fmt.Sprintf("cart-%04d", m.cartSeq),
		OwnerID:   ownerID,
		LineItems: lineItems,
		Currency:  "USD",
		Totals:    pricing.CalculateTotals(lineItems, 0, nil),
	}
	m.carts[cart.ID] = cart
	return cart, nil
}

// GetCart retrieves a cart by ID with access control.
func (m *simpleMerchant) GetCart(id, ownerID string) (*model.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cart, ok := m.carts[id]
	if !ok || !canAccessEntity(ownerID, cart.OwnerID) {
		return nil, fmt.Errorf("cart not found: %s: %w", id, merchant.ErrNotFound)
	}
	return cart, nil
}

// UpdateCart updates the line items of an existing cart.
func (m *simpleMerchant) UpdateCart(id, ownerID string, items []model.LineItemRequest) (*model.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cart, ok := m.carts[id]
	if !ok || !canAccessEntity(ownerID, cart.OwnerID) {
		return nil, fmt.Errorf("cart not found: %s: %w", id, merchant.ErrNotFound)
	}

	if len(items) > 0 {
		lineItems, err := pricing.BuildLineItems(items, m.catalog)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", err.Error(), merchant.ErrBadRequest)
		}
		cart.LineItems = lineItems
		cart.Totals = pricing.CalculateTotals(lineItems, 0, nil)
	}
	return cart, nil
}

// CancelCart cancels and removes a cart.
func (m *simpleMerchant) CancelCart(id, ownerID string) (*model.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cart, ok := m.carts[id]
	if !ok || !canAccessEntity(ownerID, cart.OwnerID) {
		return nil, fmt.Errorf("cart not found: %s: %w", id, merchant.ErrNotFound)
	}
	delete(m.carts, id)
	cart.Messages = append(cart.Messages, model.Message{Type: "info", Text: "Cart has been canceled"})
	return cart, nil
}

// --- Checkouter ---

// CreateCheckout creates a new checkout session per the UCP Shopping Checkout
// capability (dev.ucp.shopping.checkout). Returns the checkout, its hash, and
// any error.
func (m *simpleMerchant) CreateCheckout(ownerID, country string, req *model.CheckoutRequest) (*model.Checkout, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lineItems []model.LineItem
	var err error

	if req.LineItems != nil && len(req.LineItems) > 0 {
		lineItems, err = pricing.BuildLineItems(req.LineItems, m.catalog)
		if err != nil {
			return nil, "", fmt.Errorf("%s: %w", err.Error(), merchant.ErrBadRequest)
		}
	} else {
		return nil, "", fmt.Errorf("checkout must have line_items: %w", merchant.ErrBadRequest)
	}

	// Validate country availability
	if country != "" {
		for _, li := range lineItems {
			p := m.catalog.Find(li.Item.ID)
			if p != nil && len(p.AvailableCountries) > 0 && !ucp.ContainsCountry(p.AvailableCountries, ucp.NewCountry(country)) {
				return nil, "", fmt.Errorf("product %s (%s) is not available in %s: %w", p.ID, p.Title, country, merchant.ErrBadRequest)
			}
		}
	}

	m.checkoutSeq++
	coID := fmt.Sprintf("co_%04d", m.checkoutSeq)

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	co := &model.Checkout{
		ID:        coID,
		Status:    "incomplete",
		UCP:       model.UCPEnvelope{Version: "2026-01-11", Capabilities: []model.Capability{}},
		Links:     []model.Link{{Type: "application/json", URL: fmt.Sprintf("%s://localhost:%d/shopping-api/checkout-sessions/%s", m.scheme(), m.listenPort(), coID)}},
		Currency:  currency,
		LineItems: lineItems,
	}

	co.Totals = pricing.CalculateTotals(lineItems, 0, nil)
	co.Payment = mpayment.ParsePayment(req.Payment)
	co.Fulfillment = mfulfillment.ParseFulfillment(req.Fulfillment, nil, co, m.shopData, m.checkoutDestinations, m.checkoutOptionTitles, &m.addrSeqCounter, &m.addrSeqMu)
	co.Buyer = mpayment.ParseBuyer(req.Buyer)

	m.checkouts[coID] = co
	m.checkoutOwners[coID] = ownerID

	hash := computeCheckoutHashImpl(co)
	return co, hash, nil
}

// GetCheckout retrieves a checkout session by ID with access control.
func (m *simpleMerchant) GetCheckout(id, ownerID string) (*model.Checkout, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	co, ok := m.checkouts[id]
	if !ok {
		return nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	coOwner := m.checkoutOwners[id]
	if !canAccessEntity(ownerID, coOwner) {
		return nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	hash := computeCheckoutHashImpl(co)
	return co, hash, nil
}

// UpdateCheckout updates a checkout session's line items, buyer, payment,
// discounts, or fulfillment configuration.
func (m *simpleMerchant) UpdateCheckout(id, ownerID string, req *model.CheckoutRequest) (*model.Checkout, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	co, ok := m.checkouts[id]
	if !ok {
		return nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	coOwner := m.checkoutOwners[id]
	if !canAccessEntity(ownerID, coOwner) {
		return nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	if co.Status == "completed" || co.Status == "canceled" {
		return nil, "", fmt.Errorf("cannot update checkout in %s status: %w", co.Status, merchant.ErrConflict)
	}

	// Update line items if provided
	if len(req.LineItems) > 0 {
		lineItems, err := pricing.BuildLineItems(req.LineItems, m.catalog)
		if err != nil {
			return nil, "", fmt.Errorf("%s: %w", err.Error(), merchant.ErrBadRequest)
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
		co.Discounts = discount.ApplyDiscounts(req.Discounts, co.LineItems, m.shopData)
	}

	// Handle fulfillment
	if req.Fulfillment != nil {
		co.Fulfillment = mfulfillment.ParseFulfillment(req.Fulfillment, co.Buyer, co, m.shopData, m.checkoutDestinations, m.checkoutOptionTitles, &m.addrSeqCounter, &m.addrSeqMu)
		shippingCost = mfulfillment.GetCurrentShippingCost(co)
	}

	// Recalculate totals
	co.Totals = pricing.CalculateTotals(co.LineItems, shippingCost, co.Discounts)

	hash := computeCheckoutHashImpl(co)
	return co, hash, nil
}

// CompleteCheckout completes a checkout session. If approvalHash is non-empty,
// it is validated against the current checkout state (MCP flow). Returns the
// completed checkout and the created order.
func (m *simpleMerchant) CompleteCheckout(id, ownerID, country, approvalHash string, req *model.CheckoutRequest) (*model.Checkout, *model.Order, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	co, ok := m.checkouts[id]
	if !ok {
		return nil, nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	coOwner := m.checkoutOwners[id]
	if !canAccessEntity(ownerID, coOwner) {
		return nil, nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	if co.Status == "completed" {
		return nil, nil, "", fmt.Errorf("checkout already completed: %w", merchant.ErrConflict)
	}
	if co.Status == "canceled" {
		return nil, nil, "", fmt.Errorf("checkout has been canceled: %w", merchant.ErrConflict)
	}

	// Validate fulfillment is complete (REST flow)
	if approvalHash == "" && !mfulfillment.IsFulfillmentComplete(co) {
		return nil, nil, "", fmt.Errorf("Fulfillment address and option must be selected: %w", merchant.ErrBadRequest)
	}

	// Re-validate country availability at completion time
	effectiveCountry := country
	if effectiveCountry != "" {
		for _, li := range co.LineItems {
			p := m.catalog.Find(li.Item.ID)
			if p != nil && len(p.AvailableCountries) > 0 && !ucp.ContainsCountry(p.AvailableCountries, ucp.NewCountry(effectiveCountry)) {
				return nil, nil, "", fmt.Errorf("product %s (%s) is not available for delivery to %s: %w", p.ID, p.Title, effectiveCountry, merchant.ErrBadRequest)
			}
		}
	}

	// Verify approval hash if provided (MCP flow)
	if approvalHash != "" {
		expectedHash := computeCheckoutHashImpl(co)
		if approvalHash != expectedHash {
			return nil, nil, "", fmt.Errorf("checkout state changed since approval — re-fetch the checkout and request user approval again: %w", merchant.ErrConflict)
		}
	}

	// Process payment (REST flow)
	if req != nil && req.PaymentData != nil {
		if req.PaymentData.Credential != nil && req.PaymentData.Credential.Token == "fail_token" {
			return nil, nil, "", fmt.Errorf("payment failed: %w", merchant.ErrPaymentFailed)
		}
		if req.PaymentData.HandlerID != "" {
			validHandlers := map[string]bool{
				"google_pay":           true,
				"mock_payment_handler": true,
				"shop_pay":             true,
			}
			if !validHandlers[req.PaymentData.HandlerID] {
				return nil, nil, "", fmt.Errorf("unknown payment handler: %s: %w", req.PaymentData.HandlerID, merchant.ErrBadRequest)
			}
		}
	}

	// Create order
	m.orderSeq++
	orderID := fmt.Sprintf("order_%04d", m.orderSeq)

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

	// Build fulfillment expectations
	var expectations []model.Expectation
	optionTitle := m.checkoutOptionTitles[id]
	if optionTitle == "" {
		optionTitle = "Standard Shipping"
	}
	dest := m.checkoutDestinations[id]
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
		PermalinkURL: fmt.Sprintf("%s://localhost:%d/orders/%s", m.scheme(), m.listenPort(), orderID),
		LineItems:    orderLineItems,
		Fulfillment: model.OrderFulfillment{
			Expectations: expectations,
		},
		Currency: co.Currency,
		Totals:   co.Totals,
	}

	m.orders[orderID] = order
	m.orderOwners[orderID] = coOwner

	co.Status = "completed"
	co.Order = &model.OrderRef{
		ID:           orderID,
		PermalinkURL: order.PermalinkURL,
	}

	hash := computeCheckoutHashImpl(co)
	return co, order, hash, nil
}

// CancelCheckout cancels a checkout session. Cannot cancel completed checkouts.
func (m *simpleMerchant) CancelCheckout(id, ownerID string) (*model.Checkout, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	co, ok := m.checkouts[id]
	if !ok {
		return nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	coOwner := m.checkoutOwners[id]
	if !canAccessEntity(ownerID, coOwner) {
		return nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	if co.Status == "canceled" {
		return nil, "", fmt.Errorf("checkout already canceled: %w", merchant.ErrConflict)
	}
	if co.Status == "completed" {
		return nil, "", fmt.Errorf("cannot cancel completed checkout: %w", merchant.ErrConflict)
	}

	co.Status = "canceled"
	hash := computeCheckoutHashImpl(co)
	return co, hash, nil
}

// --- Orderer ---

// GetOrder retrieves an order by ID with access control.
func (m *simpleMerchant) GetOrder(id, ownerID string) (*model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ord, ok := m.orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s: %w", id, merchant.ErrNotFound)
	}
	ordOwner := m.orderOwners[id]
	if !canAccessEntity(ownerID, ordOwner) {
		return nil, fmt.Errorf("order not found: %s: %w", id, merchant.ErrNotFound)
	}
	return ord, nil
}

// ListOrders returns all orders belonging to the given owner.
func (m *simpleMerchant) ListOrders(ownerID string) ([]*model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []*model.Order
	for _, ord := range m.orders {
		ordOwner := m.orderOwners[ord.ID]
		if canAccessEntity(ownerID, ordOwner) {
			result = append(result, ord)
		}
	}
	return result, nil
}

// CancelOrder cancels an order.
func (m *simpleMerchant) CancelOrder(id, ownerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.orders[id]
	if !ok {
		return fmt.Errorf("order not found: %s: %w", id, merchant.ErrNotFound)
	}
	ordOwner := m.orderOwners[id]
	if !canAccessEntity(ownerID, ordOwner) {
		return fmt.Errorf("order not found: %s: %w", id, merchant.ErrNotFound)
	}
	return nil
}

// UpdateOrder updates an order's fulfillment events, expectations, or adjustments.
func (m *simpleMerchant) UpdateOrder(id string, req model.OrderUpdateRequest) (*model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	order, ok := m.orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s: %w", id, merchant.ErrNotFound)
	}

	if req.Fulfillment != nil {
		if req.Fulfillment.Events != nil {
			order.Fulfillment.Events = req.Fulfillment.Events
		}
		if req.Fulfillment.Expectations != nil {
			order.Fulfillment.Expectations = req.Fulfillment.Expectations
		}
	}

	if req.Adjustments != nil {
		order.Adjustments = req.Adjustments
	}

	return order, nil
}

// Reset clears all transient state (checkouts, orders, carts).
func (m *simpleMerchant) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.checkouts = map[string]*model.Checkout{}
	m.orders = map[string]*model.Order{}
	m.carts = map[string]*model.Cart{}
	m.checkoutSeq = 0
	m.orderSeq = 0
	m.cartSeq = 0
	m.orderOwners = map[string]string{}
	m.checkoutOwners = map[string]string{}
	m.checkoutDestinations = map[string]*model.FulfillmentDestination{}
	m.checkoutOptionTitles = map[string]string{}
	m.addrSeqCounter = 0

	if m.shopData != nil {
		m.shopData.ResetDynamicAddresses()
	}
}

// GetCheckoutID returns the checkout ID associated with an order.
func (m *simpleMerchant) GetCheckoutID(orderID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ord, ok := m.orders[orderID]; ok {
		return ord.CheckoutID
	}
	return ""
}

// --- Helpers ---

// canAccessEntity checks if a user can access an entity.
func canAccessEntity(userID, ownerID string) bool {
	if userID == "" {
		return ownerID == ""
	}
	return ownerID == userID
}

// computeCheckoutHashImpl produces a SHA-256 hash of the material checkout fields.
func computeCheckoutHashImpl(co *model.Checkout) string {
	type hashLineItem struct {
		ItemID   string `json:"item_id"`
		Title    string `json:"title"`
		Price    int    `json:"price"`
		Quantity int    `json:"quantity"`
	}
	type hashData struct {
		ID        string         `json:"id"`
		LineItems []hashLineItem `json:"line_items"`
		Currency  string         `json:"currency"`
		Totals    []model.Total  `json:"totals"`
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
	}
	b, _ := json.Marshal(data)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
