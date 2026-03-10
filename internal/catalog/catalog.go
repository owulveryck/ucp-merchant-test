package catalog

import (
	"strings"
)

// Product represents an item in the merchant's catalog.
type Product struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Category           string   `json:"category"`
	Brand              string   `json:"brand"`
	Price              int      `json:"price"`
	Quantity           int      `json:"quantity"`
	Rank               int      `json:"rank"`
	ImageURL           string   `json:"image_url,omitempty"`
	Description        string   `json:"description,omitempty"`
	UsageType          string   `json:"usage_type,omitempty"`
	AvailableCountries []string `json:"available_countries,omitempty"`
}

// CategoryStat holds a category name and its product count.
type CategoryStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// SearchParams holds parameters for a catalog search (per Shopify Agent Catalog spec).
type SearchParams struct {
	Query            string
	Limit            int
	MinPrice         int // cents, 0 = no min
	MaxPrice         int // cents, 0 = no max
	AvailableForSale bool
	ShipsTo          string // country code
}

// SearchResult wraps a product with computed availability metadata.
type SearchResult struct {
	Product Product `json:"product"`
	InStock bool    `json:"in_stock"`
}

// Catalog is the read-only interface for catalog operations.
type Catalog interface {
	// Find finds a product by its ID. Returns nil if not found.
	Find(id string) *Product
	// Filter returns products matching the given criteria. Empty string
	// parameters are ignored. All comparisons are case-insensitive.
	Filter(category, brand, query, usageType, country, currency, language string) []Product
	// CategoryCount returns the number of products per category, preserving
	// insertion order.
	CategoryCount() []CategoryStat
	// Lookup finds a product by ID, additionally filtering by shipping
	// destination country. Returns nil if the product doesn't exist or is not
	// available in the given country. Unlike Find, this enforces geographic
	// availability.
	Lookup(id string, shipsTo string) *Product
	// Search searches the catalog by keyword (matching title, description,
	// category) with optional price range, stock, and country filters. Returns
	// up to params.Limit results (default 10, max 300).
	Search(params SearchParams) []SearchResult
}

// ContainsCountry checks if a country code is in the list (case-insensitive).
func ContainsCountry(countries []string, country string) bool {
	for _, c := range countries {
		if strings.EqualFold(c, country) {
			return true
		}
	}
	return false
}
