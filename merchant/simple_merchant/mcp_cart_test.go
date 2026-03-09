package main

import (
	"testing"
)

func TestMCP_CreateCart(t *testing.T) {
	ts := newTestServer(t)
	result, isErr := ts.mcpToolCall("create_cart", map[string]interface{}{
		"cart": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 2,
				},
			},
		},
	}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	if result["id"] == nil {
		t.Error("expected cart id")
	}
	if result["line_items"] == nil {
		t.Error("expected line_items")
	}
	if result["totals"] == nil {
		t.Error("expected totals")
	}
}

func TestMCP_GetCart(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_cart", map[string]interface{}{
		"cart": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	cartID := createResult["id"].(string)

	result, isErr := ts.mcpToolCall("get_cart", map[string]interface{}{"id": cartID}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	if result["id"] != cartID {
		t.Errorf("expected cart id %s, got %v", cartID, result["id"])
	}
}

func TestMCP_UpdateCart(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_cart", map[string]interface{}{
		"cart": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	cartID := createResult["id"].(string)

	result, isErr := ts.mcpToolCall("update_cart", map[string]interface{}{
		"id": cartID,
		"cart": map[string]interface{}{
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
	lineItems := result["line_items"].([]interface{})
	if len(lineItems) == 0 {
		t.Fatal("expected line items")
	}
	firstItem := lineItems[0].(map[string]interface{})
	qty := firstItem["quantity"].(float64)
	if qty != 3 {
		t.Errorf("expected quantity 3, got %v", qty)
	}
}

func TestMCP_CancelCart(t *testing.T) {
	ts := newTestServer(t)
	createResult, _ := ts.mcpToolCall("create_cart", map[string]interface{}{
		"cart": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, "")
	cartID := createResult["id"].(string)

	result, isErr := ts.mcpToolCall("cancel_cart", map[string]interface{}{"id": cartID}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}

	// Cart should be gone
	_, isErr = ts.mcpToolCall("get_cart", map[string]interface{}{"id": cartID}, "")
	if !isErr {
		t.Error("expected error when getting canceled cart")
	}
}

func TestMCP_CartOwnership(t *testing.T) {
	ts := newTestServer(t)
	tokenA := ts.injectToken("cart_user_a", "US")
	tokenB := ts.injectToken("cart_user_b", "US")

	createResult, _ := ts.mcpToolCall("create_cart", map[string]interface{}{
		"cart": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, tokenA)
	cartID := createResult["id"].(string)

	// User A can access
	_, isErr := ts.mcpToolCall("get_cart", map[string]interface{}{"id": cartID}, tokenA)
	if isErr {
		t.Error("expected User A to access own cart")
	}

	// User B cannot access
	_, isErr = ts.mcpToolCall("get_cart", map[string]interface{}{"id": cartID}, tokenB)
	if !isErr {
		t.Error("expected error when User B accesses User A's cart")
	}
}
