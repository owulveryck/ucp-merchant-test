package fulfillment

import (
	"strings"
	"sync"
	"testing"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

type mockFulfillmentDS struct {
	addresses     map[string][]Address
	shippingRates []ShippingRate
	promotions    []Promotion
}

func newMockDS() *mockFulfillmentDS {
	return &mockFulfillmentDS{
		addresses: make(map[string][]Address),
	}
}

func (m *mockFulfillmentDS) FindAddressesForEmail(email string) []Address {
	return m.addresses[strings.ToLower(email)]
}

func (m *mockFulfillmentDS) SaveDynamicAddress(email string, addr Address) string {
	key := strings.ToLower(email)
	m.addresses[key] = append(m.addresses[key], addr)
	return addr.ID
}

func (m *mockFulfillmentDS) GetShippingRatesForCountry(country string) []ShippingRate {
	var result []ShippingRate
	for _, r := range m.shippingRates {
		if strings.EqualFold(r.CountryCode, country) || r.CountryCode == "default" {
			result = append(result, r)
		}
	}
	return result
}

func (m *mockFulfillmentDS) GetPromotions() []Promotion {
	return m.promotions
}

func TestParseFulfillmentNil(t *testing.T) {
	ds := newMockDS()
	counter := 0
	var mu sync.Mutex
	result := ParseFulfillment(map[string]interface{}{}, nil, nil, ds, nil, nil, &counter, &mu)
	if result != nil {
		t.Error("expected nil fulfillment when not provided")
	}
}

func TestParseFulfillmentBasic(t *testing.T) {
	ds := newMockDS()
	counter := 0
	var mu sync.Mutex

	req := map[string]interface{}{
		"fulfillment": map[string]interface{}{
			"methods": []interface{}{
				map[string]interface{}{
					"id":   "method_shipping",
					"type": "shipping",
				},
			},
		},
	}

	result := ParseFulfillment(req, nil, nil, ds, map[string]*model.FulfillmentDestination{}, map[string]string{}, &counter, &mu)
	if result == nil {
		t.Fatal("expected non-nil fulfillment")
	}
	if len(result.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(result.Methods))
	}
	if result.Methods[0].ID != "method_shipping" {
		t.Errorf("expected method_shipping, got %s", result.Methods[0].ID)
	}
}

func TestParseDestination(t *testing.T) {
	ds := newMockDS()
	counter := 0
	var mu sync.Mutex

	dMap := map[string]interface{}{
		"street_address":   "123 Main St",
		"address_locality": "NYC",
		"address_region":   "NY",
		"postal_code":      "10001",
		"address_country":  "US",
	}

	dest := ParseDestination(dMap, nil, ds, &counter, &mu)
	if dest.StreetAddress != "123 Main St" {
		t.Errorf("expected 123 Main St, got %s", dest.StreetAddress)
	}
	if dest.ID == "" {
		t.Error("expected auto-generated ID")
	}
	if counter != 1 {
		t.Errorf("expected counter=1, got %d", counter)
	}
}

func TestGetCurrentShippingCost(t *testing.T) {
	co := &model.Checkout{
		Fulfillment: &model.Fulfillment{
			Methods: []model.FulfillmentMethod{
				{
					Groups: []model.FulfillmentGroup{
						{
							SelectedOptionID: "opt_1",
							Options: []model.FulfillmentOption{
								{ID: "opt_1", Totals: []model.Total{{Type: "total", Amount: 500}}},
							},
						},
					},
				},
			},
		},
	}

	cost := GetCurrentShippingCost(co)
	if cost != 500 {
		t.Errorf("expected 500, got %d", cost)
	}
}

func TestGetCurrentShippingCostNoFulfillment(t *testing.T) {
	co := &model.Checkout{}
	cost := GetCurrentShippingCost(co)
	if cost != 0 {
		t.Errorf("expected 0, got %d", cost)
	}
}

func TestIsFulfillmentComplete(t *testing.T) {
	co := &model.Checkout{
		Fulfillment: &model.Fulfillment{
			Methods: []model.FulfillmentMethod{
				{
					SelectedDestinationID: "dest_1",
					Groups: []model.FulfillmentGroup{
						{SelectedOptionID: "opt_1"},
					},
				},
			},
		},
	}
	if !IsFulfillmentComplete(co) {
		t.Error("expected fulfillment to be complete")
	}
}

func TestIsFulfillmentIncomplete(t *testing.T) {
	co := &model.Checkout{
		Fulfillment: &model.Fulfillment{
			Methods: []model.FulfillmentMethod{
				{
					SelectedDestinationID: "dest_1",
					Groups:                []model.FulfillmentGroup{},
				},
			},
		},
	}
	if IsFulfillmentComplete(co) {
		t.Error("expected fulfillment to be incomplete (no selected option)")
	}
}

func TestGenerateShippingOptions(t *testing.T) {
	ds := newMockDS()
	ds.shippingRates = []ShippingRate{
		{ID: "r1", CountryCode: "US", ServiceLevel: "standard", Price: 500, Title: "Standard"},
		{ID: "r2", CountryCode: "US", ServiceLevel: "express", Price: 1500, Title: "Express"},
	}

	options := GenerateShippingOptions("US", nil, ds)
	if len(options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(options))
	}
}

func TestGenerateShippingOptionsFreeShipping(t *testing.T) {
	ds := newMockDS()
	ds.shippingRates = []ShippingRate{
		{ID: "r1", CountryCode: "US", ServiceLevel: "standard", Price: 500, Title: "Standard"},
	}
	ds.promotions = []Promotion{
		{ID: "p1", Type: "free_shipping", MinSubtotal: 10000},
	}

	co := &model.Checkout{
		LineItems: []model.LineItem{
			{Totals: []model.Total{{Type: "subtotal", Amount: 15000}}},
		},
	}

	options := GenerateShippingOptions("US", co, ds)
	if len(options) != 1 {
		t.Fatalf("expected 1 option, got %d", len(options))
	}
	if options[0].Totals[0].Amount != 0 {
		t.Errorf("expected free shipping (0), got %d", options[0].Totals[0].Amount)
	}
}

func TestMatchExistingAddress(t *testing.T) {
	addrs := []Address{
		{ID: "addr_1", StreetAddress: "123 Main St", City: "NYC", State: "NY", PostalCode: "10001", Country: "US"},
	}

	matched := MatchExistingAddress(addrs, "123 main st", "nyc", "ny", "10001", "us")
	if matched == nil || matched.ID != "addr_1" {
		t.Error("expected case-insensitive address match")
	}

	matched = MatchExistingAddress(addrs, "456 Oak Ave", "LA", "CA", "90001", "US")
	if matched != nil {
		t.Error("expected nil for non-matching address")
	}
}
