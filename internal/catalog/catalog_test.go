package catalog

import "testing"

func TestContainsCountry(t *testing.T) {
	countries := []string{"US", "GB", "FR"}
	if !ContainsCountry(countries, "us") {
		t.Error("expected US to match case-insensitively")
	}
	if ContainsCountry(countries, "JP") {
		t.Error("expected JP not in list")
	}
}
