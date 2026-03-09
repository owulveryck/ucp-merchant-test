package main

// UCP REST API models matching the conformance test expectations.

// RestUCP is the "ucp" envelope in checkout responses (maps to ResponseCheckout).
type RestUCP struct {
	Version      string           `json:"version"`
	Capabilities []RestCapability `json:"capabilities"`
}

type RestCapability struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// RestCheckout is the checkout model returned by REST endpoints.
type RestCheckout struct {
	ID          string           `json:"id"`
	UCP         RestUCP          `json:"ucp"`
	Status      string           `json:"status"`
	Currency    string           `json:"currency"`
	LineItems   []RestLineItem   `json:"line_items"`
	Totals      []Total          `json:"totals"`
	Links       []RestLink       `json:"links"`
	Payment     RestPayment      `json:"payment"`
	Fulfillment *RestFulfillment `json:"fulfillment,omitempty"`
	Buyer       *RestBuyer       `json:"buyer,omitempty"`
	Order       *RestOrderRef    `json:"order,omitempty"`
	Discounts   *RestDiscounts   `json:"discounts,omitempty"`
}

type RestLink struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type RestLineItem struct {
	ID       string   `json:"id"`
	Item     RestItem `json:"item"`
	Quantity int      `json:"quantity"`
	Totals   []Total  `json:"totals"`
}

type RestItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Price    int    `json:"price,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type RestOrderRef struct {
	ID           string `json:"id"`
	PermalinkURL string `json:"permalink_url"`
}

// RestFulfillment is the hierarchical fulfillment model.
type RestFulfillment struct {
	Methods []RestFulfillmentMethod `json:"methods"`
}

type RestFulfillmentMethod struct {
	ID                    string                       `json:"id"`
	Type                  string                       `json:"type"`
	LineItemIDs           []string                     `json:"line_item_ids"`
	Destinations          []RestFulfillmentDestination `json:"destinations,omitempty"`
	SelectedDestinationID string                       `json:"selected_destination_id,omitempty"`
	Groups                []RestFulfillmentGroup       `json:"groups,omitempty"`
}

type RestFulfillmentGroup struct {
	ID               string                  `json:"id"`
	LineItemIDs      []string                `json:"line_item_ids"`
	Options          []RestFulfillmentOption `json:"options,omitempty"`
	SelectedOptionID string                  `json:"selected_option_id,omitempty"`
}

type RestFulfillmentOption struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Totals []Total `json:"totals"`
}

type RestFulfillmentDestination struct {
	ID              string `json:"id,omitempty"`
	FullName        string `json:"full_name,omitempty"`
	StreetAddress   string `json:"street_address,omitempty"`
	AddressLocality string `json:"address_locality,omitempty"`
	AddressRegion   string `json:"address_region,omitempty"`
	PostalCode      string `json:"postal_code,omitempty"`
	AddressCountry  string `json:"address_country,omitempty"`
}

// RestPayment contains payment instruments and handlers.
type RestPayment struct {
	SelectedInstrumentID string                   `json:"selected_instrument_id,omitempty"`
	Instruments          []map[string]interface{} `json:"instruments"`
	Handlers             []map[string]interface{} `json:"handlers"`
}

// RestBuyer has first/last name plus consent.
type RestBuyer struct {
	FirstName string       `json:"first_name,omitempty"`
	LastName  string       `json:"last_name,omitempty"`
	FullName  string       `json:"fullName,omitempty"`
	Email     string       `json:"email,omitempty"`
	Consent   *RestConsent `json:"consent,omitempty"`
}

type RestConsent struct {
	Marketing  *bool `json:"marketing,omitempty"`
	Analytics  *bool `json:"analytics,omitempty"`
	SaleOfData *bool `json:"sale_of_data,omitempty"`
}

// RestDiscounts tracks discount codes and applied discounts.
type RestDiscounts struct {
	Codes   []string              `json:"codes,omitempty"`
	Applied []RestAppliedDiscount `json:"applied,omitempty"`
}

type RestAppliedDiscount struct {
	Code      string `json:"code,omitempty"`
	Title     string `json:"title"`
	Amount    int    `json:"amount"`
	Automatic bool   `json:"automatic,omitempty"`
}

// RestOrder is returned by GET /orders/{id}.
type RestOrder struct {
	ID           string               `json:"id"`
	UCP          RestUCP              `json:"ucp"`
	CheckoutID   string               `json:"checkout_id"`
	PermalinkURL string               `json:"permalink_url"`
	LineItems    []RestOrderLineItem  `json:"line_items"`
	Fulfillment  RestOrderFulfillment `json:"fulfillment"`
	Adjustments  []RestAdjustment     `json:"adjustments,omitempty"`
	Currency     string               `json:"currency"`
	Totals       []Total              `json:"totals"`
}

type RestOrderLineItem struct {
	ID       string            `json:"id"`
	Item     RestItem          `json:"item"`
	Quantity RestOrderQuantity `json:"quantity"`
	Totals   []Total           `json:"totals"`
	Status   string            `json:"status"`
	ParentID *string           `json:"parent_id,omitempty"`
}

type RestOrderQuantity struct {
	Total     int `json:"total"`
	Fulfilled int `json:"fulfilled"`
}

type RestOrderFulfillment struct {
	Expectations []RestExpectation      `json:"expectations,omitempty"`
	Events       []RestFulfillmentEvent `json:"events,omitempty"`
}

type RestExpectation struct {
	ID          string                     `json:"id"`
	LineItems   []RestEventLineItem        `json:"line_items"`
	MethodType  string                     `json:"method_type"`
	Destination RestFulfillmentDestination `json:"destination"`
	Description string                     `json:"description,omitempty"`
}

type RestFulfillmentEvent struct {
	ID             string              `json:"id"`
	OccurredAt     string              `json:"occurred_at"`
	Type           string              `json:"type"`
	LineItems      []RestEventLineItem `json:"line_items,omitempty"`
	TrackingNumber string              `json:"tracking_number,omitempty"`
	TrackingURL    string              `json:"tracking_url,omitempty"`
	Description    string              `json:"description,omitempty"`
}

type RestEventLineItem struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

type RestAdjustment struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	OccurredAt  string `json:"occurred_at"`
	Status      string `json:"status"`
	Amount      int    `json:"amount"`
	Description string `json:"description,omitempty"`
}
