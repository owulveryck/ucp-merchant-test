package main

import (
	"fmt"

	"github.com/owulveryck/ucp-merchant-test/pkg/jsondata"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/discount"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/fulfillment"
	"github.com/owulveryck/ucp-merchant-test/pkg/sample"
)

// Address alias used in store.go and tests.
type Address = fulfillment.Address

// shopDataSource is the interface satisfied by both sample.DataSource and jsondata.DataSource.
type shopDataSource interface {
	discount.DiscountLookup
	fulfillment.FulfillmentDataSource
	// ResetDynamicAddresses clears all dynamically saved addresses, restoring
	// the address book to its initial loaded state.
	ResetDynamicAddresses()
}

// shopDataLoader extends shopDataSource with data loading and product access.
type shopDataLoader interface {
	shopDataSource
	// Load loads product catalog and reference data from the given data directory.
	Load(dataDir string) error
	// GetProducts returns the loaded product catalog as a slice.
	GetProducts() []Product
}

// csvLoader wraps sample.DataSource to satisfy shopDataLoader.
type csvLoader struct{ *sample.DataSource }

func (l csvLoader) GetProducts() []Product { return l.Products }

// jsonLoader wraps jsondata.DataSource to satisfy shopDataLoader.
type jsonLoader struct{ *jsondata.DataSource }

func (l jsonLoader) GetProducts() []Product { return l.Products }

var shopData shopDataSource

func loadFlowerShopData(dataDir, dataFormat string) error {
	var loader shopDataLoader
	switch dataFormat {
	case "csv":
		loader = csvLoader{sample.New()}
	case "json":
		loader = jsonLoader{jsondata.New()}
	default:
		return fmt.Errorf("unknown data format: %s (use csv or json)", dataFormat)
	}

	if err := loader.Load(dataDir); err != nil {
		return err
	}

	products := loader.GetProducts()
	catalog = products
	catalogInstance.Products = products
	productSeq = len(catalog)
	shopData = loader
	return nil
}
