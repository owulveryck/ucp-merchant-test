package a2a

import (
	"context"
	"testing"

	"github.com/owulveryck/ucp-merchant-test/pkg/auth"
	"github.com/owulveryck/ucp-merchant-test/pkg/catalog"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/merchanttest"
	"github.com/owulveryck/ucp-merchant-test/pkg/ucp"
)

func newTestServer(mock *merchanttest.Mock) *Server {
	authSrv := auth.NewOAuthServer("test", func() string { return "http" }, func() int { return 8080 })
	return New(mock, authSrv)
}

func TestHandleListProducts(t *testing.T) {
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

	s := newTestServer(mock)
	ac := &actionContext{data: map[string]any{}}
	result, err := s.handleListProducts(context.Background(), ac)
	if err != nil {
		t.Fatal(err)
	}

	products, ok := result["products"]
	if !ok {
		t.Fatal("expected products in result")
	}

	prods, ok := products.([]struct {
		ID                 string        `json:"id"`
		Title              string        `json:"title"`
		Category           ucp.Category  `json:"category"`
		Brand              string        `json:"brand"`
		Price              int           `json:"price"`
		Quantity           int           `json:"quantity"`
		ImageURL           string        `json:"image_url"`
		AvailableCountries []ucp.Country `json:"available_countries,omitempty"`
	})
	if !ok {
		// Type assertion may fail with anonymous struct; check length via pagination.
		pagination, ok := result["pagination"].(map[string]any)
		if !ok {
			t.Fatal("expected pagination in result")
		}
		if pagination["total"] != 2 {
			t.Errorf("expected total=2, got %v", pagination["total"])
		}
		return
	}

	if len(prods) != 2 {
		t.Errorf("expected 2 products, got %d", len(prods))
	}
}

func TestHandleGetProductDetails_Found(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.FindFunc = func(id string) *catalog.Product {
		if id == "SKU-001" {
			return &catalog.Product{ID: "SKU-001", Title: "Roses", Price: 2999}
		}
		return nil
	}

	s := newTestServer(mock)
	ac := &actionContext{data: map[string]any{"id": "SKU-001"}}
	result, err := s.handleGetProductDetails(context.Background(), ac)
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "SKU-001" {
		t.Errorf("expected id=SKU-001, got %v", result["id"])
	}
}

func TestHandleGetProductDetails_NotFound(t *testing.T) {
	mock := merchanttest.NewMock()
	s := newTestServer(mock)

	ac := &actionContext{data: map[string]any{"id": "NONEXISTENT"}}
	_, err := s.handleGetProductDetails(context.Background(), ac)
	if err == nil {
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

	s := newTestServer(mock)
	ac := &actionContext{data: map[string]any{"query": "roses"}}
	result, err := s.handleSearchCatalog(context.Background(), ac)
	if err != nil {
		t.Fatal(err)
	}
	if result["total"] != 1 {
		t.Errorf("expected total=1, got %v", result["total"])
	}
}

func TestHandleLookupProduct(t *testing.T) {
	mock := merchanttest.NewMock()
	mock.LookupFunc = func(id string, shipsTo ucp.Country) *catalog.Product {
		return &catalog.Product{ID: id, Title: "Roses"}
	}

	s := newTestServer(mock)
	ac := &actionContext{data: map[string]any{"id": "SKU-001"}}
	result, err := s.handleLookupProduct(context.Background(), ac)
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "SKU-001" {
		t.Errorf("expected id=SKU-001, got %v", result["id"])
	}
}
