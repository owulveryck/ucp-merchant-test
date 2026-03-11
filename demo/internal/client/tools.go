package client

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

func defineTools() []*genai.Tool {
	return []*genai.Tool{{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        "search_products",
				Description: "Search for products across all merchants via the Shopping Graph",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"query": {Type: genai.TypeString, Description: "Search query (e.g. 'wireless headphones')"},
						"limit": {Type: genai.TypeInteger, Description: "Max results (default 10)"},
					},
					Required: []string{"query"},
				},
			},
			{
				Name:        "get_product_details",
				Description: "Get detailed product information from a specific merchant",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"merchant_url": {Type: genai.TypeString, Description: "Merchant base URL"},
						"product_id":   {Type: genai.TypeString, Description: "Product ID"},
					},
					Required: []string{"merchant_url", "product_id"},
				},
			},
			{
				Name:        "create_checkout",
				Description: "Create a checkout session at a merchant with line items",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"merchant_url": {Type: genai.TypeString, Description: "Merchant base URL"},
						"product_id":   {Type: genai.TypeString, Description: "Product ID to purchase"},
						"quantity":     {Type: genai.TypeInteger, Description: "Quantity (default 1)"},
					},
					Required: []string{"merchant_url", "product_id"},
				},
			},
			{
				Name:        "apply_discount_codes",
				Description: "Apply discount codes to an active checkout",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"merchant_url":   {Type: genai.TypeString, Description: "Merchant base URL"},
						"checkout_id":    {Type: genai.TypeString, Description: "Checkout session ID"},
						"discount_codes": {Type: genai.TypeString, Description: "Comma-separated discount codes"},
					},
					Required: []string{"merchant_url", "checkout_id", "discount_codes"},
				},
			},
			{
				Name:        "update_checkout",
				Description: "Set buyer information and fulfillment address on a checkout",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"merchant_url": {Type: genai.TypeString, Description: "Merchant base URL"},
						"checkout_id":  {Type: genai.TypeString, Description: "Checkout session ID"},
						"email":        {Type: genai.TypeString, Description: "Buyer email"},
						"first_name":   {Type: genai.TypeString, Description: "Buyer first name"},
						"last_name":    {Type: genai.TypeString, Description: "Buyer last name"},
					},
					Required: []string{"merchant_url", "checkout_id", "email"},
				},
			},
			{
				Name:        "get_checkout_summary",
				Description: "Get current checkout totals for comparison",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"merchant_url": {Type: genai.TypeString, Description: "Merchant base URL"},
						"checkout_id":  {Type: genai.TypeString, Description: "Checkout session ID"},
					},
					Required: []string{"merchant_url", "checkout_id"},
				},
			},
			{
				Name:        "complete_checkout",
				Description: "Complete checkout and place the order using a payment token",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"merchant_url": {Type: genai.TypeString, Description: "Merchant base URL"},
						"checkout_id":  {Type: genai.TypeString, Description: "Checkout session ID"},
						"handler_id":   {Type: genai.TypeString, Description: "Payment handler ID (e.g. mock_payment_handler)"},
						"token":        {Type: genai.TypeString, Description: "Payment token (e.g. success_token)"},
					},
					Required: []string{"merchant_url", "checkout_id", "handler_id", "token"},
				},
			},
			{
				Name:        "cancel_checkout",
				Description: "Cancel a checkout session",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"merchant_url": {Type: genai.TypeString, Description: "Merchant base URL"},
						"checkout_id":  {Type: genai.TypeString, Description: "Checkout session ID"},
					},
					Required: []string{"merchant_url", "checkout_id"},
				},
			},
		},
	}}
}

func (a *Agent) executeTool(name string, args map[string]any) (string, error) {
	switch name {
	case "search_products":
		return a.toolSearchProducts(args)
	case "get_product_details":
		return a.toolGetProductDetails(args)
	case "create_checkout":
		return a.toolCreateCheckout(args)
	case "apply_discount_codes":
		return a.toolApplyDiscountCodes(args)
	case "update_checkout":
		return a.toolUpdateCheckout(args)
	case "get_checkout_summary":
		return a.toolGetCheckoutSummary(args)
	case "complete_checkout":
		return a.toolCompleteCheckout(args)
	case "cancel_checkout":
		return a.toolCancelCheckout(args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *Agent) toolSearchProducts(args map[string]any) (string, error) {
	query, _ := args["query"].(string)
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	a.emitEvent("tool_call", fmt.Sprintf("Searching for: %s", query))

	resp, err := a.searchGraph(query, limit)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(resp)
	return string(data), nil
}

func (a *Agent) toolGetProductDetails(args map[string]any) (string, error) {
	merchantURL, _ := args["merchant_url"].(string)
	productID, _ := args["product_id"].(string)

	a.emitEvent("tool_call", fmt.Sprintf("Getting details for %s at %s", productID, merchantURL))

	result, err := a.a2aClient.SendAction(merchantURL, "get_product_details", map[string]any{
		"id": productID,
	})
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (a *Agent) toolCreateCheckout(args map[string]any) (string, error) {
	merchantURL, _ := args["merchant_url"].(string)
	productID, _ := args["product_id"].(string)
	quantity := 1
	if q, ok := args["quantity"].(float64); ok {
		quantity = int(q)
	}

	a.emitEvent("tool_call", fmt.Sprintf("Creating checkout for %s (qty:%d) at %s", productID, quantity, merchantURL))

	result, err := a.a2aClient.SendAction(merchantURL, "create_checkout", map[string]any{
		"line_items": []any{
			map[string]any{
				"product_id": productID,
				"quantity":   float64(quantity),
			},
		},
	})
	if err != nil {
		return "", err
	}

	// Extract checkout ID from response
	if checkout, ok := result["a2a.ucp.checkout"].(map[string]any); ok {
		data, _ := json.Marshal(checkout)
		return string(data), nil
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (a *Agent) toolApplyDiscountCodes(args map[string]any) (string, error) {
	merchantURL, _ := args["merchant_url"].(string)
	checkoutID, _ := args["checkout_id"].(string)
	codesStr, _ := args["discount_codes"].(string)

	codes := strings.Split(codesStr, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}

	a.emitEvent("tool_call", fmt.Sprintf("Applying codes %v to checkout %s", codes, checkoutID))

	result, err := a.a2aClient.SendAction(merchantURL, "update_checkout", map[string]any{
		"id":             checkoutID,
		"discount_codes": codes,
	})
	if err != nil {
		return "", err
	}

	if checkout, ok := result["a2a.ucp.checkout"].(map[string]any); ok {
		data, _ := json.Marshal(checkout)
		return string(data), nil
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (a *Agent) toolUpdateCheckout(args map[string]any) (string, error) {
	merchantURL, _ := args["merchant_url"].(string)
	checkoutID, _ := args["checkout_id"].(string)
	email, _ := args["email"].(string)
	firstName, _ := args["first_name"].(string)
	lastName, _ := args["last_name"].(string)

	a.emitEvent("tool_call", fmt.Sprintf("Updating checkout %s with buyer %s", checkoutID, email))

	updateData := map[string]any{
		"id": checkoutID,
		"buyer": map[string]any{
			"email":      email,
			"first_name": firstName,
			"last_name":  lastName,
		},
	}

	result, err := a.a2aClient.SendAction(merchantURL, "update_checkout", updateData)
	if err != nil {
		return "", err
	}

	if checkout, ok := result["a2a.ucp.checkout"].(map[string]any); ok {
		data, _ := json.Marshal(checkout)
		return string(data), nil
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (a *Agent) toolGetCheckoutSummary(args map[string]any) (string, error) {
	merchantURL, _ := args["merchant_url"].(string)
	checkoutID, _ := args["checkout_id"].(string)

	a.emitEvent("tool_call", fmt.Sprintf("Getting checkout summary %s", checkoutID))

	result, err := a.a2aClient.SendAction(merchantURL, "get_checkout", map[string]any{
		"id": checkoutID,
	})
	if err != nil {
		return "", err
	}

	if checkout, ok := result["a2a.ucp.checkout"].(map[string]any); ok {
		data, _ := json.Marshal(checkout)
		return string(data), nil
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (a *Agent) toolCompleteCheckout(args map[string]any) (string, error) {
	merchantURL, _ := args["merchant_url"].(string)
	checkoutID, _ := args["checkout_id"].(string)
	handlerID, _ := args["handler_id"].(string)
	token, _ := args["token"].(string)

	a.emitEvent("tool_call", fmt.Sprintf("Completing checkout %s", checkoutID))

	payment := map[string]any{
		"handler_id": handlerID,
		"token":      token,
	}

	result, err := a.a2aClient.SendActionWithPayment(merchantURL, "complete_checkout", map[string]any{
		"id": checkoutID,
	}, payment)
	if err != nil {
		return "", err
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}

func (a *Agent) toolCancelCheckout(args map[string]any) (string, error) {
	merchantURL, _ := args["merchant_url"].(string)
	checkoutID, _ := args["checkout_id"].(string)

	a.emitEvent("tool_call", fmt.Sprintf("Cancelling checkout %s", checkoutID))

	result, err := a.a2aClient.SendAction(merchantURL, "cancel_checkout", map[string]any{
		"id": checkoutID,
	})
	if err != nil {
		return "", err
	}

	if checkout, ok := result["a2a.ucp.checkout"].(map[string]any); ok {
		data, _ := json.Marshal(checkout)
		return string(data), nil
	}
	data, _ := json.Marshal(result)
	return string(data), nil
}
