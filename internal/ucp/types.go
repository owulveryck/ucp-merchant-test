package ucp

import "strings"

// Country represents an ISO 3166-1 alpha-2 country code, normalized to
// uppercase (e.g. "US", "GB", "FR"). The zero value ("") means no country
// filter and will match all products regardless of shipping destination.
type Country string

// Currency represents an ISO 4217 currency code, normalized to uppercase
// (e.g. "USD", "EUR", "GBP"). The value is not validated against the ISO list;
// it is the caller's responsibility to supply a valid code.
type Currency string

// Language represents a BCP 47 language tag (e.g. "en", "fr-CA"). It is
// currently unused by business logic but typed for forward compatibility so
// that future localization support benefits from compile-time safety.
type Language string

// Category represents a merchant-defined product category. Original casing is
// preserved so that display-facing code can render the category as the merchant
// defined it (e.g. "Fresh Flowers"). Use the Matches method for
// case-insensitive comparison rather than ==.
type Category string

// NewCountry creates a Country from a raw string, normalizing to uppercase.
// An empty input produces the zero value (""), which acts as "no filter".
func NewCountry(s string) Country {
	return Country(strings.ToUpper(s))
}

// NewCurrency creates a Currency from a raw string, normalizing to uppercase.
// An empty input produces the zero value (""), which acts as "no filter".
func NewCurrency(s string) Currency {
	return Currency(strings.ToUpper(s))
}

// Matches performs a case-insensitive comparison between two Category values.
// If either value is empty, the result is false (an empty category does not
// match anything).
func (c Category) Matches(other Category) bool {
	return strings.EqualFold(string(c), string(other))
}

// UCPService identifies a UCP service in the discovery profile.
// Values use reverse-domain notation (e.g. "dev.ucp.shopping").
type UCPService string

// Official UCP service identifiers.
const (
	ServiceShopping UCPService = "dev.ucp.shopping"
)

// ContainsCountry reports whether c is present in countries. Both the list
// entries and c must already be normalized via NewCountry (i.e. uppercase),
// because the comparison uses simple ==. Returns false if countries is empty.
func ContainsCountry(countries []Country, c Country) bool {
	for _, cc := range countries {
		if cc == c {
			return true
		}
	}
	return false
}
