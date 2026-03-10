package main

import (
	"testing"
)

func TestMCP_GetOrder(t *testing.T) {
	ts := newTestServer(t)
	_, orderID := ts.mcpCreateAndCompleteCheckout("")

	result, isErr := ts.mcpToolCall("get_order", map[string]interface{}{"id": orderID}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	if result["id"] != orderID {
		t.Errorf("expected order id %s, got %v", orderID, result["id"])
	}
}

func TestMCP_ListOrders(t *testing.T) {
	ts := newTestServer(t)
	token := ts.injectToken("order_list_user", "US")

	// Create two orders
	ts.mcpCreateAndCompleteCheckout(token)
	ts.mcpCreateAndCompleteCheckout(token)

	result, isErr := ts.mcpToolCall("list_orders", map[string]interface{}{}, token)
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	orders, ok := result["orders"].([]interface{})
	if !ok {
		t.Fatal("expected orders array")
	}
	if len(orders) != 2 {
		t.Errorf("expected 2 orders, got %d", len(orders))
	}
}

func TestMCP_CancelOrder(t *testing.T) {
	ts := newTestServer(t)
	_, orderID := ts.mcpCreateAndCompleteCheckout("")

	result, isErr := ts.mcpToolCall("cancel_order", map[string]interface{}{"id": orderID}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	if result["status"] != "canceled" {
		t.Errorf("expected status canceled, got %v", result["status"])
	}
}

func TestMCP_OrderOwnership(t *testing.T) {
	ts := newTestServer(t)
	tokenA := ts.injectToken("order_owner_a", "US")
	tokenB := ts.injectToken("order_owner_b", "US")

	_, orderID := ts.mcpCreateAndCompleteCheckout(tokenA)

	// User A can access
	_, isErr := ts.mcpToolCall("get_order", map[string]interface{}{"id": orderID}, tokenA)
	if isErr {
		t.Error("expected User A to access own order")
	}

	// User B cannot access
	_, isErr = ts.mcpToolCall("get_order", map[string]interface{}{"id": orderID}, tokenB)
	if !isErr {
		t.Error("expected error when User B accesses User A's order")
	}
}
