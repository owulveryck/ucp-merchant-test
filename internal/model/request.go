package model

// CheckoutRequest is the parsed body of a checkout create/update call.
type CheckoutRequest struct {
	Currency    string              `json:"currency,omitempty"`
	LineItems   []LineItemRequest   `json:"line_items,omitempty"`
	Payment     *PaymentRequest     `json:"payment,omitempty"`
	Buyer       *BuyerRequest       `json:"buyer,omitempty"`
	Fulfillment *FulfillmentRequest `json:"fulfillment,omitempty"`
	Discounts   *DiscountsRequest   `json:"discounts,omitempty"`
	PaymentData *PaymentDataRequest `json:"payment_data,omitempty"`
}

// LineItemRequest is a line item in an incoming request.
type LineItemRequest struct {
	ID        string   `json:"id,omitempty"`
	Item      *ItemRef `json:"item,omitempty"`
	ProductID string   `json:"product_id,omitempty"`
	Quantity  int      `json:"quantity,omitempty"`
}

// ItemRef identifies a product by ID in a request.
type ItemRef struct {
	ID string `json:"id"`
}

// PaymentRequest carries payment configuration from the caller.
type PaymentRequest struct {
	SelectedInstrumentID string                   `json:"selected_instrument_id,omitempty"`
	Instruments          []map[string]interface{} `json:"instruments,omitempty"`
	Handlers             []map[string]interface{} `json:"handlers,omitempty"`
}

// BuyerRequest carries buyer identity from the caller.
type BuyerRequest struct {
	FirstName string          `json:"first_name,omitempty"`
	LastName  string          `json:"last_name,omitempty"`
	FullName  string          `json:"fullName,omitempty"`
	Name      string          `json:"name,omitempty"`
	Email     string          `json:"email,omitempty"`
	Consent   *ConsentRequest `json:"consent,omitempty"`
}

// ConsentRequest carries buyer consent preferences.
type ConsentRequest struct {
	Marketing  *bool `json:"marketing,omitempty"`
	Analytics  *bool `json:"analytics,omitempty"`
	SaleOfData *bool `json:"sale_of_data,omitempty"`
}

// FulfillmentRequest carries fulfillment data from the caller.
type FulfillmentRequest struct {
	Methods []FulfillmentMethodRequest `json:"methods,omitempty"`
}

// FulfillmentMethodRequest is a fulfillment method in a request.
type FulfillmentMethodRequest struct {
	ID                    string                          `json:"id,omitempty"`
	Type                  string                          `json:"type,omitempty"`
	Destinations          []FulfillmentDestinationRequest `json:"destinations,omitempty"`
	SelectedDestinationID string                          `json:"selected_destination_id,omitempty"`
	Groups                []FulfillmentGroupRequest       `json:"groups,omitempty"`
}

// FulfillmentDestinationRequest is a destination address in a request.
type FulfillmentDestinationRequest struct {
	ID              string `json:"id,omitempty"`
	FullName        string `json:"full_name,omitempty"`
	StreetAddress   string `json:"street_address,omitempty"`
	AddressLocality string `json:"address_locality,omitempty"`
	AddressRegion   string `json:"address_region,omitempty"`
	PostalCode      string `json:"postal_code,omitempty"`
	AddressCountry  string `json:"address_country,omitempty"`
}

// FulfillmentGroupRequest is a fulfillment group selection in a request.
type FulfillmentGroupRequest struct {
	SelectedOptionID string `json:"selected_option_id,omitempty"`
}

// DiscountsRequest carries discount codes from the caller.
type DiscountsRequest struct {
	Codes []string `json:"codes,omitempty"`
}

// PaymentDataRequest carries payment data for checkout completion.
type PaymentDataRequest struct {
	HandlerID  string             `json:"handler_id,omitempty"`
	Credential *PaymentCredential `json:"credential,omitempty"`
}

// PaymentCredential carries payment credential data.
type PaymentCredential struct {
	Token string `json:"token,omitempty"`
}
