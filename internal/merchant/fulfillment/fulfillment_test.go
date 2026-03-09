package fulfillment

import (
	"sync"
	"testing"

	"github.com/owulveryck/ucp-merchant-test/internal/data"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func TestParseFulfillmentNil(t *testing.T) {
	ds := data.New()
	counter := 0
	var mu sync.Mutex
	result := ParseFulfillment(map[string]interface{}{}, nil, nil, ds, nil, nil, &counter, &mu)
	if result != nil {
		t.Error("expected nil fulfillment when not provided")
	}
}

func TestParseFulfillmentBasic(t *testing.T) {
	ds := data.New()
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
	ds := data.New()
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
	ds := data.New()
	ds.ShippingRates = []data.CSVShippingRate{
		{ID: "r1", CountryCode: "US", ServiceLevel: "standard", Price: 500, Title: "Standard"},
		{ID: "r2", CountryCode: "US", ServiceLevel: "express", Price: 1500, Title: "Express"},
	}

	options := GenerateShippingOptions("US", nil, ds)
	if len(options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(options))
	}
}

func TestGenerateShippingOptionsFreeShipping(t *testing.T) {
	ds := data.New()
	ds.ShippingRates = []data.CSVShippingRate{
		{ID: "r1", CountryCode: "US", ServiceLevel: "standard", Price: 500, Title: "Standard"},
	}
	ds.Promotions = []data.CSVPromotion{
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
