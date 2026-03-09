// Package catalog defines the product catalog interface for the UCP Shopping Service.
//
// In the Universal Commerce Protocol, the Catalog capability
// (dev.ucp.shopping.catalog) allows platforms to search and browse business
// product catalogs. This enables product discovery before checkout, supporting
// use cases such as:
//
//   - Free-text product search (dev.ucp.shopping.catalog.search)
//   - Category and filter-based browsing
//   - Batch product/variant retrieval by identifier (dev.ucp.shopping.catalog.lookup)
//   - Price comparison across variants
//
// # Key UCP Catalog Concepts
//
// A Product is a catalog item with a title, description, media, and one or more
// purchasable variants. In this implementation, the Product struct represents a
// single-variant product where the variant ID equals the product ID.
//
// Product and variant IDs returned by catalog operations can be used directly in
// checkout line_items[].item.id. The variant ID from catalog retrieval matches
// the item ID expected by the Checkout Capability.
//
// # Catalog Interface
//
// The Catalog interface provides read-only operations mapping to UCP capabilities:
//
//   - Find: retrieves a product by ID (maps to dev.ucp.shopping.catalog.lookup).
//     Returns nil if the product does not exist.
//
//   - Filter: searches products by category, brand, title substring, usage type,
//     country, currency, and language (maps to dev.ucp.shopping.catalog.search).
//     Supports the UCP search contract where an empty result set is not an error.
//
//   - CategoryCount: returns a list of categories with their product counts.
//
// Concrete implementations of the Catalog interface live in merchant packages
// (e.g., merchant/simple_merchant).
//
// # Context and Localization
//
// UCP catalog operations accept optional Context signals (country, currency,
// language) as Filter parameters for relevance and localization. The
// ContainsCountry helper supports case-insensitive country code matching for
// filtering products by shipping eligibility.
package catalog
