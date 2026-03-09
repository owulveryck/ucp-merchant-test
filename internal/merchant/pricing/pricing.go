package pricing

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// BuildLineItems creates line items from a raw request map.
func BuildLineItems(req map[string]interface{}, cat catalog.Catalog) ([]model.LineItem, error) {
	rawItems, _ := req["line_items"].([]interface{})
	if len(rawItems) == 0 {
		return nil, fmt.Errorf("line_items is required")
	}

	var items []model.LineItem
	for i, raw := range rawItems {
		rawMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		itemID := ""
		if itemMap, ok := rawMap["item"].(map[string]interface{}); ok {
			itemID, _ = itemMap["id"].(string)
		}
		if itemID == "" {
			itemID, _ = rawMap["product_id"].(string)
		}
		if itemID == "" {
			return nil, fmt.Errorf("line item %d: missing item.id", i)
		}

		product := cat.Find(itemID)
		if product == nil {
			return nil, fmt.Errorf("Product not found: %s", itemID)
		}

		qty := 1
		if q, ok := rawMap["quantity"].(float64); ok {
			qty = int(q)
		}
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
		if existingID, ok := rawMap["id"].(string); ok && existingID != "" {
			liID = existingID
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
