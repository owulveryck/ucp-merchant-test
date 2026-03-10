package catalog

import (
	"testing"

	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

func TestContainsCountry(t *testing.T) {
	countries := []ucp.Country{ucp.NewCountry("US"), ucp.NewCountry("GB"), ucp.NewCountry("FR")}
	if !ucp.ContainsCountry(countries, ucp.NewCountry("us")) {
		t.Error("expected US to match case-insensitively")
	}
	if ucp.ContainsCountry(countries, ucp.NewCountry("JP")) {
		t.Error("expected JP not in list")
	}
}
