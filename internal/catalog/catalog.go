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

// Catalog is the read-only interface for catalog operations.
type Catalog interface {
	Find(id string) *Product
	Filter(category, brand, query, usageType, country, currency, language string) []Product
	CategoryCount() []CategoryStat
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
