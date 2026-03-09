package pricing_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/pricing"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

type exampleCatalog struct {
	products []catalog.Product
}

func (c *exampleCatalog) Find(id string) *catalog.Product {
	for i := range c.products {
		if c.products[i].ID == id {
			return &c.products[i]
		}
	}
	return nil
}

func (c *exampleCatalog) Filter(category, brand, query, usageType, country, currency, language string) []catalog.Product {
	return nil
}

func (c *exampleCatalog) CategoryCount() []map[string]interface{} {
	return nil
}

func ExampleBuildLineItems() {
	cat := &exampleCatalog{
		products: []catalog.Product{
			{ID: "sku_roses", Title: "Bouquet of Roses", Price: 4999, Quantity: 10},
		},
	}

	reqItems := []model.LineItemRequest{
		{Item: &model.ItemRef{ID: "sku_roses"}, Quantity: 2},
	}

	items, err := pricing.BuildLineItems(reqItems, cat)
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
