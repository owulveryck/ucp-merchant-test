// Package pricing implements line item construction and total calculation for the
// UCP Shopping Service.
//
// In the Universal Commerce Protocol, checkout and cart responses include line_items
// and totals arrays that reflect the current pricing state. This package provides
// the shared pricing primitives used by both the Checkout Capability
// (dev.ucp.shopping.checkout) and the Cart Capability (dev.ucp.shopping.cart).
//
// # Line Items
//
// BuildLineItems constructs the line_items array from the platform's request. Each
// line item references a product by item.id, which must match a variant ID from
// the catalog (dev.ucp.shopping.catalog.lookup). The function validates product
// existence and stock availability, returning the UCP standard error "out_of_stock"
// for unavailable items.
//
// Each line item receives a computed subtotal (price × quantity) and a total. Line
// item IDs are auto-generated as "li_001", "li_002", etc. Per UCP, line item
// totals include at minimum "subtotal" and "total" entries.
//
// # Totals
//
// CalculateTotals computes the order-level totals array from line items, shipping
// cost, and optional discounts. UCP defines the following standardized total types:
//
//   - subtotal: sum of all line item subtotals (before discounts and shipping)
//   - items_discount: sum of discounts allocated to individual line items
//   - discount: order-level discounts (shipping discounts, flat order discounts)
//   - fulfillment: shipping/delivery cost from the selected fulfillment option
//   - tax: tax amount (not implemented in this test server)
//   - fee: additional fees (not implemented in this test server)
//   - total: final amount = subtotal - discount + fulfillment
//
// All amounts are in minor currency units (e.g., cents). Discount amounts are
// always positive integers; platforms display them as subtractive (e.g., "-$5.00").
// The total type "shipping" must never be used — UCP requires "fulfillment" instead.
package pricing
