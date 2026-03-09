package catalog

import "testing"

func TestFind(t *testing.T) {
	c := New()
	c.Products = []Product{
		{ID: "sku_1", Title: "Widget", Price: 1000, Quantity: 10},
		{ID: "sku_2", Title: "Gadget", Price: 2000, Quantity: 5},
	}

	p := c.Find("sku_1")
	if p == nil || p.Title != "Widget" {
		t.Errorf("expected Widget, got %v", p)
	}

	p = c.Find("sku_999")
	if p != nil {
		t.Errorf("expected nil, got %v", p)
	}
}

func TestFilter(t *testing.T) {
	c := New()
	c.Products = []Product{
		{ID: "1", Title: "Red Widget", Category: "Tools", Brand: "Acme", UsageType: "intensive"},
		{ID: "2", Title: "Blue Gadget", Category: "Electronics", Brand: "TechCo", UsageType: "occasional"},
		{ID: "3", Title: "Green Widget", Category: "Tools", Brand: "Acme", UsageType: "versatile"},
	}

	result := c.Filter("Tools", "", "", "", "")
	if len(result) != 2 {
		t.Errorf("expected 2 tools, got %d", len(result))
	}

	result = c.Filter("", "Acme", "", "", "")
	if len(result) != 2 {
		t.Errorf("expected 2 Acme products, got %d", len(result))
	}

	result = c.Filter("", "", "gadget", "", "")
	if len(result) != 1 {
		t.Errorf("expected 1 gadget, got %d", len(result))
	}
}

func TestContainsCountry(t *testing.T) {
	countries := []string{"US", "GB", "FR"}
	if !ContainsCountry(countries, "us") {
		t.Error("expected US to match case-insensitively")
	}
	if ContainsCountry(countries, "JP") {
		t.Error("expected JP not in list")
	}
}

func TestInit(t *testing.T) {
	c := New()
	c.Init(42)
	if len(c.Products) == 0 {
		t.Error("expected products after init")
	}
	if c.ProductSeq != len(c.Products) {
		t.Errorf("expected ProductSeq=%d, got %d", len(c.Products), c.ProductSeq)
	}
}
