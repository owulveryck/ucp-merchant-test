package mcp

import (
	"context"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"

	icatalog "github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

func (s *Server) handleListProducts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userCountry := userCountryFromContext(ctx)

	category, _ := args["category"].(string)
	brand, _ := args["brand"].(string)
	query, _ := args["query"].(string)

	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 50 {
		limit = 50
	}

	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}
	if offset < 0 {
		offset = 0
	}

	filtered := s.merchant.Filter(ucp.Category(category), brand, query, userCountry, "", "")

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Title < filtered[j].Title
	})

	categories := s.merchant.CategoryCount()

	total := len(filtered)

	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page := filtered[offset:end]

	type productInfo struct {
		ID                 string        `json:"id"`
		Title              string        `json:"title"`
		Category           ucp.Category  `json:"category"`
		Brand              string        `json:"brand"`
		Price              int           `json:"price"`
		Quantity           int           `json:"quantity"`
		ImageURL           string        `json:"image_url"`
		AvailableCountries []ucp.Country `json:"available_countries,omitempty"`
	}
	products := make([]productInfo, 0, len(page))
	for _, p := range page {
		products = append(products, productInfo{
			ID:                 p.ID,
			Title:              p.Title,
			Category:           p.Category,
			Brand:              p.Brand,
			Price:              p.Price,
			Quantity:           p.Quantity,
			ImageURL:           p.ImageURL,
			AvailableCountries: p.AvailableCountries,
		})
	}

	result := map[string]interface{}{
		"products": products,
		"pagination": map[string]interface{}{
			"total":    total,
			"offset":   offset,
			"limit":    limit,
			"has_more": end < total,
		},
		"categories": categories,
	}

	return toolResultFromJSON(result, extractImageURLs(result)), nil
}

func (s *Server) handleGetProductDetails(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return toolResultFromError(fmt.Errorf("id is required")), nil
	}

	p := s.merchant.Find(id)
	if p == nil {
		return toolResultFromError(fmt.Errorf("product not found: %s", id)), nil
	}
	result := map[string]interface{}{
		"id":          p.ID,
		"title":       p.Title,
		"category":    p.Category,
		"brand":       p.Brand,
		"price":       p.Price,
		"quantity":    p.Quantity,
		"image_url":   p.ImageURL,
		"description": p.Description,
	}
	if len(p.AvailableCountries) > 0 {
		result["available_countries"] = p.AvailableCountries
	}

	return toolResultFromJSON(result, extractImageURLs(result)), nil
}

func (s *Server) handleSearchCatalog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userCountry := userCountryFromContext(ctx)

	query, _ := args["query"].(string)
	if query == "" {
		return toolResultFromError(fmt.Errorf("query is required")), nil
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	shipsTo := userCountry

	results := s.merchant.Search(icatalog.SearchParams{
		Query:   query,
		Limit:   limit,
		ShipsTo: shipsTo,
	})

	result := map[string]interface{}{
		"results": results,
		"total":   len(results),
	}

	return toolResultFromJSON(result, extractImageURLs(result)), nil
}

func (s *Server) handleLookupProduct(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	userCountry := userCountryFromContext(ctx)

	id, _ := args["id"].(string)
	if id == "" {
		return toolResultFromError(fmt.Errorf("id is required")), nil
	}

	shipsTo := userCountry

	p := s.merchant.Lookup(id, shipsTo)
	if p == nil {
		return toolResultFromError(fmt.Errorf("product not found: %s", id)), nil
	}

	result := map[string]interface{}{
		"id":          p.ID,
		"title":       p.Title,
		"category":    p.Category,
		"brand":       p.Brand,
		"price":       p.Price,
		"quantity":    p.Quantity,
		"image_url":   p.ImageURL,
		"description": p.Description,
	}
	if len(p.AvailableCountries) > 0 {
		result["available_countries"] = p.AvailableCountries
	}

	return toolResultFromJSON(result, extractImageURLs(result)), nil
}
