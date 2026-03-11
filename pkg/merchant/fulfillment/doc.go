// Package fulfillment implements the UCP Fulfillment Extension
// (dev.ucp.shopping.fulfillment) for the Shopping Service.
//
// The Fulfillment Extension enables businesses to advertise support for physical
// goods fulfillment (shipping, pickup, etc.) and structures the delivery
// configuration within checkout sessions.
//
// # Fulfillment Structure
//
// Fulfillment adds a fulfillment field to the Checkout containing:
//
//   - Methods: fulfillment methods applicable to cart items (shipping, pickup, etc.).
//     Each method includes:
//   - LineItemIDs: which items this method fulfills
//   - Destinations: where to fulfill (shipping address, store location)
//   - Groups: business-generated packages, each with selectable Options
//
// # Fulfillment Flow
//
// The progressive fulfillment flow in a checkout session follows these steps:
//
//  1. Platform submits fulfillment.methods with type (e.g., "shipping")
//  2. Business populates destinations from known buyer addresses
//  3. Platform selects a destination via selected_destination_id
//  4. Business generates shipping options (groups with options) based on the
//     selected destination's country
//  5. Platform selects a shipping option via groups[].selected_option_id
//  6. Business computes fulfillment totals and includes them in the checkout totals
//
// # Destinations
//
// ParseDestination handles address parsing and ID assignment. When a destination
// has no ID, the function attempts to match it against existing addresses for the
// buyer's email. If no match is found, a dynamic address ID is generated and
// the address is saved for future lookups. This supports the UCP flow where
// buyers can submit new shipping addresses during checkout.
//
// # Shipping Options
//
// GenerateShippingOptions creates fulfillment options based on the destination
// country and applicable promotions. Country-specific shipping rates override
// default rates for the same service level. Free shipping promotions are applied
// when the order meets the minimum subtotal threshold or contains eligible items,
// as defined by the conformance test data.
//
// Each option includes a title (e.g., "Standard Shipping"), optional description
// (e.g., "Arrives Dec 12-15 via USPS"), and totals with "fulfillment" and "total"
// amounts. Per UCP rendering guidelines, title + description + total is sufficient
// to render any fulfillment option without understanding the specific method type.
//
// # Fulfillment Completeness
//
// IsFulfillmentComplete checks whether all required selections have been made:
// a destination must be selected (selected_destination_id) and at least one
// shipping option must be chosen (groups[].selected_option_id). The checkout
// can only transition to ready_for_complete when fulfillment is complete.
//
// # Shipping Cost
//
// GetCurrentShippingCost extracts the selected shipping option's total amount
// from the checkout's fulfillment structure. This cost is included in the
// order-level totals as the "fulfillment" total type (never "shipping", which
// is not a valid UCP total type).
package fulfillment
