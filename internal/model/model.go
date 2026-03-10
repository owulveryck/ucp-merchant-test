package model

import "github.com/owulveryck/ucp-merchant-test/internal/ucp"

// UCPEnvelope carries protocol metadata in checkout and order responses.
//
// The "ucp" field is always an object (never a plain string) containing the
// protocol version (e.g., "2026-01-11") and the capabilities array listing
// extensions active for this resource.
type UCPEnvelope struct {
	Version      string       `json:"version"`
	Capabilities []Capability `json:"capabilities"`
}

// Capability identifies a UCP extension active on a checkout or order resource.
// Each capability has a name (e.g., "dev.ucp.shopping.fulfillment") and version.
type Capability struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// Checkout represents a UCP checkout session (dev.ucp.shopping.checkout).
//
// A checkout session is initiated when a user expresses purchase intent and
// progresses through a defined status lifecycle:
//   - incomplete: missing required information, inspect Messages for context
//   - requires_escalation: buyer handoff needed via ContinueURL
//   - ready_for_complete: all information collected, Complete may be called
//   - complete_in_progress: business is processing the completion
//   - completed: order placed successfully, the Order field is populated
//   - canceled: session invalid or expired
//
// The business remains the Merchant of Record (MoR). Update operations use
// full replacement semantics.
type Checkout struct {
	ID          string       `json:"id"`
	UCP         UCPEnvelope  `json:"ucp"`
	Status      string       `json:"status"`
	Currency    ucp.Currency `json:"currency"`
	LineItems   []LineItem   `json:"line_items"`
	Totals      []Total      `json:"totals"`
	Links       []Link       `json:"links"`
	Payment     Payment      `json:"payment"`
	Fulfillment *Fulfillment `json:"fulfillment,omitempty"`
	Buyer       *Buyer       `json:"buyer,omitempty"`
	Order       *OrderRef    `json:"order,omitempty"`
	Discounts   *Discounts   `json:"discounts,omitempty"`
}

// Link represents a hypermedia link in a UCP resource.
// The Type field carries the media type (e.g., "application/json"), not a rel value.
type Link struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// LineItem represents a purchasable item with quantity and computed totals in a
// checkout or cart. Totals types are standardized by UCP: "subtotal", "items_discount",
// "discount", "fulfillment", "tax", "fee", and "total". Amounts are in minor
// currency units (e.g., cents).
type LineItem struct {
	ID       string  `json:"id"`
	Item     Item    `json:"item"`
	Quantity int     `json:"quantity"`
	Totals   []Total `json:"totals"`
}

// Item identifies a product within a line item, carrying the catalog product ID,
// title, unit price, and optional image URL.
type Item struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Price    int    `json:"price,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// Total represents a pricing total in a checkout, order, or line item.
// Type must be one of: "items_discount", "subtotal", "discount", "fulfillment",
// "tax", "fee", "total". Discount amounts are always positive (>= 0);
// platforms display them as subtractive. Amounts are in minor currency units.
type Total struct {
	Type        string `json:"type"`
	DisplayText string `json:"display_text,omitempty"`
	Amount      int    `json:"amount"`
}

// OrderRef is a reference to the created order, populated when a checkout
// reaches "completed" status. Contains the order ID and permalink URL.
type OrderRef struct {
	ID           string `json:"id"`
	PermalinkURL string `json:"permalink_url"`
}

// Fulfillment models the UCP Fulfillment extension (dev.ucp.shopping.fulfillment).
//
// It structures physical delivery into Methods (shipping, pickup, etc.), each
// containing Destinations (where to fulfill), Groups (business-generated packages),
// and Options (selectable shipping speeds with costs). Fulfillment is optional in
// the checkout to support digital goods that need no physical delivery.
type Fulfillment struct {
	Methods []FulfillmentMethod `json:"methods"`
}

// FulfillmentMethod represents a delivery method (e.g., "shipping", "pickup")
// within the fulfillment hierarchy. It links line items to destinations and
// groups, and tracks the buyer's selected destination.
type FulfillmentMethod struct {
	ID                    string                   `json:"id"`
	Type                  string                   `json:"type"`
	LineItemIDs           []string                 `json:"line_item_ids"`
	Destinations          []FulfillmentDestination `json:"destinations,omitempty"`
	SelectedDestinationID string                   `json:"selected_destination_id,omitempty"`
	Groups                []FulfillmentGroup       `json:"groups,omitempty"`
}

// FulfillmentGroup is a business-generated package grouping line items that
// ship together. Each group offers selectable shipping Options with costs.
type FulfillmentGroup struct {
	ID               string              `json:"id"`
	LineItemIDs      []string            `json:"line_item_ids"`
	Options          []FulfillmentOption `json:"options,omitempty"`
	SelectedOptionID string              `json:"selected_option_id,omitempty"`
}

// FulfillmentOption is a selectable shipping speed/cost within a fulfillment group
// (e.g., "Standard Shipping", "Express Shipping"). Totals carry the fulfillment cost.
type FulfillmentOption struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Totals []Total `json:"totals"`
}

// FulfillmentDestination is a shipping address within a fulfillment method.
// Fields follow Schema.org PostalAddress naming (street_address, address_locality, etc.).
type FulfillmentDestination struct {
	ID              string      `json:"id,omitempty"`
	FullName        string      `json:"full_name,omitempty"`
	StreetAddress   string      `json:"street_address,omitempty"`
	AddressLocality string      `json:"address_locality,omitempty"`
	AddressRegion   string      `json:"address_region,omitempty"`
	PostalCode      string      `json:"postal_code,omitempty"`
	AddressCountry  ucp.Country `json:"address_country,omitempty"`
}

// Payment models payment configuration for a checkout session.
//
// Payment handlers are discovered from the business's UCP profile at /.well-known/ucp.
// Handlers define processing specifications for collecting payment instruments
// (e.g., Google Pay, Shop Pay). When the buyer submits payment, the platform
// populates the Instruments array with collected instrument data. Payment is
// required (not optional) in checkout responses.
type Payment struct {
	SelectedInstrumentID string                   `json:"selected_instrument_id,omitempty"`
	Instruments          []map[string]interface{} `json:"instruments"`
	Handlers             []map[string]interface{} `json:"handlers"`
}

// Buyer represents the person making the purchase, with fields for name,
// email, and phone number. Buyer information enables identity linking,
// address lookup, and personalized fulfillment options.
type Buyer struct {
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	FullName  string   `json:"fullName,omitempty"`
	Email     string   `json:"email,omitempty"`
	Consent   *Consent `json:"consent,omitempty"`
}

// Consent captures buyer consent preferences for marketing, analytics,
// and sale of data as defined by UCP privacy extensions.
type Consent struct {
	Marketing  *bool `json:"marketing,omitempty"`
	Analytics  *bool `json:"analytics,omitempty"`
	SaleOfData *bool `json:"sale_of_data,omitempty"`
}

// Discounts implements the UCP Discount Extension (dev.ucp.shopping.discount).
//
// It supports code-based discounts (percentage or fixed amount) and automatic
// discounts applied by business rules. Applied discounts include human-readable
// titles and amounts. Rejected codes are communicated via the Messages array.
type Discounts struct {
	Codes   []string          `json:"codes,omitempty"`
	Applied []AppliedDiscount `json:"applied,omitempty"`
}

// AppliedDiscount represents a single discount that has been validated and applied
// to the checkout. Amount is always positive; platforms display it as subtractive.
type AppliedDiscount struct {
	Code      string `json:"code,omitempty"`
	Title     string `json:"title"`
	Amount    int    `json:"amount"`
	Automatic bool   `json:"automatic,omitempty"`
}

// Order represents a confirmed transaction resulting from a successful checkout
// completion (dev.ucp.shopping.order).
//
// Orders have three main components:
//   - LineItems: what was purchased, with quantity counts (total, fulfilled)
//   - Fulfillment: how items get delivered, including Expectations (buyer-facing
//     promises) and Events (append-only log of what actually happened)
//   - Adjustments: post-order events independent of fulfillment, typically money
//     movements such as refunds, returns, credits, disputes, and cancellations
//
// The business sends the full order entity on each update (not incremental deltas).
type Order struct {
	ID           string           `json:"id"`
	UCP          UCPEnvelope      `json:"ucp"`
	CheckoutID   string           `json:"checkout_id"`
	PermalinkURL string           `json:"permalink_url"`
	LineItems    []OrderLineItem  `json:"line_items"`
	Fulfillment  OrderFulfillment `json:"fulfillment"`
	Adjustments  []Adjustment     `json:"adjustments,omitempty"`
	Currency     ucp.Currency     `json:"currency"`
	Totals       []Total          `json:"totals"`
}

// OrderLineItem extends LineItem with order-specific fields: quantity tracking
// (total vs fulfilled), status, and an optional ParentID for item hierarchies.
type OrderLineItem struct {
	ID       string        `json:"id"`
	Item     Item          `json:"item"`
	Quantity OrderQuantity `json:"quantity"`
	Totals   []Total       `json:"totals"`
	Status   string        `json:"status"`
	ParentID *string       `json:"parent_id,omitempty"`
}

// OrderQuantity tracks the total quantity ordered and quantity fulfilled so far.
type OrderQuantity struct {
	Total     int `json:"total"`
	Fulfilled int `json:"fulfilled"`
}

// OrderFulfillment contains buyer-facing delivery promises (Expectations) and
// an append-only log of fulfillment events (Events) such as shipments and deliveries.
type OrderFulfillment struct {
	Expectations []Expectation      `json:"expectations,omitempty"`
	Events       []FulfillmentEvent `json:"events,omitempty"`
}

// Expectation is a buyer-facing delivery promise within an order, describing
// which line items will be fulfilled, by what method, and to which destination.
type Expectation struct {
	ID          string                 `json:"id"`
	LineItems   []EventLineItem        `json:"line_items"`
	MethodType  string                 `json:"method_type"`
	Destination FulfillmentDestination `json:"destination"`
	Description string                 `json:"description,omitempty"`
}

// FulfillmentEvent is an append-only record of a fulfillment action (e.g.,
// "shipped", "delivered"). Events carry timestamps, tracking information,
// and affected line items.
type FulfillmentEvent struct {
	ID             string          `json:"id"`
	OccurredAt     string          `json:"occurred_at"`
	Type           string          `json:"type"`
	LineItems      []EventLineItem `json:"line_items,omitempty"`
	TrackingNumber string          `json:"tracking_number,omitempty"`
	TrackingURL    string          `json:"tracking_url,omitempty"`
	Description    string          `json:"description,omitempty"`
}

// EventLineItem identifies a line item and quantity affected by a fulfillment event.
type EventLineItem struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

// Adjustment represents a post-order event independent of fulfillment, such as
// a refund, return, credit, dispute, or cancellation. Adjustments are typically
// money movements with a status lifecycle.
type Adjustment struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	OccurredAt  string `json:"occurred_at"`
	Status      string `json:"status"`
	Amount      int    `json:"amount"`
	Description string `json:"description,omitempty"`
}

// Cart represents a UCP cart session (dev.ucp.shopping.cart).
//
// Cart enables basket building without the complexity of checkout. While
// checkout manages payment handlers, status lifecycle, and order finalization,
// cart provides a lightweight CRUD interface for item collection before
// purchase intent is established. Cart totals are estimates; accurate pricing
// is computed at checkout.
type Cart struct {
	ID        string       `json:"id"`
	OwnerID   string       `json:"owner_id,omitempty"`
	LineItems []LineItem   `json:"line_items"`
	Currency  ucp.Currency `json:"currency"`
	Totals    []Total      `json:"totals"`
	Messages  []Message    `json:"messages,omitempty"`
}

// Message communicates business outcomes using UCP severity levels:
// "recoverable", "requires_buyer_input", "requires_buyer_review", and
// "unrecoverable".
type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
