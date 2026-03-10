package main

import (
	icatalog "github.com/owulveryck/ucp-merchant-test/internal/catalog"
)

// Product is an alias for the internal catalog.Product type.
type Product = icatalog.Product

// Global catalog instance used by the application.
var catalogInstance = newCatalogStore()

// Keep the global `catalog` slice for backward compatibility.
var catalog []Product
var productSeq int
