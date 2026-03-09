package discount_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/data"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/discount"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func ExampleApplyDiscounts() {
	ds := data.New()
	ds.Discounts = []data.CSVDiscount{
		{Code: "10OFF", Type: "percentage", Value: 10, Description: "10% Off"},
	}

	items := []model.LineItem{
		{
			ID:       "li_001",
			Item:     model.Item{ID: "sku_roses", Title: "Roses", Price: 5000},
			Quantity: 1,
			Totals:   []model.Total{{Type: "subtotal", Amount: 5000}},
		},
	}

	discountsRaw := map[string]interface{}{
		"codes": []interface{}{"10OFF"},
	}

	result := discount.ApplyDiscounts(discountsRaw, items, ds)
	fmt.Println(result.Codes[0])
	fmt.Println(result.Applied[0].Title)
	fmt.Println(result.Applied[0].Amount)
	// Output:
	// 10OFF
	// 10% Off
	// 500
}
