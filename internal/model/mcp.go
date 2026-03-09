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

// MCPToolResult is the result payload for a tools/call response.
type MCPToolResult struct {
	Content []MCPContentBlock `json:"content"`
	IsError bool              `json:"isError,omitempty"`
}

// MCPContentBlock is a single content block in an MCP tool result.
type MCPContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// MCPInitializeResult is the result payload for the initialize response.
type MCPInitializeResult struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    MCPCapabilities `json:"capabilities"`
	ServerInfo      MCPServerInfo   `json:"serverInfo"`
}

// MCPCapabilities describes MCP server capabilities.
type MCPCapabilities struct {
	Tools map[string]any `json:"tools"`
}

// MCPServerInfo identifies the MCP server.
type MCPServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MCPToolsListResult is the result payload for the tools/list response.
type MCPToolsListResult struct {
	Tools []ToolDef `json:"tools"`
}
