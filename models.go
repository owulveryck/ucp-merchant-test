package main

// UCP canonical data types used by both REST and MCP transports.

// UCPEnvelope is the "ucp" envelope in checkout/order responses.
type UCPEnvelope struct {
	Version      string       `json:"version"`
	Capabilities []Capability `json:"capabilities"`
}

type Capability struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// Checkout is the checkout model returned by both REST and MCP endpoints.
type Checkout struct {
	ID          string       `json:"id"`
	UCP         UCPEnvelope  `json:"ucp"`
	Status      string       `json:"status"`
	Currency    string       `json:"currency"`
	LineItems   []LineItem   `json:"line_items"`
	Totals      []Total      `json:"totals"`
	Links       []Link       `json:"links"`
	Payment     Payment      `json:"payment"`
	Fulfillment *Fulfillment `json:"fulfillment,omitempty"`
	Buyer       *Buyer       `json:"buyer,omitempty"`
	Order       *OrderRef    `json:"order,omitempty"`
	Discounts   *Discounts   `json:"discounts,omitempty"`
}

type Link struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type LineItem struct {
	ID       string  `json:"id"`
	Item     Item    `json:"item"`
	Quantity int     `json:"quantity"`
	Totals   []Total `json:"totals"`
}

type Item struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Price    int    `json:"price,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type Total struct {
	Type        string `json:"type"`
	DisplayText string `json:"display_text,omitempty"`
	Amount      int    `json:"amount"`
}

type OrderRef struct {
	ID           string `json:"id"`
	PermalinkURL string `json:"permalink_url"`
}

// Fulfillment is the hierarchical fulfillment model.
type Fulfillment struct {
	Methods []FulfillmentMethod `json:"methods"`
}

type FulfillmentMethod struct {
	ID                    string                   `json:"id"`
	Type                  string                   `json:"type"`
	LineItemIDs           []string                 `json:"line_item_ids"`
	Destinations          []FulfillmentDestination `json:"destinations,omitempty"`
	SelectedDestinationID string                   `json:"selected_destination_id,omitempty"`
	Groups                []FulfillmentGroup       `json:"groups,omitempty"`
}

type FulfillmentGroup struct {
	ID               string              `json:"id"`
	LineItemIDs      []string            `json:"line_item_ids"`
	Options          []FulfillmentOption `json:"options,omitempty"`
	SelectedOptionID string              `json:"selected_option_id,omitempty"`
}

type FulfillmentOption struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Totals []Total `json:"totals"`
}

type FulfillmentDestination struct {
	ID              string `json:"id,omitempty"`
	FullName        string `json:"full_name,omitempty"`
	StreetAddress   string `json:"street_address,omitempty"`
	AddressLocality string `json:"address_locality,omitempty"`
	AddressRegion   string `json:"address_region,omitempty"`
	PostalCode      string `json:"postal_code,omitempty"`
	AddressCountry  string `json:"address_country,omitempty"`
}

// Payment contains payment instruments and handlers.
type Payment struct {
	SelectedInstrumentID string                   `json:"selected_instrument_id,omitempty"`
	Instruments          []map[string]interface{} `json:"instruments"`
	Handlers             []map[string]interface{} `json:"handlers"`
}

// Buyer has first/last name plus consent.
type Buyer struct {
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	FullName  string   `json:"fullName,omitempty"`
	Email     string   `json:"email,omitempty"`
	Consent   *Consent `json:"consent,omitempty"`
}

type Consent struct {
	Marketing  *bool `json:"marketing,omitempty"`
	Analytics  *bool `json:"analytics,omitempty"`
	SaleOfData *bool `json:"sale_of_data,omitempty"`
}

// Discounts tracks discount codes and applied discounts.
type Discounts struct {
	Codes   []string          `json:"codes,omitempty"`
	Applied []AppliedDiscount `json:"applied,omitempty"`
}

type AppliedDiscount struct {
	Code      string `json:"code,omitempty"`
	Title     string `json:"title"`
	Amount    int    `json:"amount"`
	Automatic bool   `json:"automatic,omitempty"`
}

// Order is returned by GET /orders/{id}.
type Order struct {
	ID           string           `json:"id"`
	UCP          UCPEnvelope      `json:"ucp"`
	CheckoutID   string           `json:"checkout_id"`
	PermalinkURL string           `json:"permalink_url"`
	LineItems    []OrderLineItem  `json:"line_items"`
	Fulfillment  OrderFulfillment `json:"fulfillment"`
	Adjustments  []Adjustment     `json:"adjustments,omitempty"`
	Currency     string           `json:"currency"`
	Totals       []Total          `json:"totals"`
}

type OrderLineItem struct {
	ID       string        `json:"id"`
	Item     Item          `json:"item"`
	Quantity OrderQuantity `json:"quantity"`
	Totals   []Total       `json:"totals"`
	Status   string        `json:"status"`
	ParentID *string       `json:"parent_id,omitempty"`
}

type OrderQuantity struct {
	Total     int `json:"total"`
	Fulfilled int `json:"fulfilled"`
}

type OrderFulfillment struct {
	Expectations []Expectation      `json:"expectations,omitempty"`
	Events       []FulfillmentEvent `json:"events,omitempty"`
}

type Expectation struct {
	ID          string                 `json:"id"`
	LineItems   []EventLineItem        `json:"line_items"`
	MethodType  string                 `json:"method_type"`
	Destination FulfillmentDestination `json:"destination"`
	Description string                 `json:"description,omitempty"`
}

type FulfillmentEvent struct {
	ID             string          `json:"id"`
	OccurredAt     string          `json:"occurred_at"`
	Type           string          `json:"type"`
	LineItems      []EventLineItem `json:"line_items,omitempty"`
	TrackingNumber string          `json:"tracking_number,omitempty"`
	TrackingURL    string          `json:"tracking_url,omitempty"`
	Description    string          `json:"description,omitempty"`
}

type EventLineItem struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

type Adjustment struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	OccurredAt  string `json:"occurred_at"`
	Status      string `json:"status"`
	Amount      int    `json:"amount"`
	Description string `json:"description,omitempty"`
}

// Cart is used by MCP cart operations.
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
