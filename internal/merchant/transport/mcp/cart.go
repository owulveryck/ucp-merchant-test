package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleCreateCart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)

	cartData, _ := args["cart"].(map[string]interface{})
	if cartData == nil {
		return toolResultFromError(fmt.Errorf("missing cart parameter")), nil
	}

	cartLineItems := parseLineItemRequests(cartData)
	if len(cartLineItems) == 0 {
		return toolResultFromError(fmt.Errorf("cart must have at least one line item")), nil
	}

	cart, err := s.merchant.CreateCart(userID, cartLineItems)
	if err != nil {
		return toolResultFromError(err), nil
	}
	return toolResultFromJSON(cart, nil), nil
}

func (s *Server) handleGetCart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)

	id, _ := args["id"].(string)
	cart, err := s.merchant.GetCart(id, userID)
	if err != nil {
		return toolResultFromError(fmt.Errorf("cart not found: %s", id)), nil
	}
	return toolResultFromJSON(cart, nil), nil
}

func (s *Server) handleUpdateCart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)

	id, _ := args["id"].(string)

	cartData, _ := args["cart"].(map[string]interface{})
	if cartData == nil {
		return toolResultFromError(fmt.Errorf("missing cart parameter")), nil
	}
	cartLineItems := parseLineItemRequests(cartData)

	cart, err := s.merchant.UpdateCart(id, userID, cartLineItems)
	if err != nil {
		return toolResultFromError(fmt.Errorf("cart not found: %s", id)), nil
	}
	return toolResultFromJSON(cart, nil), nil
}

func (s *Server) handleCancelCart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)

	id, _ := args["id"].(string)
	cart, err := s.merchant.CancelCart(id, userID)
	if err != nil {
		return toolResultFromError(fmt.Errorf("cart not found: %s", id)), nil
	}
	return toolResultFromJSON(cart, nil), nil
}
