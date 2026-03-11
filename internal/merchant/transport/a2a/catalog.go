package a2a

import (
	"context"
	"fmt"
	"sort"

	icatalog "github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

func (s *Server) handleListProducts(_ context.Context, ac *actionContext) (map[string]any, error) {
	category, _ := ac.data["category"].(string)
	brand, _ := ac.data["brand"].(string)
	query, _ := ac.data["query"].(string)

	limit := 20
	if l, ok := ac.data["limit"].(float64); ok {
		limit = int(l)
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 50 {
		limit = 50
	}

	offset := 0
	if o, ok := ac.data["offset"].(float64); ok {
		offset = int(o)
	}
	if offset < 0 {
		offset = 0
	}

	filtered := s.merchant.Filter(ucp.Category(category), brand, query, ac.country, "", "")
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

	return map[string]any{
		"products": products,
		"pagination": map[string]any{
			"total":    total,
			"offset":   offset,
			"limit":    limit,
			"has_more": end < total,
		},
		"categories": categories,
	}, nil
}

func (s *Server) handleGetProductDetails(_ context.Context, ac *actionContext) (map[string]any, error) {
	id, _ := ac.data["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	p := s.merchant.Find(id)
	if p == nil {
		return nil, fmt.Errorf("product not found: %s", id)
	}

	result := map[string]any{
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
	return result, nil
}

func (s *Server) handleSearchCatalog(_ context.Context, ac *actionContext) (map[string]any, error) {
	query, _ := ac.data["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	limit := 10
	if l, ok := ac.data["limit"].(float64); ok {
		limit = int(l)
	}

	results := s.merchant.Search(icatalog.SearchParams{
		Query:   query,
		Limit:   limit,
		ShipsTo: ac.country,
	})

	return map[string]any{
		"results": results,
		"total":   len(results),
	}, nil
}

func (s *Server) handleLookupProduct(_ context.Context, ac *actionContext) (map[string]any, error) {
	id, _ := ac.data["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	p := s.merchant.Lookup(id, ac.country)
	if p == nil {
		return nil, fmt.Errorf("product not found: %s", id)
	}

	result := map[string]any{
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
	return result, nil
}
