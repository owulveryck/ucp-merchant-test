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
		Payment: model.Payment{Instruments: []map[string]any{}, Handlers: []map[string]any{}},
	}

	fmt.Println(co.ID)
	fmt.Println(co.Status)
	fmt.Println(co.Currency)
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

	fmt.Println(order.ID)
	fmt.Println(order.CheckoutID)
	// Output:
	// ord_001
	// co_001
}

func ExampleWebhookEvent() {
	event := model.WebhookEvent{
		EventType:  "order.created",
		CheckoutID: "co_001",
	}

	b, _ := json.Marshal(event)
	fmt.Println(string(b))
	// Output:
	// {"event_type":"order.created","checkout_id":"co_001"}
}

func ExampleUCPDiscovery() {
	d := model.UCPDiscovery{
		UCP: model.UCPDiscoveryProfile{
			Version: "2026-01-11",
			Services: map[string]model.UCPServiceEntry{
				"dev.ucp.shopping": {
					Version: "2026-01-11",
					Spec:    "https://ucp.dev/specs/shopping",
					REST:    &model.UCPRESTConfig{Endpoint: "/api/shopping"},
				},
			},
			Capabilities: []model.UCPCapabilityEntry{},
		},
		Payment: model.UCPPaymentProfile{Handlers: []map[string]any{}},
	}

	fmt.Println(d.UCP.Version)
	fmt.Println(d.UCP.Services["dev.ucp.shopping"].Spec)
	// Output:
	// 2026-01-11
	// https://ucp.dev/specs/shopping
}

func ExampleMCPToolResult() {
	result := model.MCPToolResult{
		Content: []model.MCPContentBlock{
			{Type: "text", Text: "Operation completed successfully"},
		},
	}

	b, _ := json.Marshal(result)
	fmt.Println(string(b))
	// Output:
	// {"content":[{"type":"text","text":"Operation completed successfully"}]}
}
