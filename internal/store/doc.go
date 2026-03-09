// Package store provides thread-safe, in-memory storage for UCP Shopping Service
// resources.
//
// The Universal Commerce Protocol defines several stateful resources that a
// business must manage throughout their lifecycle:
//
//   - Checkout Sessions (dev.ucp.shopping.checkout): created when a user expresses
//     purchase intent, progressing through the status lifecycle (incomplete →
//     ready_for_complete → completed). Each session is identified by a unique ID
//     and may carry associated metadata such as webhook URLs, shipping destinations,
//     and selected fulfillment option titles.
//
//   - Orders (dev.ucp.shopping.order): created upon successful checkout completion.
//     Orders are confirmed transactions with line items, fulfillment expectations,
//     and adjustments. The store tracks order shipments and owner identity for
//     access control.
//
//   - Carts (dev.ucp.shopping.cart): lightweight basket sessions for pre-purchase
//     exploration before checkout. Carts support incremental item management and
//     can be converted to checkout sessions via cart_id.
//
// The Store struct encapsulates all resource maps, mutex locks, and sequence
// counters. It replaces the global variables that would otherwise be scattered
// across the application. The Reset method clears all state, which is essential
// for the UCP conformance test suite's simulation reset endpoint.
//
// MCP-specific state (MCPCheckoutStates, MCPOrderShipments, MCPOrderOwners) tracks
// additional context needed by the Model Context Protocol transport binding, where
// tool-based interactions require stateful session tracking beyond what the REST
// transport needs.
package store
