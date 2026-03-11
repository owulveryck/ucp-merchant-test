package model

import "github.com/owulveryck/ucp-merchant-test/pkg/ucp"

// CheckoutRequest represents the incoming JSON body for UCP checkout session
// create (POST) and update (PUT) operations on the Shopping Service REST API.
//
// The structure mirrors the UCP Checkout Capability (dev.ucp.shopping.checkout)
// request schema. UCP update operations use full replacement semantics: every
// field present in the request replaces the corresponding field on the checkout
// session. Fields that are nil/absent are left unchanged on update.
//
// On create, LineItems is required. On update, any combination of fields may
// be provided. PaymentData is only used during the checkout completion call
// (POST .../complete) to submit the buyer's payment credential.
type CheckoutRequest struct {
	Currency    ucp.Currency        `json:"currency,omitempty"`
	LineItems   []LineItemRequest   `json:"line_items,omitempty"`
	Payment     *PaymentRequest     `json:"payment,omitempty"`
	Buyer       *BuyerRequest       `json:"buyer,omitempty"`
	Fulfillment *FulfillmentRequest `json:"fulfillment,omitempty"`
	Discounts   *DiscountsRequest   `json:"discounts,omitempty"`
	PaymentData *PaymentDataRequest `json:"payment_data,omitempty"`
}

// LineItemRequest represents a line item submitted by the platform in a checkout
// create or update request. Each line item identifies a product from the merchant's
// catalog and specifies a purchase quantity.
//
// The product can be referenced in two ways, tried in order:
//   - Item.ID: the canonical UCP approach, where Item is an object with an "id"
//     field matching a variant ID from the catalog (dev.ucp.shopping.catalog.lookup)
//   - ProductID: a shorthand alternative using "product_id" at the top level
//
// ID is an optional caller-assigned identifier for the line item. When absent,
// the server generates sequential IDs ("li_001", "li_002", etc.). Quantity
// defaults to 1 when zero or unset.
type LineItemRequest struct {
	ID        string   `json:"id,omitempty"`
	Item      *ItemRef `json:"item,omitempty"`
	ProductID string   `json:"product_id,omitempty"`
	Quantity  int      `json:"quantity,omitempty"`
}

// ItemRef identifies a catalog product by its variant ID within a line item
// request. This follows the UCP convention where line_items[].item.id references
// the product's catalog identifier, enabling the business to look up pricing,
// availability, and metadata from its product catalog.
type ItemRef struct {
	ID string `json:"id"`
}

// PaymentRequest carries payment configuration submitted by the platform during
// checkout create or update operations.
//
// In the UCP payment model, payment handlers are discovered from the business's
// UCP profile (/.well-known/ucp) and define processing specifications for
// collecting payment instruments. The platform may override the default handlers
// or instruments by providing them in this request.
//
// Instruments and Handlers use []map[string]interface{} because their schemas
// are handler-specific and intentionally opaque per the UCP specification — each
// payment handler defines its own instrument and configuration format.
type PaymentRequest struct {
	SelectedInstrumentID string                   `json:"selected_instrument_id,omitempty"`
	Instruments          []map[string]interface{} `json:"instruments,omitempty"`
	Handlers             []map[string]interface{} `json:"handlers,omitempty"`
}

// BuyerRequest carries buyer identity information submitted by the platform
// during checkout create or update operations.
//
// Buyer identity is foundational in UCP: it enables the Identity Linking
// capability for address lookup (populating fulfillment destinations from known
// addresses), payment instrument retrieval, and personalized pricing. The email
// field is particularly important as it serves as the primary key for address
// lookup in the fulfillment flow.
//
// Name fields support multiple conventions: FirstName/LastName for structured
// names, FullName for the "fullName" JSON field, and Name as a fallback that
// populates FullName when neither FullName nor FirstName is provided.
type BuyerRequest struct {
	FirstName string          `json:"first_name,omitempty"`
	LastName  string          `json:"last_name,omitempty"`
	FullName  string          `json:"fullName,omitempty"`
	Name      string          `json:"name,omitempty"`
	Email     string          `json:"email,omitempty"`
	Consent   *ConsentRequest `json:"consent,omitempty"`
}

// ConsentRequest carries buyer consent preferences as defined by the UCP
// privacy extensions. Each field uses a *bool to distinguish between "not
// provided" (nil) and an explicit true/false preference.
//
// Consent preferences are echoed back in the checkout response and forwarded
// to the order, allowing the business to honor the buyer's privacy choices
// throughout the transaction lifecycle.
type ConsentRequest struct {
	Marketing  *bool `json:"marketing,omitempty"`
	Analytics  *bool `json:"analytics,omitempty"`
	SaleOfData *bool `json:"sale_of_data,omitempty"`
}

// FulfillmentRequest carries fulfillment configuration submitted by the platform
// as part of the UCP Fulfillment Extension (dev.ucp.shopping.fulfillment).
//
// Fulfillment is optional in UCP checkout sessions to support digital goods that
// need no physical delivery. When present, it contains one or more methods that
// describe how items should be fulfilled.
//
// The fulfillment flow is progressive: the platform first submits a method type,
// then selects a destination, then selects a shipping option. Each step triggers
// the business to populate the next level of the fulfillment hierarchy.
type FulfillmentRequest struct {
	Methods []FulfillmentMethodRequest `json:"methods,omitempty"`
}

// FulfillmentMethodRequest represents a fulfillment method submitted by the
// platform within the fulfillment hierarchy. UCP defines several method types
// including "shipping" (physical delivery to an address) and "pickup" (buyer
// collects from a location).
//
// The progressive fulfillment flow uses this type across multiple checkout
// updates:
//  1. Initial: Type is set (e.g., "shipping"), Destinations may be submitted
//  2. Destination selection: SelectedDestinationID is set, triggering the
//     business to generate shipping options (Groups with Options)
//  3. Option selection: Groups[].SelectedOptionID is set, completing the
//     fulfillment configuration
type FulfillmentMethodRequest struct {
	ID                    string                          `json:"id,omitempty"`
	Type                  string                          `json:"type,omitempty"`
	Destinations          []FulfillmentDestinationRequest `json:"destinations,omitempty"`
	SelectedDestinationID string                          `json:"selected_destination_id,omitempty"`
	Groups                []FulfillmentGroupRequest       `json:"groups,omitempty"`
}

// FulfillmentDestinationRequest represents a shipping address submitted by the
// platform as part of the fulfillment flow. Field names follow the Schema.org
// PostalAddress convention used throughout UCP (street_address, address_locality,
// address_region, postal_code, address_country).
//
// When ID is empty, the server attempts to match the address against known
// addresses for the buyer's email. If no match is found, a dynamic address ID
// is generated and the address is saved for future lookups. When ID is provided,
// it references a previously known address (e.g., from the buyer's address book).
type FulfillmentDestinationRequest struct {
	ID              string      `json:"id,omitempty"`
	FullName        string      `json:"full_name,omitempty"`
	StreetAddress   string      `json:"street_address,omitempty"`
	AddressLocality string      `json:"address_locality,omitempty"`
	AddressRegion   string      `json:"address_region,omitempty"`
	PostalCode      string      `json:"postal_code,omitempty"`
	AddressCountry  ucp.Country `json:"address_country,omitempty"`
}

// FulfillmentGroupRequest represents the platform's selection of a shipping
// option within a fulfillment group. Groups are business-generated packages
// that group line items shipping together; each group offers selectable options
// with different speeds and costs.
//
// The platform sets SelectedOptionID to indicate which shipping option the
// buyer has chosen (e.g., "standard_shipping", "express_shipping"). This
// selection determines the fulfillment cost included in the checkout totals.
type FulfillmentGroupRequest struct {
	SelectedOptionID string `json:"selected_option_id,omitempty"`
}

// DiscountsRequest carries discount codes submitted by the platform as part of
// the UCP Discount Extension (dev.ucp.shopping.discount).
//
// Per UCP, discount code operations use replacement semantics: submitting Codes
// replaces any previously submitted codes on the checkout session. Sending an
// empty slice clears all codes. The business validates each code and returns
// the result in the checkout response's discounts.applied array, with rejected
// codes communicated via the messages array.
type DiscountsRequest struct {
	Codes []string `json:"codes,omitempty"`
}

// PaymentDataRequest carries the payment credential submitted by the platform
// when completing a checkout session (POST .../complete). This is the final
// step where the buyer's payment instrument is processed.
//
// HandlerID identifies which payment handler should process the credential
// (e.g., "google_pay", "mock_payment_handler"). The Credential contains the
// handler-specific payment token obtained by the platform from the payment
// provider during instrument collection.
type PaymentDataRequest struct {
	HandlerID  string             `json:"handler_id,omitempty"`
	Credential *PaymentCredential `json:"credential,omitempty"`
}

// PaymentCredential contains the payment token obtained by the platform from
// the payment provider. The token format is handler-specific and opaque to
// the business — it is forwarded to the payment processor for authorization.
//
// In this test implementation, the token "success_token" results in a
// successful payment (HTTP 200), while "fail_token" triggers a payment
// failure (HTTP 402 Payment Required).
type PaymentCredential struct {
	Token string `json:"token,omitempty"`
}

// OrderUpdateRequest represents the incoming JSON body for order update
// (PUT /orders/{id}) operations. It replaces the untyped map[string]interface{}
// previously used for parsing.
type OrderUpdateRequest struct {
	Fulfillment *OrderFulfillmentUpdate `json:"fulfillment,omitempty"`
	Adjustments []Adjustment            `json:"adjustments,omitempty"`
}

// OrderFulfillmentUpdate carries fulfillment updates for an order.
type OrderFulfillmentUpdate struct {
	Events       []FulfillmentEvent `json:"events,omitempty"`
	Expectations []Expectation      `json:"expectations,omitempty"`
}
