package model

import "time"

// MCPCheckoutState holds server-side MCP checkout session state.
// It wraps the canonical Checkout with MCP-specific fields: owner identity,
// checkout hash for change detection, and selected shipping option.
type MCPCheckoutState struct {
	Checkout     *Checkout       `json:"-"`
	OwnerID      string          `json:"-"`
	CheckoutHash string          `json:"-"`
	Shipping     *ShippingOption `json:"-"`
}

// Shipment tracks order shipment progress for MCP order progression,
// including tracking number, carrier, and delivery timestamps.
type Shipment struct {
	TrackingNumber string    `json:"tracking_number"`
	Carrier        string    `json:"carrier"`
	EstimatedDate  string    `json:"estimated_delivery,omitempty"`
	ShippedAt      time.Time `json:"shipped_at,omitempty"`
	DeliveredAt    time.Time `json:"delivered_at,omitempty"`
}

// ShippingOption represents a selected shipping option during MCP checkout,
// capturing the method, carrier, estimated delivery days, and price.
type ShippingOption struct {
	ID            string `json:"id"`
	Method        string `json:"method"`
	Carrier       string `json:"carrier"`
	EstimatedDays int    `json:"estimated_days"`
	Price         int    `json:"price"`
	DisplayText   string `json:"display_text"`
}
