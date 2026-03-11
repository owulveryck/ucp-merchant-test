package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleGetOrder(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)

	id, _ := args["id"].(string)

	ord, err := s.merchant.GetOrder(id, userID)
	if err != nil {
		return toolResultFromError(fmt.Errorf("order not found: %s", id)), nil
	}

	return toolResultFromJSON(ord, nil), nil
}

func (s *Server) handleListOrders(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := userIDFromContext(ctx)

	orders, err := s.merchant.ListOrders(userID)
	if err != nil {
		return toolResultFromError(err), nil
	}

	type orderSummary struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		CheckoutID   string `json:"checkout_id"`
		PermalinkURL string `json:"permalink_url"`
		Total        string `json:"total"`
	}
	var summaries []orderSummary
	for _, ord := range orders {
		totalText := ""
		for _, t := range ord.Totals {
			if t.Type == "total" {
				totalText = t.DisplayText
			}
		}
		summaries = append(summaries, orderSummary{
			ID:           ord.ID,
			Status:       "confirmed",
			CheckoutID:   ord.CheckoutID,
			PermalinkURL: ord.PermalinkURL,
			Total:        totalText,
		})
	}

	return toolResultFromJSON(map[string]interface{}{"orders": summaries}, nil), nil
}

func (s *Server) handleCancelOrder(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userID := userIDFromContext(ctx)

	id, _ := args["id"].(string)

	err := s.merchant.CancelOrder(id, userID)
	if err != nil {
		return toolResultFromError(fmt.Errorf("order not found: %s", id)), nil
	}

	return toolResultFromJSON(map[string]interface{}{"id": id, "status": "canceled", "message": "Order has been canceled"}, nil), nil
}
