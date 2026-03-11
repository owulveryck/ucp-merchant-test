package a2a

import (
	"context"
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/model"
	"github.com/owulveryck/ucp-merchant-test/pkg/ucp"
)

func (s *Server) handleCreateCheckout(_ context.Context, ac *actionContext) (map[string]any, error) {
	checkoutData, _ := ac.data["checkout"].(map[string]any)
	if checkoutData == nil {
		checkoutData = ac.data
	}

	req := &model.CheckoutRequest{
		Currency: ucp.Currency("USD"),
	}

	if cartID, ok := checkoutData["cart_id"].(string); ok && cartID != "" {
		cart, err := s.merchant.GetCart(cartID, ac.userID)
		if err != nil {
			return nil, fmt.Errorf("cart not found: %s", cartID)
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
			return nil, fmt.Errorf("checkout must have line_items or cart_id")
		}
		req.LineItems = coLineItems
	}

	if buyerData, ok := checkoutData["buyer"].(map[string]any); ok {
		req.Buyer = parseBuyerRequest(buyerData)
	}

	co, hash, err := s.merchant.CreateCheckout(ac.userID, ac.country, req)
	if err != nil {
		return nil, err
	}

	s.setSessionCheckout(ac.contextID, co.ID)

	return map[string]any{
		"a2a.ucp.checkout": a2aCheckoutResponse(co, hash),
	}, nil
}

func (s *Server) handleGetCheckout(_ context.Context, ac *actionContext) (map[string]any, error) {
	id := s.resolveCheckoutID(ac)
	if id == "" {
		return nil, fmt.Errorf("no checkout ID available")
	}

	co, hash, err := s.merchant.GetCheckout(id, ac.userID)
	if err != nil {
		return nil, fmt.Errorf("checkout not found: %s", id)
	}

	return map[string]any{
		"a2a.ucp.checkout": a2aCheckoutResponse(co, hash),
	}, nil
}

func (s *Server) handleUpdateCheckout(_ context.Context, ac *actionContext) (map[string]any, error) {
	id := s.resolveCheckoutID(ac)
	if id == "" {
		return nil, fmt.Errorf("no checkout ID available")
	}

	checkoutData, _ := ac.data["checkout"].(map[string]any)
	if checkoutData == nil {
		checkoutData = ac.data
	}

	req := &model.CheckoutRequest{}

	coLineItems := parseLineItemRequests(checkoutData)
	if len(coLineItems) > 0 {
		req.LineItems = coLineItems
	}

	if buyerData, ok := checkoutData["buyer"].(map[string]any); ok {
		req.Buyer = parseBuyerRequest(buyerData)
	}

	if fulfillmentData, ok := checkoutData["fulfillment"].(map[string]any); ok {
		req.Fulfillment = parseFulfillmentRequest(fulfillmentData)
	}

	if discountCodes, ok := checkoutData["discount_codes"].([]any); ok {
		req.Discounts = parseDiscountCodes(discountCodes)
	}

	co, hash, err := s.merchant.UpdateCheckout(id, ac.userID, req)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"a2a.ucp.checkout": a2aCheckoutResponse(co, hash),
	}, nil
}

func (s *Server) handleCompleteCheckout(_ context.Context, ac *actionContext) (map[string]any, error) {
	id := s.resolveCheckoutID(ac)
	if id == "" {
		return nil, fmt.Errorf("no checkout ID available")
	}

	// For A2A, approval hash is optional (empty = skip validation, same as REST).
	approvalHash, _ := ac.data["checkout_hash"].(string)

	req := &model.CheckoutRequest{}

	// Extract payment from a2a.ucp.checkout.payment DataPart.
	payment := extractPaymentFromParts(ac.parts)
	if payment != nil {
		req.PaymentData = parsePaymentData(payment)
	}

	co, _, hash, err := s.merchant.CompleteCheckout(id, ac.userID, ac.country, approvalHash, req)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"a2a.ucp.checkout": a2aCheckoutResponse(co, hash),
	}, nil
}

func (s *Server) handleCancelCheckout(_ context.Context, ac *actionContext) (map[string]any, error) {
	id := s.resolveCheckoutID(ac)
	if id == "" {
		return nil, fmt.Errorf("no checkout ID available")
	}

	co, hash, err := s.merchant.CancelCheckout(id, ac.userID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"a2a.ucp.checkout": a2aCheckoutResponse(co, hash),
	}, nil
}

// a2aCheckoutResponse builds the checkout response map for the A2A DataPart.
func a2aCheckoutResponse(co *model.Checkout, hash string) map[string]any {
	resp := map[string]any{
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
		resp["order"] = map[string]any{
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

// parsePaymentData converts a raw payment map from the
// a2a.ucp.checkout.payment DataPart into a PaymentDataRequest.
func parsePaymentData(data map[string]any) *model.PaymentDataRequest {
	pd := &model.PaymentDataRequest{}
	if handlerID, ok := data["handler_id"].(string); ok {
		pd.HandlerID = handlerID
	}
	if cred, ok := data["credential"].(map[string]any); ok {
		pd.Credential = &model.PaymentCredential{}
		if token, ok := cred["token"].(string); ok {
			pd.Credential.Token = token
		}
	}
	// Handle instruments array (for AP2 mandates).
	if instruments, ok := data["instruments"].([]any); ok {
		for _, inst := range instruments {
			instMap, ok := inst.(map[string]any)
			if !ok {
				continue
			}
			if cred, ok := instMap["credential"].(map[string]any); ok {
				pd.Credential = &model.PaymentCredential{}
				if token, ok := cred["token"].(string); ok {
					pd.Credential.Token = token
				}
			}
			if hID, ok := instMap["handler_id"].(string); ok {
				pd.HandlerID = hID
			}
		}
	}
	return pd
}
