package a2a

import (
	"context"

	a2alib "github.com/a2aproject/a2a-go/a2a"

	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/ucp"
)

// contextKey is used for storing auth data in request context.
type contextKey int

const (
	ctxUserID contextKey = iota
	ctxUserCountry
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
func parseLineItemRequests(data map[string]any) []model.LineItemRequest {
	rawItems, _ := data["line_items"].([]any)
	if len(rawItems) == 0 {
		return nil
	}
	items := make([]model.LineItemRequest, 0, len(rawItems))
	for _, raw := range rawItems {
		rawMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		li := model.LineItemRequest{}
		if id, ok := rawMap["id"].(string); ok {
			li.ID = id
		}
		if itemMap, ok := rawMap["item"].(map[string]any); ok {
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
func parseBuyerRequest(data map[string]any) *model.BuyerRequest {
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

// asDataPart extracts the Data map from a Part, handling both value and pointer
// forms of DataPart.
func asDataPart(p a2alib.Part) (map[string]any, bool) {
	switch v := p.(type) {
	case a2alib.DataPart:
		return v.Data, true
	case *a2alib.DataPart:
		return v.Data, true
	}
	return nil, false
}

// parseFulfillmentRequest converts a raw fulfillment map to a typed FulfillmentRequest.
func parseFulfillmentRequest(data map[string]any) *model.FulfillmentRequest {
	if data == nil {
		return nil
	}
	fr := &model.FulfillmentRequest{}

	rawMethods, _ := data["methods"].([]any)
	for _, rm := range rawMethods {
		m, ok := rm.(map[string]any)
		if !ok {
			continue
		}
		method := model.FulfillmentMethodRequest{}
		if v, ok := m["id"].(string); ok {
			method.ID = v
		}
		if v, ok := m["type"].(string); ok {
			method.Type = v
		}
		if v, ok := m["selected_destination_id"].(string); ok {
			method.SelectedDestinationID = v
		}
		if rawDests, ok := m["destinations"].([]any); ok {
			for _, rd := range rawDests {
				dm, ok := rd.(map[string]any)
				if !ok {
					continue
				}
				dest := model.FulfillmentDestinationRequest{}
				if v, ok := dm["id"].(string); ok {
					dest.ID = v
				}
				if v, ok := dm["full_name"].(string); ok {
					dest.FullName = v
				}
				if v, ok := dm["street_address"].(string); ok {
					dest.StreetAddress = v
				}
				if v, ok := dm["address_locality"].(string); ok {
					dest.AddressLocality = v
				}
				if v, ok := dm["address_region"].(string); ok {
					dest.AddressRegion = v
				}
				if v, ok := dm["postal_code"].(string); ok {
					dest.PostalCode = v
				}
				if v, ok := dm["address_country"].(string); ok {
					dest.AddressCountry = ucp.Country(v)
				}
				method.Destinations = append(method.Destinations, dest)
			}
		}
		if rawGroups, ok := m["groups"].([]any); ok {
			for _, rg := range rawGroups {
				gm, ok := rg.(map[string]any)
				if !ok {
					continue
				}
				group := model.FulfillmentGroupRequest{}
				if v, ok := gm["selected_option_id"].(string); ok {
					group.SelectedOptionID = v
				}
				method.Groups = append(method.Groups, group)
			}
		}
		fr.Methods = append(fr.Methods, method)
	}
	return fr
}

// parseDiscountCodes converts a raw []any of discount codes to a typed DiscountsRequest.
func parseDiscountCodes(codes []any) *model.DiscountsRequest {
	dr := &model.DiscountsRequest{}
	for _, c := range codes {
		if s, ok := c.(string); ok {
			dr.Codes = append(dr.Codes, s)
		}
	}
	return dr
}

// extractPaymentFromParts scans all parts for a DataPart containing
// "a2a.ucp.checkout.payment" and returns its value.
func extractPaymentFromParts(parts a2alib.ContentParts) map[string]any {
	for _, p := range parts {
		data, ok := asDataPart(p)
		if !ok {
			continue
		}
		if payment, ok := data["a2a.ucp.checkout.payment"].(map[string]any); ok {
			return payment
		}
	}
	return nil
}
