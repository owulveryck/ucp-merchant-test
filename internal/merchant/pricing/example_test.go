package pricing_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/pricing"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func ExampleBuildLineItems() {
	cat := catalog.New()
	cat.Products = []catalog.Product{
		{ID: "sku_roses", Title: "Bouquet of Roses", Price: 4999, Quantity: 10},
	}

	req := map[string]interface{}{
		"line_items": []interface{}{
			map[string]interface{}{
				"item":     map[string]interface{}{"id": "sku_roses"},
				"quantity": float64(2),
			},
		},
	}

	items, err := pricing.BuildLineItems(req, cat)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(items[0].Item.Title)
	fmt.Println(items[0].Quantity)
	fmt.Println(items[0].Totals[0].Amount)
	// Output:
	// Bouquet of Roses
	// 2
	// 9998
}

func ExampleCalculateTotals() {
	items := []model.LineItem{
		{
			ID:       "li_001",
			Item:     model.Item{ID: "sku_roses", Title: "Roses", Price: 4999},
			Quantity: 1,
			Totals:   []model.Total{{Type: "subtotal", Amount: 4999}},
		},
	}

	totals := pricing.CalculateTotals(items, 500, nil)
	for _, t := range totals {
		fmt.Printf("%s: %d\n", t.Type, t.Amount)
	}
	// Output:
	// subtotal: 4999
	// fulfillment: 500
	// total: 5499
}
