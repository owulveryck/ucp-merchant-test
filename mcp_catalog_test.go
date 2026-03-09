package main

import (
	"testing"
)

func TestMCP_ListProducts(t *testing.T) {
	ts := newTestServer(t)
	result, isErr := ts.mcpToolCall("list_products", map[string]interface{}{}, "")
	if isErr {
		t.Fatalf("expected success, got error: %v", result)
	}
	products, ok := result["products"].([]interface{})
	if !ok {
		t.Fatal("expected products array")
	}
	if len(products) == 0 {
		t.Error("expected at least one product")
	}
}

func TestMCP_ListProductsWithFilters(t *testing.T) {
	ts := newTestServer(t)

	// Filter by category
	result, isErr := ts.mcpToolCall("list_products", map[string]interface{}{
		"category": "bouquets",
	}, "")
	if isErr {
		t.Fatalf("expected success, got error: %v", result)
	}
	products := result["products"].([]interface{})
	for _, p := range products {
		pm := p.(map[string]interface{})
		if pm["category"] != "bouquets" {
			t.Errorf("expected category bouquets, got %v", pm["category"])
		}
	}

	// Filter by query
	result, isErr = ts.mcpToolCall("list_products", map[string]interface{}{
		"query": "rose",
	}, "")
	if isErr {
		t.Fatalf("query filter failed: %v", result)
	}
	if result["products"] == nil {
		t.Error("expected products in query result")
	}
}

func TestMCP_ListProductsPagination(t *testing.T) {
	ts := newTestServer(t)

	// Get first 2 products
	result, isErr := ts.mcpToolCall("list_products", map[string]interface{}{
		"limit":  2,
		"offset": 0,
	}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	products := result["products"].([]interface{})
	if len(products) != 2 {
		t.Errorf("expected 2 products, got %d", len(products))
	}
	pagination := result["pagination"].(map[string]interface{})
	if pagination["has_more"] != true {
		t.Error("expected has_more=true with limit=2")
	}

	// Get next page
	result2, isErr := ts.mcpToolCall("list_products", map[string]interface{}{
		"limit":  2,
		"offset": 2,
	}, "")
	if isErr {
		t.Fatalf("pagination offset failed: %v", result2)
	}
	products2 := result2["products"].([]interface{})
	if len(products2) == 0 {
		t.Error("expected products on second page")
	}
	// Ensure different products
	first := products[0].(map[string]interface{})["id"]
	second := products2[0].(map[string]interface{})["id"]
	if first == second {
		t.Error("expected different products on different pages")
	}
}

func TestMCP_GetProductDetails(t *testing.T) {
	ts := newTestServer(t)
	result, isErr := ts.mcpToolCall("get_product_details", map[string]interface{}{
		"id": "bouquet_roses",
	}, "")
	if isErr {
		t.Fatalf("expected success: %v", result)
	}
	if result["id"] != "bouquet_roses" {
		t.Errorf("expected id bouquet_roses, got %v", result["id"])
	}
	if result["title"] == nil || result["title"] == "" {
		t.Error("expected title in product details")
	}
	if result["price"] == nil {
		t.Error("expected price in product details")
	}
}

func TestMCP_GetProductDetailsNotFound(t *testing.T) {
	ts := newTestServer(t)
	_, isErr := ts.mcpToolCall("get_product_details", map[string]interface{}{
		"id": "nonexistent_product",
	}, "")
	if !isErr {
		t.Error("expected error for nonexistent product")
	}
}
