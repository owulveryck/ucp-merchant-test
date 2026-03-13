package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	icatalog "github.com/owulveryck/ucp-merchant-test/pkg/catalog"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/payment"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/pricing"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/ucp"
)

// MerchantConfig holds the configurable state for an arena merchant.
type MerchantConfig struct {
	mu            sync.RWMutex
	SellingPrice  int            `json:"selling_price"`
	Stock         int            `json:"stock"`
	DiscountCodes []DiscountCode `json:"discount_codes"`
	BoostScore    int            `json:"boost_score"`
}

// DiscountCode is a configurable discount.
type DiscountCode struct {
	Code            string `json:"code"`
	Type            string `json:"type"`
	Value           int    `json:"value"`
	NewCustomerOnly bool   `json:"new_customer_only"`
}

// arenaMerchant implements merchant.Merchant for a single arena tenant.
type arenaMerchant struct {
	mu       sync.Mutex
	config   *MerchantConfig
	product  icatalog.Product
	baseURL  func() string
	notifier *Notifier

	// Cost & profit tracking
	costPrice   int // prix d'achat en cents
	totalProfit int // profit net cumule (marge - cout boost) en cents
	salesCount  int // nombre de ventes completees

	// State
	checkouts      map[string]*model.Checkout
	orders         map[string]*model.Order
	carts          map[string]*model.Cart
	checkoutSeq    int
	orderSeq       int
	cartSeq        int
	checkoutOwners map[string]string
	orderOwners    map[string]string

	// New customer tracking
	purchaseHistory map[string]bool // email -> has purchased

	// Optional callback for sale events (forwarding to obs-hub)
	onSale func(SaleEvent)

	// Optional callback for activity events (forwarding to obs-hub + merchant dashboard)
	onActivity func(eventType, summary string)
}

func newArenaMerchant(productID, productName string, costPrice int, baseURL func() string, notifier *Notifier) *arenaMerchant {
	return &arenaMerchant{
		costPrice: costPrice,
		config: &MerchantConfig{
			SellingPrice: costPrice + 1000, // default: cost + $10
			Stock:        10,
			BoostScore:   50,
		},
		product: icatalog.Product{
			ID:                 productID,
			Title:              productName,
			Price:              costPrice + 1000,
			Quantity:           10,
			Description:        fmt.Sprintf("Premium %s", productName),
			AvailableCountries: []ucp.Country{"US", "FR", "GB", "DE"},
		},
		baseURL:         baseURL,
		notifier:        notifier,
		checkouts:       make(map[string]*model.Checkout),
		orders:          make(map[string]*model.Order),
		carts:           make(map[string]*model.Cart),
		checkoutOwners:  make(map[string]string),
		orderOwners:     make(map[string]string),
		purchaseHistory: make(map[string]bool),
	}
}

// notifyActivity sends an activity event to the merchant dashboard SSE and obs-hub.
func (m *arenaMerchant) notifyActivity(eventType, summary string) {
	data, _ := json.Marshal(map[string]string{
		"type":    eventType,
		"summary": summary,
	})
	go m.notifier.SendRaw(data)
	if m.onActivity != nil {
		go m.onActivity(eventType, summary)
	}
}

// currentProduct returns the product with current config values.
func (m *arenaMerchant) currentProduct() icatalog.Product {
	m.config.mu.RLock()
	defer m.config.mu.RUnlock()
	p := m.product
	p.Price = m.config.SellingPrice
	p.Quantity = m.config.Stock
	return p
}

// --- Cataloger ---

func (m *arenaMerchant) Find(id string) *icatalog.Product {
	p := m.currentProduct()
	if p.ID == id {
		m.notifyActivity("product_details", fmt.Sprintf("Détails produit consultés: %s", p.Title))
		return &p
	}
	return nil
}

func (m *arenaMerchant) Filter(category ucp.Category, brand, query string, country ucp.Country, currency ucp.Currency, language ucp.Language) []icatalog.Product {
	p := m.currentProduct()
	if p.Quantity <= 0 {
		return nil
	}
	if query != "" && !strings.Contains(strings.ToLower(p.Title), strings.ToLower(query)) {
		return nil
	}
	m.notifyActivity("catalog_browse", "Catalogue consulté")
	return []icatalog.Product{p}
}

func (m *arenaMerchant) CategoryCount() []icatalog.CategoryStat {
	return []icatalog.CategoryStat{{Name: "audio", Count: 1}}
}

func (m *arenaMerchant) Lookup(id string, shipsTo ucp.Country) *icatalog.Product {
	p := m.currentProduct()
	if p.ID == id {
		m.notifyActivity("product_details", fmt.Sprintf("Détails produit consultés: %s", p.Title))
		return &p
	}
	return nil
}

func (m *arenaMerchant) Search(params icatalog.SearchParams) []icatalog.SearchResult {
	p := m.currentProduct()
	if params.Query != "" && !strings.Contains(strings.ToLower(p.Title), strings.ToLower(params.Query)) {
		return nil
	}
	m.notifyActivity("catalog_browse", "Catalogue consulté (recherche)")
	return []icatalog.SearchResult{
		{Product: p},
	}
}

// --- Carter ---

func (m *arenaMerchant) CreateCart(ownerID string, items []model.LineItemRequest) (*model.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(items) == 0 {
		return nil, fmt.Errorf("cart must have at least one line item: %w", merchant.ErrBadRequest)
	}

	lineItems, err := m.buildLineItems(items)
	if err != nil {
		return nil, err
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
	m.notifyActivity("cart_created", fmt.Sprintf("Panier créé (%s)", cart.ID))
	return cart, nil
}

func (m *arenaMerchant) GetCart(id, ownerID string) (*model.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cart, ok := m.carts[id]
	if !ok {
		return nil, fmt.Errorf("cart not found: %s: %w", id, merchant.ErrNotFound)
	}
	return cart, nil
}

func (m *arenaMerchant) UpdateCart(id, ownerID string, items []model.LineItemRequest) (*model.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cart, ok := m.carts[id]
	if !ok {
		return nil, fmt.Errorf("cart not found: %s: %w", id, merchant.ErrNotFound)
	}
	if len(items) > 0 {
		lineItems, err := m.buildLineItems(items)
		if err != nil {
			return nil, err
		}
		cart.LineItems = lineItems
		cart.Totals = pricing.CalculateTotals(lineItems, 0, nil)
	}
	return cart, nil
}

func (m *arenaMerchant) CancelCart(id, ownerID string) (*model.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cart, ok := m.carts[id]
	if !ok {
		return nil, fmt.Errorf("cart not found: %s: %w", id, merchant.ErrNotFound)
	}
	delete(m.carts, id)
	cart.Messages = append(cart.Messages, model.Message{Type: "info", Text: "Cart has been canceled"})
	return cart, nil
}

// --- Checkouter ---

func (m *arenaMerchant) CreateCheckout(ownerID string, country ucp.Country, req *model.CheckoutRequest) (*model.Checkout, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.LineItems == nil || len(req.LineItems) == 0 {
		return nil, "", fmt.Errorf("checkout must have line_items: %w", merchant.ErrBadRequest)
	}

	lineItems, err := m.buildLineItems(req.LineItems)
	if err != nil {
		return nil, "", err
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
		Links:     []model.Link{{Type: "application/json", URL: fmt.Sprintf("%s/shopping-api/checkout-sessions/%s", m.baseURL(), coID)}},
		Currency:  currency,
		LineItems: lineItems,
	}

	co.Totals = pricing.CalculateTotals(lineItems, 0, nil)
	co.Payment = payment.ParsePayment(req.Payment)
	co.Buyer = payment.ParseBuyer(req.Buyer)

	// Set up basic fulfillment if requested
	if req.Fulfillment != nil {
		co.Fulfillment = m.parseFulfillment(req.Fulfillment, co)
	}

	m.checkouts[coID] = co
	m.checkoutOwners[coID] = ownerID

	m.notifyActivity("checkout_created", fmt.Sprintf("Checkout créé (%s)", coID))

	hash := computeHash(co)
	return co, hash, nil
}

func (m *arenaMerchant) GetCheckout(id, ownerID string) (*model.Checkout, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	co, ok := m.checkouts[id]
	if !ok {
		return nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}
	hash := computeHash(co)
	return co, hash, nil
}

func (m *arenaMerchant) UpdateCheckout(id, ownerID string, req *model.CheckoutRequest) (*model.Checkout, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	co, ok := m.checkouts[id]
	if !ok {
		return nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	if co.Status == "completed" || co.Status == "canceled" {
		return nil, "", fmt.Errorf("cannot update checkout in %s status: %w", co.Status, merchant.ErrConflict)
	}

	if len(req.LineItems) > 0 {
		lineItems, err := m.buildLineItems(req.LineItems)
		if err != nil {
			return nil, "", err
		}
		co.LineItems = lineItems
	}

	if req.Buyer != nil {
		co.Buyer = payment.ParseBuyer(req.Buyer)
	}

	if req.Payment != nil {
		co.Payment = payment.ParsePayment(req.Payment)
	}

	// Handle discounts
	if req.Discounts != nil {
		co.Discounts = m.applyDiscounts(req.Discounts, co)
	}

	// Handle fulfillment
	if req.Fulfillment != nil {
		co.Fulfillment = m.parseFulfillment(req.Fulfillment, co)
	}

	// Recalculate totals
	shippingCost := m.getShippingCost(co)
	co.Totals = pricing.CalculateTotals(co.LineItems, shippingCost, co.Discounts)

	// Emit activity with detail
	detail := "Checkout mis à jour"
	if req.Discounts != nil {
		detail = "Checkout mis à jour (codes promo)"
	} else if req.Fulfillment != nil {
		detail = "Checkout mis à jour (livraison)"
	} else if req.Buyer != nil {
		detail = "Checkout mis à jour (acheteur)"
	} else if req.Payment != nil {
		detail = "Checkout mis à jour (paiement)"
	}
	m.notifyActivity("checkout_updated", detail)

	hash := computeHash(co)
	return co, hash, nil
}

func (m *arenaMerchant) CompleteCheckout(id, ownerID string, country ucp.Country, approvalHash string, req *model.CheckoutRequest) (*model.Checkout, *model.Order, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	co, ok := m.checkouts[id]
	if !ok {
		return nil, nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	if co.Status == "completed" {
		return nil, nil, "", fmt.Errorf("checkout already completed: %w", merchant.ErrConflict)
	}
	if co.Status == "canceled" {
		return nil, nil, "", fmt.Errorf("checkout has been canceled: %w", merchant.ErrConflict)
	}

	// Verify approval hash if provided (MCP flow)
	if approvalHash != "" {
		expectedHash := computeHash(co)
		if approvalHash != expectedHash {
			return nil, nil, "", fmt.Errorf("checkout state changed since approval: %w", merchant.ErrConflict)
		}
	}

	// Process payment
	if req != nil && req.PaymentData != nil {
		if req.PaymentData.Credential != nil && req.PaymentData.Credential.Token == "fail_token" {
			return nil, nil, "", fmt.Errorf("payment failed: %w", merchant.ErrPaymentFailed)
		}
	}

	// Decrement stock
	m.config.mu.Lock()
	qty := 0
	for _, li := range co.LineItems {
		qty += li.Quantity
	}
	if m.config.Stock < qty {
		m.config.mu.Unlock()
		return nil, nil, "", fmt.Errorf("insufficient stock: %w", merchant.ErrBadRequest)
	}
	m.config.Stock -= qty
	m.config.mu.Unlock()

	// Record buyer email for new customer tracking
	if co.Buyer != nil && co.Buyer.Email != "" {
		m.purchaseHistory[co.Buyer.Email] = true
	}

	// Create order
	m.orderSeq++
	orderID := fmt.Sprintf("order_%04d", m.orderSeq)

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
		CheckoutID:   id,
		PermalinkURL: fmt.Sprintf("%s/orders/%s", m.baseURL(), orderID),
		LineItems:    orderLineItems,
		Currency:     co.Currency,
		Totals:       co.Totals,
	}

	m.orders[orderID] = order
	m.orderOwners[orderID] = m.checkoutOwners[id]

	co.Status = "completed"
	co.Order = &model.OrderRef{
		ID:           orderID,
		PermalinkURL: order.PermalinkURL,
	}

	hash := computeHash(co)

	// Calculate profit with boost cost
	m.config.mu.RLock()
	boostScore := m.config.BoostScore
	sellingPrice := m.config.SellingPrice
	m.config.mu.RUnlock()

	margin := sellingPrice - m.costPrice
	boostCost := boostScore * margin / 100
	netProfit := (margin - boostCost) * qty
	m.totalProfit += netProfit
	m.salesCount++

	// Send sale notification
	buyerEmail := ""
	if co.Buyer != nil {
		buyerEmail = co.Buyer.Email
	}
	total := 0
	for _, t := range co.Totals {
		if t.Type == "total" {
			total = t.Amount
		}
	}
	saleEvent := SaleEvent{
		Type:        "sale",
		OrderID:     orderID,
		Buyer:       buyerEmail,
		Total:       total,
		BoostCost:   boostCost * qty,
		NetProfit:   netProfit,
		TotalProfit: m.totalProfit,
	}
	go m.notifier.Send(saleEvent)
	if m.onSale != nil {
		go m.onSale(saleEvent)
	}

	return co, order, hash, nil
}

func (m *arenaMerchant) CancelCheckout(id, ownerID string) (*model.Checkout, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	co, ok := m.checkouts[id]
	if !ok {
		return nil, "", fmt.Errorf("checkout not found: %s: %w", id, merchant.ErrNotFound)
	}

	if co.Status == "canceled" {
		return nil, "", fmt.Errorf("checkout already canceled: %w", merchant.ErrConflict)
	}
	if co.Status == "completed" {
		return nil, "", fmt.Errorf("cannot cancel completed checkout: %w", merchant.ErrConflict)
	}

	co.Status = "canceled"
	m.notifyActivity("checkout_canceled", fmt.Sprintf("Checkout annulé (%s)", id))
	hash := computeHash(co)
	return co, hash, nil
}

// --- Orderer ---

func (m *arenaMerchant) GetOrder(id, ownerID string) (*model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ord, ok := m.orders[id]
	if !ok {
		return nil, fmt.Errorf("order not found: %s: %w", id, merchant.ErrNotFound)
	}
	return ord, nil
}

func (m *arenaMerchant) ListOrders(ownerID string) ([]*model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*model.Order
	for _, ord := range m.orders {
		result = append(result, ord)
	}
	return result, nil
}

func (m *arenaMerchant) CancelOrder(id, ownerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.orders[id]; !ok {
		return fmt.Errorf("order not found: %s: %w", id, merchant.ErrNotFound)
	}
	return nil
}

func (m *arenaMerchant) UpdateOrder(id string, req model.OrderUpdateRequest) (*model.Order, error) {
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

// Reset clears all transient state.
func (m *arenaMerchant) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkouts = make(map[string]*model.Checkout)
	m.orders = make(map[string]*model.Order)
	m.carts = make(map[string]*model.Cart)
	m.checkoutSeq = 0
	m.orderSeq = 0
	m.cartSeq = 0
	m.checkoutOwners = make(map[string]string)
	m.orderOwners = make(map[string]string)
}

// --- Helpers ---

func (m *arenaMerchant) buildLineItems(items []model.LineItemRequest) ([]model.LineItem, error) {
	p := m.currentProduct()
	var lineItems []model.LineItem
	for i, req := range items {
		pid := ""
		if req.Item != nil {
			pid = req.Item.ID
		}
		if pid == "" {
			pid = req.ProductID
		}
		if pid != p.ID {
			return nil, fmt.Errorf("product not found: %s: %w", pid, merchant.ErrBadRequest)
		}
		if p.Quantity <= 0 {
			return nil, fmt.Errorf("product out of stock: %s: %w", pid, merchant.ErrBadRequest)
		}
		qty := req.Quantity
		if qty <= 0 {
			qty = 1
		}

		liID := fmt.Sprintf("li_%03d", i+1)
		subtotal := p.Price * qty
		lineItems = append(lineItems, model.LineItem{
			ID: liID,
			Item: model.Item{
				ID:    p.ID,
				Title: p.Title,
				Price: p.Price,
			},
			Quantity: qty,
			Totals: []model.Total{
				{Type: "subtotal", Amount: subtotal, DisplayText: fmt.Sprintf("$%.2f", float64(subtotal)/100)},
				{Type: "total", Amount: subtotal, DisplayText: fmt.Sprintf("$%.2f", float64(subtotal)/100)},
			},
		})
	}
	return lineItems, nil
}

// applyDiscounts handles discount code application with new customer logic.
func (m *arenaMerchant) applyDiscounts(req *model.DiscountsRequest, co *model.Checkout) *model.Discounts {
	if req == nil || len(req.Codes) == 0 {
		return nil
	}

	m.config.mu.RLock()
	codes := make([]DiscountCode, len(m.config.DiscountCodes))
	copy(codes, m.config.DiscountCodes)
	m.config.mu.RUnlock()

	result := &model.Discounts{
		Codes: req.Codes,
	}

	subtotal := 0
	for _, li := range co.LineItems {
		for _, t := range li.Totals {
			if t.Type == "subtotal" {
				subtotal += t.Amount
			}
		}
	}

	buyerEmail := ""
	if co.Buyer != nil {
		buyerEmail = co.Buyer.Email
	}

	for _, submittedCode := range req.Codes {
		for _, dc := range codes {
			if !strings.EqualFold(dc.Code, submittedCode) {
				continue
			}
			// Check new customer restriction
			if dc.NewCustomerOnly && buyerEmail != "" && m.purchaseHistory[buyerEmail] {
				continue
			}

			var amount int
			var title string
			switch dc.Type {
			case "percentage":
				amount = subtotal * dc.Value / 100
				title = fmt.Sprintf("%d%% off", dc.Value)
			case "fixed":
				amount = dc.Value
				title = fmt.Sprintf("$%.2f off", float64(dc.Value)/100)
			default:
				continue
			}

			result.Applied = append(result.Applied, model.AppliedDiscount{
				Code:   dc.Code,
				Title:  title,
				Amount: amount,
			})
			break
		}
	}

	return result
}

// parseFulfillment creates a basic fulfillment structure.
func (m *arenaMerchant) parseFulfillment(req *model.FulfillmentRequest, co *model.Checkout) *model.Fulfillment {
	if req == nil {
		return nil
	}

	var lineItemIDs []string
	for _, li := range co.LineItems {
		lineItemIDs = append(lineItemIDs, li.ID)
	}

	f := &model.Fulfillment{}

	for _, mr := range req.Methods {
		method := model.FulfillmentMethod{
			ID:          "method_1",
			Type:        "shipping",
			LineItemIDs: lineItemIDs,
		}

		// Handle destinations
		for _, dr := range mr.Destinations {
			dest := model.FulfillmentDestination{
				ID: dr.ID,
			}
			if dr.FullName != "" {
				dest.FullName = dr.FullName
			}
			if dr.StreetAddress != "" {
				dest.StreetAddress = dr.StreetAddress
			}
			if dr.AddressLocality != "" {
				dest.AddressLocality = dr.AddressLocality
			}
			if dr.AddressRegion != "" {
				dest.AddressRegion = dr.AddressRegion
			}
			if dr.PostalCode != "" {
				dest.PostalCode = dr.PostalCode
			}
			if dr.AddressCountry != "" {
				dest.AddressCountry = dr.AddressCountry
			}
			if dest.ID == "" {
				dest.ID = "addr_1"
			}
			method.Destinations = append(method.Destinations, dest)
		}

		// Handle buyer address lookup
		if co.Buyer != nil && co.Buyer.Email != "" && len(method.Destinations) == 0 {
			method.Destinations = []model.FulfillmentDestination{
				{
					ID:              "addr_default",
					FullName:        fmt.Sprintf("%s %s", co.Buyer.FirstName, co.Buyer.LastName),
					StreetAddress:   "123 Main St",
					AddressLocality: "Anytown",
					AddressRegion:   "CA",
					PostalCode:      "90210",
					AddressCountry:  "US",
				},
			}
		}

		if mr.SelectedDestinationID != "" {
			method.SelectedDestinationID = mr.SelectedDestinationID

			// Generate shipping options when destination is selected
			shippingCost := 499
			method.Groups = []model.FulfillmentGroup{
				{
					ID:          "group_1",
					LineItemIDs: lineItemIDs,
					Options: []model.FulfillmentOption{
						{
							ID:    "option_standard",
							Title: "Standard Shipping",
							Totals: []model.Total{
								{Type: "fulfillment", Amount: shippingCost, DisplayText: fmt.Sprintf("$%.2f", float64(shippingCost)/100)},
								{Type: "total", Amount: shippingCost, DisplayText: fmt.Sprintf("$%.2f", float64(shippingCost)/100)},
							},
						},
					},
				},
			}
		}

		// Handle selected option
		for _, gr := range mr.Groups {
			if gr.SelectedOptionID != "" && len(method.Groups) > 0 {
				method.Groups[0].SelectedOptionID = gr.SelectedOptionID
			}
		}

		f.Methods = append(f.Methods, method)
	}

	if len(f.Methods) == 0 {
		f.Methods = []model.FulfillmentMethod{
			{
				ID:          "method_1",
				Type:        "shipping",
				LineItemIDs: lineItemIDs,
			},
		}
	}

	return f
}

func (m *arenaMerchant) getShippingCost(co *model.Checkout) int {
	if co.Fulfillment == nil {
		return 0
	}
	for _, method := range co.Fulfillment.Methods {
		for _, group := range method.Groups {
			if group.SelectedOptionID != "" {
				for _, opt := range group.Options {
					if opt.ID == group.SelectedOptionID {
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

func computeHash(co *model.Checkout) string {
	type hashLineItem struct {
		ItemID   string `json:"item_id"`
		Title    string `json:"title"`
		Price    int    `json:"price"`
		Quantity int    `json:"quantity"`
	}
	type hashData struct {
		ID        string         `json:"id"`
		LineItems []hashLineItem `json:"line_items"`
		Currency  ucp.Currency   `json:"currency"`
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
