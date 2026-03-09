package model_test

import (
	"encoding/json"
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func ExampleCheckout() {
	co := model.Checkout{
		ID:       "co_001",
		UCP:      model.UCPEnvelope{Version: "2026-01-11", Capabilities: []model.Capability{}},
		Status:   "incomplete",
		Currency: "USD",
		LineItems: []model.LineItem{
			{
				ID:       "li_001",
				Item:     model.Item{ID: "sku_roses", Title: "Bouquet of Roses", Price: 4999},
				Quantity: 2,
				Totals:   []model.Total{{Type: "subtotal", Amount: 9998}, {Type: "total", Amount: 9998}},
			},
		},
		Totals:  []model.Total{{Type: "subtotal", Amount: 9998}, {Type: "total", Amount: 9998}},
		Payment: model.Payment{Instruments: []map[string]interface{}{}, Handlers: []map[string]interface{}{}},
	}

	b, _ := json.Marshal(co)
	var out map[string]interface{}
	json.Unmarshal(b, &out)

	fmt.Println(out["id"])
	fmt.Println(out["status"])
	fmt.Println(out["currency"])
	// Output:
	// co_001
	// incomplete
	// USD
}

func ExampleOrder() {
	order := model.Order{
		ID:         "ord_001",
		UCP:        model.UCPEnvelope{Version: "2026-01-11", Capabilities: []model.Capability{}},
		CheckoutID: "co_001",
		Currency:   "USD",
		LineItems: []model.OrderLineItem{
			{
				ID:       "li_001",
				Item:     model.Item{ID: "sku_roses", Title: "Bouquet of Roses", Price: 4999},
				Quantity: model.OrderQuantity{Total: 2, Fulfilled: 0},
				Totals:   []model.Total{{Type: "subtotal", Amount: 9998}, {Type: "total", Amount: 9998}},
				Status:   "active",
			},
		},
		Fulfillment: model.OrderFulfillment{
			Expectations: []model.Expectation{
				{
					ID:         "exp_001",
					MethodType: "shipping",
					LineItems:  []model.EventLineItem{{ID: "li_001", Quantity: 2}},
					Destination: model.FulfillmentDestination{
						ID:             "addr_1",
						StreetAddress:  "123 Main St",
						AddressCountry: "US",
					},
				},
			},
		},
		Totals: []model.Total{{Type: "subtotal", Amount: 9998}, {Type: "total", Amount: 9998}},
	}

	b, _ := json.Marshal(order)
	var out map[string]interface{}
	json.Unmarshal(b, &out)

	fmt.Println(out["id"])
	fmt.Println(out["checkout_id"])
	// Output:
	// ord_001
	// co_001
}
