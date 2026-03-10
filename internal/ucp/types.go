package ucp

import "strings"

// Country represents an ISO 3166-1 alpha-2 country code, normalized to uppercase.
// The zero value represents no country filter.
type Country string

// Currency represents an ISO 4217 currency code, normalized to uppercase.
type Currency string

// Language represents a BCP 47 language tag.
type Language string

// Category represents a merchant-defined product category. Original casing is
// preserved for display; use Matches for case-insensitive comparison.
type Category string

// NewCountry creates a Country from a raw string, normalizing to uppercase.
func NewCountry(s string) Country {
	return Country(strings.ToUpper(s))
}

// NewCurrency creates a Currency from a raw string, normalizing to uppercase.
func NewCurrency(s string) Currency {
	return Currency(strings.ToUpper(s))
}

// Matches performs a case-insensitive comparison between two Category values.
func (c Category) Matches(other Category) bool {
	return strings.EqualFold(string(c), string(other))
}

// ContainsCountry checks if a country is in the list. Both the list entries and
// the search value must be pre-normalized via NewCountry, so simple == is used.
func ContainsCountry(countries []Country, c Country) bool {
	for _, cc := range countries {
		if cc == c {
			return true
		}
	}
	return false
}
