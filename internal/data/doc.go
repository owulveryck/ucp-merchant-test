// Package data loads and manages the test dataset for the UCP conformance test
// suite.
//
// The Universal Commerce Protocol defines a conformance test suite that validates
// merchant implementations against a standardized dataset. This package loads
// CSV and JSON files from the test data directory (e.g., flower_shop) into an
// in-memory DataSource that the merchant server queries at runtime.
//
// # Test Data Structure
//
// The conformance test data consists of:
//
//   - Products (products.csv): catalog items with ID, title, price, and image URL.
//     Prices are in minor currency units (cents).
//
//   - Inventory (inventory.csv): stock quantities per product ID. Products with
//     quantity 0 are out of stock and must be rejected by checkout. The UCP spec
//     defines the standard error code "out_of_stock" for this case.
//
//   - Customers (customers.csv): test buyer identities with ID, name, and email.
//     Used by the Identity Linking capability (dev.ucp.common.identity_linking)
//     to look up authenticated users.
//
//   - Addresses (addresses.csv): shipping destinations linked to customer IDs.
//     Used by the Fulfillment extension (dev.ucp.shopping.fulfillment) to populate
//     fulfillment method destinations when a buyer's email is known.
//
//   - Payment Instruments (payment_instruments.csv): test payment methods with
//     token-based processing. Tokens like "success_token" and "fail_token"
//     simulate successful and failed payment processing as defined by the
//     payment handler specification.
//
//   - Discounts (discounts.csv): discount codes with type (percentage or
//     fixed_amount) and value. Implements the UCP Discount Extension
//     (dev.ucp.shopping.discount) which supports code-based discounts submitted
//     via checkout create/update operations.
//
//   - Shipping Rates (shipping_rates.csv): fulfillment costs per country and
//     service level. Country-specific rates override default rates for the same
//     service level.
//
//   - Promotions (promotions.csv): automatic discount rules such as free shipping
//     for orders above a minimum subtotal or containing specific eligible items.
//     These are surfaced as automatic discounts per the UCP discount specification.
//
//   - Conformance Input (conformance_input.json): test expectations defining
//     currency, available items with expected prices, out-of-stock items, and
//     non-existent item IDs for negative testing.
//
// # Dynamic Addresses
//
// When a buyer submits a new shipping address during checkout that does not match
// any existing address, the DataSource stores it as a dynamic address. This
// supports the UCP fulfillment flow where destinations can be created on the fly
// and must be retrievable in subsequent checkout updates.
//
// # Lookup Functions
//
// The DataSource provides lookup functions that map to UCP operations:
// FindCustomerByEmail (identity linking), FindAddressesForEmail (fulfillment
// destination population), FindDiscountByCode (discount validation),
// FindPaymentInstrumentByID/ByToken (payment processing), and
// GetShippingRatesForCountry (fulfillment option generation).
package data
