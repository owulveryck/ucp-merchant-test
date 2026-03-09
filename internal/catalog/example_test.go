package catalog_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
)

func ExampleCatalog_Find() {
	c := catalog.New()
	c.Products = []catalog.Product{
		{ID: "sku_roses", Title: "Bouquet of Roses", Price: 4999, Quantity: 10},
		{ID: "sku_tulips", Title: "Tulip Bundle", Price: 2999, Quantity: 5},
	}

	p := c.Find("sku_roses")
	fmt.Println(p.Title)
	fmt.Println(p.Price)
	// Output:
	// Bouquet of Roses
	// 4999
}

func ExampleCatalog_Filter() {
	c := catalog.New()
	c.Products = []catalog.Product{
		{ID: "1", Title: "Red Widget", Category: "Tools", Brand: "Acme"},
		{ID: "2", Title: "Blue Gadget", Category: "Electronics", Brand: "TechCo"},
		{ID: "3", Title: "Green Widget", Category: "Tools", Brand: "Acme"},
	}

	results := c.Filter("Tools", "", "", "", "")
	fmt.Println(len(results))
	for _, p := range results {
		fmt.Println(p.Title)
	}
	// Output:
	// 2
	// Red Widget
	// Green Widget
}
