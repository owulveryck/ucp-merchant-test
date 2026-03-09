package main

import (
	"testing"
)

func TestMCP_GuestAccess(t *testing.T) {
	ts := newTestServer(t)
	// Guest (no token) can list products
	result, isErr := ts.mcpToolCall("list_products", map[string]interface{}{}, "")
	if isErr {
		t.Fatalf("expected success, got error: %v", result)
	}
	if result["products"] == nil {
		t.Error("expected products in response")
	}
}

func TestMCP_AuthenticatedAccess(t *testing.T) {
	ts := newTestServer(t)
	token := ts.injectToken("user_alice", "US")

	// Authenticated user can list products
	result, isErr := ts.mcpToolCall("list_products", map[string]interface{}{}, token)
	if isErr {
		t.Fatalf("expected success, got error: %v", result)
	}
	if result["products"] == nil {
		t.Error("expected products in response")
	}

	// Create a checkout scoped to this user
	createResult, isErr := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, token)
	if isErr {
		t.Fatalf("create_checkout failed: %v", createResult)
	}
	if createResult["id"] == nil {
		t.Error("expected checkout id")
	}
}

func TestMCP_ExpiredToken(t *testing.T) {
	ts := newTestServer(t)
	token := ts.injectExpiredToken("user_expired", "US")

	resp, _ := ts.mcpRequest("tools/call", map[string]interface{}{
		"name":      "list_products",
		"arguments": map[string]interface{}{},
	}, token)

	if resp.StatusCode != 401 {
		t.Fatalf("expected 401 for expired token, got %d", resp.StatusCode)
	}
}

func TestMCP_UserScopingCheckout(t *testing.T) {
	ts := newTestServer(t)
	tokenA := ts.injectToken("user_a", "US")
	tokenB := ts.injectToken("user_b", "US")

	// User A creates a checkout
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, tokenA)
	checkoutID := createResult["id"].(string)

	// User B cannot get User A's checkout
	_, isErr := ts.mcpToolCall("get_checkout", map[string]interface{}{"id": checkoutID}, tokenB)
	if !isErr {
		t.Error("expected error when User B accesses User A's checkout")
	}

	// User B cannot update User A's checkout
	_, isErr = ts.mcpToolCall("update_checkout", map[string]interface{}{
		"id":       checkoutID,
		"checkout": map[string]interface{}{"shipping_option_id": "standard"},
	}, tokenB)
	if !isErr {
		t.Error("expected error when User B updates User A's checkout")
	}

	// User B cannot complete User A's checkout
	_, isErr = ts.mcpToolCall("complete_checkout", map[string]interface{}{
		"id":       checkoutID,
		"approval": map[string]interface{}{"checkout_hash": "fake"},
	}, tokenB)
	if !isErr {
		t.Error("expected error when User B completes User A's checkout")
	}
}

func TestMCP_UserScopingCart(t *testing.T) {
	ts := newTestServer(t)
	tokenA := ts.injectToken("user_a", "US")
	tokenB := ts.injectToken("user_b", "US")

	// User A creates a cart
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

	// User B cannot get User A's cart
	_, isErr := ts.mcpToolCall("get_cart", map[string]interface{}{"id": cartID}, tokenB)
	if !isErr {
		t.Error("expected error when User B accesses User A's cart")
	}

	// User B cannot update User A's cart
	_, isErr = ts.mcpToolCall("update_cart", map[string]interface{}{
		"id": cartID,
		"cart": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 2,
				},
			},
		},
	}, tokenB)
	if !isErr {
		t.Error("expected error when User B updates User A's cart")
	}
}

func TestMCP_UserScopingOrder(t *testing.T) {
	ts := newTestServer(t)
	tokenA := ts.injectToken("user_a", "US")
	tokenB := ts.injectToken("user_b", "US")

	// User A creates and completes a checkout
	_, orderID := ts.mcpCreateAndCompleteCheckout(tokenA)

	// User B cannot get User A's order
	_, isErr := ts.mcpToolCall("get_order", map[string]interface{}{"id": orderID}, tokenB)
	if !isErr {
		t.Error("expected error when User B accesses User A's order")
	}
}

func TestMCP_GuestCannotAccessUserEntities(t *testing.T) {
	ts := newTestServer(t)
	token := ts.injectToken("user_private", "US")

	// Authenticated user creates a checkout
	createResult, _ := ts.mcpToolCall("create_checkout", map[string]interface{}{
		"checkout": map[string]interface{}{
			"line_items": []interface{}{
				map[string]interface{}{
					"item":     map[string]interface{}{"id": "bouquet_roses", "title": "Bouquet of Roses"},
					"quantity": 1,
				},
			},
		},
	}, token)
	checkoutID := createResult["id"].(string)

	// Guest cannot access authenticated user's checkout
	_, isErr := ts.mcpToolCall("get_checkout", map[string]interface{}{"id": checkoutID}, "")
	if !isErr {
		t.Error("expected error when guest accesses authenticated user's checkout")
	}
}
