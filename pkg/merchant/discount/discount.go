package discount

import (
	"github.com/owulveryck/ucp-merchant-test/pkg/model"
)

// Discount represents a discount code definition in the merchant's discount
// catalog. Each discount has a code (e.g., "10OFF", "WELCOME20", "FIXED500"),
// a type indicating the allocation method, a value, and a human-readable
// description shown to the buyer.
//
// Type must be one of:
//   - "percentage": Value is applied as a percentage of the remaining subtotal
//     (e.g., Value=10 means 10% off)
//   - "fixed_amount": Value is a fixed amount in minor currency units
//     (e.g., Value=500 means $5.00 off)
type Discount struct {
	Code        string
	Type        string // "percentage" or "fixed_amount"
	Value       int
	Description string
}

// DiscountLookup provides access to the merchant's discount code catalog.
// Implementations look up discount definitions by code string, returning nil
// when the code is not recognized. Code matching should be case-insensitive
// per UCP conventions.
type DiscountLookup interface {
	// FindDiscountByCode looks up a discount definition by its code string
	// (case-insensitive). Returns nil when the code is not recognized.
	FindDiscountByCode(code string) *Discount
}

// ApplyDiscounts processes discount codes submitted by the platform via the
// UCP Discount Extension (dev.ucp.shopping.discount) and computes the applied
// discount amounts against the current line item subtotals.
//
// When req is nil or contains no codes, nil is returned (no discounts section
// in the checkout response). Otherwise, each code is looked up via the
// [DiscountLookup] interface:
//   - Recognized codes are applied sequentially against the remaining subtotal.
//     Percentage discounts compound: each subsequent percentage discount operates
//     on the subtotal after prior discounts have been deducted.
//   - Unrecognized codes are still echoed back in the result's Codes slice
//     (per UCP, rejected codes appear in discounts.codes but not in
//     discounts.applied), allowing the platform to surface rejection messages.
//
// The returned [model.Discounts] contains:
//   - Codes: all submitted codes (both valid and invalid), preserving order
//   - Applied: successfully applied discounts with code, title, and amount
//
// All discount amounts are positive integers in minor currency units. Platforms
// display them as subtractive (e.g., "-$5.00").
//
// This function is called during checkout update when the platform submits or
// modifies discount codes.
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
