package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	gomcp "github.com/mark3labs/mcp-go/mcp"

	"github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/merchanttest"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

func newTestMCPServer(mock *merchanttest.Mock) *Server {
	return &Server{merchant: mock}
}

func makeToolRequest(args map[string]interface{}) gomcp.CallToolRequest {
	raw, _ := json.Marshal(args)
	var req gomcp.CallToolRequest
	req.Params.Arguments = make(map[string]interface{})
	json.Unmarshal(raw, &req.Params.Arguments)
	// Direct assignment is simpler
	req.Params.Arguments = args
	return req
}

func TestHandleListProducts_Default(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.FilterFunc = func(category ucp.Category, brand, query string, country ucp.Country, currency ucp.Currency, language ucp.Language) []catalog.Product {
		return []catalog.Product{
			{ID: "SKU-001", Title: "Roses", Price: 2999},
			{ID: "SKU-002", Title: "Tulips", Price: 1999},
		}
	}
	mock.CategoryCountFunc = func() []catalog.CategoryStat {
		return []catalog.CategoryStat{{Name: "Flowers", Count: 2}}
	}

	s := newTestMCPServer(mock)
	result, err := s.handleListProducts(context.Background(), makeToolRequest(map[string]interface{}{}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}

	text := extractText(t, result)
	if !strings.Contains(text, "SKU-001") || !strings.Contains(text, "SKU-002") {
		t.Error("expected both products in response")
	}
	if !strings.Contains(text, "pagination") {
		t.Error("expected pagination in response")
	}
}

func TestHandleListProducts_WithFilters(t *testing.T) {
	mock := merchanttest.NewMock()
	var gotCategory ucp.Category
	var gotBrand string
	mock.FilterFunc = func(category ucp.Category, brand, query string, country ucp.Country, currency ucp.Currency, language ucp.Language) []catalog.Product {
		gotCategory = category
		gotBrand = brand
		return nil
	}
	mock.CategoryCountFunc = func() []catalog.CategoryStat { return nil }

	s := newTestMCPServer(mock)
	s.handleListProducts(context.Background(), makeToolRequest(map[string]interface{}{
		"category": "Fresh Flowers",
		"brand":    "Rose Garden",
	}))

	if string(gotCategory) != "Fresh Flowers" {
		t.Errorf("expected category Fresh Flowers, got %s", gotCategory)
	}
	if gotBrand != "Rose Garden" {
		t.Errorf("expected brand Rose Garden, got %s", gotBrand)
	}
}

func TestHandleListProducts_Pagination(t *testing.T) {
	mock := merchanttest.NewMock()
	products := make([]catalog.Product, 5)
	for i := range products {
		products[i] = catalog.Product{ID: "SKU", Title: string(rune('A' + i))}
	}
	mock.FilterFunc = func(ucp.Category, string, string, ucp.Country, ucp.Currency, ucp.Language) []catalog.Product {
		return products
	}
	mock.CategoryCountFunc = func() []catalog.CategoryStat { return nil }

	s := newTestMCPServer(mock)
	result, _ := s.handleListProducts(context.Background(), makeToolRequest(map[string]interface{}{
		"limit":  float64(2),
		"offset": float64(1),
	}))

	text := extractText(t, result)
	if !strings.Contains(text, `"has_more": true`) {
		t.Error("expected has_more=true")
	}
}

func TestHandleGetProductDetails_Found(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.FindFunc = func(id string) *catalog.Product {
		if id == "SKU-001" {
			return &catalog.Product{ID: "SKU-001", Title: "Roses", Price: 2999, Brand: "Test"}
		}
		return nil
	}

	s := newTestMCPServer(mock)
	result, err := s.handleGetProductDetails(context.Background(), makeToolRequest(map[string]interface{}{
		"id": "SKU-001",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
	text := extractText(t, result)
	if !strings.Contains(text, "SKU-001") {
		t.Error("expected product ID in response")
	}
}

func TestHandleGetProductDetails_NotFound(t *testing.T) {
	mock := merchanttest.NewMock()
	s := newTestMCPServer(mock)

	result, err := s.handleGetProductDetails(context.Background(), makeToolRequest(map[string]interface{}{
		"id": "NONEXISTENT",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for not found product")
	}
}

func TestHandleSearchCatalog(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.SearchFunc = func(params catalog.SearchParams) []catalog.SearchResult {
		if params.Query != "roses" {
			t.Errorf("expected query 'roses', got %q", params.Query)
		}
		return []catalog.SearchResult{
			{Product: catalog.Product{ID: "SKU-001", Title: "Red Roses"}},
		}
	}

	s := newTestMCPServer(mock)
	result, err := s.handleSearchCatalog(context.Background(), makeToolRequest(map[string]interface{}{
		"query": "roses",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
	text := extractText(t, result)
	if !strings.Contains(text, "SKU-001") {
		t.Error("expected search result in response")
	}
}

func TestHandleLookupProduct(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.LookupFunc = func(id string, shipsTo ucp.Country) *catalog.Product {
		return &catalog.Product{ID: id, Title: "Roses"}
	}

	s := newTestMCPServer(mock)
	result, err := s.handleLookupProduct(context.Background(), makeToolRequest(map[string]interface{}{
		"id": "SKU-001",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Error("expected no error")
	}
}

func extractText(t *testing.T, result *gomcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no content in result")
	}
	b, _ := json.Marshal(result.Content[0])
	var tc struct {
		Text string `json:"text"`
	}
	json.Unmarshal(b, &tc)
	return tc.Text
}
