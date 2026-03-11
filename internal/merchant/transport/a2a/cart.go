package a2a

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *Server) handleCreateCart(_ context.Context, ac *actionContext) (map[string]any, error) {
	cartData, _ := ac.data["cart"].(map[string]any)
	if cartData == nil {
		cartData = ac.data
	}

	cartLineItems := parseLineItemRequests(cartData)
	if len(cartLineItems) == 0 {
		return nil, fmt.Errorf("cart must have at least one line item")
	}

	cart, err := s.merchant.CreateCart(ac.userID, cartLineItems)
	if err != nil {
		return nil, err
	}
	return toMap(cart)
}

func (s *Server) handleGetCart(_ context.Context, ac *actionContext) (map[string]any, error) {
	id, _ := ac.data["id"].(string)
	cart, err := s.merchant.GetCart(id, ac.userID)
	if err != nil {
		return nil, fmt.Errorf("cart not found: %s", id)
	}
	return toMap(cart)
}

func (s *Server) handleUpdateCart(_ context.Context, ac *actionContext) (map[string]any, error) {
	id, _ := ac.data["id"].(string)

	cartData, _ := ac.data["cart"].(map[string]any)
	if cartData == nil {
		cartData = ac.data
	}
	cartLineItems := parseLineItemRequests(cartData)

	cart, err := s.merchant.UpdateCart(id, ac.userID, cartLineItems)
	if err != nil {
		return nil, fmt.Errorf("cart not found: %s", id)
	}
	return toMap(cart)
}

func (s *Server) handleCancelCart(_ context.Context, ac *actionContext) (map[string]any, error) {
	id, _ := ac.data["id"].(string)
	cart, err := s.merchant.CancelCart(id, ac.userID)
	if err != nil {
		return nil, fmt.Errorf("cart not found: %s", id)
	}
	return toMap(cart)
}

// toMap converts a struct to map[string]any via JSON round-trip.
func toMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	return m, err
}
