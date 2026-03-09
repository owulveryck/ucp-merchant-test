package main

import (
	icatalog "github.com/owulveryck/ucp-merchant-test/internal/catalog"
)

// Product is an alias for the internal catalog.Product type.
type Product = icatalog.Product

// Global catalog instance used by the application.
var catalogInstance = icatalog.New()

func initCatalog(seed int64) {
	catalogInstance.Init(seed)
	catalog = catalogInstance.Products
}

func loadCatalogFromFile(path string) error {
	if err := catalogInstance.LoadFromFile(path); err != nil {
		return err
	}
	catalog = catalogInstance.Products
	return nil
}

func findProduct(id string) *Product {
	return catalogInstance.Find(id)
}

func filterCatalog(category, brand, query, usageType, country string) []Product {
	return catalogInstance.Filter(category, brand, query, usageType, country)
}

func categoryCount(products []Product) []map[string]interface{} {
	counts := map[string]int{}
	order := []string{}
	for _, p := range products {
		if _, seen := counts[p.Category]; !seen {
			order = append(order, p.Category)
		}
		counts[p.Category]++
	}
	result := make([]map[string]interface{}, 0, len(order))
	for _, name := range order {
		result = append(result, map[string]interface{}{
			"name":  name,
			"count": counts[name],
		})
	}
	return result
}

func containsCountry(countries []string, country string) bool {
	return icatalog.ContainsCountry(countries, country)
}

// Keep the global `catalog` slice for backward compatibility.
var catalog []Product
var productSeq int
