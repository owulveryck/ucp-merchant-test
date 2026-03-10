// Package ucp provides typed domain values for Universal Commerce Protocol
// context signals: country, currency, language, and category.
//
// # Why Named Types
//
// UCP operations accept multiple string-like parameters (country code, currency
// code, language tag, category name). Using distinct named types instead of raw
// strings gives compile-time safety against transposed arguments and makes
// function signatures self-documenting:
//
//	func Filter(category ucp.Category, country ucp.Country, currency ucp.Currency) // clear
//	func Filter(category, country, currency string)                                 // error-prone
//
// # Normalization Rules
//
//   - Country and Currency are uppercased on construction via NewCountry /
//     NewCurrency, so comparisons with == are safe after creation.
//   - Category preserves original casing for display purposes. Use the Matches
//     method for case-insensitive comparison.
//   - Language is stored as-is (BCP 47 tags are case-insensitive by spec but
//     conventionally mixed-case, e.g. "fr-CA").
//
// # Relationship to JSON Wire Format
//
// UCP model structs (model.Checkout, model.Order, etc.) use plain strings for
// JSON serialization. Conversion to/from typed values happens at handler and
// data-loading boundaries, keeping serialization simple while enforcing type
// safety in business logic.
//
// # Usage
//
//	country := ucp.NewCountry("us")  // Country("US")
//	currency := ucp.NewCurrency("eur") // Currency("EUR")
//	cat := ucp.Category("Flowers")
//	cat.Matches(ucp.Category("flowers")) // true
//	ucp.ContainsCountry([]ucp.Country{"US", "GB"}, country) // true
package ucp
