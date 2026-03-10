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
//   - Filter: searches products by category, brand, title substring,
//     country, currency, and language (maps to dev.ucp.shopping.catalog.search).
//     Supports the UCP search contract where an empty result set is not an error.
//
//   - CategoryCount: returns a list of CategoryStat with category names and product counts.
//
// Concrete implementations of the Catalog interface live in merchant packages
// (e.g., merchant/simple_merchant).
//
// # Query Semantics
//
// UCP does not mandate a structured query language (no boolean operators,
// field selectors, or wildcards). This implementation uses plain
// case-insensitive substring matching:
//
//   - Filter.query matches against the product Title only.
//   - Search.query matches against Title, Description, or Category (OR logic:
//     a product is returned if any of those fields contains the query substring).
//   - An empty query means "no filter" — all products pass.
//
// For example, the query "rose" matches a product titled "Bouquet of Red Roses"
// because strings.Contains(strings.ToLower("Bouquet of Red Roses"), "rose")
// is true.
//
// # Context and Localization
//
// UCP catalog operations accept optional Context signals as typed values from
// the ucp package (ucp.Country, ucp.Currency, ucp.Language) for relevance and
// localization. The ucp.ContainsCountry helper supports country code matching
// for filtering products by shipping eligibility; values must be pre-normalized
// via ucp.NewCountry.
package catalog
