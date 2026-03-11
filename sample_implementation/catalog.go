package main

import (
	"sync"

	icatalog "github.com/owulveryck/ucp-merchant-test/pkg/catalog"
)

// Product is an alias for the internal catalog.Product type.
type Product = icatalog.Product

// Global catalog instance used by the application.
var catalogInstance = newCatalogStore()

// Keep the global `catalog` slice for backward compatibility (dashboard).
var catalog []Product
var productSeq int

// catalogMu protects the global catalog slice used by the dashboard API.
var catalogMu sync.Mutex
