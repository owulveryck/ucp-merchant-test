package mcp

import (
	"context"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

func userIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserID).(string)
	return v
}

func userCountryFromContext(ctx context.Context) ucp.Country {
	v, _ := ctx.Value(ctxUserCountry).(ucp.Country)
	return v
}

// parseLineItemRequests converts a raw map's "line_items" field to typed requests.
func parseLineItemRequests(data map[string]interface{}) []model.LineItemRequest {
	rawItems, _ := data["line_items"].([]interface{})
	if len(rawItems) == 0 {
		return nil
	}
	items := make([]model.LineItemRequest, 0, len(rawItems))
	for _, raw := range rawItems {
		rawMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		li := model.LineItemRequest{}
		if id, ok := rawMap["id"].(string); ok {
			li.ID = id
		}
		if itemMap, ok := rawMap["item"].(map[string]interface{}); ok {
			if id, ok := itemMap["id"].(string); ok {
				li.Item = &model.ItemRef{ID: id}
			}
		}
		if pid, ok := rawMap["product_id"].(string); ok {
			li.ProductID = pid
		}
		if q, ok := rawMap["quantity"].(float64); ok {
			li.Quantity = int(q)
		}
		items = append(items, li)
	}
	return items
}

// parseBuyerRequest converts a raw buyer map to a typed BuyerRequest.
func parseBuyerRequest(data map[string]interface{}) *model.BuyerRequest {
	if data == nil {
		return nil
	}
	b := &model.BuyerRequest{}
	if v, ok := data["first_name"].(string); ok {
		b.FirstName = v
	}
	if v, ok := data["last_name"].(string); ok {
		b.LastName = v
	}
	if v, ok := data["fullName"].(string); ok {
		b.FullName = v
	}
	if v, ok := data["name"].(string); ok {
		b.Name = v
	}
	if v, ok := data["email"].(string); ok {
		b.Email = v
	}
	return b
}
