package catalog

import (
	"github.com/owulveryck/ucp-merchant-test/pkg/ucp"
)

// Product represents an item in the merchant's catalog.
type Product struct {
	ID       string       `json:"id"`
	Title    string       `json:"title"`
	Category ucp.Category `json:"category"`
	// Brand is the free-text manufacturer or brand name (e.g. "Rose Garden Co.").
	Brand    string `json:"brand"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
	ImageURL string `json:"image_url,omitempty"`
	// Description is an optional long-form product description used for
	// keyword search matching.
	Description        string        `json:"description,omitempty"`
	AvailableCountries []ucp.Country `json:"available_countries,omitempty"`
}

// CategoryStat holds a category name and its product count.
type CategoryStat struct {
	Name  ucp.Category `json:"name"`
	Count int          `json:"count"`
}

// SearchParams holds parameters for a catalog search (per Shopify Agent Catalog spec).
type SearchParams struct {
	// Query is a free-text search string matched case-insensitively against
	// product title, description, and category. Matching is plain substring
	// comparison — no boolean operators, field syntax, or wildcards are
	// supported.
	Query string
	// Limit is the maximum number of results to return. Defaults to 10 if
	// zero; capped at 300.
	Limit int
	// ShipsTo filters products by shipping destination country code.
	ShipsTo ucp.Country
}

// SearchResult wraps a product with computed availability metadata.
type SearchResult struct {
	Product Product `json:"product"`
}

// Catalog is the read-only interface for catalog operations.
type Catalog interface {
	// Find finds a product by its ID. Returns nil if not found.
	Find(id string) *Product
	// Filter returns products matching the given criteria. Empty/zero-value
	// parameters are ignored. All comparisons are case-insensitive.
	//
	// Parameters:
	//   - category: matches Product.Category via case-insensitive comparison.
	//   - brand: case-insensitive exact match against Product.Brand.
	//   - query: plain case-insensitive substring match against Product.Title
	//     (no boolean operators, field syntax, or wildcards).
	//   - country: filters to products whose AvailableCountries includes this code.
	//   - currency: reserved for future price-currency filtering (currently unused).
	//   - language: reserved for future localization (currently unused).
	Filter(category ucp.Category, brand, query string, country ucp.Country, currency ucp.Currency, language ucp.Language) []Product
	// CategoryCount returns the number of products per category, preserving
	// insertion order.
	CategoryCount() []CategoryStat
	// Lookup finds a product by ID, additionally filtering by shipping
	// destination country. Returns nil if the product doesn't exist or is not
	// available in the given country. Unlike Find, this enforces geographic
	// availability.
	Lookup(id string, shipsTo ucp.Country) *Product
	// Search searches the catalog by keyword (matching title, description,
	// category) with optional price range, stock, and country filters. Returns
	// up to params.Limit results (default 10, max 300).
	Search(params SearchParams) []SearchResult
}
