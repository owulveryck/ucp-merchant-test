package main

import (
	idata "github.com/owulveryck/ucp-merchant-test/internal/data"
)

// CSVAddress alias used in store.go and tests.
type CSVAddress = idata.CSVAddress

var shopData = idata.New()

func loadFlowerShopData(dataDir string) error {
	if err := shopData.Load(dataDir); err != nil {
		return err
	}
	catalog = shopData.Products
	catalogInstance.Products = shopData.Products
	productSeq = len(catalog)
	return nil
}
