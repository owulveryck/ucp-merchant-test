package fulfillment_test

import (
	"fmt"
	"strings"

	"github.com/owulveryck/ucp-merchant-test/internal/merchant/fulfillment"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

type exampleFulfillmentDS struct {
	shippingRates []fulfillment.ShippingRate
	promotions    []fulfillment.Promotion
}

func (m *exampleFulfillmentDS) FindAddressesForEmail(email string) []fulfillment.Address {
	return nil
}

func (m *exampleFulfillmentDS) SaveDynamicAddress(email string, addr fulfillment.Address) string {
	return addr.ID
}

func (m *exampleFulfillmentDS) GetShippingRatesForCountry(country string) []fulfillment.ShippingRate {
	var result []fulfillment.ShippingRate
	for _, r := range m.shippingRates {
		if strings.EqualFold(r.CountryCode, country) {
			result = append(result, r)
		}
	}
	return result
}

func (m *exampleFulfillmentDS) GetPromotions() []fulfillment.Promotion {
	return m.promotions
}

func ExampleGenerateShippingOptions() {
	ds := &exampleFulfillmentDS{
		shippingRates: []fulfillment.ShippingRate{
			{ID: "rate_1", CountryCode: "US", ServiceLevel: "standard", Price: 500, Title: "Standard Shipping"},
			{ID: "rate_2", CountryCode: "US", ServiceLevel: "express", Price: 1500, Title: "Express Shipping"},
		},
	}

	options := fulfillment.GenerateShippingOptions("US", nil, ds)
	for _, opt := range options {
		fmt.Printf("%s: %d\n", opt.Title, opt.Totals[0].Amount)
	}
	// Output:
	// Standard Shipping: 500
	// Express Shipping: 1500
}

func ExampleIsFulfillmentComplete() {
	co := &model.Checkout{
		Fulfillment: &model.Fulfillment{
			Methods: []model.FulfillmentMethod{
				{
					ID:                    "method_shipping",
					Type:                  "shipping",
					SelectedDestinationID: "addr_1",
					Groups: []model.FulfillmentGroup{
						{ID: "group_1", SelectedOptionID: "rate_1"},
					},
				},
			},
		},
	}

	fmt.Println(fulfillment.IsFulfillmentComplete(co))
	// Output:
	// true
}
