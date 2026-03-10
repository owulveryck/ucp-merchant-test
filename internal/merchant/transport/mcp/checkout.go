package mcp

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func (s *Server) handleCreateCheckout(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	checkoutData, _ := args["checkout"].(map[string]interface{})
	if checkoutData == nil {
		return nil, fmt.Errorf("missing checkout parameter")
	}

	req := &model.CheckoutRequest{
		Currency: "USD",
	}

	// Check if creating from a cart
	if cartID, ok := checkoutData["cart_id"].(string); ok && cartID != "" {
		// Get cart line items via the merchant
		cart, err := s.merchant.GetCart(cartID, userID)
		if err != nil {
			return nil, fmt.Errorf("cart not found: %s", cartID)
		}
		// Convert cart line items to request format
		for _, li := range cart.LineItems {
			req.LineItems = append(req.LineItems, model.LineItemRequest{
				ID:       li.ID,
				Item:     &model.ItemRef{ID: li.Item.ID},
				Quantity: li.Quantity,
			})
		}
	} else {
		coLineItems := parseLineItemRequests(checkoutData)
		if len(coLineItems) == 0 {
			return nil, fmt.Errorf("checkout must have line_items or cart_id")
		}
		req.LineItems = coLineItems
	}

	// Check for buyer info
	if buyerData, ok := checkoutData["buyer"].(map[string]interface{}); ok {
		req.Buyer = parseBuyerRequest(buyerData)
	}

	co, hash, err := s.merchant.CreateCheckout(userID, userCountry, req)
	if err != nil {
		return nil, err
	}

	return mcpCheckoutResponse(co, hash), nil
}

func (s *Server) handleGetCheckout(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	id, _ := args["id"].(string)

	co, hash, err := s.merchant.GetCheckout(id, userID)
	if err != nil {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}

	return mcpCheckoutResponse(co, hash), nil
}

func (s *Server) handleUpdateCheckout(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	id, _ := args["id"].(string)

	checkoutData, _ := args["checkout"].(map[string]interface{})
	if checkoutData == nil {
		return nil, fmt.Errorf("missing checkout parameter")
	}

	req := &model.CheckoutRequest{}

	coLineItems := parseLineItemRequests(checkoutData)
	if len(coLineItems) > 0 {
		req.LineItems = coLineItems
	}

	if buyerData, ok := checkoutData["buyer"].(map[string]interface{}); ok {
		req.Buyer = parseBuyerRequest(buyerData)
	}

	co, hash, err := s.merchant.UpdateCheckout(id, userID, req)
	if err != nil {
		return nil, err
	}

	return mcpCheckoutResponse(co, hash), nil
}

func (s *Server) handleCompleteCheckout(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	id, _ := args["id"].(string)

	// Verify approval hash
	approval, _ := args["approval"].(map[string]interface{})
	if approval == nil {
		return nil, fmt.Errorf("approval is required: the platform must present the checkout to the user for approval before completing")
	}
	approvalHash, _ := approval["checkout_hash"].(string)
	if approvalHash == "" {
		return nil, fmt.Errorf("approval.checkout_hash is required")
	}

	req := &model.CheckoutRequest{}

	// Parse optional payment from checkout data
	if checkoutData, ok := args["checkout"].(map[string]interface{}); ok {
		if _, hasPayment := checkoutData["payment"]; hasPayment {
			// Payment is present but we don't need to do anything special
			// for MCP — payment processing is simulated
		}
	}

	co, _, hash, err := s.merchant.CompleteCheckout(id, userID, userCountry, approvalHash, req)
	if err != nil {
		return nil, err
	}

	return mcpCheckoutResponse(co, hash), nil
}

func (s *Server) handleCancelCheckout(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	id, _ := args["id"].(string)

	co, hash, err := s.merchant.CancelCheckout(id, userID)
	if err != nil {
		return nil, err
	}

	return mcpCheckoutResponse(co, hash), nil
}

// mcpCheckoutResponse builds the MCP response for a checkout, adding the checkout hash.
func mcpCheckoutResponse(co *model.Checkout, hash string) interface{} {
	resp := map[string]interface{}{
		"id":         co.ID,
		"status":     co.Status,
		"currency":   co.Currency,
		"line_items": co.LineItems,
		"totals":     co.Totals,
		"links":      co.Links,
	}
	if hash != "" {
		resp["checkout_hash"] = hash
	}
	if co.Buyer != nil {
		resp["buyer"] = co.Buyer
	}
	if co.Order != nil {
		resp["order"] = map[string]interface{}{
			"id":            co.Order.ID,
			"permalink_url": co.Order.PermalinkURL,
		}
	}
	if co.Fulfillment != nil {
		resp["fulfillment"] = co.Fulfillment
	}
	if co.Discounts != nil {
		resp["discounts"] = co.Discounts
	}
	return resp
}
