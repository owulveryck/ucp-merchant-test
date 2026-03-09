package main

import (
	"testing"
)

func TestMCP_CreateCheckout(t *testing.T) {
	ts := newTestServer(t)
	result, isErr := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	if result["id"] == nil {
		t.Error("expected checkout id")
	}
	if result["checkout_hash"] == nil || result["checkout_hash"] == "" {
		t.Error("expected checkout_hash")
	}
	if result["status"] != "incomplete" {
		t.Errorf("expected status incomplete, got %v", result["status"])
	}
}

func TestMCP_CreateCheckoutFromCart(t *testing.T) {
	ts := newTestServer(t)
	// Create a cart first
	cartResult, _ := ts.mcpToolCall("create_cart", map[string]interface{}{
		"cart": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	cartID := cartResult["id"].(string)

	// Create checkout from cart
	result, isErr := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"cart_id": cartID,
		},
	}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	if result["id"] == nil {
		t.Error("expected checkout id from cart")
	}
	if result["line_items"] == nil {
		t.Error("expected line_items inherited from cart")
	}
}

func TestMCP_GetCheckout(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	checkoutID := createResult["id"].(string)

	result, isErr := ts.mcpToolCall("get_checkout", map[string]interface{}{"id": checkoutID}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	if result["id"] != checkoutID {
		t.Errorf("expected id %s, got %v", checkoutID, result["id"])
	}
	if result["checkout_hash"] == nil {
		t.Error("expected checkout_hash in get response")
	}
}

func TestMCP_UpdateCheckoutLineItems(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	checkoutID := createResult["id"].(string)
	originalHash := createResult["checkout_hash"].(string)

	result, isErr := ts.mcpToolCall("update_checkout", map[string]interface{}{
		"id": checkoutID,
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 3,
				},
			},
		},
	}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	newHash := result["checkout_hash"].(string)
	if newHash == originalHash {
		t.Error("expected checkout_hash to change after update")
	}
}

func TestMCP_UpdateCheckoutBuyer(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	checkoutID := createResult["id"].(string)

	result, isErr := ts.mcpToolCall("update_checkout", map[string]interface{}{
		"id": checkoutID,
		"checkout": map[string]interface{}{
			"buyer": map[string]interface{}{
				"fullName": "John Doe",
				"email":    "john@example.com",
			},
		},
	}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	buyer, ok := result["buyer"].(map[string]interface{})
	if !ok {
		t.Fatal("expected buyer in response")
	}
	if buyer["fullName"] != "John Doe" {
		t.Errorf("expected buyer name John Doe, got %v", buyer["full_name"])
	}
}

func TestMCP_UpdateCheckoutShipping(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	checkoutID := createResult["id"].(string)

	result, isErr := ts.mcpToolCall("update_checkout", map[string]interface{}{
		"id": checkoutID,
		"checkout": map[string]interface{}{
			"shipping_option_id": "express",
		},
	}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	if result["selected_shipping"] == nil {
		t.Error("expected selected_shipping in response")
	}
	shipping := result["selected_shipping"].(map[string]interface{})
	if shipping["id"] != "express" {
		t.Errorf("expected shipping id express, got %v", shipping["id"])
	}
}

func TestMCP_GetShippingOptions(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	checkoutID := createResult["id"].(string)

	result, isErr := ts.mcpToolCall("get_shipping_options", map[string]interface{}{
		"checkout_id": checkoutID,
	}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	options, ok := result["options"].([]interface{})
	if !ok {
		t.Fatal("expected options array")
	}
	if len(options) != 3 {
		t.Errorf("expected 3 shipping options, got %d", len(options))
	}
}

func TestMCP_CompleteCheckout(t *testing.T) {
	ts := newTestServer(t)
	checkoutID, orderID := ts.mcpCreateAndCompleteCheckout("")

	if checkoutID == "" {
		t.Error("expected checkout ID")
	}
	if orderID == "" {
		t.Error("expected order ID")
	}
}

func TestMCP_CompleteCheckoutWrongHash(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	checkoutID := createResult["id"].(string)

	_, isErr := ts.mcpToolCall("complete_checkout", map[string]interface{}{
		"id": checkoutID,
		"approval": map[string]interface{}{
			"checkout_hash": "wrong_hash_value",
		},
	}, "")
	if !isErr {
		t.Error("expected error for wrong checkout hash")
	}
}

func TestMCP_CompleteCheckoutMissingApproval(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	checkoutID := createResult["id"].(string)

	_, isErr := ts.mcpToolCall("complete_checkout", map[string]interface{}{
		"id": checkoutID,
	}, "")
	if !isErr {
		t.Error("expected error for missing approval")
	}
}

func TestMCP_CompleteCheckoutStaleHash(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	checkoutID := createResult["id"].(string)
	staleHash := createResult["checkout_hash"].(string)

	// Update the checkout to invalidate the hash
	ts.mcpToolCall("update_checkout", map[string]interface{}{
		"id": checkoutID,
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 5,
				},
			},
		},
	}, "")

	// Try to complete with the stale hash
	_, isErr := ts.mcpToolCall("complete_checkout", map[string]interface{}{
		"id": checkoutID,
		"approval": map[string]interface{}{
			"checkout_hash": staleHash,
		},
	}, "")
	if !isErr {
		t.Error("expected error for stale checkout hash")
	}
}

func TestMCP_CancelCheckout(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	checkoutID := createResult["id"].(string)

	result, isErr := ts.mcpToolCall("cancel_checkout", map[string]interface{}{"id": checkoutID}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	if result["status"] != "canceled" {
		t.Errorf("expected status canceled, got %v", result["status"])
	}
}

func TestMCP_CancelCompletedCheckout(t *testing.T) {
	ts := newTestServer(t)
	checkoutID, _ := ts.mcpCreateAndCompleteCheckout("")

	_, isErr := ts.mcpToolCall("cancel_checkout", map[string]interface{}{"id": checkoutID}, "")
	if !isErr {
		t.Error("expected error when canceling completed checkout")
	}
}

func TestMCP_UpdateCompletedCheckout(t *testing.T) {
	ts := newTestServer(t)
	checkoutID, _ := ts.mcpCreateAndCompleteCheckout("")

	_, isErr := ts.mcpToolCall("update_checkout", map[string]interface{}{
		"id": checkoutID,
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 5,
				},
			},
		},
	}, "")
	if !isErr {
		t.Error("expected error when updating completed checkout")
	}
}
