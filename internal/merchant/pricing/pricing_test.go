package pricing

import (
	"testing"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

type testCatalog struct {
	products []catalog.Product
}

func (c *testCatalog) Find(id string) *catalog.Product {
	for i := range c.products {
		if c.products[i].ID == id {
			return &c.products[i]
		}
	}
	return nil
}

func (c *testCatalog) Filter(category, brand, query, usageType, country, currency, language string) []catalog.Product {
	return nil
}

func (c *testCatalog) CategoryCount() []map[string]interface{} {
	return nil
}

func TestBuildLineItems(t *testing.T) {
	cat := &testCatalog{
		products: []catalog.Product{
			{ID: "prod_1", Title: "Test Product", Price: 1000, Quantity: 10},
		},
	}

	req := map[string]interface{}{
		"line_items": []interface{}{
			map[string]interface{}{
				"item":     map[string]interface{}{"id": "prod_1"},
				"quantity": float64(2),
			},
		},
	}

	items, err := BuildLineItems(req, cat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Quantity != 2 {
		t.Errorf("expected qty 2, got %d", items[0].Quantity)
	}
	if items[0].Item.Price != 1000 {
		t.Errorf("expected price 1000, got %d", items[0].Item.Price)
	}
}

func TestBuildLineItemsOutOfStock(t *testing.T) {
	cat := &testCatalog{
		products: []catalog.Product{
			{ID: "prod_1", Title: "Out of Stock", Price: 1000, Quantity: 0},
		},
	}

	req := map[string]interface{}{
		"line_items": []interface{}{
			map[string]interface{}{
				"item":     map[string]interface{}{"id": "prod_1"},
				"quantity": float64(1),
			},
		},
	}

	_, err := BuildLineItems(req, cat)
	if err == nil {
		t.Fatal("expected error for out of stock product")
	}
}

func TestCalculateTotals(t *testing.T) {
	items := []model.LineItem{
		{
			ID:       "li_001",
			Quantity: 2,
			Totals:   []model.Total{{Type: "subtotal", Amount: 2000}},
		},
	}

	totals := CalculateTotals(items, 500, nil)

	subtotal := findTotal(totals, "subtotal")
	if subtotal != 2000 {
		t.Errorf("expected subtotal 2000, got %d", subtotal)
	}

	shipping := findTotal(totals, "fulfillment")
	if shipping != 500 {
		t.Errorf("expected fulfillment 500, got %d", shipping)
	}

	total := findTotal(totals, "total")
	if total != 2500 {
		t.Errorf("expected total 2500, got %d", total)
	}
}

func TestCalculateTotalsWithDiscount(t *testing.T) {
	items := []model.LineItem{
		{Totals: []model.Total{{Type: "subtotal", Amount: 10000}}},
	}
	discounts := &model.Discounts{
		Applied: []model.AppliedDiscount{{Amount: 1000}},
	}

	totals := CalculateTotals(items, 0, discounts)
	total := findTotal(totals, "total")
	if total != 9000 {
		t.Errorf("expected total 9000 after discount, got %d", total)
	}
}

func findTotal(totals []model.Total, typ string) int {
	for _, t := range totals {
		if t.Type == typ {
			return t.Amount
		}
	}
	return -1
}
