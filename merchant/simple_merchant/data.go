package main

import (
	idata "github.com/owulveryck/ucp-merchant-test/internal/data"
)

// Type aliases for backward compatibility.
type CSVCustomer = idata.CSVCustomer
type CSVAddress = idata.CSVAddress
type CSVPaymentInstrument = idata.CSVPaymentInstrument
type CSVDiscount = idata.CSVDiscount
type CSVShippingRate = idata.CSVShippingRate
type CSVPromotion = idata.CSVPromotion
type ConformanceInput = idata.ConformanceInput
type FlowerShopData = idata.DataSource

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

func findCustomerByEmail(email string) *CSVCustomer {
	return shopData.FindCustomerByEmail(email)
}

func findAddressesByCustomerID(customerID string) []CSVAddress {
	return shopData.FindAddressesByCustomerID(customerID)
}

func findAddressesForEmail(email string) []CSVAddress {
	return shopData.FindAddressesForEmail(email)
}

func findDiscountByCode(code string) *CSVDiscount {
	return shopData.FindDiscountByCode(code)
}

func findPaymentInstrumentByID(id string) *CSVPaymentInstrument {
	return shopData.FindPaymentInstrumentByID(id)
}

func findPaymentInstrumentByToken(token string) *CSVPaymentInstrument {
	return shopData.FindPaymentInstrumentByToken(token)
}

func getShippingRatesForCountry(country string) []CSVShippingRate {
	return shopData.GetShippingRatesForCountry(country)
}

func matchExistingAddress(addrs []CSVAddress, street, locality, region, postal, country string) *CSVAddress {
	return idata.MatchExistingAddress(addrs, street, locality, region, postal, country)
}

func saveDynamicAddress(email string, addr CSVAddress) string {
	return shopData.SaveDynamicAddress(email, addr)
}
