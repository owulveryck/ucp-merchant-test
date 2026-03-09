// Package discount implements the UCP Discount Extension
// (dev.ucp.shopping.discount) for the Shopping Service.
//
// The Discount Extension allows businesses to indicate that they support discount
// codes on checkout sessions, and specifies how discount codes are shared between
// the platform and the business. It extends the Checkout Capability
// (dev.ucp.shopping.checkout) by adding a discounts object to the checkout.
//
// # Key Features
//
//   - Submit one or more discount codes via checkout create/update operations
//   - Receive applied discounts with human-readable titles and amounts
//   - Rejected codes communicated via the messages array with detailed error codes
//   - Automatic discounts surfaced alongside code-based discounts
//
// # Discount Types
//
// UCP supports two allocation methods for discounts:
//
//   - percentage: applied as a percentage of the subtotal (e.g., "10OFF" → 10% off)
//   - fixed_amount: applied as a fixed value in minor currency units (e.g., "FIXED500" → $5 off)
//
// # Stacking and Priority
//
// When multiple discounts are applied, the order matters because percentage
// discounts compound differently depending on when they're applied. Lower priority
// numbers are applied first. Each subsequent discount operates on the remaining
// subtotal after prior discounts.
//
// # Request/Response Behavior
//
// Per UCP, discount code operations use replacement semantics: submitting
// discounts.codes replaces any previously submitted codes. Sending an empty
// array clears all codes. Codes are matched case-insensitively.
//
// The response includes discounts.applied with all active discounts (code-based
// and automatic). Discount amounts are reflected in the totals array using
// "items_discount" (for line-item-allocated discounts) and "discount" (for
// order-level discounts). All discount amounts are positive integers in minor
// currency units.
//
// # Rejected Codes
//
// When a submitted code cannot be applied, it still appears in discounts.codes
// (echoed back) but not in discounts.applied. The rejection is communicated via
// the messages array with standard error codes such as discount_code_expired,
// discount_code_invalid, discount_code_already_applied, and others defined by
// the UCP specification.
package discount
