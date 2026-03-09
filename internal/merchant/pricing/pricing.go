package pricing

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// BuildLineItems creates line items from typed request data.
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

// CalculateTotals computes checkout/cart totals from line items, shipping cost, and discounts.
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
