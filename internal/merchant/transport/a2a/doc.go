// Package a2a implements the UCP Shopping Service A2A (Agent-to-Agent Protocol)
// transport binding.
//
// It uses the [a2a-go] library to provide a JSON-RPC 2.0 endpoint that
// exposes UCP shopping capabilities through the A2A protocol. All business
// logic is delegated to the [merchant.Merchant] interface.
//
// # Protocol Overview
//
// The A2A transport follows the UCP A2A Checkout Binding specification.
// Platforms interact with the merchant agent by sending A2A messages
// containing structured [a2a.DataPart] objects with an "action" field that
// identifies the UCP operation (e.g., "list_products", "create_checkout").
//
// Supported actions mirror the UCP Shopping Service capabilities:
//
//   - Catalog: list_products, get_product_details, search_catalog, lookup_product
//   - Cart: create_cart, get_cart, update_cart, cancel_cart
//   - Checkout: create_checkout, get_checkout, update_checkout, complete_checkout, cancel_checkout
//   - Orders: get_order, list_orders, cancel_order
//
// # Checkout Wrapping
//
// Per the UCP A2A spec, checkout objects are returned wrapped in a DataPart
// with key "a2a.ucp.checkout":
//
//	{"a2a.ucp.checkout": { ...checkoutObject }}
//
// Payment data for checkout completion is expected in a separate DataPart
// with key "a2a.ucp.checkout.payment".
//
// # Agent Card
//
// The server exposes an Agent Card at /.well-known/agent-card.json that
// advertises the UCP extension URI and shopping capabilities. The UCP
// discovery endpoint (/.well-known/ucp) references this agent card as
// the A2A transport endpoint.
//
// # References
//
//   - A2A Protocol: https://a2a-protocol.org/latest/
//   - UCP A2A Checkout Binding: docs/specification/checkout-a2a.md
//   - a2a-go library: https://github.com/a2aproject/a2a-go
//   - merchant.Merchant interface: internal/merchant/merchant.go
//
// [a2a-go]: https://github.com/a2aproject/a2a-go
// [a2a.DataPart]: https://pkg.go.dev/github.com/a2aproject/a2a-go/a2a#DataPart
// [merchant.Merchant]: https://pkg.go.dev/github.com/owulveryck/ucp-merchant-test/internal/merchant#Merchant
package a2a
