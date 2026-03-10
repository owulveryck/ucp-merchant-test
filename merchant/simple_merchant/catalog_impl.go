package main

import (
	"strings"

	icatalog "github.com/owulveryck/ucp-merchant-test/internal/catalog"
)

type catalogStore struct {
	Products   []icatalog.Product
	ProductSeq int
}

func newCatalogStore() *catalogStore {
	return &catalogStore{}
}

func (c *catalogStore) Find(id string) *icatalog.Product {
	for i := range c.Products {
		if c.Products[i].ID == id {
			return &c.Products[i]
		}
	}
	return nil
}

func (c *catalogStore) Filter(category, brand, query, usageType, country, currency, language string) []icatalog.Product {
	var result []icatalog.Product
	for _, p := range c.Products {
		if category != "" && !strings.EqualFold(p.Category, category) {
			continue
		}
		if brand != "" && !strings.EqualFold(p.Brand, brand) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(p.Title), strings.ToLower(query)) {
			continue
		}
		if usageType != "" && !strings.EqualFold(p.UsageType, usageType) {
			continue
		}
		if country != "" && len(p.AvailableCountries) > 0 {
			if !icatalog.ContainsCountry(p.AvailableCountries, country) {
				continue
			}
		}
		result = append(result, p)
	}
	return result
}

func (c *catalogStore) CategoryCount() []icatalog.CategoryStat {
	counts := map[string]int{}
	order := []string{}
	for _, p := range c.Products {
		if _, seen := counts[p.Category]; !seen {
			order = append(order, p.Category)
		}
		counts[p.Category]++
	}
	result := make([]icatalog.CategoryStat, 0, len(order))
	for _, name := range order {
		result = append(result, icatalog.CategoryStat{
			Name:  name,
			Count: counts[name],
		})
	}
	return result
}

func (c *catalogStore) Lookup(id string, shipsTo string) *icatalog.Product {
	p := c.Find(id)
	if p == nil {
		return nil
	}
	if shipsTo != "" && len(p.AvailableCountries) > 0 {
		if !icatalog.ContainsCountry(p.AvailableCountries, shipsTo) {
			return nil
		}
	}
	return p
}

func (c *catalogStore) Search(params icatalog.SearchParams) []icatalog.SearchResult {
	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 300 {
		limit = 300
	}

	query := strings.ToLower(params.Query)
	var results []icatalog.SearchResult
	for _, p := range c.Products {
		if query != "" {
			titleMatch := strings.Contains(strings.ToLower(p.Title), query)
			descMatch := strings.Contains(strings.ToLower(p.Description), query)
			catMatch := strings.Contains(strings.ToLower(p.Category), query)
			if !titleMatch && !descMatch && !catMatch {
				continue
			}
		}
		if params.MinPrice > 0 && p.Price < params.MinPrice {
			continue
		}
		if params.MaxPrice > 0 && p.Price > params.MaxPrice {
			continue
		}
		if params.AvailableForSale && p.Quantity <= 0 {
			continue
		}
		if params.ShipsTo != "" && len(p.AvailableCountries) > 0 {
			if !icatalog.ContainsCountry(p.AvailableCountries, params.ShipsTo) {
				continue
			}
		}
		results = append(results, icatalog.SearchResult{
			Product: p,
			InStock: p.Quantity > 0,
		})
		if len(results) >= limit {
			break
		}
	}
	return results
}
