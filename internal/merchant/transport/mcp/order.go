package mcp

import "fmt"

func (s *Server) handleGetOrder(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	id, _ := args["id"].(string)

	ord, err := s.merchant.GetOrder(id, userID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	return ord, nil
}

func (s *Server) handleListOrders(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	orders, err := s.merchant.ListOrders(userID)
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

	return map[string]interface{}{"orders": summaries}, nil
}

func (s *Server) handleCancelOrder(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	id, _ := args["id"].(string)

	err := s.merchant.CancelOrder(id, userID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %s", id)
	}

	return map[string]interface{}{"id": id, "status": "canceled", "message": "Order has been canceled"}, nil
}
