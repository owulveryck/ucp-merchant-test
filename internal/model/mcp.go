package model

// MCPCheckoutState holds server-side MCP checkout session state.
// It wraps the canonical Checkout with MCP-specific fields: owner identity
// and checkout hash for change detection.
type MCPCheckoutState struct {
	Checkout     *Checkout `json:"-"`
	OwnerID      string    `json:"-"`
	CheckoutHash string    `json:"-"`
}
