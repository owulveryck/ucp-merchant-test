package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

func (s *Server) handleCreateCheckout(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)
	userCountry := userCountryFromContext(ctx)

	checkoutData, _ := args["checkout"].(map[string]interface{})
	if checkoutData == nil {
		return toolResultFromError(fmt.Errorf("missing checkout parameter")), nil
	}

	req := &model.CheckoutRequest{
		Currency: "USD",
	}

	// Check if creating from a cart
	if cartID, ok := checkoutData["cart_id"].(string); ok && cartID != "" {
		cart, err := s.merchant.GetCart(cartID, userID)
		if err != nil {
			return toolResultFromError(fmt.Errorf("cart not found: %s", cartID)), nil
		}
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
			return toolResultFromError(fmt.Errorf("checkout must have line_items or cart_id")), nil
		}
		req.LineItems = coLineItems
	}

	if buyerData, ok := checkoutData["buyer"].(map[string]interface{}); ok {
		req.Buyer = parseBuyerRequest(buyerData)
	}

	co, hash, err := s.merchant.CreateCheckout(userID, userCountry, req)
	if err != nil {
		return toolResultFromError(err), nil
	}

	return toolResultFromJSON(mcpCheckoutResponse(co, hash), nil), nil
}

func (s *Server) handleGetCheckout(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)

	id, _ := args["id"].(string)

	co, hash, err := s.merchant.GetCheckout(id, userID)
	if err != nil {
		return toolResultFromError(fmt.Errorf("checkout not found: %s", id)), nil
	}

	return toolResultFromJSON(mcpCheckoutResponse(co, hash), nil), nil
}

func (s *Server) handleUpdateCheckout(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)

	id, _ := args["id"].(string)

	checkoutData, _ := args["checkout"].(map[string]interface{})
	if checkoutData == nil {
		return toolResultFromError(fmt.Errorf("missing checkout parameter")), nil
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
		return toolResultFromError(err), nil
	}

	return toolResultFromJSON(mcpCheckoutResponse(co, hash), nil), nil
}

func (s *Server) handleCompleteCheckout(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)
	userCountry := userCountryFromContext(ctx)

	id, _ := args["id"].(string)

	approval, _ := args["approval"].(map[string]interface{})
	if approval == nil {
		return toolResultFromError(fmt.Errorf("approval is required: the platform must present the checkout to the user for approval before completing")), nil
	}
	approvalHash, _ := approval["checkout_hash"].(string)
	if approvalHash == "" {
		return toolResultFromError(fmt.Errorf("approval.checkout_hash is required")), nil
	}

	req := &model.CheckoutRequest{}

	co, _, hash, err := s.merchant.CompleteCheckout(id, userID, userCountry, approvalHash, req)
	if err != nil {
		return toolResultFromError(err), nil
	}

	return toolResultFromJSON(mcpCheckoutResponse(co, hash), nil), nil
}

func (s *Server) handleCancelCheckout(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)

	id, _ := args["id"].(string)

	co, hash, err := s.merchant.CancelCheckout(id, userID)
	if err != nil {
		return toolResultFromError(err), nil
	}

	return toolResultFromJSON(mcpCheckoutResponse(co, hash), nil), nil
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
