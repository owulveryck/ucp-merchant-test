package a2a

import (
	"context"
	"fmt"
)

func (s *Server) handleGetOrder(_ context.Context, ac *actionContext) (map[string]any, error) {
	id, _ := ac.data["id"].(string)

	ord, err := s.merchant.GetOrder(id, ac.userID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	return toMap(ord)
}

func (s *Server) handleListOrders(_ context.Context, ac *actionContext) (map[string]any, error) {
	orders, err := s.merchant.ListOrders(ac.userID)
	if err != nil {
		return nil, err
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

	return map[string]any{"orders": summaries}, nil
}

func (s *Server) handleCancelOrder(_ context.Context, ac *actionContext) (map[string]any, error) {
	id, _ := ac.data["id"].(string)

	err := s.merchant.CancelOrder(id, ac.userID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	return map[string]any{
		"id":      id,
		"status":  "canceled",
		"message": "Order has been canceled",
	}, nil
}
