package discount

import (
	"github.com/owulveryck/ucp-merchant-test/internal/data"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// DiscountLookup provides access to discount code data.
type DiscountLookup interface {
	FindDiscountByCode(code string) *data.CSVDiscount
}

// ApplyDiscounts processes discount codes against line items.
func ApplyDiscounts(discountsRaw interface{}, lineItems []model.LineItem, dl DiscountLookup) *model.Discounts {
	dMap, ok := discountsRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	codesRaw, _ := dMap["codes"].([]interface{})
	if len(codesRaw) == 0 {
		return nil
	}

	subtotal := 0
	for _, li := range lineItems {
		for _, t := range li.Totals {
			if t.Type == "subtotal" {
				subtotal += t.Amount
			}
		}
	}

	result := &model.Discounts{}
	for _, cRaw := range codesRaw {
		code, _ := cRaw.(string)
		if code == "" {
			continue
		}
		result.Codes = append(result.Codes, code)

		d := dl.FindDiscountByCode(code)
		if d == nil {
			continue
		}

		var amount int
		switch d.Type {
		case "percentage":
			amount = subtotal * d.Value / 100
			subtotal -= amount
		case "fixed_amount":
			amount = d.Value
			subtotal -= amount
		}

		result.Applied = append(result.Applied, model.AppliedDiscount{
			Code:   d.Code,
			Title:  d.Description,
			Amount: amount,
		})
	}

	return result
}
