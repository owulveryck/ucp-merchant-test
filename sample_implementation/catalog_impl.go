package main

import (
	"strings"

	icatalog "github.com/owulveryck/ucp-merchant-test/internal/catalog"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
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

func (c *catalogStore) Filter(category ucp.Category, brand, query string, country ucp.Country, currency ucp.Currency, language ucp.Language) []icatalog.Product {
	var result []icatalog.Product
	for _, p := range c.Products {
		if category != "" && !p.Category.Matches(category) {
			continue
		}
		if brand != "" && !strings.EqualFold(p.Brand, brand) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(p.Title), strings.ToLower(query)) {
			continue
		}
		if country != "" && len(p.AvailableCountries) > 0 {
			if !ucp.ContainsCountry(p.AvailableCountries, country) {
				continue
			}
		}
		result = append(result, p)
	}
	return result
}

func (c *catalogStore) CategoryCount() []icatalog.CategoryStat {
	counts := map[ucp.Category]int{}
	order := []ucp.Category{}
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

func (c *catalogStore) Lookup(id string, shipsTo ucp.Country) *icatalog.Product {
	p := c.Find(id)
	if p == nil {
		return nil
	}
	if shipsTo != "" && len(p.AvailableCountries) > 0 {
		if !ucp.ContainsCountry(p.AvailableCountries, shipsTo) {
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
			catMatch := strings.Contains(strings.ToLower(string(p.Category)), query)
			if !titleMatch && !descMatch && !catMatch {
				continue
			}
		}
		if params.ShipsTo != "" && len(p.AvailableCountries) > 0 {
			if !ucp.ContainsCountry(p.AvailableCountries, params.ShipsTo) {
				continue
			}
		}
		results = append(results, icatalog.SearchResult{
			Product: p,
		})
		if len(results) >= limit {
			break
		}
	}
	return results
}
