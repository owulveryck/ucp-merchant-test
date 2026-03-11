package discount_test

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/discount"
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
)

type exampleDiscountLookup struct {
	discounts []discount.Discount
}

func (m *exampleDiscountLookup) FindDiscountByCode(code string) *discount.Discount {
	for i := range m.discounts {
		if m.discounts[i].Code == code {
			return &m.discounts[i]
		}
	}
	return nil
}

func ExampleApplyDiscounts() {
	dl := &exampleDiscountLookup{
		discounts: []discount.Discount{
			{Code: "10OFF", Type: "percentage", Value: 10, Description: "10% Off"},
		},
	}

	items := []model.LineItem{
		{
			ID:       "li_001",
			Item:     model.Item{ID: "sku_roses", Title: "Roses", Price: 5000},
			Quantity: 1,
			Totals:   []model.Total{{Type: "subtotal", Amount: 5000}},
		},
	}

	req := &model.DiscountsRequest{Codes: []string{"10OFF"}}

	result := discount.ApplyDiscounts(req, items, dl)
	fmt.Println(result.Codes[0])
	fmt.Println(result.Applied[0].Title)
	fmt.Println(result.Applied[0].Amount)
	// Output:
	// 10OFF
	// 10% Off
	// 500
}
