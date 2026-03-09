package discount

import (
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// Discount represents a discount code with its type and value.
type Discount struct {
	Code        string
	Type        string // "percentage" or "fixed_amount"
	Value       int
	Description string
}

// DiscountLookup provides access to discount code data.
type DiscountLookup interface {
	FindDiscountByCode(code string) *Discount
}

// ApplyDiscounts processes discount codes against line items.
func ApplyDiscounts(req *model.DiscountsRequest, lineItems []model.LineItem, dl DiscountLookup) *model.Discounts {
	if req == nil || len(req.Codes) == 0 {
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
	for _, code := range req.Codes {
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
