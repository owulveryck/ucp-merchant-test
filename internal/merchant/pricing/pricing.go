package pricing

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// BuildLineItems constructs the UCP line_items array from a platform's checkout
// or cart request. It resolves each [model.LineItemRequest] against the product
// catalog, validates stock availability, and computes per-item subtotals.
//
// Product resolution tries Item.ID first, then falls back to ProductID. Both
// must match a variant ID from the merchant's catalog
// (dev.ucp.shopping.catalog.lookup). If the product is not found or is out of
// stock, an error is returned.
//
// Each resulting [model.LineItem] includes:
//   - A server-assigned ID ("li_001", "li_002", …) unless the caller provides one
//   - The resolved product metadata (title, unit price, image URL)
//   - The requested quantity (minimum 1)
//   - Totals with "subtotal" (price × quantity) and "total" entries
//
// This function is called during checkout create, checkout update (when
// line_items are provided), and cart create/update operations.
func BuildLineItems(reqItems []model.LineItemRequest, cat catalog.Catalog) ([]model.LineItem, error) {
	if len(reqItems) == 0 {
		return nil, fmt.Errorf("line_items is required")
	}

	var items []model.LineItem
	for i, li := range reqItems {
		itemID := ""
		if li.Item != nil {
			itemID = li.Item.ID
		}
		if itemID == "" {
			itemID = li.ProductID
		}
		if itemID == "" {
			return nil, fmt.Errorf("line item %d: missing item.id", i)
		}

		product := cat.Find(itemID)
		if product == nil {
			return nil, fmt.Errorf("Product not found: %s", itemID)
		}

		qty := li.Quantity
		if qty < 1 {
			qty = 1
		}

		if product.Quantity <= 0 {
			return nil, fmt.Errorf("Insufficient stock for product %s", itemID)
		}
		if qty > product.Quantity {
			return nil, fmt.Errorf("Insufficient stock for product %s: requested %d, available %d", itemID, qty, product.Quantity)
		}

		lineTotal := product.Price * qty

		liID := fmt.Sprintf("li_%03d", i+1)
		if li.ID != "" {
			liID = li.ID
		}

		items = append(items, model.LineItem{
			ID: liID,
			Item: model.Item{
				ID:       product.ID,
				Title:    product.Title,
				Price:    product.Price,
				ImageURL: product.ImageURL,
			},
			Quantity: qty,
			Totals: []model.Total{
				{Type: "subtotal", Amount: lineTotal},
				{Type: "total", Amount: lineTotal},
			},
		})
	}
	return items, nil
}

// CalculateTotals computes the order-level totals array for a UCP checkout or
// cart response. It aggregates line item subtotals, applies discounts, adds the
// fulfillment cost, and produces the final total.
//
// The returned totals array follows the UCP-defined ordering:
//   - "subtotal": sum of all line item subtotals (before discounts and shipping)
//   - "discount": total discount amount (only present when discounts are applied;
//     the amount is always a positive integer, displayed as subtractive by platforms)
//   - "fulfillment": shipping/delivery cost from the selected fulfillment option
//     (only present when shippingCost > 0; UCP requires this type name, never "shipping")
//   - "total": final amount = subtotal − discount + fulfillment
//
// All amounts are in minor currency units (e.g., cents). Each total includes a
// DisplayText formatted as "$X.XX" for human-readable rendering.
//
// This function is called after every checkout or cart mutation that could affect
// pricing: line item changes, discount application, and fulfillment option selection.
func CalculateTotals(items []model.LineItem, shippingCost int, discounts *model.Discounts) []model.Total {
	subtotal := 0
	for _, li := range items {
		for _, t := range li.Totals {
			if t.Type == "subtotal" {
				subtotal += t.Amount
			}
		}
	}

	total := subtotal

	var totals []model.Total
	totals = append(totals, model.Total{
		Type:        "subtotal",
		DisplayText: fmt.Sprintf("$%.2f", float64(subtotal)/100),
		Amount:      subtotal,
	})

	if discounts != nil {
		discountAmount := 0
		for _, d := range discounts.Applied {
			discountAmount += d.Amount
		}
		if discountAmount > 0 {
			total -= discountAmount
			totals = append(totals, model.Total{
				Type:        "discount",
				DisplayText: fmt.Sprintf("-$%.2f", float64(discountAmount)/100),
				Amount:      discountAmount,
			})
		}
	}

	if shippingCost > 0 {
		total += shippingCost
		totals = append(totals, model.Total{
			Type:        "fulfillment",
			DisplayText: fmt.Sprintf("$%.2f", float64(shippingCost)/100),
			Amount:      shippingCost,
		})
	}

	totals = append(totals, model.Total{
		Type:        "total",
		DisplayText: fmt.Sprintf("$%.2f", float64(total)/100),
		Amount:      total,
	})

	return totals
}
