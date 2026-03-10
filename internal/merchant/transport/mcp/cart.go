package mcp

import "fmt"

func (s *Server) handleCreateCart(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	cartData, _ := args["cart"].(map[string]interface{})
	if cartData == nil {
		return nil, fmt.Errorf("missing cart parameter")
	}

	cartLineItems := parseLineItemRequests(cartData)
	if len(cartLineItems) == 0 {
		return nil, fmt.Errorf("cart must have at least one line item")
	}

	cart, err := s.merchant.CreateCart(userID, cartLineItems)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *Server) handleGetCart(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	id, _ := args["id"].(string)
	cart, err := s.merchant.GetCart(id, userID)
	if err != nil {
		return nil, fmt.Errorf("cart not found: %s", id)
	}
	return cart, nil
}

func (s *Server) handleUpdateCart(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	id, _ := args["id"].(string)

	cartData, _ := args["cart"].(map[string]interface{})
	if cartData == nil {
		return nil, fmt.Errorf("missing cart parameter")
	}
	cartLineItems := parseLineItemRequests(cartData)

	cart, err := s.merchant.UpdateCart(id, userID, cartLineItems)
	if err != nil {
		return nil, fmt.Errorf("cart not found: %s", id)
	}
	return cart, nil
}

func (s *Server) handleCancelCart(args map[string]interface{}, userID, userCountry string) (interface{}, error) {
	id, _ := args["id"].(string)
	cart, err := s.merchant.CancelCart(id, userID)
	if err != nil {
		return nil, fmt.Errorf("cart not found: %s", id)
	}
	return cart, nil
}
